package directstream

import (
	"seanime/internal/mediastream/mkvparser"

	"github.com/anacrolix/torrent"
)

// AppendSubtitleFile finds the subtitle file for the torrent and appends it as a track to the metadata
//   - If there's only one subtitle file, use it
//   - If there are multiple subtitle files, use the one that matches the name of the selected torrent file
//   - If there are no subtitle files, do nothing
//
// If the subtitle file is not ASS/SSA, it will be converted to ASS/SSA.
func (s *TorrentStream) AppendSubtitleFile(t *torrent.Torrent, file *torrent.File, metadata *mkvparser.Metadata) {

}
