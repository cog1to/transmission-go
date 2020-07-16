package tui

// #include <termios.h>
// #include <stdio.h>
// #include <unistd.h>
import "C"

func SetRaw(enabled bool) {
  SetParam(C.OPOST, enabled)
}

func SetEcho(enabled bool) {
  SetParam(C.ECHO, enabled)
}

func SetParam(param C.uint, enabled bool) {
  termios := C.struct_termios{}

  C.tcgetattr(C.STDIN_FILENO, &termios)
  if (enabled) {
    termios.c_lflag = termios.c_lflag & param
  } else {
    termios.c_lflag = termios.c_lflag &^ param
  }
  C.tcsetattr(C.STDIN_FILENO, C.TCSAFLUSH, &termios)
}
