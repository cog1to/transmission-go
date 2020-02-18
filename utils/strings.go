package utils

import (
  "math/rand"
)

func RandomString(length int) string {
  runes := make([]rune, length)
  for i := 0; i < length; i++ {
    runes[i] = rune(rand.Int() % 94 + 32)
  }
  return string(runes)
}
