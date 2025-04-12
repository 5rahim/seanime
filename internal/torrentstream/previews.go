package torrentstream

import (
	"fmt"
	"github.com/5rahim/habari"
	"github.com/anacrolix/torrent"
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
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

	fileMetadataMap := make(map[string]*habari.Metadata)
	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}
	wg.Add(len(selectedTorrent.Files()))
	for _, file := range selectedTorrent.Files() {
		go func(file *torrent.File) {
			defer wg.Done()
			defer util.HandlePanicInModuleThen("debridstream/GetTorrentFilePreviewsFromManualSelection", func() {})

			metadata := habari.Parse(file.DisplayPath())
			mu.Lock()
			fileMetadataMap[file.Path()] = metadata
			mu.Unlock()
		}(file)
	}
	wg.Wait()

	containsAbsoluteEps := false
	for _, metadata := range fileMetadataMap {
		if len(metadata.EpisodeNumber) == 1 {
			ep := util.StringToIntMust(metadata.EpisodeNumber[0])
			if ep > opts.Media.GetTotalEpisodeCount() {
				containsAbsoluteEps = true
				break
			}
		}
	}

	wg = sync.WaitGroup{}
	mu2 := sync.Mutex{}

	for i, file := range selectedTorrent.Files() {
		wg.Add(1)
		go func(i int, file *torrent.File) {
			defer wg.Done()
			defer util.HandlePanicInModuleThen("torrentstream/GetTorrentFilePreviewsFromManualSelection", func() {})

			mu.RLock()
			metadata := fileMetadataMap[file.Path()]
			mu.RUnlock()

			displayTitle := file.DisplayPath()

			isLikely := false
			parsedEpisodeNumber := -1

			if metadata != nil && !comparison.ValueContainsSpecial(displayTitle) && !comparison.ValueContainsNC(displayTitle) {
				if len(metadata.EpisodeNumber) == 1 {
					ep := util.StringToIntMust(metadata.EpisodeNumber[0])
					parsedEpisodeNumber = ep
					displayTitle = fmt.Sprintf("Episode %d", ep)
					if metadata.EpisodeTitle != "" {
						displayTitle = fmt.Sprintf("%s - %s", displayTitle, metadata.EpisodeTitle)
					}
				}
			}

			if !containsAbsoluteEps {
				isLikely = parsedEpisodeNumber == opts.EpisodeNumber
			}

			mu2.Lock()
			// Get the file preview
			ret = append(ret, &FilePreview{
				Path:          file.Path(),
				DisplayPath:   file.DisplayPath(),
				DisplayTitle:  displayTitle,
				EpisodeNumber: parsedEpisodeNumber,
				IsLikely:      isLikely,
				Index:         i,
			})
			mu2.Unlock()
		}(i, file)
	}

	wg.Wait()

	r.logger.Debug().Str("hash", opts.Torrent.InfoHash).Msg("torrentstream: Got file previews for torrent selection, dropping torrent")
	go selectedTorrent.Drop()

	return
}
