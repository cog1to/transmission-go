package transform

import (
	"fmt"
	"transmission"
	"list"
	"sort"
)

func GeneralizeTorrents(
	items []transmission.TorrentListItem,
	sorted bool,
) []list.Identifiable {
	output := make([]list.Identifiable, len(items))
	for ind, item := range items {
		output[ind] = item
	}

	if !sorted {
		return output
	}

	sort.Slice(output, func(l, r int) bool {
		return output[l].(transmission.TorrentListItem).AddedDate < output[r].(transmission.TorrentListItem).AddedDate
	})

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

func MapToString(slice []transmission.TorrentListItem) []string {
	output := make([]string, len(slice))
	for index, element := range slice {
		output[index] = fmt.Sprintf("%d", element.Id())
	}
	return output
}

func MapToIds(slice []transmission.TorrentListItem) []int {
	output := make([]int, len(slice))
	for index, item := range slice {
		output[index] = item.Id()
	}
	return output
}

func IdsAndNextState(torrents []transmission.TorrentListItem) ([]int, bool) {
	isActive := false
	ids := make([]int, len(torrents))
	for i, torrent := range torrents {
		isActive = isActive || torrent.Status != transmission.TR_STATUS_STOPPED
		ids[i] = torrent.Id()
	}

	return ids, !isActive
}
