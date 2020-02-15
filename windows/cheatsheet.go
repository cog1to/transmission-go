package windows

import (
  gc "../goncurses"
)

/* Data */

type HelpItem struct {
  Input string
  Description string
}

/* Window */

func CheatsheetWindow(parent *gc.Window, reader *InputReader, items []HelpItem) {
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
      left = minInt(10, maxInt(left, len([]rune(item.Input))))
    }
    return left
  }()

  maxWidth := func() int {
    width := 0
    for _, item := range items {
      width = maxInt(len([]rune(item.Description)), width)
    }
    return width
  }()

  if maxWidth == 0 {
    return
  }

  rows, cols := parent.MaxYX()
  maxWindowWidth := cols - OUTER_PADDING_X
  windowWidth := minInt(maxWindowWidth, maxWidth + INNER_PADDING + DELIMITER_WIDTH + maxLeft)

  maxHeight := func() int {
    y, maxRight := 0, windowWidth - INNER_PADDING + DELIMITER_WIDTH - maxLeft
    for _, item := range items {
      y += maxInt(1, len([]rune(item.Description)) / maxRight)
    }
    return y
  }()
  windowHeight := minInt(rows - OUTER_PADDING_Y, maxHeight + MARGINS_VERTICAL)

  // Spawn new sub-window.
  window := parent.Derived(windowHeight, windowWidth, (rows - windowHeight) / 2, (cols - windowWidth) / 2)
  defer window.Delete()

  // Handle user input.
  observer := make(chan gc.Key)
  reader.AddObserver(observer)
  defer reader.RemoveObserver(observer)

  drawHelpWindow(window, items)
  <-observer
}

func drawHelpWindow(window *gc.Window, items []HelpItem) {
  window.Clear()
  window.Box(gc.ACS_VLINE, gc.ACS_HLINE)

  _, col := window.MaxYX()
  startX, width := 2, col-4

  maxLeft := func() (int) {
    var left int
    for _, item := range items {
      left = minInt(10, maxInt(left, len([]rune(item.Input))))
    }
    return left
  }()

  x, y := startX, 1
  for _, item := range items {
    // Input
    index, inputY, input := 0, 0, []rune(item.Input)
    for index < len(input) {
      line := string(input[index:minInt(index + maxLeft, len(input))])
      window.MovePrint(y + inputY, x, line)
      index += len(input[index:minInt(index + maxLeft, len(input))])
      inputY += 1
    }

    // Transition
    withAttribute(window, gc.A_BOLD, func(window *gc.Window) {
      window.MovePrint(y, startX + maxLeft, " :: ")
    })

    // Meaning
    index, textY, textX, textWidth := 0, 0, x + maxLeft + 4, width - maxLeft - 4
    for index < len(item.Description) {
      line := item.Description[index:minInt(index + textWidth, len(item.Description))]
      window.MovePrint(y + textY, textX, string(line))
      index += len(line)
      textY += 1
    }

    // Advance to next item
    y += maxInt(inputY, textY)
  }

  window.Refresh()
}
