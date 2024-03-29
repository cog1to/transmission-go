package windows

import (
	"os"
	"os/signal"
	"syscall"
	"tui"
)

type InputReader interface {
	OnInput(tui.Key)
}

type Window interface {
	Draw()
	Resize()
	IsFullScreen()bool
	SetActive(bool)
	InputReader
}

type WindowManager struct {
	root tui.Drawable
	windows []Window
	inputReaders []InputReader
	signals chan os.Signal
	resize bool
	input chan tui.Key
	Exit chan bool
	Draw chan bool
}

func NewWindowManager(root tui.Drawable) *WindowManager {
	manager := &WindowManager{
		root,
		make([]Window, 0),
		make([]InputReader, 0),
		make(chan os.Signal, 1),
		false,
		make(chan tui.Key),
		make(chan bool),
		make(chan bool)}

	// Signal channel.
	signal.Notify(manager.signals, syscall.SIGWINCH)
	signal.Notify(manager.signals, os.Interrupt)
	go func() {
		for {
			sig := <-manager.signals
			if sig == syscall.SIGWINCH {
				manager.resize = true
				manager.Draw <- true
			} else if sig == os.Interrupt {
				manager.Exit <- true
			}
		}
	}()

	// Input channel.
	tui.StartListening()
	go func() {
		for {
			input := <-tui.Input
			manager.input <- input
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

	go func() {
		manager.Draw <- true
	}()
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

	go func() {
		manager.Draw <- true
	}()
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
	var fullScreenIndex int = -1
	for ind := len(manager.windows) - 1; ind >= 0; ind-- {
		if manager.windows[ind].IsFullScreen() {
			fullScreenIndex = ind
		}
	}

	var windows []Window
	if fullScreenIndex > 0 {
		windows = manager.windows[fullScreenIndex:]
	} else {
		windows = manager.windows
	}

	for _, window := range windows {
		window.Draw()
	}
}

func (manager *WindowManager) DrawTop() {
	if top := manager.Top(); top != nil {
		top.Draw()
	}
}

func (manager *WindowManager) Resize() {
	for _, window := range manager.windows {
		window.Resize()
	}
}

func (manager *WindowManager) Top() Window {
	if len(manager.windows) > 0 {
		return manager.windows[len(manager.windows) - 1]
	}
	return nil
}

func (manager *WindowManager) Start() {
	for {
		select {
		case <-manager.Draw:
			if manager.resize {
				manager.resize = false
				manager.root.Refresh()
				manager.Resize()
				manager.Redraw()
			}
			manager.DrawTop()
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
