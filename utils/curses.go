package utils

import (
  "tui"
)

func WithColor(window tui.Drawable, front, back tui.Color, block func(tui.Drawable)) {
  tui.ColorOn(front, back)
  block(window)
  tui.ColorOff()
}

