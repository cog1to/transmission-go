package windows

import (
  gc "../goncurses"
  "../transmission"
  "fmt"
  "strings"
  "os"
  "../suggestions"
  "../utils"
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
  Cursor int
  Width int
  Result NewTorResult
  Reader *InputReader
  Input chan InputFieldResult
}

const (
  FOCUS_URL int = 0
  FOCUS_PATH = 1
  FOCUS_CONFIRM = 2
  FOCUS_CANCEL = 3
)

/* Input handling */

type NewTorrentInput interface {
  UpdateState(state *NewTorrentWindowState)
}

type NewTorBlank struct { }
func (blank NewTorBlank) UpdateState(state *NewTorrentWindowState) { }

type NewTorMove struct {
  Direction int
}

func (move NewTorMove) UpdateState(state *NewTorrentWindowState) {
  if move.Direction > 0 {
    state.Focus = (state.Focus + move.Direction) % 4
  } else {
    state.Focus = state.Focus + move.Direction
    if state.Focus < 0 {
      state.Focus = 3
    }
  }

  switch state.Focus {
  case FOCUS_URL:
    go state.UrlField.Activate(state.Reader, state.Input, false)
  case FOCUS_PATH:
    go state.PathField.Activate(state.Reader, state.Input, false)
  }
}

type NewTorAction struct {
  Confirmed bool
}

func (action NewTorAction) UpdateState(state *NewTorrentWindowState) {
  switch action.Confirmed {
  case true:
    state.Result = NEW_RESULT_CONFIRM
  default:
    state.Result = NEW_RESULT_CANCEL
  }
}

type NewTorNotRecognized struct { }
func (empty NewTorNotRecognized) UpdateState(state *NewTorrentWindowState) { }

type NewTorRefresh struct {}
func (empty NewTorRefresh) UpdateState(state *NewTorrentWindowState) { }

/* Main loop */

func NewTorrentWindow(source *gc.Window, reader *InputReader, client *transmission.Client, errorDrawer func(error)) {
  rows, cols := source.MaxYX()

  height, width := 10, utils.MinInt(cols, utils.MaxInt(60, cols * 3 / 4))
  y, x := (rows - height) / 2, (cols - width) / 2

  window, err := gc.NewWindow(height, width, y, x)
  window.Keypad(true)

  if err != nil {
    errorDrawer(err)
    return
  }

  // Window state
  state := &NewTorrentWindowState{
    Width: width - 4,
    UrlField: &InputField{ 2, 4, width - 4, 0, 0, true, []rune{}, 0, "", suggestions.GetSuggestedPaths, nil },
    PathField: &InputField{ 2, 6, width - 4, 0, 0, false, []rune{}, 0, "", suggestions.GetSuggestedPaths, nil },
    Reader: reader,
    Input: make(chan InputFieldResult)}

  // Handle  user input.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)

  get_input := func() NewTorrentInput {
    select {
      case ch := <-observer:
        switch ch {
        case gc.KEY_DOWN, gc.KEY_TAB:
          return NewTorMove{ 1 }
        case gc.KEY_UP:
          return NewTorMove{ -1 }
        case gc.KEY_LEFT, gc.KEY_RIGHT:
          if state.Focus == FOCUS_CANCEL {
            return NewTorMove { -1 }
          } else {
            return NewTorMove { 1 }
          }
        case '\n':
          return NewTorAction{ state.Focus == FOCUS_CONFIRM }
        }
      case fr := <-state.Input:
        switch fr {
        case FOCUS_FORWARD:
          return NewTorMove { 1 }
        case FOCUS_BACKWARD:
          return NewTorMove { -1 }
        case UPDATE:
          return NewTorRefresh{}
        }
        break
    }

    return NewTorNotRecognized{}
  }

  // Initial draw.
  drawNewTorrentWindow(window, *state)

  // First field is activated by default.
  go state.UrlField.Activate(state.Reader, state.Input, false)

  for {
    input := get_input()
    input.UpdateState(state)

    if state.Result == NEW_RESULT_CONFIRM {
      state.Result = NEW_RESULT_NO

      url, path := expandHome(string(state.UrlField.Value)), expandHome(string(state.PathField.Value))

      err := client.AddTorrent(url, path)
      if err != nil {
        errorDrawer(fmt.Errorf("Error: %s", err))
      } else {
        break
      }
    } else if state.Result == NEW_RESULT_CANCEL {
      break
    } else {
      drawNewTorrentWindow(window, *state)
    }
  }

  reader.RemoveObserver(observer)
  window.Delete()
}

func drawNewTorrentWindow(window *gc.Window, state NewTorrentWindowState) {
  window.Erase()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.MaxYX()
  startX, width := 2, col-4

  // Header
  window.MovePrint(1, startX, "Add torrent")
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // URL
  window.MovePrint(3, startX, "Torrent file or URL:")
  state.UrlField.Draw(window)

  // Path
  window.MovePrintf(5, startX, "Download path:")
  state.PathField.Draw(window)

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

func expandHome(input string) string {
  if strings.HasPrefix(input, "~") {
    return strings.Replace(input, "~", os.Getenv("HOME"), 1)
  }
  return input
}
