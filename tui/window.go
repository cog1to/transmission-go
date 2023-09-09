package tui

import "strings"

type Window struct {
	row, col int
	width, height int
	parent Drawable
}

var screen *Screen;

func Init() Drawable {
	size := Termsize()

	window := Window{
		row: 0, col: 0, width: size.Cols, height: size.Rows}

	cells := make([][]cell, size.Rows)
	for i := range cells {
		cells[i] = make([]cell, size.Cols)
	}

	buffer := make([][]cell, size.Rows)
	for i := range buffer {
		buffer[i] = make([]cell, size.Cols)
	}

	screen = &Screen{
		cells, buffer, make([]Attribute, 0), false, nil, window}

	return screen
}

func (win *Window) Sub(row, col, height, width int) Drawable {
	window := &Window{
		row: row, col: col, width: width, height: height, parent: win}
	return window
}

func (win *Window) WithAttributes(attrs []Attribute, block func()()) {
	if win.parent != nil {
		win.parent.WithAttributes(attrs, block)
	}
}

func (win *Window) WithColor(front, back Color, block func()()) {
	if win.parent != nil {
		win.parent.WithColor(front, back, block)
	}
}

func (win *Window) WithAttribute(attr Attribute, block func()()) {
	if win.parent != nil {
		attributes := []Attribute{attr}
		win.parent.WithAttributes(attributes, block)
	}
}

func (win *Window) Destroy() {
	win.parent = nil
}

func (win *Window) absCoordinates(row, col int) (int, int) {
	if win.parent == nil {
		return row, col
	} else {
		return win.parent.absCoordinates(win.row + row, win.col + col)
	}
}

func (win *Window) parentCoordinates(row, col int) (int, int) {
	if win.parent == nil {
		return row, col
	} else {
		return win.row + row, win.col + col
	}
}

func (win *Window) Row() int {
	return win.row
}

func (win *Window) Col() int {
	return win.col
}

func (win *Window) Height() int {
	return win.height
}

func (win *Window) SetHeight(height int) {
	win.height = height
}

func (win *Window) Width() int {
	return win.width
}

func (win *Window) SetWidth(width int) {
	win.width = width
}

func (win *Window) Parent() Drawable {
	return win.parent
}

func (win *Window) MoveTo(row, col int) {
	absRow, absCol := win.absCoordinates(row, col)
	MoveTo(absRow, absCol)
}

func (win *Window) MovePrint(row, col int, str string) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.MovePrintf(absRow, absCol, "%s", str)
	}
}

func (win *Window) MovePrintf(row, col int, format string, args ...interface{}) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.MovePrintf(absRow, absCol, format, args...)
	}
}

func (win *Window) Box() {
	win.LineBox(0, 0, win.height, win.width)
}

func (win *Window) LineBox(row, col, height, width int) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.LineBox(absRow, absCol, height, width)
	}
}

func (win *Window) Clear() {
	for idx := 0; idx < win.height; idx++ {
		win.MovePrintf(win.row + idx, win.col, strings.Repeat(" ", win.width))
	}
}

func (win *Window) Move(row, col int) {
	var size Winsize
	if (win.parent != nil) {
		size = Winsize{ Rows: win.parent.Height(), Cols: win.parent.Width() }
	} else {
		size = Termsize()
	}

	win.col = min(max(0, col), size.Cols - win.width)
	win.row = min(max(0, row), size.Rows - win.height)
}

func (win *Window) Resize(height, width int) {
	win.SetHeight(height)
	win.SetWidth(width)
}

func (win *Window) Refresh() {}

func (win *Window) Redraw() {
	if win.parent != nil {
		if scr, ok := win.parent.(*Screen); ok {
			scr.Redraw()
		} else {
			win.Parent().Redraw()
		}
	}
}

func (win *Window) MaxYX() (int, int) {
	return win.height, win.width
}

func (win *Window) HLine(row, col, width int) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.HLine(absRow, absCol, width)
	}
}

func (win *Window) Line(row, col int, symbol rune, width int) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.Line(absRow, absCol, symbol, width)
	}
}

func (win *Window) Erase() {
	win.ClearBox(win.row, win.col, win.height, win.width)
}

func (win *Window) ClearBox(row, col, height, width int) {
	if win.parent != nil {
		absRow, absCol := win.parentCoordinates(row, col)
		win.parent.ClearBox(absRow, absCol, height, width)
	}
}
