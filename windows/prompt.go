package windows

import (
  "fmt"
  "strconv"
  tui "../tui"
  "../utils"
  "../suggestions"
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
  parent *tui.Window
  window *tui.Window
  manager *WindowManager
  state *PromptState
}

func (window *Prompt) OnInput(key tui.Key) {
  // Prompt doesn't do anything, it should be captured by input field
}

func (window *Prompt) SetActive(active bool) {
  if active {
    tui.ShowCursor()
    window.manager.AddInputReader(window.state.Field)
  } else {
    tui.HideCursor()
    window.manager.RemoveInputReader(window.state.Field)
  }
}

func (window *Prompt) IsFullScreen() bool {
  return false
}

func (window *Prompt) Draw() {
  window.window.Box()

  _, col := window.window.MaxYX()
  startX, width := 2, col-4

  // Prompt.
  window.window.MovePrint(1, startX, window.state.Title + " ")
  window.state.Field.Draw()

  // Delimiter.
  window.window.HLine(2, 1, col-2)

  // Controls reminder.
  window.window.MovePrint(3, startX + (width - len(CONTROLS_TEXT)) / 2, CONTROLS_TEXT)

  // Cursor.
  window.state.Field.SetCursor(window.window)
}

func (window *Prompt) Resize() {
  height, width, y, x := MeasurePrompt(window.parent, window.state.Title, window.state.Limit)
  window.window.Move(y, x)
  window.window.Resize(height, width)
}

func MeasurePrompt(parent *tui.Window, title string, limit int) (int, int, int, int) {
  rows, cols := parent.MaxYX()

  promptDecorationLength := 4 + len(title) + 1
  maxWidth := utils.MinInt(cols, utils.MaxInt(60, cols * 3 / 4))

  promptLength := utils.MaxInt(0, maxWidth - promptDecorationLength)
  if (limit > 0) {
    promptLength  = utils.MinInt(limit + 1, maxWidth - promptDecorationLength)
  }

  height, width := 5, utils.MaxInt(promptDecorationLength + promptLength, len(CONTROLS_TEXT) + 4)
  y, x := (rows - height) / 2, (cols - width) / 2

  return height, width, y, x
}

func NewPrompt(
  parent *tui.Window,
  manager *WindowManager,
  title string,
  limit int,
  charset string,
  initial string,
  completion func(string),
  cancel func(),
  suggester Suggester) *Prompt {
  height, width, y, x := MeasurePrompt(parent, title, limit)
  prompt := parent.Sub(y, x, height, width)

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

  var inputLength = width - len(title + " ") - 4

  state := &PromptState{
    title,
    limit,
    &InputField{
      2 + len(title + " "),
      1,
      inputLength,
      true,
      utils.MaxInt(0, length - (limit + 1)),
      length,
      true,
      initialRunes,
      limit,
      charset,
      suggester,
      nil,
      nil,
      prompt,
      onResult},
    make(chan InputFieldResult)}

  state.Field.UpdateSuggestion()

  return &Prompt{
    parent,
    prompt,
    manager,
    state}
}

/* Public helpers */

func IntPrompt(parent *tui.Window, manager *WindowManager, title string, value int, flag bool, onFinish func(int), onError func(error)) {
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
    },
    nil)
  manager.AddWindow(prompt)
}

func PathPrompt(parent *tui.Window, manager *WindowManager, title string, initial string, onFinish func(string), onError func(error)) {
  var initialValue = initial
  if initialValue == "" {
    initialValue = "~"
  }

  var prompt *Prompt
  prompt = NewPrompt(
    parent,
    manager,
    title,
    0,
    "",
    initialValue,
    func(output string) {
      manager.RemoveWindow(prompt)
      onFinish(output)
    },
    func() {
      manager.RemoveWindow(prompt)
    },
    suggestions.GetSuggestedPaths)
  manager.AddWindow(prompt)
}
