package windows

import (
  gc "github.com/rthornton128/goncurses"
)

func minInt(x, y int) int {
  if x < y {
    return x
  } else {
    return y
  }
}

func maxInt(x, y int) int {
  if x > y {
    return x
  } else {
    return y
  }
}

func withAttribute(window *gc.Window, attr gc.Char, block func(*gc.Window)) {
  window.AttrOn(attr)
  block(window)
  window.AttrOff(attr)
}
