package windows

import (
  "fmt"
  gc "../goncurses"
  "../utils"
  "../logger"
)

/* Data */

type HelpItem struct {
  Input string
  Description string
}

/* Window */

type Cheatsheet struct {
  parent *gc.Window
  window *gc.Window
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
  const (
    OUTER_PADDING_X, OUTER_PADDING_Y = 3, 5
    INNER_PADDING = 4
    MARGINS_VERTICAL = 2
    DELIMITER_WIDTH = len(" :: ")
  )

  // Measuring window's width and height
  maxLeft := func() (int) {
    var left int
    for _, item := range window.items {
      left = utils.MinInt(10, utils.MaxInt(left, len([]rune(item.Input))))
    }
    return left
  }()

  maxWidth := func() int {
    width := 0
    for _, item := range window.items {
      width = utils.MaxInt(len([]rune(item.Description)), width)
    }
    return width
  }()

  rows, cols := window.parent.MaxYX()
  maxWindowWidth := cols - OUTER_PADDING_X
  windowWidth := utils.MinInt(maxWindowWidth, maxWidth + INNER_PADDING + DELIMITER_WIDTH + maxLeft)

  maxHeight := func() int {
    y, maxRight := 0, windowWidth - INNER_PADDING + DELIMITER_WIDTH - maxLeft
    for _, item := range window.items {
      y += utils.MaxInt(1, len([]rune(item.Description)) / maxRight)
    }
    return y
  }()
  windowHeight := utils.MinInt(rows - OUTER_PADDING_Y, maxHeight + MARGINS_VERTICAL)

  // Spawn new sub-window.
  logger.Log(fmt.Sprintf("New pos: %d, %d. New size: %d, %d", (rows - windowHeight) / 2, (cols - windowWidth) / 2, windowHeight, windowWidth))
  window.window.MoveWindow((rows - windowHeight) / 2, (cols - windowWidth) / 2)
  window.window.Resize(windowHeight, windowWidth)
}

func (window *Cheatsheet) OnInput(key gc.Key) {
  window.manager.RemoveWindow(window)
}

func NewCheatsheet(parent *gc.Window, items []HelpItem, manager *WindowManager) *Cheatsheet {
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

  // Spawn new sub-window.
  window := parent.Derived(windowHeight, windowWidth, (rows - windowHeight) / 2, (cols - windowWidth) / 2)

  return &Cheatsheet{
    parent,
    window,
    items,
    manager}
}

/* Drawing */

func drawCheatsheet(window *gc.Window, items []HelpItem) {
  window.Clear()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

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
    utils.WithAttribute(window, gc.A_BOLD, func(window *gc.Window) {
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

  window.Refresh()
}
