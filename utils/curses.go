package utils

import (
  tui "../tui"
)

func WithColor(window *tui.Window, front, back tui.Color, block func(*tui.Window)) {
  tui.ColorOn(front, back)
  block(window)
  tui.ColorOff()
}

