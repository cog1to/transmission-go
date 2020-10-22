package windows

import (
  tui "../tui"
  "fmt"
  "strings"
  "../utils"
)

type Suggester = func(string)([]string)

type InputField struct {
  X, Y, Length int
  IsModal bool
  Offset int
  Cursor int
  IsActive bool
  Value []rune
  Limit int
  Charset string
  Suggester Suggester
  Suggestion *string
  Manager *WindowManager
  Parent tui.Drawable
  OnResult func(*InputField, InputFieldResult)
}

type InputFieldResult int

const (
  FOCUS_FORWARD InputFieldResult = 1
  FOCUS_BACKWARD InputFieldResult = -1
  UPDATE InputFieldResult = 0
  CONFIRM InputFieldResult = 2
  CANCEL InputFieldResult = 3
)

func (field *InputField) Draw() {
  window := field.Parent
  runes := field.Value

  // Draw current value.
  if len(runes) > 0 {
    window.WithColor(tui.COLOR_4BIT_BLACK, tui.COLOR_4BIT_CYAN, func() {
      start, end := field.Offset, utils.MinInt(field.Offset + field.Length, len(field.Value))
      window.MovePrint(field.Y, field.X, string(runes[start:end]))
    })
  }

  // Draw suggestion.
  visible := len(field.Value) - field.Offset
  if field.IsActive && field.Length > visible && field.Suggestion != nil {
    suggestionRunes := []rune(*field.Suggestion)
    tail := suggestionRunes[len(field.Value):utils.MinInt(len(suggestionRunes), len(field.Value)+(field.Length-visible))]
    window.WithColor(tui.COLOR_4BIT_WHITE, tui.COLOR_4BIT_CYAN, func() {
      window.MovePrint(field.Y, field.X + visible, string(tail))
    })
    visible += len(tail)
  }

  // Clear the rest of the field.
  if field.Length > visible {
    window.WithColor(tui.COLOR_4BIT_BLACK, tui.COLOR_4BIT_CYAN, func() {
      window.MovePrint(field.Y, field.X + visible, strings.Repeat(" ", field.Length - visible))
    })
  }
}

func (field *InputField) OnInput(c tui.Key) {
  // Focus helper.
  moveFocus := func(result InputFieldResult) {
    field.IsActive = false
    field.OnResult(field, result)
  }

  if c.Rune != nil {
    field.NewChar(*c.Rune)
    field.OnResult(field, UPDATE)
  } else if c.EscapeSeq != nil {
    switch *c.EscapeSeq {
    case tui.ESC_DOWN:
      if field.IsModal { return }
      moveFocus(FOCUS_FORWARD)
    case tui.ESC_UP:
      if field.IsModal { return }
      moveFocus(FOCUS_BACKWARD)
    case tui.ESC_LEFT:
      field.Cursor = utils.MaxInt(0, field.Cursor - 1)
      if field.Offset > field.Cursor {
        field.Offset -= 1
      }
      field.OnResult(field, UPDATE)
    case tui.ESC_RIGHT:
      field.Cursor = utils.MinInt(len(field.Value), field.Cursor + 1)
      if (field.Offset + field.Length) < field.Cursor {
        field.Offset += 1
      }
      field.OnResult(field, UPDATE)
    case tui.ESC_HOME, tui.ESC_PGUP:
      field.Offset = 0
      field.Cursor = 0
      field.OnResult(field, UPDATE)
    case tui.ESC_END, tui.ESC_PGDOWN:
      field.Offset = utils.MaxInt(0, len(field.Value) - field.Length + 1)
      field.Cursor = len(field.Value)
      field.OnResult(field, UPDATE)
    }
  } else if c.ControlCode != 0 {
    switch c.ControlCode {
    case tui.ASC_ESC:
      if field.IsModal {
        field.OnResult(field, CANCEL)
      }
    case tui.ASC_TAB:
      if field.Suggestion != nil && string(field.Value) != (*field.Suggestion) {
        field.Value = []rune(*field.Suggestion)
        field.Offset = utils.MaxInt(0, len(field.Value) - field.Length + 1)
        field.Cursor = len(field.Value)
        field.UpdateSuggestion()
        field.OnResult(field, UPDATE)
      } else {
        moveFocus(FOCUS_FORWARD)
      }
    case tui.ASC_BACKSPACE, tui.ASC_DELETE:
      field.ControlCode(c.ControlCode)
    case tui.ASC_ENTER:
      field.IsActive = false
      if field.IsModal {
        field.OnResult(field, CONFIRM)
      } else {
        field.OnResult(field, FOCUS_FORWARD)
      }
    }
  }
}

func (field *InputField) NewChar(c rune) {
  var trimmed []rune
  var runes []rune = field.Value

  if (field.Limit > 0) && (len(field.Value) >= field.Limit) {
    return
  }

  if len(field.Charset) > 0 && !strings.Contains(field.Charset, fmt.Sprintf("%c", c)) {
    return
  }

  prefix, suffix := runes[:field.Cursor], runes[field.Cursor:]
  output := fmt.Sprintf("%s%c%s", string(prefix), c, string(suffix))

  trimmed = []rune(output)
  field.Cursor += 1

  field.Value = trimmed
  field.Cursor = utils.MaxInt(utils.MinInt(len(trimmed), field.Cursor), 0)
  if field.Cursor == len(trimmed) {
    field.Offset = utils.MaxInt(0, len(field.Value) - field.Length + 1)
  }

  field.UpdateSuggestion()
}

func (field *InputField) ControlCode(code int) {
  var trimmed []rune
  var runes []rune = field.Value
  var index int

  if len(field.Value) == 0 {
    return
  }

  switch code {
  case tui.ASC_BACKSPACE:
    if field.Cursor == 0 {
      return
    } else {
      index = field.Cursor - 1
    }
    trimmed = utils.Remove(runes, index)
    field.Cursor -= 1
  case tui.ASC_DELETE:
    if field.Cursor == len(runes) {
      index = field.Cursor - 1
    } else {
      index = field.Cursor
    }
    trimmed = utils.Remove(runes, index)
  }

  field.Value = trimmed
  field.Cursor = utils.MaxInt(utils.MinInt(len(trimmed), field.Cursor), 0)
  if field.Cursor == len(trimmed) {
    field.Offset = utils.MaxInt(0, len(field.Value) - field.Length + 1)
  }

  field.UpdateSuggestion()
}

func (field *InputField) UpdateSuggestion() {
  var suggestion *string
  if (field.Suggester != nil) && len(field.Value) > 0 {
    if suggestions := field.Suggester(string(field.Value)); len(suggestions) > 0 {
      suggestion = &suggestions[0]
    }
  }
  field.Suggestion = suggestion
}

func (field *InputField) SetCursor(window tui.Drawable) {
  window.MoveTo(field.Y, field.X + (field.Cursor - field.Offset))
}
