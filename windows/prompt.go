package windows

import (
  gc "../goncurses"
  "fmt"
  "strconv"
  "../utils"
)

type PromptState struct {
  Title string
  Limit int
  Field *InputField
  Input chan InputFieldResult
}

const CONTROLS_TEXT = "RETURN - Confirm | ESC - Cancel"

func Prompt(
  parent *gc.Window,
  reader *InputReader,
  title string,
  limit int,
  charset string,
  initial string) string {
  rows, cols := parent.MaxYX()
  height, width := 5, utils.MaxInt(4 + len(title) + 1 + limit + 1, len(CONTROLS_TEXT) + 4)
  y, x := (rows - height) / 2, (cols - width) / 2

  prompt, err := gc.NewWindow(height, width, y, x)
  if err != nil {
    return ""
  }
  defer prompt.Delete()

  // Enable cursor.
  gc.Cursor(1)
  defer gc.Cursor(0)

  initialRunes := []rune(initial)
  length := len(initialRunes)

  state := &PromptState{
    title,
    limit,
    &InputField{
      2 + len(title + " "),
      1,
      limit + 1,
      utils.MaxInt(0, length - (limit + 1)),
      length,
      true,
      initialRunes,
      limit,
      charset,
      nil,
      nil},
    make(chan InputFieldResult)}

  // Field is active by default.
  go state.Field.Activate(reader, state.Input, true)

  Loop: for {
    drawPrompt(prompt, *state)

    input := <-state.Input
    switch input {
    case CONFIRM:
      break Loop
    case CANCEL:
      state.Field.Value = []rune{}
      break Loop
    }
  }

  return string(state.Field.Value)
}

func drawPrompt(window *gc.Window, state PromptState) {
  window.Erase()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.MaxYX()
  startX, width := 2, col-4

  // Prompt.
  window.MovePrint(1, startX, state.Title + " ")
  state.Field.Draw(window)

  // Delimiter.
  window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // Controls reminder.
  window.MovePrint(3, startX + (width - len(CONTROLS_TEXT)) / 2, CONTROLS_TEXT)

  // Cursor.
  state.Field.SetCursor(window)

  window.Refresh()
}

/* Public helpers */

func intPrompt(window *gc.Window, reader *InputReader, title string, value int, flag bool, onFinish func(int), onError func(error)) {
  var initialValue string
  if flag && value > 0 {
    initialValue = fmt.Sprintf("%d", value)
  }

  output := Prompt(window, reader, title, 6, "0123456789", initialValue)
  if len(output) > 0 {
    limit, e := strconv.Atoi(output)
    if e != nil {
      onError(e)
    } else {
      onFinish(limit)
    }
  } else {
    onFinish(0)
  }
}

