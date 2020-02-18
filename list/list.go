package list

import (
  gc "../goncurses"
  wchar "../cgo.wchar"
  "../utils"
)

/* Common item interface */

type Identifiable interface {
  Id() int
}

/* List */

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
  Items []Identifiable
}

func (drawer *List) Draw() {
  if len(drawer.Items) == 0 {
    drawer.Window.Clear()
    drawer.Window.Refresh()
    return
  }

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
    if utils.Contains(drawer.Selection, drawer.Items[index + drawer.Offset].Id()) {
      attribute = attribute | gc.A_BOLD
    }

    utils.WithAttribute(drawer.Window, attribute, func(window *gc.Window) {
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
    drawer.Cursor = utils.MaxInt(0, drawer.Cursor - 1)
  } else if (direction > 0) {
    drawer.Cursor = utils.MinInt(drawer.Cursor + 1, len(drawer.Items) - 1)
  }
  drawer.UpdateOffset()
}

func (drawer *List) Page(direction int) {
  rows, _ := drawer.Window.MaxYX()
  rows = rows - drawer.MarginTop - drawer.MarginBottom

  if (direction < 0) {
    drawer.Offset = utils.MaxInt(drawer.Offset - (rows), 0)
    drawer.Cursor = utils.MaxInt(drawer.Cursor - (rows), 0)
  } else if (direction > 0) {
    drawer.Offset = utils.MinInt(drawer.Offset + rows, utils.MaxInt(len(drawer.Items) - rows, 0))
    drawer.Cursor = utils.MinInt(drawer.Cursor + rows, len(drawer.Items) - 1)
  }

  drawer.UpdateOffset()
}

func (drawer *List) Select() {
  id := drawer.Items[drawer.Cursor].Id()
  if utils.Contains(drawer.Selection, id) {
    drawer.Selection = utils.RemoveInt(drawer.Selection, id)
  } else {
    drawer.Selection = append(drawer.Selection, id)
  }
}

func (drawer *List) ClearSelection() {
  drawer.Selection = []int{}
}

func (drawer *List) SetItems(items []Identifiable) {
  drawer.Items = items
  drawer.UpdateSelection()
  drawer.Cursor = utils.MinInt(len(drawer.Items) - 1, drawer.Cursor)
  drawer.UpdateOffset()
}

func (drawer *List) UpdateSelection() {
  if len(drawer.Selection) == 0 {
    return
  }

  newSelection := make([]int, 0, len(drawer.Selection))
  for _, id := range drawer.Selection {
    if item(drawer.Items, id) != nil {
      newSelection = append(newSelection, id)
    }
  }

  drawer.Selection = newSelection
}

func (drawer *List) UpdateOffset() {
  rows, _ := drawer.Window.MaxYX()

  if drawer.Cursor > (rows - drawer.MarginTop - drawer.MarginBottom) + drawer.Offset - 1 {
    drawer.Offset = utils.MinInt(drawer.Offset + 1, len(drawer.Items) - (rows - drawer.MarginTop - drawer.MarginBottom))
  } else if drawer.Cursor < drawer.Offset {
    drawer.Offset = utils.MaxInt(drawer.Offset - 1, 0)
  }
}

func (drawer *List) GetSelection() []Identifiable {
  if len(drawer.Selection) > 0 {
    items := make([]Identifiable, 0, len(drawer.Selection))
    for _, id := range drawer.Selection {
      if existingItem := item(drawer.Items, id); existingItem != nil {
        items = append(items, *existingItem)
      }
    }
    return items
  }

  return drawer.Items[drawer.Cursor:drawer.Cursor+1]
}

func item(items []Identifiable, id int) *Identifiable {
  for _, item := range items {
    if item.Id() == id {
      return &item
    }
  }

  return nil
}
