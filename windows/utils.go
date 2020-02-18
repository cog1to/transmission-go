package windows

import (
  gc "../goncurses"
  transmission "../transmission"
  "math/rand"
)

func minInt(x, y int) int {
  if x < y {
    return x
  } else {
    return y
  }
}

func maxInt(x, y int) int {
  if x > y {
    return x
  } else {
    return y
  }
}

func withAttribute(window *gc.Window, attr gc.Char, block func(*gc.Window)) {
  window.AttrOn(attr)
  block(window)
  window.AttrOff(attr)
}

func withColor(window *gc.Window, color int16, block func(*gc.Window)) {
  window.ColorOn(color)
  block(window)
  window.ColorOff(color)
}

func remove(slice []rune, s int) []rune {
  return append(slice[:s], slice[s+1:]...)
}

func removeInt(slice []int, el int) []int {
    for index, element := range slice {
    if element == el {
      return append(slice[:index], slice[index+1:]...)
    }
  }
  return slice
}

func contains(slice []int, el int) bool {
  for _, element := range slice {
    if element == el {
      return true
    }
  }
  return false
}

func generalizeTorrents(items []transmission.TorrentListItem) []Identifiable {
  output := make([]Identifiable, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func generalizeFiles(items []transmission.TorrentFile) []Identifiable {
  output := make([]Identifiable, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func toTorrentList(items []Identifiable) []transmission.TorrentListItem {
  output := make([]transmission.TorrentListItem, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentListItem)
  }
  return output
}

func toFileList(items []Identifiable) []transmission.TorrentFile {
  output := make([]transmission.TorrentFile, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentFile)
  }
  return output
}

func randomString(length int) string {
  runes := make([]rune, length)
  for i := 0; i < length; i++ {
    runes[i] = rune(rand.Int() % 94 + 32)
  }
  return string(runes)
}
