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
)

type ListOperation struct {
  Operation input
  Items []transmission.TorrentListItem
}

type AddOperation struct {
  Url string
  Path string
}

type ListWindowState struct {
  Rows, Cols int
  PendingOperation *ListOperation
  Error error
  List List
}

const (
  INFO_HEIGHT = 4
  HEADER_HEIGHT = 2
  FOOTER_HEIGHT = 2
)

func NewListWindow(screen *gc.Window, client *transmission.Client) {
  reader := NewInputReader(screen)
  observer := make(chan gc.Key)
  reader.AddObserver(observer)

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

  formatter := func(torrent interface{}, width int) string {
    item := torrent.(transmission.TorrentListItem)

    maxTitleLength := maxInt(0, width - 63)
    title := []rune(item.Name)

    croppedTitleLength := minInt(maxTitleLength, len(title))
    croppedTitle := title[0:croppedTitleLength]
    spacesLength := maxTitleLength - croppedTitleLength

    format := fmt.Sprintf("%%5d %%s%%s %%-6s %%-9s %%-12s %%-6.3f %%-9s %%-9s")
    return fmt.Sprintf(format,
      item.Id,
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
      []interface{}{}}}

  // Handle updates and control.
  func(control chan input, err chan error, list chan torrents) {
    for {
      select {
      case e := <-err:
        state.Error = e
        drawList(screen, *state)
      case items := <-list:
        if items != nil {
          state.List.SetItems(generalizeTorrents(*items))
        } else {
          state.List.SetItems([]interface{}{})
        }
        drawList(screen, *state)
      case c := <-control:
        switch c {
        case EXIT:
          return
        case DELETE, DELETE_WITH_DATA:
          if state.List.Cursor == -1 {
            break
          }

          if op := state.PendingOperation; op != nil {
            if len(state.List.Selection) == 0 {
              go handleOperation(screen, client, *op, list, err)
              state.PendingOperation = nil
            } else {
              state.List.Selection = []int{}
              go handleOperation(screen, client, *op, list, err)
              state.PendingOperation = nil
            }
          } else {
            if len(state.List.Selection) == 0 {
              state.PendingOperation = &ListOperation{ c, toTorrentList(state.List.Items[state.List.Cursor:state.List.Cursor+1]) }
            } else {
              items := make([]transmission.TorrentListItem, len(state.List.Selection))
              for i, index := range state.List.Selection {
                items[i] = state.List.Items[index].(transmission.TorrentListItem)
              }
              state.PendingOperation = &ListOperation{ c, items }
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
          gc.End()
          screen.Refresh()
          state.List.UpdateOffset()
        case ADD:
          state.PendingOperation = nil
          errorDrawer := func(err error) {
            drawError(screen, err)
          }
          NewTorrentWindow(screen, reader, client, errorDrawer)
          go updateList(client, list, err)
        case SELECT:
          state.List.Select()
          state.PendingOperation = nil
        case CLEAR_SELECT:
          state.List.ClearSelection()
          state.PendingOperation = nil
        case DETAILS:
          if state.List.Cursor > -1 {
            item := state.List.Items[state.List.Cursor]
            torrent := item.(transmission.TorrentListItem)
            errorDrawer := func(err error) {
              drawError(screen, err)
            }
            TorrentDetailsWindow(screen, reader, client, errorDrawer, torrent.Id)
          }
        }

        // Redraw.
        drawList(screen, *state)
      }
    }
  }(out, err, items)

  reader.RemoveObserver(observer)
}

func drawList(window *gc.Window, state ListWindowState) {
  row, col := window.MaxYX()

  maxTitleLength := maxInt(0, col - 63)

  // Legend.
  legendFormat := fmt.Sprintf("%%5s %%-%ds %%-6s %%-9s %%-12s %%-6s %%-9s %%-9s", maxTitleLength)
  window.MovePrintf(0, 0, legendFormat, "Id", "Name", "Done", "Size", "Status", "Ratio", "Down", "Up")
  window.HLine(1, 0, gc.ACS_HLINE, col)

  state.List.Draw()

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, gc.ACS_HLINE, col)
  if op := state.PendingOperation; op != nil {
    var idsString string
    if len(op.Items) == 1 {
      idsString = fmt.Sprintf("torrent %d", op.Items[0].Id)
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

func handleOperation(screen *gc.Window, client *transmission.Client, operation interface{}, items chan torrents, err chan error) {
  var e error

  switch operation.(type) {
  case ListOperation:
    lop := operation.(ListOperation)
    ids := func() []int64 {
      output := make([]int64, len(lop.Items))
      for index, item := range lop.Items {
        output[index] = item.Id
      }
      return output
    }()
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

func mapToString(slice []transmission.TorrentListItem) []string {
  output := make([]string, len(slice))
  for index, element := range slice {
    output[index] = fmt.Sprintf("%d", element.Id)
  }
  return output
}
