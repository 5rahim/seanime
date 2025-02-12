package debrid_client

import (
	"fmt"
	"github.com/5rahim/habari"
	"seanime/internal/api/anilist"
	"seanime/internal/debrid/debrid"
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
		FileId                string `json:"fileId"`
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
	defer util.HandlePanicInModuleWithError("debrid_client/GetTorrentFilePreviewsFromManualSelection", &err)

	if opts.Torrent == nil || opts.Magnet == "" || opts.Media == nil {
		return nil, fmt.Errorf("torrentstream: Invalid options")
	}

	r.logger.Trace().Str("hash", opts.Torrent.InfoHash).Msg("debridstream: Getting file previews for torrent selection")

	torrentInfo, err := r.GetTorrentInfo(debrid.GetTorrentInfoOptions{
		MagnetLink: opts.Magnet,
		InfoHash:   opts.Torrent.InfoHash,
	})
	if err != nil {
		r.logger.Error().Err(err).Msgf("debridstream: Error adding torrent %s", opts.Magnet)
		return nil, err
	}

	fileMetadataMap := make(map[string]*habari.Metadata)
	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}
	wg.Add(len(torrentInfo.Files))
	for _, file := range torrentInfo.Files {
		go func(file *debrid.TorrentItemFile) {
			defer wg.Done()
			defer util.HandlePanicInModuleThen("debridstream/GetTorrentFilePreviewsFromManualSelection", func() {})

			metadata := habari.Parse(file.Path)
			mu.Lock()
			fileMetadataMap[file.Path] = metadata
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

	for i, file := range torrentInfo.Files {
		wg.Add(1)
		go func(i int, file *debrid.TorrentItemFile) {
			defer wg.Done()
			defer util.HandlePanicInModuleThen("debridstream/GetTorrentFilePreviewsFromManualSelection", func() {})

			mu.RLock()
			metadata, found := fileMetadataMap[file.Path]
			mu.RUnlock()

			displayTitle := file.Path

			isLikely := false
			parsedEpisodeNumber := -1

			if found && !comparison.ValueContainsSpecial(file.Name) && !comparison.ValueContainsNC(file.Name) {
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
				Path:          file.Path,
				DisplayPath:   file.Path,
				DisplayTitle:  displayTitle,
				EpisodeNumber: parsedEpisodeNumber,
				IsLikely:      isLikely,
				FileId:        file.ID,
				Index:         i,
			})
			mu2.Unlock()
		}(i, file)
	}

	wg.Wait()

	return
}
