package main

// #include <locale.h>
import "C"

import (
  "./transmission"
  "./windows"
  tui "./tui"
  "flag"
  "./logger"
)

func main() {
  C.setlocale(C.int(0), C.CString(""))
  logger.Open("log.log")

  // Command line arguments.
  var host = flag.String("h", "localhost", "Hostname")
  var port = flag.Int("p", 9091, "Port")
  var obfuscate = flag.Bool("o", false, "Obfuscate torrent and file names")
  flag.Parse()

  // Initialize daemon connection
  client := transmission.NewClient(*host, int32(*port))

  // Initialize curses.
  stdscr := tui.Init()
  defer tui.ShowCursor()

  // Screen init.
  stdscr.Refresh()

  // Basic setup.
  tui.SetRaw(true)
  tui.SetEcho(false)
  tui.HideCursor()

  // Initialize window manager.
  manager := windows.NewWindowManager(stdscr)
  listWindow := windows.NewListWindow(stdscr, client, *obfuscate, manager)
  manager.AddWindow(listWindow)
  manager.Start()
}

