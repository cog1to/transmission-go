package tui

// #include <termios.h>
// #include <stdio.h>
// #include <unistd.h>
import "C"

var termios *C.struct_termios

func SetRaw(enabled bool) {
	if (termios == nil) {
		termios = &C.struct_termios{}
		C.tcgetattr(C.STDIN_FILENO, termios)
	}

	if (enabled) {
		copy := *termios
		copy.c_lflag = copy.c_lflag & C.OPOST
		copy.c_lflag = copy.c_lflag &^ C.ECHO
		C.tcsetattr(C.STDIN_FILENO, C.TCSAFLUSH, &copy)
	} else {
		C.tcsetattr(C.STDIN_FILENO, C.TCSAFLUSH, termios)
	}
}

