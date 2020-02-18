package main

// #include <locale.h>
import "C"

import (
  "log"
  "./transmission"
  "./windows"
  gc "./goncurses"
  "flag"
)

func main() {
  C.setlocale(C.int(0), C.CString(""))

  // Command line arguments.
  var host = flag.String("h", "localhost", "Hostname")
  var port = flag.Int("p", 9091, "Port")
  var obfuscate = flag.Bool("o", false, "Obfuscate torrent and file names")
  flag.Parse()

  // Initialize daemon connection
  client := transmission.NewClient(*host, int32(*port))

  // Initialize curses.
  stdscr, gcerr := gc.Init()
  if gcerr != nil {
    log.Fatal(gcerr)
  }

  defer gc.End()

  // Colors.
  gc.StartColor()
  gc.UseDefaultColors()
  gc.InitPair(1, gc.C_BLACK, gc.C_CYAN)
  gc.InitPair(2, 23, gc.C_CYAN)

  // Basic setup.
  gc.Raw(true)
  gc.Echo(false)
  gc.Cursor(0)

  // Arrow keys support.
  stdscr.Keypad(true)

  // Show list.
  windows.NewListWindow(stdscr, client, *obfuscate)
}

