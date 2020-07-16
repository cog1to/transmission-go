package tui

import "strings"
import "fmt"

type Window struct {
  Row, Col int
  Width, Height int
  Parent *Window
}

func Init() *Window {
  size := Termsize()
  return &Window{
    Row: 0, Col: 0, Width: size.Cols, Height: size.Rows}
}

func (win *Window) Sub(row, col, height, width int) *Window {
  window := &Window{
    Row: row, Col: col, Width: width, Height: height, Parent: win}
  return window
}

func (win *Window) Destroy() {
  win.Parent = nil
}

func (win *Window) absCoordinates(row, col int) (int, int) {
  if win.Parent == nil {
    return row, col
  } else {
    return win.Parent.absCoordinates(win.Row + row, win.Col + col)
  }
}

func (win *Window) MoveTo(row, col int) {
  absRow, absCol := win.absCoordinates(row, col)
  MoveTo(absRow, absCol)
}

func (win *Window) MovePrint(row, col int, str string) {
  absRow, absCol := win.absCoordinates(row, col)
  MovePrintf(absRow, absCol, "%s", str)
}

func (win *Window) MovePrintf(row, col int, format string, args ...interface{}) {
  absRow, absCol := win.absCoordinates(row, col)
  MovePrintf(absRow, absCol, format, args...)
}

func (win *Window) Box() {
  absRow, absCol := win.absCoordinates(0, 0)
  Box(absRow, absCol, win.Height, win.Width)
}

func (win *Window) Clear() {
  for idx := 0; idx < win.Height; idx++ {
    win.MovePrintf(win.Row + idx, win.Col, strings.Repeat(" ", win.Width))
  }
}

func (win *Window) Move(row, col int) {
  var size Winsize
  if (win.Parent != nil) {
    size = Winsize{ Rows: win.Parent.Height, Cols: win.Parent.Width }
  } else {
    size = Termsize()
  }

  win.Col = min(max(0, col), size.Cols - win.Width)
  win.Row = min(max(0, row), size.Rows - win.Height)
}

func (win *Window) Resize(height, width int) {
  win.Height = height
  win.Width = width
}

func (win *Window) Refresh() {
  var size Winsize
  var row, col int

  if win.Parent == nil {
    size = Termsize()

    win.Width = size.Cols
    win.Height = size.Rows
  } else {
    size = Winsize{ Rows: win.Parent.Height, Cols: win.Parent.Width }
    row, col = win.Parent.Row, win.Parent.Col
    win.Width = min(win.Width, max(size.Cols - col, 0))
    win.Height = min(win.Height, max(size.Rows - row, 0))
  }
}

func (win *Window) MaxYX() (int, int) {
  return win.Height, win.Width
}

func (win *Window) HLine(row, col, width int) {
  fmt.Printf(
    ESC_BEGIN_BOXDRAW +
    fmt.Sprintf(string(ESC_MOVE_TO), row + 1, col + 1) +
    strings.Repeat(ASC_HLINE, width) +
    ESC_END_BOXDRAW)
}

func (win *Window) Line(row, col int, symbol rune, width int) {
  fmt.Printf(
    fmt.Sprintf(string(ESC_MOVE_TO), row + 1, col + 1) +
    strings.Repeat(string(symbol), width))
}

func (win *Window) Erase() {
  ClearBox(win.Row, win.Col, win.Height, win.Width)
}
