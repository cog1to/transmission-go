package utils

import (
  "tui"
)

func WithColor(window tui.Drawable, front, back tui.Color, block func(tui.Drawable)) {
  tui.ColorOn(front, back)
  block(window)
  tui.ColorOff()
}

func RuneSlice(data []rune, offset, cellLength int) (int, int) {
  end, cellOffset := offset, 0
  for j := offset; j < len(data); j++ {
    runeWidth := 1
    if tui.IsWide(data[j]) {
      runeWidth = 2
    }

    cellOffset += runeWidth
    if cellOffset > cellLength {
      break
    }

    end += 1
  }
  return offset, end
}

func OffsetToFit(data []rune, cellLength int) int {
  end, runeWidth, cellWidth := len(data) - 1, 1, 0
  for ; end >= 0; end-- {
    runeWidth = 1
    if tui.IsWide(data[end]) {
      runeWidth = 2
    }

    cellWidth += runeWidth
    if cellWidth > cellLength {
      break
    }
  }

  return end
}

func CellsTo(data []rune, from, to int) int {
  count := 0
  for j := from; j < len(data) && j < to; j++ {
    if tui.IsWide(data[j]) {
      count += 2
    } else {
      count += 1
    }
  }
  return count
}
