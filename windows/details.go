package windows

import (
  gc "../goncurses"
  "../transmission"
  wchar "../cgo.wchar"
  "time"
  "fmt"
  "strings"
)

const DETAILS_HEADER_HEIGHT = 4

type DetailsWindowState struct {
  Torrent *transmission.TorrentDetails
  List List
}


func TorrentDetailsWindow(
  source *gc.Window,
  reader *InputReader,
  client *transmission.Client,
  errorDrawer func(error),
  id int) {
  rows, cols := source.MaxYX()

  window, err := gc.NewWindow(rows-2, cols, 0, 0)
  window.Keypad(true)

  if err != nil {
    errorDrawer(err)
    return
  }

  // Handle user input.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)

  // Handle details update.
  details, e := make(chan *transmission.TorrentDetails), make(chan error)
  go func() {
    // First poll.
    getDetails(client, id, details, e)

    for {
      <-time.After(time.Duration(3) * time.Second)
      getDetails(client, id, details, e)
    }
  }()

  var state *DetailsWindowState

  formatter := func(file interface{}, width int) string {
    item := file.(transmission.TorrentFile)

    var filename string = item.Name
    if state.Torrent != nil && strings.HasPrefix(filename, state.Torrent.Name + "/") {
      filename = strings.TrimPrefix(filename, state.Torrent.Name + "/")
    }

    // Format: # - Done - Priority - Get - Size - Name
    maxTitleLength := width - 31
    title := []rune(filename)
    croppedTitleLength := minInt(maxTitleLength, len(title))
    croppedTitle := title[0:croppedTitleLength]
    spacesLength := maxTitleLength - croppedTitleLength

    format := "%3d %-6s %-8s %-3s %-9s %s%s"
    return fmt.Sprintf(format,
      item.Number,
      fmt.Sprintf("%3.0f%%", (float32(item.BytesCompleted)/float32(item.Length))*100.0),
      formatPriority(item.Priority),
      formatFlag(item.Wanted),
      formatSize(item.Length),
      string(croppedTitle),
      strings.Repeat(" ", spacesLength))
  }

  state = &DetailsWindowState{
    List: List{
      window,
      formatter,
      DETAILS_HEADER_HEIGHT+2,
      0,
      0,
      0,
      0,
      []int{},
      0,
      []Identifiable{}}}

  // Initial draw.
  drawTorrentDetailsWindow(window, *state)

  Loop: for {
    select {
    case ch := <-observer:
      switch ch {
      case 'q', 'h', gc.KEY_LEFT:
        break Loop
      case ' ':
        state.List.Select()
      case 'j', gc.KEY_DOWN:
        state.List.MoveCursor(1)
      case 'k', gc.KEY_UP:
        state.List.MoveCursor(-1)
      case gc.KEY_PAGEUP:
        state.List.Page(-1)
      case gc.KEY_PAGEDOWN:
        state.List.Page(1)
      case 'c':
        state.List.ClearSelection()
      case 'p':
        // Change priority.
        items := state.List.GetSelection()
        if len(items) > 0 {
          files := toFileList(items)
          ids, priority := idsAndNextPriority(files)
          go updatePriority(client, id, ids, priority, details, e)
        }
      case 'g':
        // Change 'wanted' status.
        items := state.List.GetSelection()
        if len(items) > 0 {
          files := toFileList(items)
          ids, wanted := idsAndNextWanted(files)
          go updateWanted(client, id, ids, wanted, details, e)
        }
      case 'L':
        // Change download limit.
        intPrompt(window, reader, "Set download limit (KB):",
          state.Torrent.DownloadLimit, state.Torrent.DownloadLimited,
          func(limit int) { go setDownloadLimit(client, id, limit, details, e) },
          func(err error) { go func() { e <- err }() })
      case 'U':
        // Change upload limit.
        intPrompt(window, reader, "Set upload limit (KB):",
          state.Torrent.UploadLimit, state.Torrent.UploadLimited,
          func(limit int) { go setUploadLimit(client, id, limit, details, e) },
          func(err error) { go func() { e <- err }() })
      case gc.KEY_F1:
        items := []HelpItem{
          HelpItem{ "qh←", "Go back to torrent list" },
          HelpItem{ "jk↑↓", "Move cursor up and down" },
          HelpItem{ "Space", "Toggle selection" },
          HelpItem{ "c", "Clear selection" },
          HelpItem{ "g", "Download/Don't download selected file(s)" },
          HelpItem{ "p", "Change priority of selected file(s)" },
          HelpItem{ "L", "Set torrent's download speed limit" },
          HelpItem{ "U", "Set torrent's upload speed limit" }}
        CheatsheetWindow(window, reader, items)
      }
    case torrent := <-details:
      state.Torrent = torrent
      if torrent != nil {
        state.List.Items = generalizeFiles(torrent.Files)
      } else {
        state.List.Items = []Identifiable{}
      }
    case detailsError := <-e:
      errorDrawer(detailsError)
    }

    drawTorrentDetailsWindow(window, *state)
  }

  reader.RemoveObserver(observer)
  window.Delete()
}

func drawTorrentDetailsWindow(window *gc.Window, state DetailsWindowState) {
  window.Erase()
  _, col := window.MaxYX()

  // Header
  if (state.Torrent != nil) {
    item := *state.Torrent

    // Name.
    ws, convertError := wchar.FromGoString(item.Name)
    if (convertError == nil) {
      withAttribute(window, gc.A_BOLD, func(window *gc.Window) {
        window.MovePrintW(0, 0, ws)
      })
    }

    // Data.
    done := fmt.Sprintf(
      "%3.0f%%",
      (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0)
    size := formatSize(item.SizeWhenDone)
    ratio := item.Ratio
    status := formatStatus(item.Status)
    dataString := fmt.Sprintf("Size: %s | Done: %s | Ratio: %.3f | Status: %s", size, done, ratio, status)
    window.MovePrint(1, 0, dataString)

    // Speeds.
    downSpeed := formatSpeed(item.DownloadSpeed)
    downLimit := formatSpeedWithFlag(float32(item.DownloadLimit * 1024), item.DownloadLimited)
    upSpeed := formatSpeed(item.UploadSpeed)
    upLimit := formatSpeedWithFlag(float32(item.UploadLimit * 1024), item.UploadLimited)

    speedString := fmt.Sprintf("Down: %s (Limit: %s) | Up: %s (Limit: %s)",
      downSpeed,
      downLimit,
      upSpeed,
      upLimit)
    window.MovePrintf(2, 0, "%s%s", speedString, strings.Repeat(" ", col - len(speedString)))

    // Separator.
    window.HLine(3, 0, gc.ACS_HLINE, col)
  }

  // Legend: # - Done - Priority - Get - Size - Name
  legendFormat := "%3s %-6s %-8s %-3s %-9s %s"
  window.MovePrintf(DETAILS_HEADER_HEIGHT, 0, legendFormat, "#", "Done", "Priority", "Get", "Size", "Name")
  window.HLine(DETAILS_HEADER_HEIGHT + 1, 0, gc.ACS_HLINE, col)

  state.List.Draw()

  window.Refresh()
}

func getDetails(client *transmission.Client, id int, tor chan *transmission.TorrentDetails, err chan error) {
  item, e := client.TorrentDetails(id)
  err <- e
  tor <- item
}

func updatePriority(client *transmission.Client, id int, ids []int, priority int, tor chan *transmission.TorrentDetails, err chan error) {
  e := client.SetPriority(id, ids, priority)

  if e != nil {
    err <- e
  } else {
    getDetails(client, id, tor, err)
  }
}

func updateWanted(client *transmission.Client, id int, ids []int, wanted bool, tor chan *transmission.TorrentDetails, err chan error) {
  e := client.SetWanted(id, ids, wanted)

  if e != nil {
    err <- e
  } else {
    getDetails(client, id, tor, err)
  }
}

func setDownloadLimit(client *transmission.Client, id int, limit int, tor chan *transmission.TorrentDetails, err chan error) {
  e := client.SetDownloadLimit(id, limit)
  if e != nil {
    err <- e
  } else {
    getDetails(client, id, tor, err)
  }
}

func setUploadLimit(client *transmission.Client, id int, limit int, tor chan *transmission.TorrentDetails, err chan error) {
  e := client.SetUploadLimit(id, limit)
  if e != nil {
    err <- e
  } else {
    getDetails(client, id, tor, err)
  }
}

func idsAndNextPriority(files []transmission.TorrentFile) ([]int, int) {
  minPriority := 99
  ids := make([]int, len(files))
  for i, file := range files {
    minPriority = minInt(minPriority, file.Priority)
    ids[i] = file.Number
  }

  nextPriority := transmission.TR_PRIORITY_LOW
  if minPriority != transmission.TR_PRIORITY_HIGH {
    nextPriority = minPriority + 1
  }

  return ids, nextPriority
}

func idsAndNextWanted(files []transmission.TorrentFile) ([]int, bool) {
  wanted := true
  ids := make([]int, len(files))
  for i, file := range files {
    wanted = wanted && file.Wanted
    ids[i] = file.Number
  }

  return ids, !wanted
}