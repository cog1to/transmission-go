package tui

// #include <unistd.h>
import "C"
import(
  "os"
  "unicode/utf8"
)

const(
  ESC_F1 = "\033[OP"
  ESC_F1_S = "\033OP"
  ESC_HOME = "\033[1~"
  ESC_INSERT = "\033[2~"
  ESC_DELETE = "\033[3~"
  ESC_END = "\033[4~"
  ESC_PGUP = "\033[5~"
  ESC_PGDOWN = "\033[6~"
  ESC_UP = "\033[A"
  ESC_DOWN = "\033[B"
  ESC_LEFT = "\033[D"
  ESC_RIGHT = "\033[C"
  ESC_MOUSE_PREFIX = "\033[M"
)

const(
  ASC_NULL int = 0
  ASC_BEL = 7
  ASC_BACKSPACE = 8
  ASC_DELETE = 127
  ASC_ENTER = 19
  ASC_TAB = 9
  ASC_CR = 13
  ASC_ESC = 27
)

const(
  BUTTON_1_PRESS = 0
  BUTTON_2_PRESS = 1
  BUTTON_3_PRESS = 2
  BUTTON_RELEASE = 3
)

type Mouse struct {
  Button int
  Shift bool
  Meta bool
  Control bool
  X int
  Y int
}

type Key struct {
  ControlCode int
  EscapeSeq *string
  Mouse *Mouse
  Rune *rune
}

const (
  mouse_encoding_offset = 32
  mouse_button_mask = 3
  mouse_shift_mask = 4
  mouse_meta_mask = 8
  mouse_control_mask = 16
)

var Input chan Key

func StartListening() {
  Input = make(chan Key)
  const bufferLength = 6

  go func() {
    var buffer []byte = make([]byte, bufferLength)
    var index = 0

    for {
      // Read new chunk of input.
      var s []byte = make([]byte, bufferLength - index)
      length, err := os.Stdin.Read(s)
      if err != nil {
        os.Exit(1)
      }

      // Append to any existing data.
      toCopy := min(bufferLength - index, length)
      for i := 0; i < toCopy; i++ {
        buffer[index + i] = s[i]
      }
      index += toCopy

      // Iterate over buffer content and send everything.
      success := true
      for success {
        // Check for escape sequence.
        if buffer[0] == 27 && (buffer[1] == 91 || buffer[1] == 79) {
          // Mouse events.
          if buffer[2] == 'M' {
            sequence := string(buffer[:6])
            // Parse the data.
            raw_button := int(sequence[3]) - mouse_encoding_offset
            button := raw_button & mouse_button_mask
            shift := (raw_button & mouse_shift_mask) > 0
            meta := (raw_button & mouse_meta_mask) > 0
            control := (raw_button & mouse_control_mask) > 0
            x := int(sequence[4]) - mouse_encoding_offset
            y := int(sequence[5]) - mouse_encoding_offset

            // Move the buffer.
            shiftLeft(buffer, bufferLength, 6)
            index -= 6

            // Report the event.
            Input <- Key{Mouse: &Mouse{Button: button, Shift: shift, Meta: meta, Control: control, X: x, Y: y}}
          } else {
            // Find closing byte and extract everything from beginning to closing byte.
            var closeByteIndex = -1
            for i := 2; i < bufferLength; i++ {
              if buffer[i] >= 0x40 && buffer[i] <= 0x7E {
                closeByteIndex = i
                break
              }
            }

            // Terminator not found, just resetting the buffer.
            if closeByteIndex == -1 {
              for i := 0; i < bufferLength; i++ { buffer[i] = 0 }
              success = false
            } else {
              sequence := string(buffer[:(closeByteIndex + 1)])
              if (sequence == ESC_F1_S) { sequence = ESC_F1 }
              shiftLeft(buffer, bufferLength, closeByteIndex + 1)
              index -= (closeByteIndex + 1)
              Input <- Key{EscapeSeq: &sequence}
            }
          }
        // Check if we have a control code, like Tab or Enter.
        } else if buffer[0] < 32 || buffer[0] == 127 {
          code := buffer[0]
          if code == 10 {
            code = ASC_ENTER
          }

          shiftLeft(buffer, bufferLength, 1)
          index -= 1
          Input <- Key{ControlCode: int(code)}
        // Everything else should be interpreted as runes.
        } else {
          r, size := utf8.DecodeRune(buffer)
          if r == utf8.RuneError {
            success = false
          } else {
            shiftLeft(buffer, bufferLength, size)
            index -= size
            Input <- Key{Rune: &r}
          }
        }

        if (index == 0) {
          success = false
        }
      }
    }
  }()
}

/* Utils. */

// Returns lesser of two integers.
func min(a, b int) int {
  if a < b {
    return a
  }
  return b
}

// Returns higher of two integers.
func max(a, b int) int {
  if a > b {
    return a
  }
  return b
}

// Shifts contents of array to the left by `from` number of elements.
// For example shiftLeft([0, 1, 2, 3, 4], 2) -> [2, 3, 4, 0, 0]
func shiftLeft(arr []byte, length, from int) {
  rest := arr[from:]
  copy(arr, rest)
  for i := length - from; i < length; i++ {
    arr[i] = 0
  }
}
