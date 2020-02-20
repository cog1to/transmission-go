package logger

import (
  "log"
  "os"
)

var f *os.File

func Open(file string) {
  f, err := os.OpenFile(file, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    log.Fatalf("error opening file: %v", err)
  }
  log.SetOutput(f)
}

func Deinit() {
  f.Close()
}

func Log(line string) {
  log.Println(line)
}
