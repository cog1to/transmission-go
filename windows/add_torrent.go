package windows

import (
  gc "../goncurses"
  "../transmission"
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
  Cursor int
  Width int
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
  var trimmed []rune
  var runes []rune
  var index int

  switch state.Focus {
  case FOCUS_URL:
    runes = []rune(state.Url)
  case FOCUS_PATH:
    runes = []rune(state.Path)
  }

  if len(runes) == 0 && (char.Input == gc.KEY_BACKSPACE || char.Input == gc.KEY_DC) {
    return
  }

  switch char.Input {
  case gc.KEY_BACKSPACE:
    if state.Cursor == 0 {
      return
    } else {
      index = state.Cursor - 1
    }
    trimmed = remove(runes, index)
    state.Cursor -= 1
  case gc.KEY_DC:
    if state.Cursor == len(runes) {
      index = state.Cursor - 1
    } else {
      index = state.Cursor
    }
    trimmed = remove(runes, index)
  default:
    trimmed = []rune(string(runes) + fmt.Sprintf("%c", char.Input))
    state.Cursor += 1
  }


  switch state.Focus {
  case FOCUS_URL:
    state.Url = string(trimmed)
  case FOCUS_PATH:
    state.Path = string(trimmed)
  }

  state.Cursor = maxInt(minInt(len(trimmed), state.Cursor), 0)
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

  if state.Focus == FOCUS_URL {
    state.Cursor = minInt(state.Cursor, len([]rune(state.Url)))
  } else if state.Focus == FOCUS_PATH {
    state.Cursor = minInt(state.Cursor, len([]rune(state.Path)))
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

type NewTorCursorMove struct {
  Direction int
}

func (action NewTorCursorMove) UpdateState(state *NewTorrentWindowState) {
  newPosition := state.Cursor + action.Direction

  var input string
  if state.Focus == FOCUS_URL {
    input = state.Url
  } else if state.Focus == FOCUS_PATH {
    input = state.Path
  }

  length := len([]rune(input))
  offset := maxInt(0, length - state.Width + 1)
  newPosition = maxInt(offset, minInt(newPosition, length))
  state.Cursor = newPosition
}

type NewTorCursorJump struct {
  Direction int
}

func (action NewTorCursorJump) UpdateState(state *NewTorrentWindowState) {
  var input []rune
  switch state.Focus {
  case FOCUS_URL:
    input = []rune(state.Url)
  case FOCUS_PATH:
    input = []rune(state.Path)
  }

  switch action.Direction {
  case 1:
    state.Cursor = len(input)
  case -1:
    offset := maxInt(0, len(input) - state.Width + 1)
    state.Cursor = maxInt(offset, 0)
  }
}

/* Main loop */

func NewTorrentWindow(source *gc.Window, reader *InputReader, client *transmission.Client, errorDrawer func(error)) {
  rows, cols := source.MaxYX()

  height, width := 10, minInt(cols, maxInt(60, cols * 3 / 4))
  y, x := (rows - height) / 2, (cols - width) / 2

  window, err := gc.NewWindow(height, width, y, x)
  window.Keypad(true)

  if err != nil {
    errorDrawer(err)
    return
  }

  // Window state
  state := &NewTorrentWindowState{ Width: width - 4 }

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
      case gc.KEY_LEFT:
        return NewTorCursorMove { -1 }
      case gc.KEY_RIGHT:
        return NewTorCursorMove { 1 }
      case gc.KEY_HOME:
        return NewTorCursorJump { -1 }
      case gc.KEY_END:
        return NewTorCursorJump { 1 }
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
      state.Result = NEW_RESULT_NO
      err := client.AddTorrent(state.Url, state.Path)
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
  window.MovePrintf(1, startX, "Add torrent")
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // URL
  window.MovePrintf(3, startX, "Torrent file or URL:")
  window.ColorOn(1)
  urlRunes := []rune(state.Url)
  urlOffset := maxInt(0, len(urlRunes) - width + 1)
  window.MovePrintf(4, startX, string(urlRunes[urlOffset:]))
  if urlOffset == 0 {
    window.MovePrintf(4, startX + len(urlRunes), strings.Repeat(" ", width - len(urlRunes)))
  } else {
    window.MovePrintf(4, startX + width - 1, " ")
  }
  window.ColorOff(1)

  // Path
  window.MovePrintf(5, startX, "Download path:")
  window.ColorOn(1)
  pathRunes := []rune(state.Path)
  pathOffset := maxInt(0, len(pathRunes) - width + 1)
  window.MovePrintf(6, startX, string(pathRunes[pathOffset:]))
  if pathOffset == 0 {
    window.MovePrintf(6, startX + len(pathRunes), strings.Repeat(" ", width - len(pathRunes)))
  } else {
    window.MovePrintf(6, startX + width - 1, " ")
  }
  window.ColorOff(1)

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
  withAttribute(window, attribute, func(window *gc.Window) {
    window.MovePrintf(8, startX + (buttonWidth - len("Confirm")) / 2, "Confirm")
  })

  // Cancel
  if state.Focus == FOCUS_CANCEL {
    attribute = gc.A_REVERSE
  } else {
    attribute = gc.A_NORMAL
  }
  withAttribute(window, attribute, func(window *gc.Window) {
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
    window.Move(4, startX + minInt(state.Cursor - urlOffset, width - 1))
  case FOCUS_PATH:
    window.Move(6, startX + minInt(state.Cursor - pathOffset, width - 1))
  }

  window.Refresh()
}

