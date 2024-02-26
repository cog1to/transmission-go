package tui

// #include <sys/ioctl.h>
import "C"
import(
	"syscall"
	"unsafe"
	"fmt"
)

type Winsize struct {
	Rows, Cols int
}

func Termsize() Winsize {
	ts := C.struct_winsize{
		ws_col: 0,
		ws_row: 0,
	}

	syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ts)),
	)
	return Winsize{ int(ts.ws_row), int(ts.ws_col) }
}

func Clear() {
	fmt.Printf(string(ESC_CLEAR_SCREEN))
}

func MoveTo(row, col int) {
	fmt.Printf(string(ESC_MOVE_TO), row + 1, col + 1)
}

func AttributeOn(attr Attribute) {
	fmt.Printf("\033[%sm", string(attr))
}

func AttributeOff(attr Attribute) {
	sequence := "2" + string(attr)
	if (attr == ATTR_BOLD) {
		sequence = "22"
	}

	fmt.Printf("\033[%sm", sequence)
}

func ColorOn(foreground, background Color) {
	fmt.Printf("\033[3%s;4%sm", string(foreground), string(background))
}

func ColorOff() {
	fmt.Printf(string(ESC_CLEAR_COLOR))
}

func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func MovePrintf(x, y int, format string, args ...interface{}) {
	MoveTo(x, y)
	fmt.Printf(format, args...)
}

func HideCursor() {
	fmt.Print(ESC_HIDE_CURSOR)
}

func ShowCursor() {
	fmt.Print(ESC_SHOW_CURSOR)
}

func WithAttribute(attr Attribute, block func()()) {
	AttributeOn(attr)
	block()
	AttributeOff(attr)
}

func WithAttributes(attrs []Attribute, block func()()) {
	for _, attr := range attrs {
		AttributeOn(attr)
	}
	block()
	for _, attr := range attrs {
		AttributeOff(attr)
	}
}

