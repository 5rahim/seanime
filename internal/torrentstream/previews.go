package torrentstream

import (
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/anacrolix/torrent"
	"seanime/internal/api/anilist"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/seanime-parser"
	"sync"
)

type (
	FilePreview struct {
		Path                  string `json:"path"`
		DisplayPath           string `json:"displayPath"`
		DisplayTitle          string `json:"displayTitle"`
		EpisodeNumber         int    `json:"episodeNumber"`
		RelativeEpisodeNumber int    `json:"relativeEpisodeNumber"`
		IsLikely              bool   `json:"isLikely"`
		Index                 int    `json:"index"`
	}

	GetTorrentFilePreviewsOptions struct {
		Torrent        *hibiketorrent.AnimeTorrent
		Magnet         string
		EpisodeNumber  int
		AbsoluteOffset int
		Media          *anilist.BaseAnime
	}
)

func (r *Repository) GetTorrentFilePreviewsFromManualSelection(opts *GetTorrentFilePreviewsOptions) (ret []*FilePreview, err error) {
	defer util.HandlePanicInModuleWithError("torrentstream/GetTorrentFilePreviewsFromManualSelection", &err)

	if opts.Torrent == nil || opts.Magnet == "" || opts.Media == nil {
		return nil, fmt.Errorf("torrentstream: Invalid options")
	}

	r.logger.Trace().Str("hash", opts.Torrent.InfoHash).Msg("torrentstream: Getting file previews for torrent selection")

	selectedTorrent, err := r.client.AddTorrent(opts.Magnet)
	if err != nil {
		r.logger.Error().Err(err).Msgf("torrentstream: Error adding torrent %s", opts.Magnet)
		return nil, err
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for i, file := range selectedTorrent.Files() {
		wg.Add(1)
		go func(i int, file *torrent.File) {
			defer wg.Done()
			defer util.HandlePanicInModuleThen("torrentstream/GetTorrentFilePreviewsFromManualSelection", func() {})

			metadata := seanime_parser.Parse(file.DisplayPath())

			displayTitle := file.Path()

			isLikely := false
			parsedEpisodeNumber := -1

			if metadata != nil && !comparison.ValueContainsSpecial(file.DisplayPath()) {
				if len(metadata.EpisodeNumber) == 1 {
					ep := util.StringToIntMust(metadata.EpisodeNumber[0])
					parsedEpisodeNumber = ep
					displayTitle = fmt.Sprintf("Episode %d", ep)
					if metadata.EpisodeTitle != "" {
						displayTitle = fmt.Sprintf("%s - %s", displayTitle, metadata.EpisodeTitle)
					}
				}
			}

			relativeEpisode := parsedEpisodeNumber
			if relativeEpisode > opts.Media.GetTotalEpisodeCount() {
				relativeEpisode = relativeEpisode - opts.AbsoluteOffset
			}

			isLikely = relativeEpisode == opts.EpisodeNumber

			mu.Lock()
			// Get the file preview
			ret = append(ret, &FilePreview{
				Path:                  file.Path(),
				DisplayPath:           file.DisplayPath(),
				DisplayTitle:          displayTitle,
				EpisodeNumber:         parsedEpisodeNumber,
				RelativeEpisodeNumber: relativeEpisode,
				IsLikely:              isLikely,
				Index:                 i,
			})
			mu.Unlock()
		}(i, file)
	}

	wg.Wait()

	r.logger.Debug().Str("hash", opts.Torrent.InfoHash).Msg("torrentstream: Got file previews for torrent selection, dropping torrent")
	go selectedTorrent.Drop()

	return
}
