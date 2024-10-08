package offline

import (
	"github.com/rs/zerolog"
	"seanime/internal/library/anime"
	"seanime/internal/util/image_downloader"
	"sync"
	"time"
)

type (
	assetsHandler struct {
		logger          *zerolog.Logger
		imageDownloader *image_downloader.ImageDownloader
	}
)

func newAssetsHandler(logger *zerolog.Logger, imageDownloader *image_downloader.ImageDownloader) *assetsHandler {
	return &assetsHandler{
		logger:          logger,
		imageDownloader: imageDownloader,
	}
}

func (h *assetsHandler) DownloadAssets(
	animeEntries []*AnimeEntry,
	mangaEntries []*MangaEntry,
	user *anime.User,
	ids []int, // Media to download assets for
) (ret *AssetMapImageMap, err error) {

	h.imageDownloader.DeleteDownloads()

	ret = &AssetMapImageMap{}
	mu := sync.Mutex{}
	cancelCh := make(chan struct{})
	errCh := make(chan error)

	userAvatarUrls := make([]string, 0)
	if user.Viewer.GetAvatar().GetLarge() != nil {
		userAvatarUrls = append(userAvatarUrls, *user.Viewer.GetAvatar().GetLarge())
	}
	if user.Viewer.GetAvatar().GetMedium() != nil {
		userAvatarUrls = append(userAvatarUrls, *user.Viewer.GetAvatar().GetMedium())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = h.imageDownloader.DownloadImages(userAvatarUrls)
		imageMap, err := h.imageDownloader.GetImageFilenamesByUrls(userAvatarUrls)
		if err == nil {
			for url, filename := range imageMap {
				(*ret)[url] = filename
			}
		}
	}()
	wg.Wait()

	wg1 := sync.WaitGroup{}
	for _, animeEntry := range animeEntries {
		//if !slices.Contains(ids, animeEntry.MediaId) {
		//	continue
		//}

		wg1.Add(1)
		go func(entry *AnimeEntry) {
			defer wg1.Done()

			select {
			case <-cancelCh:
				return
			default:
				// Download the anime entry's assets
				assetMap, err := h.downloadAnimeEntryAssets(entry)
				if err != nil {
					errCh <- err
					return
				}

				mu.Lock()
				for url, filename := range assetMap {
					(*ret)[url] = filename
				}
				mu.Unlock()
			}
		}(animeEntry)
	}

	wg2 := sync.WaitGroup{}
	for _, mangaEntry := range mangaEntries {
		//if !slices.Contains(ids, mangaEntry.MediaId) {
		//	continue
		//}

		wg2.Add(1)
		go func(entry *MangaEntry) {
			defer wg2.Done()

			select {
			case <-cancelCh:
				return
			default:
				// Download the manga entry's assets
				assetMap, err := h.downloadMangaEntryAssets(entry)
				if err != nil {
					errCh <- err
					return
				}

				mu.Lock()
				for url, filename := range assetMap {
					(*ret)[url] = filename
				}
				mu.Unlock()
			}
		}(mangaEntry)
	}

	go func() {
		wg1.Wait()
		wg2.Wait()
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			close(cancelCh)
			return nil, err
		}
	}

	close(cancelCh)

	return ret, nil
}

func (h *assetsHandler) downloadAnimeEntryAssets(entry *AnimeEntry) (ret AssetMapImageMap, err error) {
	ret = AssetMapImageMap{}

	urls := make([]string, 0)

	urlMap := make(map[string]bool)

	// Get the anime entry's images
	coverUrl := entry.Media.GetCoverImageSafe()
	if coverUrl != "" {
		urls = append(urls, coverUrl)
		urlMap[coverUrl] = true
	}
	bannerUrl := entry.Media.GetBannerImageSafe()
	if bannerUrl != "" {
		if _, ok := urlMap[bannerUrl]; !ok {
			urls = append(urls, bannerUrl)
			urlMap[bannerUrl] = true
		}
	}

	// Get the anime entry's episode images
	for _, episode := range entry.Episodes {
		if episode.EpisodeMetadata == nil {
			continue
		}
		if _, ok := urlMap[episode.EpisodeMetadata.Image]; !ok {
			urls = append(urls, episode.EpisodeMetadata.Image)
			urlMap[episode.EpisodeMetadata.Image] = true
		}
	}

	ret, err = h.downloadImages(urls, entry.MediaId)

	return
}

func (h *assetsHandler) downloadMangaEntryAssets(entry *MangaEntry) (ret AssetMapImageMap, err error) {
	ret = AssetMapImageMap{}

	urls := make([]string, 0)

	urlMap := make(map[string]bool)

	// Get the manga entry's images
	coverUrl := entry.Media.GetCoverImageSafe()
	if coverUrl != "" {
		urls = append(urls, coverUrl)
		urlMap[coverUrl] = true
	}
	bannerUrl := entry.Media.GetBannerImageSafe()
	if bannerUrl != "" {
		if _, ok := urlMap[bannerUrl]; !ok {
			urls = append(urls, bannerUrl)
			urlMap[bannerUrl] = true
		}
	}

	ret, err = h.downloadImages(urls, entry.MediaId)

	return
}

func (h *assetsHandler) downloadImages(urls []string, mId int) (ret AssetMapImageMap, err error) {
	ret = AssetMapImageMap{}

	retryCount := 0
	for {
		err = h.imageDownloader.DownloadImages(urls)
		if err != nil {
			if retryCount < 2 { // Retry for up to 2 times
				h.logger.Error().Err(err).Int("mediaId", mId).Msg("offline hub: Failed to download anime entry assets, retrying")
				retryCount++
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}

		// Download the images
		imageMap, err := h.imageDownloader.GetImageFilenamesByUrls(urls)
		if err != nil {
			if retryCount < 2 { // Retry for up to 2 times
				h.logger.Error().Err(err).Int("mediaId", mId).Msg("offline hub: Failed to retrieved downloaded image filenames, retrying")
				retryCount++
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}

		// Add the image map to the return value
		for url, filename := range imageMap {
			ret[url] = filename
		}
		h.logger.Debug().Int("mediaId", mId).Msg("offline hub: Downloaded anime entry assets")
		break
	}

	return ret, nil
}
