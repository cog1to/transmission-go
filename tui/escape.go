package tui

type EscapeSequence string

const(
  ESC_CLEAR_SCREEN EscapeSequence = "\033c"
  ESC_MOVE_TO = "\033[%d;%dH"
  ESC_CLEAR_COLOR = "\033[39m\033[49m"
  ESC_BEGIN_BOXDRAW = "\033(0"
  ESC_END_BOXDRAW = "\033(B"
  ESC_HIDE_CURSOR = "\033[?25l"
  ESC_SHOW_CURSOR = "\033[?25h"
)

type Attribute string

const(
  ATTR_BOLD Attribute = "1"
  ATTR_REVERSED       = "7"
  ATTR_NORMAL         = "27"
)

type Color string

const(
  COLOR_4BIT_BLACK Color = "0"
  COLOR_4BIT_RED         = "1"
  COLOR_4BIT_GREEN       = "2"
  COLOR_4BIT_YELLOW      = "3"
  COLOR_4BIT_BLUE        = "4"
  COLOR_4BIT_MAGENTA     = "5"
  COLOR_4BIT_CYAN        = "6"
  COLOR_4BIT_WHITE       = "7"
)
