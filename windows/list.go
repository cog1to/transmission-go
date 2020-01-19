package windows

import (
  "fmt"
  "time"
  "../transmission"
  gc "github.com/rthornton128/goncurses"
  "os"
  "os/signal"
  "syscall"
)

type torrents = *[]transmission.TorrentListItem
type input int

const (
  EXIT input = iota
  RESIZE = iota
  CURSOR_UP = iota
  CURSOR_DOWN = iota
  DETAILS = iota
  DELETE = iota
  DELETE_WITH_DATA = iota
  ADD = iota
)

type ListOperation struct {
  Operation input
  Item *transmission.TorrentListItem
}

type AddOperation struct {
  Url string
  Path string
}

type ListWindowState struct {
  Items torrents
  Offset int
  Cursor int
  Selection []int
  Rows, Cols int
  PendingOperation *ListOperation
  Error error
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
      case gc.KEY_UP:
        control <- CURSOR_UP
      case gc.KEY_DOWN:
        control <- CURSOR_DOWN
      case 'a':
        control <- ADD
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
  items, err := make(chan *[]transmission.TorrentListItem), make(chan error)
  go func() {
    // First poll.
    list, e := client.List()
    err <- e
    items <- list

    for {
      <-time.After(time.Duration(3) * time.Second)
      list, e := client.List()
      err <- e
      items <- list
    }
  }()

  // Initial window state.
  state := &ListWindowState{}
  state.Offset, state.Cursor = 0, -1
  state.Rows, state.Cols = screen.MaxYX()

  // Handle updates and control.
  func(control chan input, err chan error, list chan torrents) {
    for {
      select {
      case e := <-err:
        state.Error = e
      case l := <-list:
        state.Items = l
        state.Cursor = minInt(len(*state.Items) - 1, state.Cursor)
        drawList(screen, *state)
      case c := <-control:
        switch c {
        case EXIT:
          return
        case DELETE, DELETE_WITH_DATA:
          if state.Cursor == -1 {
            break
          }

          if op := state.PendingOperation; op != nil && op.Operation == c && op.Item.Id == (*state.Items)[state.Cursor].Id {
            go handleOperation(screen, client, *op, list, err)
            state.PendingOperation = nil
          } else {
            state.PendingOperation = &ListOperation{ c, &(*state.Items)[state.Cursor] }
          }
        case CURSOR_UP:
          state.Cursor, state.PendingOperation = maxInt(0, state.Cursor - 1), nil
        case CURSOR_DOWN:
          state.Cursor, state.PendingOperation = minInt(state.Cursor + 1, len(*state.Items) - 1), nil
        case RESIZE:
          gc.End()
          screen.Refresh()
          state.Rows, state.Cols = screen.MaxYX()
        case ADD:
          res := make(chan error)
          NewTorrentWindow(screen, reader, res)
        }

        // Update offset if needed.
        if state.Cursor > (state.Rows - INFO_HEIGHT) + state.Offset - 1 {
          state.Offset = minInt(state.Offset + 1, len(*state.Items) - (state.Rows - INFO_HEIGHT))
        } else if state.Cursor < state.Offset {
          state.Offset = maxInt(state.Offset - 1, 0)
        }

        // Redraw.
        drawList(screen, *state)
      }
    }
  }(out, err, items)

  reader.RemoveObserver(observer)
}

func drawList(window *gc.Window, state ListWindowState) {
  if state.Items == nil {
    return
  }

  row, col := state.Rows, state.Cols
  titleLength := col - 63
  format := fmt.Sprintf("%%5d %%-%ds %%-6s %%-9s %%-12s %%-6.3f %%-9s %%-9s", titleLength)

  // Legend.
  legendFormat := fmt.Sprintf("%%5s %%-%ds %%-6s %%-9s %%-12s %%-6s %%-9s %%-9s", titleLength)
  window.MovePrintf(0, 0, legendFormat, "Id", "Name", "Done", "Size", "Status", "Ratio", "Down", "Up")
  window.HLine(1, 0, gc.ACS_HLINE, col)

  // List.
  x, y := 0, HEADER_HEIGHT
  for index, item := range (*state.Items)[state.Offset:] {
    var attribute gc.Char
    if index + state.Offset == state.Cursor {
      attribute = gc.A_REVERSE
    } else {
      attribute = gc.A_NORMAL
    }

    withAttribute(window, attribute, func(window *gc.Window) {
      // Format: ID | Name | Done | Size | Status
      window.MovePrintf(y, x, format,
        item.Id,
        item.Name[0:minInt(len(item.Name), titleLength)],
        fmt.Sprintf("%d%%", ((item.SizeWhenDone - item.LeftUntilDone)/item.SizeWhenDone*100)),
        formatSize(item.SizeWhenDone),
        formatStatus(item.Status),
        item.Ratio,
        formatSpeed(item.DownloadSpeed),
        formatSpeed(item.UploadSpeed))
    })

    y += 1
    if y >= row - FOOTER_HEIGHT {
      break
    }
  }

  // Clear remaining lines if needed.
  for index := y; index < row - FOOTER_HEIGHT; index++ {
    window.HLine(index, 0, ' ', col)
  }

  // Status.
  window.HLine(row - FOOTER_HEIGHT, 0, gc.ACS_HLINE, col)
  if op := state.PendingOperation; op != nil {
    switch op.Operation {
    case DELETE:
      window.MovePrintf(row - FOOTER_HEIGHT + 1, 0,
        "Removing torrent %d from the list. Press 'd' again to confirm.", op.Item.Id)
    case DELETE_WITH_DATA:
      window.MovePrintf(row - FOOTER_HEIGHT + 1, 0,
        "Deleting torrent %d along with data. Press 'D' again to confirm.", op.Item.Id)
    }
  } else if state.Error != nil {
    window.MovePrintf(row - FOOTER_HEIGHT + 1, 0, "%s", state.Error)
  } else {
    window.HLine(row - FOOTER_HEIGHT + 1, 0, ' ', col)
  }

  window.Refresh()
}

func formatSize(size int64) string {
  switch true {
  case size >= (1024 * 1024 * 1024):
    return fmt.Sprintf("%.2fGB", float64(size)/float64(1024 * 1024 * 1024))
  case size >= (1024 * 1024):
    return fmt.Sprintf("%.2fMB", float64(size)/float64(1024 * 1024))
  case size >= 1024:
    return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
  default:
    return fmt.Sprintf("%dB", size)
  }
}

func formatSpeed(speed float32) string {
  switch true {
  case speed == 0:
    return "0"
  case speed >= (1024 * 1024 * 1024):
    return fmt.Sprintf("%.2fGB", speed/float32(1024 * 1024 * 1024))
  case speed >= (1024 * 1024):
    return fmt.Sprintf("%.2fMB", speed/float32(1024 * 1024))
  case speed >= 1024:
    return fmt.Sprintf("%.2fKB", speed/1024)
  default:
    return fmt.Sprintf("%.2fB", speed)
  }
}

func formatStatus(status int8) string {
  switch status {
  case transmission.TR_STATUS_STOPPED:
    return "Stopped"
  case transmission.TR_STATUS_CHECK_WAIT:
    return "Check queue"
  case transmission.TR_STATUS_CHECK:
    return "Checking"
  case transmission.TR_STATUS_DOWNLOAD_WAIT:
    return "In Queue"
  case transmission.TR_STATUS_DOWNLOAD:
    return "Download"
  case transmission.TR_STATUS_SEED_WAIT:
    return "Seed queue"
  case transmission.TR_STATUS_SEED:
    return "Seeding"
  }
  return "Unknown"
}

func handleOperation(screen *gc.Window, client *transmission.Client, operation interface{}, items chan *[]transmission.TorrentListItem, err chan error) {
  var e error

  switch operation.(type) {
  case ListOperation:
    lop := operation.(ListOperation)
    switch lop.Operation {
    case DELETE:
      e = client.Delete([]int64{ lop.Item.Id }, false)
    case DELETE_WITH_DATA:
      e = client.Delete([]int64{ lop.Item.Id }, true)
    default:
      e = fmt.Errorf("Unknown list operation type")
    }
  default:
    e = fmt.Errorf("Unknown operation type")
  }

  if e != nil {
    err <- e
  } else {
    list, e := client.List()
    items <- list
    err <- e
  }
}

