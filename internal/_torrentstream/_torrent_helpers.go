package torrentstream

import (
	"github.com/anacrolix/torrent"
)

func GetLargestFile(t *torrent.Torrent) *torrent.File {
	var curr *torrent.File
	for _, file := range t.Files() {
		if curr == nil || file.Length() > curr.Length() {
			curr = file
		}
	}
	return curr
}

func GetPercentageComplete(t *torrent.Torrent) float64 {
	info := t.Info()
	if info == nil {
		return 0
	}
	return float64(t.BytesCompleted()) / float64(info.TotalLength()) * 100
}
