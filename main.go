package main

// #include <locale.h>
import "C"

import (
	"os/exec"
	"flag"
	"transmission"
	"windows"
	"tui"
)

func main() {
	C.setlocale(C.int(0), C.CString(""))

	// Command line arguments.
	var host = flag.String("h", "localhost", "Hostname")
	var port = flag.Int("p", 9091, "Port")
	var obfuscate = flag.Bool("o", false, "Obfuscate torrent and file names")
	var launch = flag.Bool("s", false, "Launch `transmission-daemon` before starting the client")
	flag.Parse()

	// Initialize daemon connection
	client := transmission.NewClient(*host, int32(*port))

	// If launch was requested in the arguments,
	// check for existing daemon first, and launch a new instance if needed.
	if *launch == true {
		_, err := client.GetSessionSettings()
		if err != nil {
			cmd := exec.Command("transmission-daemon")
			err := cmd.Run()
			if err != nil {
				panic(err)
			}
		}
	}

	// Initialize curses.
	stdscr := tui.Init()

	// Screen init.
	stdscr.Refresh()

	// Basic setup.
	tui.SetRaw(true)
	tui.HideCursor()
	defer func() {
		tui.ShowCursor()
		tui.SetRaw(false)
		tui.Clear()
	}()

	// Initialize window manager.
	manager := windows.NewWindowManager(stdscr)
	listWindow := windows.NewListWindow(stdscr, client, *obfuscate, manager)
	manager.AddWindow(listWindow)
	manager.Start()
}

