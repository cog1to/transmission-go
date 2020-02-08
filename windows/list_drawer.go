package windows

import (
  gc "../goncurses"
  wchar "../cgo.wchar"
)

type Formatter = func(item interface{}, width int) string

type List struct {
  // Public
  Window *gc.Window
  Formatter Formatter
  MarginTop, MarginBottom, MarginLeft, MarginRight int

  // Private
  Cursor int
  Selection []int
  Offset int
  Items []interface{}
}

func (drawer *List) Draw() {
  rows, cols := drawer.Window.MaxYX()

  // Draw items.
  x, y := drawer.MarginLeft, drawer.MarginTop
  for index, item := range drawer.Items[drawer.Offset:] {
    itemString := drawer.Formatter(item, cols - drawer.MarginLeft - drawer.MarginRight)

    var attribute gc.Char
    if index + drawer.Offset == drawer.Cursor {
      attribute = gc.A_REVERSE
    } else {
      attribute = gc.A_NORMAL
    }
    if contains(drawer.Selection, index + drawer.Offset) {
      attribute = attribute | gc.A_BOLD
    }

    withAttribute(drawer.Window, attribute, func(window *gc.Window) {
      ws, convertError := wchar.FromGoString(itemString)
      if (convertError == nil) {
        window.MovePrintW(y, x, ws)
      }
    })

    y += 1
    if y >= (rows - drawer.MarginBottom) {
      break
    }
  }

  // Clear remaining lines if needed.
  for index := y; index < rows - drawer.MarginBottom; index++ {
    drawer.Window.HLine(index, drawer.MarginLeft, ' ', cols - drawer.MarginLeft - drawer.MarginRight)
  }
}

func (drawer *List) MoveCursor(direction int) {
  if (direction < 0) {
    drawer.Cursor = maxInt(0, drawer.Cursor - 1)
  } else if (direction > 0) {
    drawer.Cursor = minInt(drawer.Cursor + 1, len(drawer.Items) - 1)
  }
  drawer.UpdateOffset()
}

func (drawer *List) Page(direction int) {
  rows, _ := drawer.Window.MaxYX()
  rows = rows - drawer.MarginTop - drawer.MarginBottom

  if (direction < 0) {
    drawer.Offset = maxInt(drawer.Offset - (rows), 0)
    drawer.Cursor = maxInt(drawer.Cursor - (rows), 0)
  } else if (direction > 0) {
    drawer.Offset = minInt(drawer.Offset + rows, maxInt(len(drawer.Items) - rows, 0))
    drawer.Cursor = minInt(drawer.Cursor + rows, len(drawer.Items) - 1)
  }

  drawer.UpdateOffset()
}

func (drawer *List) Select() {
  if contains(drawer.Selection, drawer.Cursor) {
    drawer.Selection = removeInt(drawer.Selection, drawer.Cursor)
  } else {
    drawer.Selection = append(drawer.Selection, drawer.Cursor)
  }
}

func (drawer *List) ClearSelection() {
  drawer.Selection = []int{}
}

func (drawer *List) SetItems(items []interface{}) {
  drawer.Items = items
  drawer.Cursor = minInt(len(drawer.Items) - 1, drawer.Cursor)
  drawer.UpdateOffset()
}

func (drawer *List) UpdateOffset() {
  rows, _ := drawer.Window.MaxYX()

  if drawer.Cursor > (rows - drawer.MarginTop - drawer.MarginBottom) + drawer.Offset - 1 {
    drawer.Offset = minInt(drawer.Offset + 1, len(drawer.Items) - (rows - drawer.MarginTop - drawer.MarginBottom))
  } else if drawer.Cursor < drawer.Offset {
    drawer.Offset = maxInt(drawer.Offset - 1, 0)
  }
}

func (drawer *List) GetSelection() []interface{} {
  if len(drawer.Selection) > 0 {
    items := make([]interface{}, len(drawer.Selection))
    for i, index := range drawer.Selection {
      items[i] = drawer.Items[index]
    }
    return items
  }

  return drawer.Items[drawer.Cursor:drawer.Cursor+1]
}
