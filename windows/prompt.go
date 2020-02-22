package windows

import (
  "fmt"
  "strconv"
  gc "../goncurses"
  "../utils"
)

/* Data */

const CONTROLS_TEXT = "RETURN - Confirm | ESC - Cancel"

type PromptState struct {
  Title string
  Limit int
  Field *InputField
  Input chan InputFieldResult
}

/* Window */

type Prompt struct {
  parent *gc.Window
  window *gc.Window
  manager *WindowManager
  state *PromptState
}

func (window *Prompt) OnInput(key gc.Key) {
  // Prompt doesn't do anything, it should be captured by input field
}

func (window *Prompt) SetActive(active bool) {
  if active {
    gc.Cursor(1)
    window.manager.AddInputReader(window.state.Field)
  } else {
    gc.Cursor(0)
    window.manager.RemoveInputReader(window.state.Field)
  }
}

func (window *Prompt) IsFullScreen() bool {
  return false
}

func (window *Prompt) Draw() {
  window.window.Erase()
  window.window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.window.MaxYX()
  startX, width := 2, col-4

  // Prompt.
  window.window.MovePrint(1, startX, window.state.Title + " ")
  window.state.Field.Draw()

  // Delimiter.
  window.window.HLine(2, 1, gc.ACS_HLINE, col-2)

  // Controls reminder.
  window.window.MovePrint(3, startX + (width - len(CONTROLS_TEXT)) / 2, CONTROLS_TEXT)

  // Cursor.
  window.state.Field.SetCursor(window.window)

  window.window.Refresh()
}

func (window *Prompt) Resize() {
  height, width, y, x := MeasurePrompt(window.parent, window.state.Title, window.state.Limit)
  window.window.MoveWindow(y, x)
  window.window.Resize(height, width)
}

func MeasurePrompt(parent *gc.Window, title string, limit int) (int, int, int, int) {
  rows, cols := parent.MaxYX()
  height, width := 5, utils.MaxInt(4 + len(title) + 1 + limit + 1, len(CONTROLS_TEXT) + 4)
  y, x := (rows - height) / 2, (cols - width) / 2

  return height, width, y, x
}

func NewPrompt(
  parent *gc.Window,
  manager *WindowManager,
  title string,
  limit int,
  charset string,
  initial string,
  completion func(string),
  cancel func()) *Prompt {
  height, width, y, x := MeasurePrompt(parent, title, limit)
  prompt, _ := gc.NewWindow(height, width, y, x)

  initialRunes := []rune(initial)
  length := len(initialRunes)

  onResult := func(field *InputField, result InputFieldResult) {
    go func() {
      switch result {
      case UPDATE:
        manager.Draw <- true
      case CONFIRM:
        completion(string(field.Value))
      case CANCEL:
        cancel()
      }
    }()
  }

  state := &PromptState{
    title,
    limit,
    &InputField{
      2 + len(title + " "),
      1,
      limit + 1,
      true,
      utils.MaxInt(0, length - (limit + 1)),
      length,
      true,
      initialRunes,
      limit,
      charset,
      nil,
      nil,
      nil,
      prompt,
      onResult},
    make(chan InputFieldResult)}

  return &Prompt{
    parent,
    prompt,
    manager,
    state}
}

/* Public helpers */

func IntPrompt(parent *gc.Window, manager *WindowManager, title string, value int, flag bool, onFinish func(int), onError func(error)) {
  var initialValue string
  if flag && value > 0 {
    initialValue = fmt.Sprintf("%d", value)
  }

  handleResult := func(output string) {
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

  var prompt *Prompt
  prompt = NewPrompt(
    parent,
    manager,
    title,
    6,
    "0123456789",
    initialValue,
    func(output string) {
      manager.RemoveWindow(prompt)
      handleResult(output)
    },
    func() {
      manager.RemoveWindow(prompt)
    })
  manager.AddWindow(prompt)
}

