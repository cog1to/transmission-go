package utils

import (
  gc "../goncurses"
)

func WithAttribute(window *gc.Window, attr gc.Char, block func(*gc.Window)) {
  window.AttrOn(attr)
  block(window)
  window.AttrOff(attr)
}

func WithColor(window *gc.Window, color int16, block func(*gc.Window)) {
  window.ColorOn(color)
  block(window)
  window.ColorOff(color)
}