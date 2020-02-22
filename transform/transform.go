package transform

import (
  "../transmission"
  "../list"
)

func GeneralizeTorrents(items []transmission.TorrentListItem) []list.Identifiable {
  output := make([]list.Identifiable, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func GeneralizeFiles(items []transmission.TorrentFile) []list.Identifiable {
  output := make([]list.Identifiable, len(items))
  for ind, item := range items {
    output[ind] = item
  }
  return output
}

func ToTorrentList(items []list.Identifiable) []transmission.TorrentListItem {
  output := make([]transmission.TorrentListItem, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentListItem)
  }
  return output
}

func ToFileList(items []list.Identifiable) []transmission.TorrentFile {
  output := make([]transmission.TorrentFile, len(items))
  for ind, item := range items {
    output[ind] = item.(transmission.TorrentFile)
  }
  return output
}

func IdsAndNextPriority(files []transmission.TorrentFile) ([]int, int) {
  minPriority := 99
  ids := make([]int, len(files))
  for i, file := range files {
    if file.Priority < minPriority || minPriority == 99 {
      minPriority = file.Priority
    }
    ids[i] = file.Number
  }

  nextPriority := transmission.TR_PRIORITY_LOW
  if minPriority != 99 && minPriority != transmission.TR_PRIORITY_HIGH {
    nextPriority = minPriority + 1
  }

  return ids, nextPriority
}

func IdsAndNextWanted(files []transmission.TorrentFile) ([]int, bool) {
  wanted := true
  ids := make([]int, len(files))
  for i, file := range files {
    wanted = wanted && file.Wanted
    ids[i] = file.Number
  }

  return ids, !wanted
}
