package windows

import (
  "fmt"
  "os"
  "os/signal"
  "syscall"
  "strings"
  gc "../goncurses"
  "../transmission"
  ls "../list"
  "../utils"
  "../transform"
  "../worker"
)

type torrents = *[]transmission.TorrentListItem
type settings = *transmission.SessionSettings
type Input int

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
  Rows, Cols int
  PendingOperation *ListOperation
  Error error
  List ls.List
  Settings settings
}

/* Render loop */

const (
  INFO_HEIGHT = 4
  HEADER_HEIGHT = 2
  FOOTER_HEIGHT = 2
)

func NewListWindow(screen *gc.Window, client *transmission.Client, obfuscated bool) {
  reader := NewInputReader(screen)
  observer := make(chan gc.Key)
  reader.AddObserver(observer)
  defer reader.RemoveObserver(observer)

  // Handle resizing.
  sigs := make(chan os.Signal)
  signal.Notify(sigs, syscall.SIGWINCH)

  // Handle list update.
  items, err := make(chan torrents), make(chan error)
  listWorker := worker.Repeating(3, func() {
    updateList(client, items, err)
  })

  // Handle session update.
  session := make(chan settings)
  sessionWorker := worker.Repeating(3, func() {
    updateSession(client, session, err)
  })

  // List of workers.
  workers := worker.WorkerList{ listWorker, sessionWorker }

  // Item formatter.
  formatter := func(torrent interface{}, width int) string {
    return formatTorrentListItem(torrent, width, obfuscated)
  }

  // Initial window state.
  state := &ListWindowState{
    List: ls.List{
      screen,
      formatter,
      HEADER_HEIGHT,
      FOOTER_HEIGHT,
      0,
      0,
      0,
      []int{},
      0,
      []ls.Identifiable{}}}

  // Start polling.
  workers.Start()
  defer workers.Stop()

  // Main loop.
  for {
    select {
    case sig := <-sigs:
      if sig == syscall.SIGWINCH {
        // Resize window.
        gc.End()
        screen.Refresh()
        state.List.UpdateOffset()
        drawList(screen, *state)
      }
    case e := <-err:
      state.Error = e
      drawList(screen, *state)
    case s := <-session:
      state.Settings = s
    case list := <-items:
      if list != nil {
        state.List.SetItems(transform.GeneralizeTorrents(*list))
      } else {
        state.List.SetItems([]ls.Identifiable{})
      }
      drawList(screen, *state)
    case inp := <-observer:
      c := control(inp)
      switch c {
      case EXIT:
        return
      case DELETE, DELETE_WITH_DATA:
        // Schedule selected items' deletion. To prevent accidental deletes, command needs to be confirmed.
        if op := state.PendingOperation; op != nil && op.Operation == c {
          state.List.Selection = []int{}
          go handleOperation(screen, client, *op, items, err)
          state.PendingOperation = nil
        } else {
          items := state.List.GetSelection()
          if len(items) > 0 {
            state.PendingOperation = &ListOperation{
              c,
              transform.ToTorrentList(items)}
          }
        }
      case CURSOR_UP:
        state.List.MoveCursor(-1)
        state.PendingOperation = nil
      case CURSOR_DOWN:
        state.List.MoveCursor(1)
        state.PendingOperation = nil
      case CURSOR_PAGEUP:
        state.List.Page(-1)
        state.PendingOperation = nil
      case CURSOR_PAGEDOWN:
        state.List.Page(1)
        state.PendingOperation = nil
      case ADD:
        // Open new torrent dialog.
        state.PendingOperation = nil
        errorDrawer := func(err error) {
          drawError(screen, err)
        }
        NewTorrentWindow(screen, reader, client, errorDrawer)
        go updateList(client, items, err)
      case SELECT:
        // Toggle selection for item under cursor.
        state.List.Select()
        state.PendingOperation = nil
      case CLEAR_SELECT:
        // Clear selection.
        state.List.ClearSelection()
        state.PendingOperation = nil
      case DETAILS:
        // Go to torrent details.
        if state.List.Cursor >= 0 {
          item := state.List.Items[state.List.Cursor]
          torrent := item.(transmission.TorrentListItem)
          errorDrawer := func(err error) {
            drawError(screen, err)
          }
          worker.WithSuspended(workers, func() {
            TorrentDetailsWindow(screen, reader, client, errorDrawer, torrent.Id(), obfuscated)
          })
        }
      case PAUSE:
        // Pause/Start selected torrents.
        list := state.List.GetSelection()
        if len(list) > 0 {
          torrents := transform.ToTorrentList(list)
          _, isActive := idsAndNextState(torrents)
          op := ListActiveOperation{
              isActive,
              ListOperation{
                c, torrents}}
          go handleOperation(screen, client, op, items, err)
        }
      case DOWN_LIMIT:
        worker.WithSuspended(workers, func() {
          // Set global download limit.
          intPrompt(screen, reader, "Global download limit (KB):",
            state.Settings.DownloadSpeedLimit, state.Settings.DownloadSpeedLimitEnabled,
            func(limit int) { go setGlobalDownloadLimit(client, limit, session, err) },
            func(err error) { drawError(screen, err) })
        })
      case UP_LIMIT:
        worker.WithSuspended(workers, func() {
          // Set global upload limit.
          intPrompt(screen, reader, "Global upload limit (KB):",
            state.Settings.UploadSpeedLimit, state.Settings.UploadSpeedLimitEnabled,
            func(limit int) { go setGlobalUploadLimit(client, limit, session, err) },
            func(err error) { drawError(screen, err) })
        })
      case HELP:
        worker.WithSuspended(workers, func() {
          showMainCheatSheet(screen, reader)
        })
      }

      // Redraw.
      drawList(screen, *state)
    }
  }
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
  } else {
    window.HLine(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  }

  window.Refresh()
}

func drawError(window *gc.Window, err error) {
  row, col := window.MaxYX()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, gc.ACS_HLINE, col)
  window.HLine(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  if err != nil {
    window.MovePrintf(row - FOOTER_HEIGHT + 1, 0, "%s", err)
  }

  window.Refresh()
}

func showMainCheatSheet(parent *gc.Window, reader *InputReader) {
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

  CheatsheetWindow(parent, reader, items)
}

/* Network */

func handleOperation(screen *gc.Window, client *transmission.Client, operation interface{}, items chan torrents, err chan error) {
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
    err <- e
  } else {
    updateList(client, items, err)
  }
}

func updateList(client *transmission.Client, items chan torrents, err chan error) {
  list, e := client.List()
  err <- e
  items <- list
}

func setGlobalDownloadLimit(client *transmission.Client, limit int, session chan settings, err chan error) {
  e := client.SetGlobalDownloadLimit(limit)

  if e != nil {
    err <- e
  } else {
    updateSession(client, session, err)
  }
}

func setGlobalUploadLimit(client *transmission.Client, limit int, session chan settings, err chan error) {
  e := client.SetGlobalUploadLimit(limit)

  if e != nil {
    err <- e
  } else {
    updateSession(client, session, err)
  }
}

func updateSession(client *transmission.Client, session chan settings, err chan error) {
  s, e := client.GetSessionSettings()
  err <- e
  session <- s
}

/* Utils */

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
