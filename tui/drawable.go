package tui

type Drawable interface {
  // Structure.
  Parent() Drawable
  Row() int
  Col() int
  Width() int
  Height() int
  SetWidth(int)
  SetHeight(int)
  MaxYX() (int, int)
  Sub(y, x, height, width int) Drawable
  absCoordinates(int, int) (int, int)

  // Drawing.
  MoveTo(row, col int)
  MovePrint(row, col int, str string)
  MovePrintf(row, col int, format string, args ...interface{})
  Clear()
  HLine(row, col, width int)
  Line(row, col int, symbol rune, width int)
  LineBox(row, col, height, width int)
  Erase()
  ClearBox(row, col, height, width int)
  Box()
  WithAttributes([]Attribute, func()())
  WithAttribute(Attribute, func()())
  WithColor(Color, Color, func()())

  // Movement.
  Move(row, col int)
  Resize(height, width int)
  Refresh()
  Redraw()
}
