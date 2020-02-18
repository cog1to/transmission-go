package windows

import (
  "fmt"
  "time"
  "../transmission"
  gc "../goncurses"
  "os"
  "os/signal"
  "syscall"
  "strings"
)

type torrents = *[]transmission.TorrentListItem
type settings = *transmission.SessionSettings
type input int

const (
  EXIT input = iota
  RESIZE = iota
  CURSOR_UP = iota
  CURSOR_DOWN = iota
  CURSOR_PAGEDOWN = iota
  CURSOR_PAGEUP = iota
  DETAILS = iota
  DELETE = iota
  DELETE_WITH_DATA = iota
  ADD = iota
  SELECT = iota
  CLEAR_SELECT = iota
  PAUSE = iota
  DOWN_LIMIT = iota
  UP_LIMIT = iota
  HELP = iota
)

/* Operations */

type ListOperation struct {
  Operation input
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
  List List
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

  // Handle user input.
  out := make(chan input)
  go func(control chan input) {
    for {
      c := <-observer
      switch c {
      case 'q':
        control <- EXIT
      case 'd':
        control <- DELETE
      case 'D':
        control <- DELETE_WITH_DATA
      case gc.KEY_RESIZE:
        control <- RESIZE
      case gc.KEY_UP, 'k':
        control <- CURSOR_UP
      case gc.KEY_DOWN, 'j':
        control <- CURSOR_DOWN
      case gc.KEY_PAGEDOWN:
        control <- CURSOR_PAGEDOWN
      case gc.KEY_PAGEUP:
        control <- CURSOR_PAGEUP
      case 'a':
        control <- ADD
      case ' ':
        control <- SELECT
      case 'c':
        control <- CLEAR_SELECT
      case 'l', gc.KEY_RIGHT, gc.KEY_RETURN:
        control <- DETAILS
      case 'p':
        control <- PAUSE
      case 'L':
        control <- DOWN_LIMIT
      case 'U':
        control <- UP_LIMIT
      case gc.KEY_F1:
        control <- HELP
      }
    }
  }(out)

  // Handle resizing.
  sigs := make(chan os.Signal, 1)
  signal.Notify(sigs, syscall.SIGWINCH)
  go func(control chan input) {
    for {
      sig := <-sigs
      if sig == syscall.SIGWINCH {
        control <- RESIZE
      }
    }
  }(out)

  // Handle list update.
  items, err := make(chan torrents), make(chan error)
  go func() {
    // First poll.
    updateList(client, items, err)

    for {
      <-time.After(time.Duration(3) * time.Second)
      updateList(client, items, err)
    }
  }()

  // Handle session update.
  session := make(chan settings)
  go func() {
    // First poll.
    updateSession(client, session, err)

    for {
      <-time.After(time.Duration(3) * time.Second)
      updateSession(client, session, err)
    }
  }()

  formatter := func(torrent interface{}, width int) string {
    item := torrent.(transmission.TorrentListItem)

    maxTitleLength := maxInt(0, width - 63)
    title := []rune(item.Name)

    var croppedTitle []rune
    croppedTitleLength := minInt(maxTitleLength, len(title))
    if (obfuscated) {
      croppedTitle = []rune(randomString(croppedTitleLength))
    } else {
      croppedTitle = title[0:croppedTitleLength]
    }
    spacesLength := maxTitleLength - croppedTitleLength

    format := fmt.Sprintf("%%5d %%s%%s %%-6s %%-9s %%-12s %%-6.3f %%-9s %%-9s")
    return fmt.Sprintf(format,
      item.Id(),
      string(croppedTitle),
      strings.Repeat(" ", spacesLength),
      fmt.Sprintf("%3.0f%%", (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0),
      formatSize(item.SizeWhenDone),
      formatStatus(item.Status),
      item.Ratio,
      formatSpeed(item.DownloadSpeed),
      formatSpeed(item.UploadSpeed))
  }

  // Initial window state.
  state := &ListWindowState{
    List: List{
      screen,
      formatter,
      HEADER_HEIGHT,
      FOOTER_HEIGHT,
      0,
      0,
      0,
      []int{},
      0,
      []Identifiable{}}}

  // Handle updates and control.
  func(control chan input, err chan error, list chan torrents) {
    for {
      select {
      case e := <-err:
        state.Error = e
        drawList(screen, *state)
      case s := <-session:
        state.Settings = s
      case items := <-list:
        if items != nil {
          state.List.SetItems(generalizeTorrents(*items))
        } else {
          state.List.SetItems([]Identifiable{})
        }
        drawList(screen, *state)
      case c := <-control:
        switch c {
        case EXIT:
          return
        case DELETE, DELETE_WITH_DATA:
          // Schedule selected items' deletion. To prevent accidental deletes, command needs to be confirmed.
          if op := state.PendingOperation; op != nil && op.Operation == c {
            state.List.Selection = []int{}
            go handleOperation(screen, client, *op, list, err)
            state.PendingOperation = nil
          } else {
            items := state.List.GetSelection()
            if len(items) > 0 {
              state.PendingOperation = &ListOperation{
                c,
                toTorrentList(items)}
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
        case RESIZE:
          // Resize window.
          gc.End()
          screen.Refresh()
          state.List.UpdateOffset()
        case ADD:
          // Open new torrent dialog.
          state.PendingOperation = nil
          errorDrawer := func(err error) {
            drawError(screen, err)
          }
          NewTorrentWindow(screen, reader, client, errorDrawer)
          go updateList(client, list, err)
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
            TorrentDetailsWindow(screen, reader, client, errorDrawer, torrent.Id(), obfuscated)
          }
        case PAUSE:
          // Pause/Start selected torrents.
          items := state.List.GetSelection()
          if len(items) > 0 {
            torrents := toTorrentList(items)
            _, isActive := idsAndNextState(torrents)
            op := ListActiveOperation{
                isActive,
                ListOperation{
                  c, torrents}}
            go handleOperation(screen, client, op, list, err)
          }
        case DOWN_LIMIT:
          // Set global download limit.
          intPrompt(screen, reader, "Global download limit (KB):",
            state.Settings.DownloadSpeedLimit, state.Settings.DownloadSpeedLimitEnabled,
            func(limit int) { go setGlobalDownloadLimit(client, limit, session, err) },
            func(err error) { drawError(screen, err) })
        case UP_LIMIT:
          // Set global upload limit.
          intPrompt(screen, reader, "Global upload limit (KB):",
            state.Settings.UploadSpeedLimit, state.Settings.UploadSpeedLimitEnabled,
            func(limit int) { go setGlobalUploadLimit(client, limit, session, err) },
            func(err error) { drawError(screen, err) })
        case HELP:
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
          CheatsheetWindow(screen, reader, items)
        }

        // Redraw.
        drawList(screen, *state)
      }
    }
  }(out, err, items)

}

func drawList(window *gc.Window, state ListWindowState) {
  row, col := window.MaxYX()

  maxTitleLength := maxInt(0, col - 63)

  // Legend.
  legendDown := "Down"
  if state.Settings != nil && state.Settings.DownloadSpeedLimitEnabled && state.Settings.DownloadSpeedLimit > 0 {
    legendDown = legendDown +  " *"
  }

  legendUp := "Up"
  if state.Settings != nil && state.Settings.UploadSpeedLimitEnabled && state.Settings.UploadSpeedLimit > 0 {
    legendUp = legendUp +  " *"
  }

  legendFormat := fmt.Sprintf("%%5s %%-%ds %%-6s %%-9s %%-12s %%-6s %%-9s %%-9s", maxTitleLength)
  window.MovePrintf(0, 0, legendFormat, "Id", "Name", "Done", "Size", "Status", "Ratio", legendDown, legendUp)
  window.HLine(1, 0, gc.ACS_HLINE, col)

  // List.
  state.List.Draw()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, gc.ACS_HLINE, col)
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
  if err != nil {
    window.MovePrintf(row - FOOTER_HEIGHT + 1, 0, "%s", err)
  } else {
    window.HLine(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  }

  window.Refresh()
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

