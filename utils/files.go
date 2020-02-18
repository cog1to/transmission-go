package utils

import (
  "os"
  "strings"
)

func ExpandHome(input string) string {
  if strings.HasPrefix(input, "~") {
    return strings.Replace(input, "~", os.Getenv("HOME"), 1)
  }
  return input
}
