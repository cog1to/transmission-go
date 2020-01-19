package main

// #include <locale.h>
import "C"

import (
  "log"
  "./transmission"
  "./windows"
  gc "github.com/rthornton128/goncurses"
)

func main() {
  C.setlocale(C.int(0), C.CString("en_US.UTF-8"))

  // Initialize daemon connection
  client := transmission.NewClient("localhost", 9091)

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

  // Basic setup.
  gc.Raw(true)
  gc.Echo(false)
  gc.Cursor(0)

  // Arrow keys support.
  stdscr.Keypad(true)

  // Show list.
  windows.NewListWindow(stdscr, client)
}

