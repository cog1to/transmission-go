package windows

import (
  gc "../goncurses"
  transmission "../transmission"
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

func generalizeTorrents(items []transmission.TorrentListItem) []interface{} {
  output := make([]interface{}, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func generalizeFiles(items []transmission.TorrentFile) []interface{} {
  output := make([]interface{}, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func toTorrentList(items []interface{}) []transmission.TorrentListItem {
  output := make([]transmission.TorrentListItem, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentListItem)
  }
  return output
}

func toFileList(items []interface{}) []transmission.TorrentFile {
  output := make([]transmission.TorrentFile, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentFile)
  }
  return output
}
