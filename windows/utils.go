package windows

import (
  gc "../goncurses"
  transmission "../transmission"
  "fmt"
  "strconv"
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

func intPrompt(window *gc.Window, reader *InputReader, title string, value int, flag bool, onFinish func(int), onError func(error)) {
  var initialValue string
  if flag && value > 0 {
    initialValue = fmt.Sprintf("%d", value)
  }

  output := Prompt(window, reader, title, 6, "0123456789", initialValue)
  if len(output) > 0 {
    limit, e := strconv.Atoi(output)
    if e != nil {
      onError(e)
    } else {
      onFinish(limit)
    }
  }
}

