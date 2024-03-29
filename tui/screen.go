package tui

import "fmt"
import "strings"

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

	printLine := func(line string, symbols []cell, row, start int) {
		newCell := symbols[0]

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

		MovePrintf(row, start, "%s", line)
	}

	differentParams := func(left, right cell) bool {
		return left.box != right.box ||
			!Same(left.attributes, right.attributes) ||
			!SameColor(left.color, right.color) ||
			left.wide != right.wide
	}

	// Draw.
	for i := 0; i < screen.height; i++ {
		// Container for uninterrupted sequence of symbols.
		// This slice accumulates the line of cells with the exact same params.
		lineStart := 0;
		lineEnd := 0;
		symbols := []cell{}
		line := ""

		for j := 0; j <= screen.width; j++ {
			if (j == screen.width) {
				if (len(symbols) > 0) {
					printLine(line, symbols, i, lineStart)
				}
			} else {
				newCell := screen.cells[i][j]
				bufferCell := screen.buffer[i][j]

				if newCell.symbol != bufferCell.symbol ||
						differentParams(newCell, bufferCell) ||
						(len(symbols) > 0 && differentParams(newCell, symbols[0])) {
					if len(symbols) == 0 {
						// Starting new segment.
						lineStart = j
						lineEnd = j + 1
						symbols = []cell{newCell}
						if !newCell.wide {
							line += string(newCell.symbol)
						}
					} else if differentParams(newCell, bufferCell) ||
							(differentParams(newCell, symbols[0])) {
						// Cursor mode change, print and re-initialize.
						printLine(line, symbols, i, lineStart)
						lineStart = j
						lineEnd = j + 1
						symbols = []cell{newCell}
						if !newCell.wide {
							line = string(newCell.symbol)
						} else {
							line = ""
						}
					} else {
						// Same params, but different symbol, append to rewrite buffer.
						lineEnd += 1
						symbols = append(symbols, newCell)
						if !newCell.wide {
							line += string(newCell.symbol)
						}
					}
				} else if len(symbols) > 0 {
					// Rewrite buffer is not empty, print and clear it.
					printLine(line, symbols, i, lineStart)
					symbols = []cell{}
					line = ""
				}

				screen.buffer[i][j] = newCell
			}
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

func (screen *Screen) MovePrintf(
	row, col int,
	format string,
	args ...interface{},
) {
	data := []rune(fmt.Sprintf(format, args...))
	maxCol := MinInt(screen.width - col, CellLength(data))

	columnIndex := 0
	for j := 0; j < len(data); j++ {
		screen.cells[row][col + columnIndex].wide = false
		screen.cells[row][col + columnIndex].symbol = data[j]

		// Box mode.
		screen.cells[row][col + columnIndex].box = screen.box

		// Attributes.
		screen.cells[row][col + columnIndex].attributes = make(
			[]Attribute,
			len(screen.attributes),
		)
		copy(screen.cells[row][col + columnIndex].attributes, screen.attributes)

		// Color.
		screen.cells[row][col + columnIndex].color = screen.color

		// Advance to the next column.
		columnIndex = columnIndex + 1

		// If symbol is a wide character, advance column index once more and mark
		// next column as 'wide'.
		// 'Wide' columns are skipped when redrawing the buffer.
		// Currently we're only checking for ideographs and Japanese alphabet.
		// 3040-309F: hiragana, 30A0-30FF: katakana
		if IsWide(data[j]) {
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
			screen.cells[i][j].wide = false
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
