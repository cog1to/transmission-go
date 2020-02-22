package windows

import (
  "fmt"
  gc "../goncurses"
  "../utils"
  "../suggestions"
  "../transmission"
)

type NewTorResult int

const (
  NEW_RESULT_NO NewTorResult = iota
  NEW_RESULT_CONFIRM = iota
  NEW_RESULT_CANCEL = iota
)

type NewTorrentWindowState struct {
  UrlField *InputField
  PathField *InputField
  Focus int
  Result NewTorResult
}

const (
  FOCUS_URL int = 0
  FOCUS_PATH = 1
  FOCUS_CONFIRM = 2
  FOCUS_CANCEL = 3
)

/* Window */

type AddTorrentWindow struct {
  client *transmission.Client
  parent *gc.Window
  window *gc.Window
  manager *WindowManager
  state *NewTorrentWindowState
  onError func(error)
}

func (window *AddTorrentWindow) IsFullScreen() bool {
  return false
}

func (window *AddTorrentWindow) SetActive(active bool) {
  if active {
    window.state.UrlField.IsActive = true
    window.manager.AddInputReader(window.state.UrlField)
  } else {
    if window.state.Focus == FOCUS_URL {
      window.manager.RemoveInputReader(window.state.UrlField)
    } else if window.state.Focus == FOCUS_PATH {
      window.manager.RemoveInputReader(window.state.PathField)
    }
    gc.Cursor(0)
  }
}

func (dialog *AddTorrentWindow) Draw() {
  window, state := dialog.window, dialog.state

  window.Erase()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.MaxYX()
  startX, width := 2, col-4

  // Header
  window.MovePrint(1, startX, "Add torrent")
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // URL
  window.MovePrint(3, startX, "Torrent file or URL:")
  state.UrlField.Draw()

  // Path
  window.MovePrintf(5, startX, "Download path:")
  state.PathField.Draw()

  // Controls delimiter
  window.HLine(7, 1, gc.ACS_HLINE, col-2)

  buttonWidth := width / 2
  attribute := gc.A_NORMAL

  // Confirm
  if state.Focus == FOCUS_CONFIRM {
    attribute = gc.A_REVERSE
  } else {
    attribute = gc.A_NORMAL
  }
  utils.WithAttribute(window, attribute, func(window *gc.Window) {
    window.MovePrintf(8, startX + (buttonWidth - len("Confirm")) / 2, "Confirm")
  })

  // Cancel
  if state.Focus == FOCUS_CANCEL {
    attribute = gc.A_REVERSE
  } else {
    attribute = gc.A_NORMAL
  }
  utils.WithAttribute(window, attribute, func(window *gc.Window) {
    window.MovePrintf(8, startX + buttonWidth + (buttonWidth - len("Cancel")) / 2, "Cancel")
  })

  // Enable cursor on input fields.
  if (state.Focus < 2) {
    gc.Cursor(1)
  } else {
    gc.Cursor(0)
  }

  // Move cursor if needed.
  switch state.Focus {
  case FOCUS_URL:
    state.UrlField.SetCursor(window)
  case FOCUS_PATH:
    state.PathField.SetCursor(window)
  }

  window.Refresh()
}

func (window *AddTorrentWindow) Resize() {
  height, width, y, x := MeasureAddTorrentWindow(window.parent)
  window.state.UrlField.Length = width - 4
  window.state.PathField.Length = width - 4
  window.window.MoveWindow(y, x)
  window.window.Resize(height, width)
}

func MeasureAddTorrentWindow(parent *gc.Window) (int, int, int, int) {
  rows, cols := parent.MaxYX()

  height, width := 10, utils.MinInt(cols, utils.MaxInt(60, cols * 3 / 4))
  y, x := (rows - height) / 2, (cols - width) / 2
  return height, width, y, x
}

func (window *AddTorrentWindow) OnInput(key gc.Key) {
  switch key {
  case gc.KEY_DOWN, gc.KEY_TAB:
    window.UpdateFocus(nil, 1)
  case gc.KEY_UP:
    window.UpdateFocus(nil, -1)
  case '\n':
    if window.state.Focus == FOCUS_CANCEL {
      window.manager.RemoveWindow(window)
    } else {
      url, path := utils.ExpandHome(string(window.state.UrlField.Value)), utils.ExpandHome(string(window.state.PathField.Value))
      err := window.client.AddTorrent(url, path)
      if err != nil {
        window.onError(fmt.Errorf("Error: %s", err))
      } else {
        window.manager.RemoveWindow(window)
      }
    }
  }
}

func (window *AddTorrentWindow) HandleInputFieldUpdate(field *InputField, result InputFieldResult) {
  switch result {
  case FOCUS_FORWARD:
    window.UpdateFocus(field, 1)
  case FOCUS_BACKWARD:
    window.UpdateFocus(field, -1)
  case UPDATE:
    go func() {
      window.manager.Draw <- true
    }()
  }
}

func (window *AddTorrentWindow) UpdateFocus(source *InputField, direction int) {
  if (source != nil) {
    source.IsActive = false
    window.manager.RemoveInputReader(source)
  }

  if direction > 0 {
    window.state.Focus = (window.state.Focus + direction) % 4
  } else {
    window.state.Focus = window.state.Focus + direction
    if window.state.Focus < 0 {
      window.state.Focus = 3
    }
  }

  var newInput *InputField
  switch window.state.Focus {
  case FOCUS_URL:
    newInput = window.state.UrlField
  case FOCUS_PATH:
    newInput = window.state.PathField
  }

  if newInput != nil {
    newInput.IsActive = true
    window.manager.AddInputReader(newInput)
  }

  go func() {
    window.manager.Draw <- true
  }()
}

func NewAddTorrentWindow(client *transmission.Client, parent *gc.Window, manager *WindowManager, onError func(error)) *AddTorrentWindow {
  height, width, y, x := MeasureAddTorrentWindow(parent)
  window, _ := gc.NewWindow(height, width, y, x)
  window.Keypad(true)

  // Window state
  state := &NewTorrentWindowState{
    UrlField: &InputField{ 2, 4, width - 4, false, 0, 0, true, []rune{}, 0, "", suggestions.GetSuggestedPaths, nil, manager, window, nil },
    PathField: &InputField{ 2, 6, width - 4, false, 0, 0, false, []rune{}, 0, "", suggestions.GetSuggestedPaths, nil, manager, window, nil }}

  dialog := &AddTorrentWindow{
    client,
    parent,
    window,
    manager,
    state,
    onError}

  // Hook up input field listeners.
  state.UrlField.OnResult = dialog.HandleInputFieldUpdate
  state.PathField.OnResult = dialog.HandleInputFieldUpdate

  return dialog
}

