package windows

import (
  gc "../goncurses"
  "strings"
  "fmt"
)

type PromptState struct {
  Value string
  Title string
  Limit int
}

const CONTROLS_TEXT = "RETURN - Confirm | ESC - Cancel"

func Prompt(parent *gc.Window, reader *InputReader, title string, limit int, charset string) string {
  rows, cols := parent.MaxYX()
  height, width := 5, maxInt(4 + len(title) + 1 + limit + 1, len(CONTROLS_TEXT) + 4)
  y, x := (rows - height) / 2, (cols - width) / 2

  prompt, err := gc.NewWindow(height, width, y, x)
  if err != nil {
    return ""
  }
  defer prompt.Delete()

  // Handle user input.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)
  defer reader.RemoveObserver(observer)

  // Enable cursor.
  gc.Cursor(1)
  defer gc.Cursor(0)

  state := &PromptState{ "", title, limit }
  Loop: for {
    drawPrompt(prompt, *state)

    // Read input
    input := <-observer

    switch input {
    case gc.KEY_RETURN:
      break Loop
    case gc.KEY_ESC:
      state.Value = ""
      break Loop
    case gc.KEY_BACKSPACE, gc.KEY_DC, 0x7f:
      if len(state.Value) > 0 {
        state.Value = string(([]rune(state.Value))[:len(state.Value)-1])
      }
    default:
      str := fmt.Sprintf("%c", input)
      if strings.Contains(charset, str) && len(state.Value) < limit {
        state.Value = state.Value + str
      }
    }
  }

  return state.Value
}

func drawPrompt(window *gc.Window, state PromptState) {
  window.Erase()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.MaxYX()
  startX, width := 2, col-4

  // Prompt.
  window.MovePrint(1, startX, state.Title + " ")
  window.ColorOn(1)
  window.Print(state.Value + strings.Repeat(" ", (state.Limit + 1) - len(state.Value)))
  window.ColorOff(1)

  // Delimiter.
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // Controls reminder.
  window.MovePrint(3, startX + (width - len(CONTROLS_TEXT)) / 2, CONTROLS_TEXT)

  // Cursor.
  gc.Cursor(1)
  window.Move(1, len(state.Title + " ") + len(state.Value) + 2)

  window.Refresh()
}
