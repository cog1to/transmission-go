package windows

import (
  gc "../goncurses"
  "fmt"
  "strings"
)

type NewTorResult int

const (
  NEW_RESULT_NO NewTorResult = iota
  NEW_RESULT_CONFIRM = iota
  NEW_RESULT_CANCEL = iota
)

type NewTorrentWindowState struct {
  Url string
  Path string
  Focus int
  Result NewTorResult
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

type NewTorChar struct {
  Input gc.Key
}

func (char NewTorChar) UpdateState(state *NewTorrentWindowState) {
  switch state.Focus {
  case FOCUS_URL:
    if char.Input == gc.KEY_BACKSPACE || char.Input == gc.KEY_DC {
      if len(state.Url) > 0 {
        runes := []rune(state.Url)
        trimmed := runes[0:len(runes)-1]
        state.Url = string(trimmed)
      }
    } else {
      state.Url = state.Url + fmt.Sprintf("%c", char.Input)
    }
  case FOCUS_PATH:
    if char.Input == gc.KEY_BACKSPACE || char.Input == gc.KEY_DC {
      if len(state.Path) > 0 {
        runes := []rune(state.Path)
        trimmed := runes[0:len(state.Path)-1]
        state.Path = string(trimmed)
      }
    } else {
      state.Path = state.Path + fmt.Sprintf("%c", char.Input)
    }
  }
}

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

/* Main loop */

func NewTorrentWindow(source *gc.Window, reader *InputReader, result chan error) {
  rows, cols := source.MaxYX()

  height, width := 11, minInt(cols, maxInt(60, cols * 3 / 4))
  y, x := (rows - height) / 2, (cols - width) / 2

  window, err := gc.NewWindow(height, width, y, x)
  window.Keypad(true)

  if err != nil {
    result <- err
    return
  }

  // Window state
  state := &NewTorrentWindowState{}

  // Handle user input.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)

  get_input := func() NewTorrentInput {
    ch := <-observer
    switch state.Focus {
    case FOCUS_URL, FOCUS_PATH:
      switch ch {
      case '\n', gc.KEY_DOWN, gc.KEY_TAB:
        return NewTorMove{ 1 }
      case gc.KEY_UP:
        return NewTorMove{ -1 }
      default:
        return NewTorChar{ ch }
      }
    default:
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
    }

    return NewTorNotRecognized{}
  }

  // Initial draw.
  drawNewTorrentWindow(window, *state)

  for {
    input := get_input()
    input.UpdateState(state)

    if state.Result == NEW_RESULT_CONFIRM {
      break
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
  window.MovePrintf(1, startX, "Add torrent")
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // URL
  window.MovePrintf(3, startX, "URL: ")
  window.ColorOn(1)
  window.MovePrintf(4, startX, state.Url)
  urlRunes := []rune(state.Url)
  window.MovePrintf(4, startX + len(urlRunes), strings.Repeat(" ", width - len(urlRunes)))
  window.ColorOff(1)

  // Path
  window.MovePrintf(5, startX, "Path: ")
  window.ColorOn(1)
  window.MovePrintf(6, startX, state.Path)
  pathRunes := []rune(state.Path)
  window.MovePrintf(6, startX + len(pathRunes), strings.Repeat(" ", width - len(pathRunes)))
  window.ColorOff(1)

  // Controls delimiter
  window.HLine(8, 1, gc.ACS_HLINE, col-2)

  buttonWidth := width / 2
  attribute := gc.A_NORMAL

  // Confirm
  if state.Focus == FOCUS_CONFIRM {
    attribute = gc.A_REVERSE
  } else {
    attribute = gc.A_NORMAL
  }
  withAttribute(window, attribute, func(window *gc.Window) {
    window.MovePrintf(9, startX + (buttonWidth - len("Confirm")) / 2, "Confirm")
  })

  // Cancel
  if state.Focus == FOCUS_CANCEL {
    attribute = gc.A_REVERSE
  } else {
    attribute = gc.A_NORMAL
  }
  withAttribute(window, attribute, func(window *gc.Window) {
    window.MovePrintf(9, startX + buttonWidth + (buttonWidth - len("Cancel")) / 2, "Cancel")
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
    window.Move(4, startX + len([]rune(state.Url)))
  case FOCUS_PATH:
    window.Move(6, startX + len(state.Path))
  }

  window.Refresh()
}

