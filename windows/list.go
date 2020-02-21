package windows

import (
  "strings"
  "fmt"
  gc "../goncurses"
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
  window *gc.Window
  client *transmission.Client
  state *ListWindowState
  workers worker.WorkerList
  manager *WindowManager
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
  // We're fullscreen, so we don't need to resize.
}

func (window *ListWindow) OnInput(key gc.Key) {
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
    case PAUSE:
      // Pause/Start selected torrents.
      items := window.state.List.GetSelection()
      if len(items) > 0 {
        torrents := transform.ToTorrentList(items)
        _, isActive := idsAndNextState(torrents)
        op := ListActiveOperation{
            isActive,
            ListOperation{
              command, torrents}}
        handleOperation(window.client, op, window.state)
      }
    case HELP:
      showListCheatsheet(window.window, window.manager)
    }

    window.manager.Draw <- true
  }()
}

func NewListWindow(parent *gc.Window, client *transmission.Client, obfuscated bool, manager *WindowManager) *ListWindow {
  rows, cols := parent.MaxYX()
  window := parent.Sub(rows, cols, 0, 0)

  // Item formatter.
  formatter := func(torrent interface{}, width int) string {
    return formatTorrentListItem(torrent, width, obfuscated)
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
    manager}
}

/* Drawing */

func formatTorrentListItem(torrent interface{}, width int, obfuscated bool) string {
  item := torrent.(transmission.TorrentListItem)

  maxTitleLength := utils.MaxInt(0, width - 71)
  title := []rune(item.Name)

  var croppedTitle []rune
  croppedTitleLength := utils.MinInt(maxTitleLength, len(title))
  if (obfuscated) {
    croppedTitle = []rune(utils.RandomString(croppedTitleLength))
  } else {
    croppedTitle = title[0:croppedTitleLength]
  }
  spacesLength := maxTitleLength - croppedTitleLength

  // Format: ID - Title - %Done - ETA - Full size - Status - Ratio - Down speed - Up speed
  format := fmt.Sprintf("%%5d %%s%%s %%-6s %%-7s %%-9s %%-12s %%-6.3f %%-9s %%-9s")

  // %Done. Handle unknown state.
  var done string
  if item.SizeWhenDone == 0 {
    done = fmt.Sprintf("%3.0f%%", 0)
  } else {
    done = fmt.Sprintf("%3.0f%%",
      (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0)
  }

  return fmt.Sprintf(format,
    item.Id(),
    string(croppedTitle),
    strings.Repeat(" ", spacesLength),
    done,
    formatTime(item.Eta, (item.SizeWhenDone > 0 && item.LeftUntilDone == 0)),
    formatSize(item.SizeWhenDone),
    formatStatus(item.Status),
    utils.MaxFloat32(0, item.Ratio),
    formatSpeed(item.DownloadSpeed),
    formatSpeed(item.UploadSpeed))
}

func drawList(window *gc.Window, state ListWindowState) {
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
  window.HLine(1, 0, gc.ACS_HLINE, col)

  // List.
  state.List.Draw()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, gc.ACS_HLINE, col)
  window.HLine(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  if op := state.PendingOperation; op != nil {
    var idsString string
    if len(op.Items) == 1 {
      idsString = fmt.Sprintf("torrent %d", op.Items[0].Id())
    } else {
      idsString = fmt.Sprintf("torrents %s", strings.Join(mapToString(op.Items), ", "))
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

/* Network */

func updateList(client *transmission.Client, state *ListWindowState) {
  list, err := client.List()

  if list != nil {
    state.List.Items = transform.GeneralizeTorrents(*list)
  }

  state.Error = err
}

func updateSession(client *transmission.Client, state *ListWindowState) {
  settings, err := client.GetSessionSettings()

  if settings != nil {
    state.Settings = settings
  }

  state.Error = err
}

func handleOperation(client *transmission.Client, operation interface{}, state *ListWindowState) {
  var e error

  switch operation.(type) {
  case ListActiveOperation:
    lop := operation.(ListActiveOperation)
    ids := mapToIds(lop.Items)
    e = client.UpdateActive(ids, lop.Active)
  case ListOperation:
    lop := operation.(ListOperation)
    ids := mapToIds(lop.Items)
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

func showListCheatsheet(parent *gc.Window, manager *WindowManager) {
  items := []HelpItem{
    HelpItem{ "q", "Exit" },
    HelpItem{ "jk↑↓", "Move cursor up and down" },
    HelpItem{ "l→", "Go to torrent details" },
    HelpItem{ "Space", "Toggle selection" },
    HelpItem{ "c", "Clear selection" },
    HelpItem{ "d", "Remove torrent(s) from the list (keep data)" },
    HelpItem{ "D", "Delete torrent(s) along with the data" },
    HelpItem{ "p", "Start/stop selected torrent(s)" },
    HelpItem{ "L", "Set global download speed limit" },
    HelpItem{ "U", "Set global upload speed limit" }}

  cheatsheet := NewCheatsheet(parent, items, manager)
  manager.AddWindow(cheatsheet)
}

/* Utils */

func control(char gc.Key) Input {
  switch char {
  case 'q':
    return EXIT
  case 'd':
    return DELETE
  case 'D':
    return DELETE_WITH_DATA
  case gc.KEY_RESIZE:
    return RESIZE
  case gc.KEY_UP, 'k':
    return CURSOR_UP
  case gc.KEY_DOWN, 'j':
    return CURSOR_DOWN
  case gc.KEY_PAGEDOWN:
    return CURSOR_PAGEDOWN
  case gc.KEY_PAGEUP:
    return CURSOR_PAGEUP
  case 'a':
    return ADD
  case ' ':
    return SELECT
  case 'c':
    return CLEAR_SELECT
  case 'l', gc.KEY_RIGHT, gc.KEY_RETURN:
    return DETAILS
  case 'p':
    return PAUSE
  case 'L':
    return DOWN_LIMIT
  case 'U':
    return UP_LIMIT
  case gc.KEY_F1:
    return HELP
  }

  return UNKNOWN
}

func mapToString(slice []transmission.TorrentListItem) []string {
  output := make([]string, len(slice))
  for index, element := range slice {
    output[index] = fmt.Sprintf("%d", element.Id())
  }
  return output
}

func mapToIds(slice []transmission.TorrentListItem) []int {
  output := make([]int, len(slice))
  for index, item := range slice {
    output[index] = item.Id()
  }
  return output
}

func idsAndNextState(torrents []transmission.TorrentListItem) ([]int, bool) {
  isActive := false
  ids := make([]int, len(torrents))
  for i, torrent := range torrents {
    isActive = isActive || torrent.Status != transmission.TR_STATUS_STOPPED
    ids[i] = torrent.Id()
  }

  return ids, !isActive
}
