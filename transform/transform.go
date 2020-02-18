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

