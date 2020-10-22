package tui

import "fmt"
import "strings"
import "unicode"

type colorPair struct {
  front, back Color
}

type cell struct {
  symbol rune
  box bool
  attributes []Attribute
  wide bool
  color *colorPair
}

type Screen struct {
  // Data
  cells [][]cell
  buffer [][]cell

  // Current params
  attributes []Attribute
  box bool
  color *colorPair

  // Inherit Window stuff
  Window
}

func (screen *Screen) Redraw() {
  // Current attributes.
  attributes := make([]Attribute, 0)
  box := false
  var color *colorPair

  // Draw.
  for i := 0; i < screen.height; i++ {
    for j := 0; j < screen.width; j++ {

      newCell := screen.cells[i][j]
      bufferCell := screen.buffer[i][j]

      if newCell.symbol != bufferCell.symbol ||
          newCell.box != bufferCell.box ||
          newCell.wide != bufferCell.wide  ||
          !Same(newCell.attributes, bufferCell.attributes) ||
          !SameColor(newCell.color, bufferCell.color) {
        if newCell.box != box {
          box = newCell.box
          if newCell.box {
            fmt.Printf(ESC_BEGIN_BOXDRAW)
          } else {
            fmt.Printf(ESC_END_BOXDRAW)
          }
        }

        for _, attr := range(newCell.attributes) {
          if !Contains(attributes, attr) {
            AttributeOn(attr)
            attributes = append(attributes, attr)
          }
        }

        for _, attr := range(attributes) {
          if !Contains(newCell.attributes, attr) {
            AttributeOff(attr)
            attributes = Remove(attributes, attr)
          }
        }

        if !SameColor(newCell.color, color) {
          if newCell.color != nil {
            color = newCell.color
            ColorOn(color.front, color.back)
          } else {
            color = nil
            ColorOff()
          }
        }

        if !newCell.wide {
          MovePrintf(i, j, "%c", newCell.symbol)
        }
      }

      screen.buffer[i][j] = newCell
    }
  }

  // Reset all settings.
  for _, attr := range(attributes) {
    AttributeOff(attr)
  }
  fmt.Printf(ESC_END_BOXDRAW)
  ColorOff()
}

func (screen *Screen) Sub(row, col, height, width int) Drawable {
  window := &Window{
    row: row, col: col, width: width, height: height, parent: screen}
  return window
}

func (screen *Screen) SetHeight(height int) {
  screen.height = height
}

func (screen *Screen) SetWidth(width int) {
  screen.width = width
}

func (screen *Screen) MovePrint(row, col int, str string) {
  screen.MovePrintf(row, col, "%s", str)
}

func (screen *Screen) Refresh() {
  size := Termsize()

  cells := make([][]cell, size.Rows)
  for i := range cells {
    cells[i] = make([]cell, size.Cols)
  }

  buffer := make([][]cell, size.Rows)
  for i := range buffer {
    buffer[i] = make([]cell, size.Cols)
  }

  screen.buffer = buffer
  screen.cells = cells
  screen.SetHeight(size.Rows)
  screen.SetWidth(size.Cols)
}

func (screen *Screen) MovePrintf(row, col int, format string, args ...interface{}) {
  data := []rune(fmt.Sprintf(format, args...))
  maxCol := MinInt(screen.width - col, len(data))

  columnIndex := 0
  for j := 0; j < len(data); j++ {
    screen.cells[row][col + columnIndex].wide = false
    screen.cells[row][col + columnIndex].symbol = data[j]

    // Box mode.
    screen.cells[row][col + columnIndex].box = screen.box

    // Attributes.
    screen.cells[row][col + columnIndex].attributes = make([]Attribute, len(screen.attributes))
    copy(screen.cells[row][col + columnIndex].attributes, screen.attributes)

    // Color.
    screen.cells[row][col + columnIndex].color = screen.color

    // Advance to the next column.
    columnIndex = columnIndex + 1

    // If symbol is a wide character, advance column index once more and mark next column as 'wide'.
    // 'Wide' columns are skipped when redrawing the buffer.
    // Currently we're only checking for ideographs and Japanese alphabet.
    // 3040-309F: hiragana, 30A0-30FF: katakana
    if ((data[j] >= 0x3040 && data[j] <= 0x30FF) || unicode.Is(unicode.Unified_Ideograph, data[j])) {
      screen.cells[row][min(col + columnIndex, maxCol - 1)].wide = true
      screen.cells[row][min(col + columnIndex, maxCol - 1)].symbol = ' '
      columnIndex = columnIndex + 1
    } else {
      screen.cells[row][min(col + columnIndex, maxCol - 1)].wide = false
    }

    if columnIndex >= maxCol {
      break
    }
  }
}

func (screen *Screen) HLine(row, col, width int) {
  screen.box = true
  screen.MovePrint(row, col, strings.Repeat(ASC_HLINE, width))
  screen.box = false
}

func (screen *Screen) Line(row, col int, symbol rune, width int) {
  screen.MovePrint(row, col, strings.Repeat(string(symbol), width))
}

func (screen *Screen) LineBox(row, col, height, width int) {
  lines := func(startx, starty, count int) {
    text := ASC_VLINE + strings.Repeat(" ", width - 1) + ASC_VLINE
    for line := 0; line < count; line++ {
      screen.MovePrint(startx + line, starty, text)
    }
  }

  screen.box = true

  // Top line.
  screen.MovePrint(row, col, ASC_TOPLEFT_CORNER)
  screen.MovePrint(row, col + 1, strings.Repeat(ASC_HLINE, width - 1))
  screen.MovePrint(row, col + width, ASC_TOPRIGHT_CORNER)

  // Lines row + 1 to row + height - 1.
  lines(row + 1, col, height - 2)

  // Bottom line.
  screen.MovePrint(row + height - 1, col, ASC_BOTTOMLEFT_CORNER)
  screen.MovePrint(row + height - 1, col + 1, strings.Repeat(ASC_HLINE, width - 1))
  screen.MovePrint(row + height - 1, col + width, ASC_BOTTOMRIGHT_CORNER)

  screen.box = false
}


func (screen *Screen) Erase() {
  screen.ClearBox(screen.row, screen.col, screen.height, screen.width)
}

func (screen *Screen) ClearBox(row, col, height, width int) {
  screen.box = false
  for i := row; i < height; i++ {
    for j := col; j < width; j++ {
      screen.cells[i][j].symbol = rune(' ')
      screen.cells[i][j].box = false
      screen.cells[i][j].attributes = []Attribute{}
      screen.cells[i][j].color = nil
    }
  }
}

func (screen *Screen) WithAttributes(attrs []Attribute, block func()()) {
  screen.attributes = attrs
  block()
  screen.attributes = make([]Attribute, 0)
}

func (screen *Screen) WithColor(front, back Color, block func()()) {
  screen.color = &colorPair{ front, back }
  block()
  screen.color = nil
}

func MinInt(x, y int) int {
  if x < y {
    return x
  } else {
    return y
  }
}
