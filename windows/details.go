package windows

import (
  "strings"
  "fmt"
  "../transmission"
  gc "../goncurses"
  "../worker"
  "../utils"
  "../list"
  "../transform"
)

const DETAILS_HEADER_HEIGHT = 4
const DETAILS_FOOTER_HEIGHT = 2

type TorrentDetailsState struct {
  Torrent *transmission.TorrentDetails
  List list.List
  Obfuscated bool
  Error error
}

type TorrentDetailsWindow struct {
  client *transmission.Client
  workers worker.WorkerList
  window *gc.Window
  manager *WindowManager
  state *TorrentDetailsState
}

func (window *TorrentDetailsWindow) IsFullScreen() bool {
  return true
}

func (window *TorrentDetailsWindow) SetActive(active bool) {
  if active {
    window.workers.Start()
  } else {
    window.workers.Stop()
  }
}

func (window *TorrentDetailsWindow) OnInput(key gc.Key) {
  state := window.state

  switch key {
  case 'q', 'h', gc.KEY_LEFT:
    window.manager.RemoveWindow(window)
    return
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
      files := transform.ToFileList(items)
      ids, priority := transform.IdsAndNextPriority(files)
      go func() {
        updatePriority(window.client, state.Torrent.Id, ids, priority, state)
        window.manager.Draw <- true
      }()
    }
  case 'g':
    // Change 'wanted' status.
    items := state.List.GetSelection()
    if len(items) > 0 {
      files := transform.ToFileList(items)
      ids, wanted := transform.IdsAndNextWanted(files)
      go func() {
        updateWanted(window.client, state.Torrent.Id, ids, wanted, state)
        window.manager.Draw <- true
      }()
    }
  case 'L':
    // Change download limit.
    IntPrompt(
      window.window,
      window.manager,
      "Set download limit (KB):",
      state.Torrent.DownloadLimit,
      state.Torrent.DownloadLimited,
      func(limit int) {
        go func() {
          setDownloadLimit(window.client, state.Torrent.Id, limit, window.state)
          window.manager.Draw <- true
        }()
      },
      func(err error) {
        state.Error = err
      })
  case 'U':
    // Change upload limit.
    IntPrompt(
      window.window,
      window.manager,
      "Set upload limit (KB):",
      state.Torrent.UploadLimit,
      state.Torrent.UploadLimited,
      func(limit int) {
        go func() {
          setUploadLimit(window.client, state.Torrent.Id, limit, window.state)
          window.manager.Draw <- true
        }()
      },
      func(err error) {
        state.Error = err
      })
  case gc.KEY_F1:
    showDetailsCheatsheet(window.window, window.manager)
  }

  go func() {
    window.manager.Draw <- true
  }()
}

func (window *TorrentDetailsWindow) Draw() {
  drawDetails(window.window, window.state)
}

func (window *TorrentDetailsWindow) Resize() {
  window.window.Refresh()
}

func NewTorrentDetailsWindow(client *transmission.Client, id int, obfuscated bool, parent *gc.Window, manager *WindowManager) *TorrentDetailsWindow {
  rows, cols := parent.MaxYX()

  window, err := gc.NewWindow(rows, cols, 0, 0)
  window.Keypad(true)

  if err != nil {
    return nil
  }

  var state *TorrentDetailsState

  formatter := func(file interface{}, width int, printer func(int, string)) {
    formatFile(file, width, obfuscated, state.Torrent, printer)
  }

  state = &TorrentDetailsState{
    List: list.List{
      window,
      formatter,
      DETAILS_HEADER_HEIGHT+2,
      DETAILS_FOOTER_HEIGHT,
      0,
      0,
      0,
      []int{},
      0,
      []list.Identifiable{}},
    Obfuscated: obfuscated}

  workers := worker.WorkerList{
    worker.Repeating(
      3,
      func() {
        getDetails(client, id, state)
        manager.Draw <- true
      })}

  return &TorrentDetailsWindow{
    client,
    workers,
    window,
    manager,
    state}
}

/* Drawing */

func formatFile(file interface{}, width int, obfuscated bool, torrent *transmission.TorrentDetails, printer func(int, string)) {
  item := file.(transmission.TorrentFile)

  var filename string = item.Name
  if torrent != nil && strings.HasPrefix(filename, torrent.Name + "/") {
    filename = strings.TrimPrefix(filename, torrent.Name + "/")
  }

  // Format: # - Done - Priority - Get - Size - Name
  maxTitleLength := width - 31
  title := []rune(filename)

  var croppedTitle []rune
  croppedTitleLength := utils.MinInt(maxTitleLength, len(title))
  if obfuscated {
    croppedTitle = []rune(utils.RandomString(croppedTitleLength))
  } else {
    croppedTitle = title[0:croppedTitleLength]
  }
  spacesLength := maxTitleLength - croppedTitleLength

  format := "%3d %-6s %-8s %-3s %-9s %s%s"
  details := fmt.Sprintf(format,
    item.Number,
    fmt.Sprintf("%3.0f%%", (float32(item.BytesCompleted)/float32(item.Length))*100.0),
    formatPriority(item.Priority),
    formatFlag(item.Wanted),
    formatSize(item.Length),
    string(croppedTitle),
    strings.Repeat(" ", spacesLength))
  printer(0, details)
}

func drawDetails(window *gc.Window, state *TorrentDetailsState) {
  window.Erase()
  _, col := window.MaxYX()

  // Header
  if (state.Torrent != nil) {
    item := *state.Torrent

    // Name.
    utils.WithAttribute(window, gc.A_BOLD, func(window *gc.Window) {
      if state.Obfuscated {
        window.MovePrint(0, 0, utils.RandomString(len([]rune(item.Name))))
      } else {
        window.MovePrint(0, 0, item.Name)
      }
    })

    // %Done. Handle unknown torrent size.
    var done string
    if item.SizeWhenDone == 0 {
      done = fmt.Sprintf("%3.0f%%", 0)
    } else {
      done = fmt.Sprintf(
        "%3.0f%%",
        (float32(item.SizeWhenDone - item.LeftUntilDone)/float32(item.SizeWhenDone))*100.0)
    }

    // Rest of the data.
    size := formatSize(item.SizeWhenDone)
    ratio := utils.MaxFloat32(item.Ratio, 0)
    status := formatStatus(item.Status)

    dataString := fmt.Sprintf("Size: %s | Done: %s | Ratio: %.3f | Status: %s", size, done, ratio, status)
    window.MovePrint(1, 0, dataString)

    // Speed values.
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

  // Draw List.
  state.List.Draw()

  // Draw Error.
  drawError(window, state.Error)

  window.Refresh()
}

func showDetailsCheatsheet(parent *gc.Window, manager *WindowManager) {
  items := []HelpItem{
    HelpItem{ "qh←", "Go back to torrent list" },
    HelpItem{ "jk↑↓", "Move cursor up and down" },
    HelpItem{ "Space", "Toggle selection" },
    HelpItem{ "c", "Clear selection" },
    HelpItem{ "g", "Download/Don't download selected file(s)" },
    HelpItem{ "p", "Change priority of selected file(s)" },
    HelpItem{ "L", "Set torrent's download speed limit" },
    HelpItem{ "U", "Set torrent's upload speed limit" }}

  cheatsheet := NewCheatsheet(parent, items, manager)
  manager.AddWindow(cheatsheet)
}

/* Network */

func getDetails(client *transmission.Client, id int, state *TorrentDetailsState) {
  torrent, e := client.TorrentDetails(id)

  state.Error = e
  state.Torrent = torrent
  if torrent != nil {
    state.List.Items = transform.GeneralizeFiles(torrent.Files)
  } else {
    state.List.Items = []list.Identifiable{}
  }
}

func updatePriority(client *transmission.Client, id int, ids []int, priority int, state *TorrentDetailsState) {
  e := client.SetPriority(id, ids, priority)

  state.Error = e
  if e == nil {
    getDetails(client, id, state)
  }
}

func updateWanted(client *transmission.Client, id int, ids []int, wanted bool, state *TorrentDetailsState) {
  e := client.SetWanted(id, ids, wanted)

  state.Error = e
  if e == nil {
    getDetails(client, id, state)
  }
}

func setDownloadLimit(client *transmission.Client, id int, limit int, state *TorrentDetailsState) {
  e := client.SetDownloadLimit(id, limit)

  state.Error = e
  if e == nil {
    getDetails(client, id, state)
  }
}

func setUploadLimit(client *transmission.Client, id int, limit int, state *TorrentDetailsState) {
  e := client.SetUploadLimit(id, limit)

  state.Error = e
  if e == nil {
    getDetails(client, id, state)
  }
}

