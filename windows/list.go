package windows

import (
  "strings"
  "fmt"
  tui "../tui"
  "../transmission"
  "../transform"
  "../list"
  "../worker"
  "../utils"
)

type Input int
type Settings *transmission.SessionSettings

const (
  INFO_HEIGHT = 4
  HEADER_HEIGHT = 2
  FOOTER_HEIGHT = 2
)

const (
  EXIT Input = iota
  RESIZE
  CURSOR_UP
  CURSOR_DOWN
  CURSOR_PAGEDOWN
  CURSOR_PAGEUP
  DETAILS
  DELETE
  DELETE_WITH_DATA
  ADD
  SELECT
  CLEAR_SELECT
  PAUSE
  DOWN_LIMIT
  UP_LIMIT
  HELP
  MOVE
  SELECT_ALL
  INVERT_SELECT
  UNKNOWN
)

/* Operations */

type ListOperation struct {
  Operation Input
  Items []transmission.TorrentListItem
}

type ListActiveOperation struct {
  Active bool
  ListOperation
}

type ListWindowState struct {
  PendingOperation *ListOperation
  Error error
  List list.List
  Settings Settings
}


type ListWindow struct {
  window *tui.Window
  client *transmission.Client
  state *ListWindowState
  workers worker.WorkerList
  manager *WindowManager
  obfuscated bool
}

func (window *ListWindow) IsFullScreen() bool {
  return true
}

func (window *ListWindow) SetActive(active bool) {
  if active {
    window.workers.Start()
  } else {
    window.workers.Stop()
  }
}

func (window *ListWindow) Draw() {
  drawList(window.window, *window.state)
}

func (window *ListWindow) Resize() {
  window.window.Refresh()
}

func (window *ListWindow) OnInput(key tui.Key) {
  go func() {
    command := control(key)
    switch command {
    case EXIT:
      window.manager.Exit <- true
      return
    case DELETE, DELETE_WITH_DATA:
      // Schedule selected items' deletion. To prevent accidental deletes, command needs to be confirmed.
      if op := window.state.PendingOperation; op != nil && op.Operation == command {
        window.state.List.Selection = []int{}
        handleOperation(window.client, *op, window.state)
        window.state.PendingOperation = nil
      } else {
        items := window.state.List.GetSelection()
        if len(items) > 0 {
          window.state.PendingOperation = &ListOperation{
            command,
            transform.ToTorrentList(items)}
        }
      }
    case CURSOR_UP:
      window.state.List.MoveCursor(-1)
      window.state.PendingOperation = nil
    case CURSOR_DOWN:
      window.state.List.MoveCursor(1)
      window.state.PendingOperation = nil
    case CURSOR_PAGEUP:
      window.state.List.Page(-1)
      window.state.PendingOperation = nil
    case CURSOR_PAGEDOWN:
      window.state.List.Page(1)
      window.state.PendingOperation = nil
    case SELECT:
      // Toggle selection for item under cursor.
      window.state.List.Select()
      window.state.PendingOperation = nil
    case CLEAR_SELECT:
      // Clear selection.
      window.state.List.ClearSelection()
      window.state.PendingOperation = nil
    case INVERT_SELECT:
      // Invert selection.
      window.state.List.InvertSelection()
      window.state.PendingOperation = nil
    case SELECT_ALL:
      // Select all items.
      window.state.List.SelectAll()
      window.state.PendingOperation = nil
    case PAUSE:
      // Pause/Start selected torrents.
      items := window.state.List.GetSelection()
      if len(items) > 0 {
        torrents := transform.ToTorrentList(items)
        _, isActive := transform.IdsAndNextState(torrents)
        op := ListActiveOperation{
            isActive,
            ListOperation{
              command, torrents}}
        handleOperation(window.client, op, window.state)
      }
    case HELP:
      showListCheatsheet(window.window, window.manager)
    case MOVE:
      torrents := transform.ToTorrentList(window.state.List.GetSelection())
      ids, idsString := transform.MapToIds(torrents), idsString(torrents)

      PathPrompt(
        window.window,
        window.manager,
        fmt.Sprintf("New location for %s:", idsString),
        "",
        func(location string) {
          go func() {
            setListLocation(window.client, ids, location, window.state)
            window.manager.Draw <- true
          }()
        },
        func(err error) {
          window.state.Error = err
          window.manager.Draw <- true
        })
    case UP_LIMIT:
      IntPrompt(
        window.window,
        window.manager,
        "Global upload limit (KB):",
        window.state.Settings.UploadSpeedLimit,
        window.state.Settings.UploadSpeedLimitEnabled,
        func(limit int) { go setGlobalUploadLimit(window.client, limit, window.state) },
        func(err error) { window.state.Error = err })
    case DOWN_LIMIT:
      IntPrompt(
        window.window,
        window.manager,
        "Global download limit (KB):",
        window.state.Settings.DownloadSpeedLimit,
        window.state.Settings.DownloadSpeedLimitEnabled,
        func(limit int) { go setGlobalDownloadLimit(window.client, limit, window.state) },
        func(err error) { window.state.Error = err })
    case ADD:
      // Open new torrent dialog.
      window.state.PendingOperation = nil
      dialog := NewAddTorrentWindow(window.client, window.window, window.manager, func(err error) { drawError(window.window, err) })
      window.manager.AddWindow(dialog)
    case DETAILS:
      // Go to torrent details.
      if window.state.List.Cursor >= 0 {
        item := window.state.List.Items[window.state.List.Cursor]
        torrent := item.(transmission.TorrentListItem)
        details := NewTorrentDetailsWindow(window.client, torrent.Id(), window.obfuscated, window.window, window.manager)
        window.manager.AddWindow(details)
      }
    }

    go func() {
      window.manager.Draw <- true
    }()
  }()
}

func NewListWindow(parent *tui.Window, client *transmission.Client, obfuscated bool, manager *WindowManager) *ListWindow {
  rows, cols := parent.MaxYX()
  window := parent.Sub(0, 0, rows, cols)

  // Item formatter.
  formatter := func(torrent interface{}, width int, printer func(int, string)) {
    formatTorrentListItem(torrent, width, obfuscated, printer)
  }

  // State.
  state := &ListWindowState{
    List: list.List{
      window,
      formatter,
      HEADER_HEIGHT,
      FOOTER_HEIGHT,
      0,
      0,
      0,
      []int{},
      0,
      []list.Identifiable{}}}

  // Handle list update.
  listWorker := worker.Repeating(3, func() {
    updateList(client, state)
    manager.Draw <- true
  })

  // Handle session update.
  sessionWorker := worker.Repeating(3, func() {
    updateSession(client, state)
    manager.Draw <- true
  })

  // List of workers.
  workers := worker.WorkerList{ listWorker, sessionWorker }

  return &ListWindow{
    window,
    client,
    state,
    workers,
    manager,
    obfuscated}
}

/* Drawing */

func formatTorrentListItem(torrent interface{}, width int, obfuscated bool, printer func(int, string)) {
  item := torrent.(transmission.TorrentListItem)

  maxTitleLength := utils.MaxInt(0, width - 72)
  title := []rune(item.Name)

  var croppedTitle []rune
  croppedTitleLength := len(title)
  if (obfuscated) {
    croppedTitle = []rune(utils.RandomString(croppedTitleLength))
  } else {
    croppedTitle = title
  }
  spacesLength := utils.MaxInt(0, width - croppedTitleLength)

  // Format: ID - Title - %Done - ETA - Full size - Status - Ratio - Down speed - Up speed
  idAndNameFormat := "%5d %s%s"
  detailsFormat := " %-6s %-7s %-9s %-12s %-6.3f %-9s %-9s"

  // %Done. Handle unknown state.
  var done string
  if item.SizeWhenDone == 0 {
    done = "  0%"
  } else {
    done = fmt.Sprintf("%3.0f%%",
      (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0)
  }

  idAndName := fmt.Sprintf(idAndNameFormat, item.Id(), string(croppedTitle), strings.Repeat(" ", spacesLength))
  printer(0, idAndName)

  details := fmt.Sprintf(detailsFormat,
    done,
    formatTime(item.Eta, (item.SizeWhenDone > 0 && item.LeftUntilDone == 0)),
    formatSize(item.SizeWhenDone),
    formatStatus(item.Status),
    utils.MaxFloat32(0, item.Ratio),
    formatSpeed(item.DownloadSpeed),
    formatSpeed(item.UploadSpeed))
  printer(maxTitleLength + 7, details)
}

func drawList(window *tui.Window, state ListWindowState) {
  window.Erase()
  row, col := window.MaxYX()

  maxTitleLength := utils.MaxInt(0, col - 71)

  // Legend.
  legendDown := "Down"
  if state.Settings != nil && state.Settings.DownloadSpeedLimitEnabled && state.Settings.DownloadSpeedLimit > 0 {
    legendDown = legendDown +  " *"
  }

  legendUp := "Up"
  if state.Settings != nil && state.Settings.UploadSpeedLimitEnabled && state.Settings.UploadSpeedLimit > 0 {
    legendUp = legendUp +  " *"
  }

  legendFormat := fmt.Sprintf("%%5s %%-%ds %%-6s %%-7s %%-9s %%-12s %%-6s %%-9s %%-9s", maxTitleLength)
  window.MovePrintf(0, 0, legendFormat, "Id", "Name", "Done", "ETA", "Size", "Status", "Ratio", legendDown, legendUp)
  window.HLine(1, 0, col)

  // List.
  state.List.Draw()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, col)
  window.Line(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  if op := state.PendingOperation; op != nil {
    var idsString string
    if len(op.Items) == 1 {
      idsString = fmt.Sprintf("torrent %d", op.Items[0].Id())
    } else {
      idsString = fmt.Sprintf("torrents %s", strings.Join(transform.MapToString(op.Items), ", "))
    }

    switch op.Operation {
    case DELETE:
      window.MovePrintf(row - FOOTER_HEIGHT + 1, 0,
        "Removing %s from the list. Press 'd' again to confirm.", idsString)
    case DELETE_WITH_DATA:
      window.MovePrintf(row - FOOTER_HEIGHT + 1, 0,
        "Deleting %s along with data. Press 'D' again to confirm.", idsString)
    }
  } else if state.Error != nil {
    window.MovePrintf(row - FOOTER_HEIGHT + 1, 0, "%s", state.Error)
  }

  window.Refresh()
}

func drawError(window *tui.Window, err error) {
  row, col := window.MaxYX()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, col)
  window.Line(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  if err != nil {
    window.MovePrintf(row - FOOTER_HEIGHT + 1, 0, "%s", err)
  }
}

/* Network */

func updateList(client *transmission.Client, state *ListWindowState) {
  list, err := client.List()

  if list != nil {
    state.List.Items = transform.GeneralizeTorrents(*list)
  }

  if err != nil {
    state.Error = err
  }
}

func updateSession(client *transmission.Client, state *ListWindowState) {
  settings, err := client.GetSessionSettings()

  if settings != nil {
    state.Settings = settings
  }

  if err != nil {
    state.Error = err
  }
}

func setGlobalDownloadLimit(client *transmission.Client, limit int, state *ListWindowState) {
  e := client.SetGlobalDownloadLimit(limit)

  if e != nil {
    state.Error = e
  } else {
    updateSession(client, state)
  }
}

func setGlobalUploadLimit(client *transmission.Client, limit int, state *ListWindowState) {
  e := client.SetGlobalUploadLimit(limit)

  if e != nil {
    state.Error = e
  } else {
    updateSession(client, state)
  }
}

func setListLocation(client *transmission.Client, ids []int, location string, state *ListWindowState) {
  if len(ids) == 0 {
    return
  }

  e := client.SetLocation(ids, utils.ExpandHome(location))

  if e != nil {
    state.Error = e
  } else {
    updateList(client, state)
  }
}

func handleOperation(client *transmission.Client, operation interface{}, state *ListWindowState) {
  var e error

  switch operation.(type) {
  case ListActiveOperation:
    lop := operation.(ListActiveOperation)
    ids := transform.MapToIds(lop.Items)
    e = client.UpdateActive(ids, lop.Active)
  case ListOperation:
    lop := operation.(ListOperation)
    ids := transform.MapToIds(lop.Items)
    switch lop.Operation {
    case DELETE:
      e = client.Delete(ids, false)
    case DELETE_WITH_DATA:
      e = client.Delete(ids, true)
    default:
      e = fmt.Errorf("Unknown list operation type")
    }
  default:
    e = fmt.Errorf("Unknown operation type")
  }

  if e != nil {
    state.Error = e
  } else {
    updateList(client, state)
  }
}


/* Navigation */

func showListCheatsheet(parent *tui.Window, manager *WindowManager) {
  items := []HelpItem{
    HelpItem{ "q", "Exit" },
    HelpItem{ "jk↑↓", "Move cursor up and down" },
    HelpItem{ "l→", "Go to torrent details" },
    HelpItem{ "a", "Add new torrent" },
    HelpItem{ "Space", "Toggle selection" },
    HelpItem{ "c", "Clear selection" },
    HelpItem{ "A", "Select all items" },
    HelpItem{ "i", "Invert selection" },
    HelpItem{ "d", "Remove torrent(s) from the list (keep data)" },
    HelpItem{ "D", "Delete torrent(s) along with the data" },
    HelpItem{ "p", "Start/stop selected torrent(s)" },
    HelpItem{ "L", "Set global download speed limit" },
    HelpItem{ "U", "Set global upload speed limit" },
    HelpItem{ "m", "Move selected torrent(s) to a new location" }}

  cheatsheet := NewCheatsheet(parent, items, manager)
  manager.AddWindow(cheatsheet)
}

/* Utils */

func control(char tui.Key) Input {
  if char.Rune != nil {
    switch *char.Rune {
    case 'q':
      return EXIT
    case 'd':
      return DELETE
    case 'D':
      return DELETE_WITH_DATA
    case 'a':
      return ADD
    case ' ':
      return SELECT
    case 'c':
      return CLEAR_SELECT
    case 'l':
      return DETAILS
    case 'p':
      return PAUSE
    case 'L':
      return DOWN_LIMIT
    case 'U':
      return UP_LIMIT
    case 'm':
      return MOVE
    case 'k':
      return CURSOR_UP
    case 'j':
      return CURSOR_DOWN
    case 'A':
      return SELECT_ALL
    case 'i':
      return INVERT_SELECT
    }
  } else if char.EscapeSeq != nil {
    switch *char.EscapeSeq {
    case tui.ESC_UP:
      return CURSOR_UP
    case tui.ESC_DOWN:
      return CURSOR_DOWN
    case tui.ESC_PGDOWN:
      return CURSOR_PAGEDOWN
    case tui.ESC_PGUP:
      return CURSOR_PAGEUP
    case tui.ESC_F1:
      return HELP
    case tui.ESC_RIGHT:
      return DETAILS
    }
  } else if char.ControlCode != 0 {
    switch char.ControlCode {
    case tui.ASC_ENTER:
      return DETAILS
    }
  }

  return UNKNOWN
}
