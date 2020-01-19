package windows

// #include <locale.h>
import "C"

import (
  gc "github.com/rthornton128/goncurses"
)

type InputReader struct {
  observerList [](chan gc.Key)
}

func NewInputReader(screen *gc.Window) *InputReader {
  reader := &InputReader{ make([](chan gc.Key), 10) }

  go func() {
    for {
      ch := screen.GetChar()
      if len(reader.observerList) > 0 {
        lastObserver := reader.observerList[len(reader.observerList) - 1]
        lastObserver <- ch
      }
    }
  }()

  return reader
}

func (reader *InputReader) AddObserver(channel chan gc.Key) {
  reader.observerList = append(reader.observerList, channel)
}

func (reader *InputReader) RemoveObserver(toDelete chan gc.Key) {
  var index = 0
  for idx, channel := range reader.observerList {
    if channel == toDelete {
      index = idx
    }
  }

  copy(reader.observerList[index:], reader.observerList[index+1:])
  reader.observerList[len(reader.observerList)-1] = nil
  reader.observerList = reader.observerList[:len(reader.observerList)-1]
}

