package windows

import (
  tui "../tui"
  "../utils"
)

/* Data */

type HelpItem struct {
  Input string
  Description string
}

/* Window */

type Cheatsheet struct {
  parent *tui.Window
  window *tui.Window
  items []HelpItem
  manager *WindowManager
}

func (window *Cheatsheet) IsFullScreen() bool {
  return false
}

func (window *Cheatsheet) SetActive(active bool) { }

func (window *Cheatsheet) Draw() {
  drawCheatsheet(window.window, window.items)
}

func (window *Cheatsheet) Resize() {
  height, width, y, x := measureCheatsheet(window.parent, window.items)

  // Spawn new sub-window.
  window.window.Move(y, x)
  window.window.Resize(height, width)
}

func (window *Cheatsheet) OnInput(key tui.Key) {
  window.manager.RemoveWindow(window)
}

func NewCheatsheet(parent *tui.Window, items []HelpItem, manager *WindowManager) *Cheatsheet {
  height, width, y, x := measureCheatsheet(parent, items)
  window := parent.Sub(height, width, y, x)

  return &Cheatsheet{
    parent,
    window,
    items,
    manager}
}

/* Drawing */

func drawCheatsheet(window *tui.Window, items []HelpItem) {
  window.Box()

  _, col := window.MaxYX()
  startX, width := 2, col-4

  maxLeft := func() (int) {
    var left int
    for _, item := range items {
      left = utils.MinInt(10, utils.MaxInt(left, len([]rune(item.Input))))
    }
    return left
  }()

  x, y := startX, 1
  for _, item := range items {
    // Input
    index, inputY, input := 0, 0, []rune(item.Input)
    for index < len(input) {
      line := string(input[index:utils.MinInt(index + maxLeft, len(input))])
      window.MovePrint(y + inputY, x, line)
      index += len(input[index:utils.MinInt(index + maxLeft, len(input))])
      inputY += 1
    }

    // Transition
    tui.WithAttribute(tui.ATTR_BOLD, func() {
      window.MovePrint(y, startX + maxLeft, " :: ")
    })

    // Meaning
    index, textY, textX, textWidth := 0, 0, x + maxLeft + 4, width - maxLeft - 4
    for index < len(item.Description) {
      line := item.Description[index:utils.MinInt(index + textWidth, len(item.Description))]
      window.MovePrint(y + textY, textX, string(line))
      index += len(line)
      textY += 1
    }

    // Advance to next item
    y += utils.MaxInt(inputY, textY)
  }
}

/* Measuring */

func measureCheatsheet(parent *tui.Window, items []HelpItem) (int, int, int, int) {
  const (
    OUTER_PADDING_X, OUTER_PADDING_Y = 3, 5
    INNER_PADDING = 4
    MARGINS_VERTICAL = 2
    DELIMITER_WIDTH = len(" :: ")
  )

  // Measuring window's width and height
  maxLeft := func() (int) {
    var left int
    for _, item := range items {
      left = utils.MinInt(10, utils.MaxInt(left, len([]rune(item.Input))))
    }
    return left
  }()

  maxWidth := func() int {
    width := 0
    for _, item := range items {
      width = utils.MaxInt(len([]rune(item.Description)), width)
    }
    return width
  }()

  rows, cols := parent.MaxYX()
  maxWindowWidth := cols - OUTER_PADDING_X
  windowWidth := utils.MinInt(maxWindowWidth, maxWidth + INNER_PADDING + DELIMITER_WIDTH + maxLeft)

  maxHeight := func() int {
    y, maxRight := 0, windowWidth - INNER_PADDING + DELIMITER_WIDTH - maxLeft
    for _, item := range items {
      y += utils.MaxInt(1, len([]rune(item.Description)) / maxRight)
    }
    return y
  }()
  windowHeight := utils.MinInt(rows - OUTER_PADDING_Y, maxHeight + MARGINS_VERTICAL)

  return windowHeight, windowWidth, (rows - windowHeight) / 2, (cols - windowWidth) / 2
}
