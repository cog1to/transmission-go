package main

import (
  "log"
  "time"
  "./transmission"
  "./windows"
  gc "github.com/rthornton128/goncurses"
)

func main() {
  // Initialize daemon connection
  client := transmission.NewClient("localhost", 9091)

  stop, op := make(chan bool), make(chan interface{})
  items, err := poll(client, 2, stop)
  go handleOperation(client, op, items, err)

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

  gc.Raw(true)
  gc.NewLines(true)
  gc.Echo(false)
  gc.Cursor(0)

  // Arrow keys support.
  stdscr.Keypad(true)

  // Show list.
  windows.NewListWindow(stdscr, items, err, op)
}

func handleOperation(client *transmission.Client, op chan interface{}, items chan *[]transmission.TorrentListItem, err chan error) {
  for {
    var e error

    operation := <-op
    switch operation.(type) {
    case windows.ListOperation:
      lop := operation.(windows.ListOperation)
      switch lop.Operation {
      case windows.DELETE:
        e = client.Delete([]int64{ lop.Item.Id }, false)
      case windows.DELETE_WITH_DATA:
        e = client.Delete([]int64{ lop.Item.Id }, true)
      }
    }

    if e != nil {
      err <- e
    } else {
      list, e := client.List()
      err <- e
      items <- list
    }
  }
}

func poll(client *transmission.Client, interval int, stop chan bool) (items chan *[]transmission.TorrentListItem, err chan error) {
  items, err = make(chan *[]transmission.TorrentListItem), make(chan error)

  // Go to background.
  go func() {
    // First poll to initialize.
    list, e := client.List()
    err <- e
    items <- list

    for {
      select {
        case <-stop:
          return
        case <-time.After(time.Duration(interval) * time.Second):
          list, e := client.List()
          err <- e
          items <- list
      }
    }
  }()

  return items, err
}
