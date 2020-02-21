package windows

import (
  gc "../goncurses"
  "os"
  "os/signal"
  "syscall"
)

type InputReader interface {
  OnInput(gc.Key)
}

type Window interface {
  Draw()
  Resize()
  IsFullScreen()bool
  SetActive(bool)
  InputReader
}

type WindowManager struct {
  root *gc.Window
  windows []Window
  inputReaders []InputReader
  signals chan os.Signal
  input chan gc.Key
  Exit chan bool
  Draw chan bool
}

func NewWindowManager(root *gc.Window) *WindowManager {
  manager := &WindowManager{
    root,
    make([]Window, 0),
    make([]InputReader, 0),
    make(chan os.Signal),
    make(chan gc.Key),
    make(chan bool),
    make(chan bool)}

  // Signal channel. 
  signal.Notify(manager.signals, syscall.SIGWINCH)

  // Input channel.
  go func() {
    for {
      ch := root.GetWChar()
      manager.input <- ch
    }
  }()

  return manager
}

func (manager *WindowManager) AddWindow(win Window) {
  if len(manager.windows) > 0 {
    manager.windows[len(manager.windows) - 1].SetActive(false)
  }

  manager.windows = append(manager.windows, win)
  manager.inputReaders = append(manager.inputReaders, win)
  win.SetActive(true)
  manager.Redraw()
}

func (manager *WindowManager) RemoveWindow(win Window) {
  win.SetActive(false)

  var index = -1
  for idx, window := range manager.windows {
    if window == win {
      index = idx
    }
  }

  if index >= 0 {
    newStack := manager.windows[:index]
    if index < len(manager.windows) - 1 {
      newStack = append(newStack, manager.windows[index+1:]...)
    }
    manager.windows = newStack
  }

  manager.RemoveInputReader(win)
  if len(manager.windows) > 0 {
    manager.windows[len(manager.windows) - 1].SetActive(true)
  }
  manager.Redraw()
}

func (manager *WindowManager) AddInputReader(reader InputReader) {
  manager.inputReaders = append(manager.inputReaders, reader)
}

func (manager *WindowManager) RemoveInputReader(rdr InputReader) {
  var index = -1
  for idx, reader := range manager.inputReaders {
    if reader == rdr {
      index = idx
    }
  }

  if index >= 0 {
    newStack := manager.inputReaders[:index]
    if index < len(manager.inputReaders) - 1 {
      newStack = append(newStack, manager.inputReaders[index+1:]...)
    }
    manager.inputReaders = newStack
  }
}

func (manager *WindowManager) Redraw() {
  for _, window := range manager.windows {
    window.Draw()
  }
}

func (manager *WindowManager) Resize() {
  for _, window := range manager.windows {
    window.Resize()
    window.Draw()
  }
}

func (manager *WindowManager) Start() {
  defer gc.End()

  for {
    select {
    case <-manager.Draw:
      manager.Redraw()
    case sig := <-manager.signals:
      if sig == syscall.SIGWINCH {
        gc.End()
        manager.root.Refresh()
        manager.Resize()
        manager.Redraw()
      }
    case input := <-manager.input:
      if len(manager.inputReaders) > 0 {
       lastReader := manager.inputReaders[len(manager.inputReaders) - 1]
       lastReader.OnInput(input)
      }
    case <-manager.Exit:
      return
    }
  }
}
