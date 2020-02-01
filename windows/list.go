package windows

import (
  "fmt"
  "time"
  "../transmission"
  gc "../goncurses"
  wchar "../cgo.wchar"
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
      case gc.KEY_PAGEDOWN:
        control <- CURSOR_PAGEDOWN
      case gc.KEY_PAGEUP:
        control <- CURSOR_PAGEUP
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
  items, err := make(chan torrents), make(chan error)
  go func() {
    // First poll.
    updateList(client, items, err)

    for {
      <-time.After(time.Duration(3) * time.Second)
      updateList(client, items, err)
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
        drawList(screen, *state)
      case items := <-list:
        if items != nil {
          state.Items = items
        } else {
          state.Items = &[]transmission.TorrentListItem{}
        }
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
        case CURSOR_PAGEUP:
          state.Offset = maxInt(state.Offset - (state.Rows - INFO_HEIGHT), 0)
          state.Cursor = maxInt(state.Cursor - (state.Rows - INFO_HEIGHT), 0)
        case CURSOR_PAGEDOWN:
          state.Offset = minInt(state.Offset + state.Rows - INFO_HEIGHT, maxInt(len(*state.Items) - (state.Rows - INFO_HEIGHT), 0))
          state.Cursor = minInt(state.Cursor + state.Rows - INFO_HEIGHT, len(*state.Items) - 1)
        case RESIZE:
          gc.End()
          screen.Refresh()
          state.Rows, state.Cols = screen.MaxYX()
        case ADD:
          errorDrawer := func(err error) {
            drawError(screen, err)
          }
          NewTorrentWindow(screen, reader, client, errorDrawer)
          go updateList(client, list, err)
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
  maxTitleLength := maxInt(0, col - 63)
  format := fmt.Sprintf("%%5d %%s%%s %%-6s %%-9s %%-12s %%-6.3f %%-9s %%-9s")

  // Legend.
  legendFormat := fmt.Sprintf("%%5s %%-%ds %%-6s %%-9s %%-12s %%-6s %%-9s %%-9s", maxTitleLength)
  window.MovePrintf(0, 0, legendFormat, "Id", "Name", "Done", "Size", "Status", "Ratio", "Down", "Up")
  window.HLine(1, 0, gc.ACS_HLINE, col)

  // List.
  x, y := 0, HEADER_HEIGHT
  for index, item := range (*state.Items)[state.Offset:] {
    title := []rune(item.Name)

    croppedTitleLength := minInt(maxTitleLength, len(title))
    croppedTitle := title[0:croppedTitleLength]
    spacesLength := maxTitleLength - croppedTitleLength

    var attribute gc.Char
    if index + state.Offset == state.Cursor {
      attribute = gc.A_REVERSE
    } else {
      attribute = gc.A_NORMAL
    }

    withAttribute(window, attribute, func(window *gc.Window) {
      output := fmt.Sprintf(format,
        item.Id,
        string(croppedTitle),
        strings.Repeat(" ", spacesLength),
        fmt.Sprintf("%3.0f%%", (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0),
        formatSize(item.SizeWhenDone),
        formatStatus(item.Status),
        item.Ratio,
        formatSpeed(item.DownloadSpeed),
        formatSpeed(item.UploadSpeed))

        ws, convertError := wchar.FromGoString(output)
        if (convertError == nil) {
          window.MovePrintW(y, x, ws)
        }
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

func handleOperation(screen *gc.Window, client *transmission.Client, operation interface{}, items chan torrents, err chan error) {
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
    updateList(client, items, err)
  }
}

func updateList(client *transmission.Client, items chan torrents, err chan error) {
    list, e := client.List()
    err <- e
    items <- list
}
