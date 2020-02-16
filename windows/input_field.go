package windows

import (
  gc "../goncurses"
  "fmt"
  "strings"
)

type InputField struct {
  X, Y, Length int
  Offset int
  Cursor int
  IsActive bool
  Value []rune
  Limit int
  Charset string
}

type InputFieldResult int

const (
  FOCUS_FORWARD InputFieldResult = 1
  FOCUS_BACKWARD InputFieldResult = -1
  UPDATE InputFieldResult = 0
  CONFIRM InputFieldResult = 2
  CANCEL InputFieldResult = 3
)

func (field *InputField) Activate(reader *InputReader, input chan InputFieldResult, isModal bool) {
  field.IsActive = true

  // Input listener.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)
  defer reader.RemoveObserver(observer)

  Loop: for {
    c := <-observer
    switch c {
    case '\n':
      field.IsActive = false
      if isModal {
        input <- CONFIRM
        break Loop
      } else {
        input <- FOCUS_FORWARD
      }
    case gc.KEY_ESC:
      if isModal {
        input <- CANCEL
        break Loop
      }
    case gc.KEY_DOWN, gc.KEY_TAB:
      if isModal { break }
      field.IsActive = false
      input <- FOCUS_FORWARD
      break Loop
    case gc.KEY_UP:
      if isModal { break }
      field.IsActive = false
      input <- FOCUS_BACKWARD
      break Loop
    case gc.KEY_LEFT:
      field.Cursor = maxInt(0, field.Cursor - 1)
      if field.Offset > field.Cursor {
        field.Offset -= 1
      }
      input <- UPDATE
    case gc.KEY_RIGHT:
      field.Cursor = minInt(len(field.Value), field.Cursor + 1)
      if (field.Offset + field.Length) < field.Cursor {
        field.Offset += 1
      }
      input <- UPDATE
    case gc.KEY_HOME, gc.KEY_PAGEUP:
      field.Offset = 0
      field.Cursor = 0
      input <- UPDATE
    case gc.KEY_END, gc.KEY_PAGEDOWN:
      field.Offset = maxInt(0, len(field.Value) - field.Length + 1)
      field.Cursor = len(field.Value)
      input <- UPDATE
    case gc.KEY_F1, gc.KEY_F2, gc.KEY_F3, gc.KEY_F4, gc.KEY_F5, gc.KEY_F6, gc.KEY_F7, gc.KEY_F8, gc.KEY_F9, gc.KEY_F10, gc.KEY_F11, gc.KEY_F12:
      break
    default:
      field.NewChar(c)
      input <- UPDATE
    }
  }
}

func (field *InputField) Draw(window *gc.Window) {
  withColor(window, 1, func(window *gc.Window) {
    runes := field.Value

    if len(runes) > 0 {
      start, end := field.Offset, minInt(field.Offset + field.Length, len(field.Value))
      window.MovePrint(field.Y, field.X, string(runes[start:end]))
    }

    visible := len(field.Value) - field.Offset
    if field.Length > visible {
      window.MovePrint(field.Y, field.X + visible, strings.Repeat(" ", field.Length - visible))
    }
  })
}

func (field *InputField) NewChar(c gc.Key) {
  var trimmed []rune
  var runes []rune = field.Value
  var index int

  if len(runes) == 0 && (c == gc.KEY_BACKSPACE || c == gc.KEY_DC) {
    return
  }

  switch c {
  case gc.KEY_BACKSPACE, 0x7f:
    if field.Cursor == 0 {
      return
    } else {
      index = field.Cursor - 1
    }
    trimmed = remove(runes, index)
    field.Cursor -= 1
  case gc.KEY_DC:
    if field.Cursor == len(runes) {
      index = field.Cursor - 1
    } else {
      index = field.Cursor
    }
    trimmed = remove(runes, index)
  default:
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
  }

  field.Value = trimmed
  field.Cursor = maxInt(minInt(len(trimmed), field.Cursor), 0)
  if field.Cursor == len(trimmed) {
    field.Offset = maxInt(0, len(field.Value) - field.Length + 1)
  }
}

func (field *InputField) SetCursor(window *gc.Window) {
  window.Move(field.Y, field.X + (field.Cursor - field.Offset))
}
