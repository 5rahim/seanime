package sync

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/image_downloader"
)

// FormatAssetUrl formats the asset URL for the given mediaId and filename.
//
//	FormatAssetUrl(123, "cover.jpg") -> "{{LOCAL_ASSETS}}/123/cover.jpg"
func FormatAssetUrl(mediaId int, filename string) *string {
	// {{LOCAL_ASSETS}} should be replaced in the client with the actual URL
	// e.g. http://<hostname>/local_assets/123/cover.jpg
	a := fmt.Sprintf("{{LOCAL_ASSETS}}/%d/%s", mediaId, filename)
	return &a
}

// BaseAnimeDeepCopy creates a deep copy of the given base anime struct.
func BaseAnimeDeepCopy(animeCollection *anilist.BaseAnime) *anilist.BaseAnime {
	bytes, err := json.Marshal(animeCollection)
	if err != nil {
		return nil
	}

	deepCopy := &anilist.BaseAnime{}
	err = json.Unmarshal(bytes, deepCopy)
	if err != nil {
		return nil
	}

	deepCopy.NextAiringEpisode = nil

	return deepCopy
}

// BaseMangaDeepCopy creates a deep copy of the given base manga struct.
func BaseMangaDeepCopy(animeCollection *anilist.BaseManga) *anilist.BaseManga {
	bytes, err := json.Marshal(animeCollection)
	if err != nil {
		return nil
	}

	deepCopy := &anilist.BaseManga{}
	err = json.Unmarshal(bytes, deepCopy)
	if err != nil {
		return nil
	}

	return deepCopy
}

func ToNewPointer[A any](a *A) *A {
	if a == nil {
		return nil
	}
	t := *a
	return &t
}

func IntPointerValue[A int](a *A) A {
	if a == nil {
		return 0
	}
	return *a
}

func Float64PointerValue[A float64](a *A) A {
	if a == nil {
		return 0
	}
	return *a
}

func MediaListStatusPointerValue(a *anilist.MediaListStatus) anilist.MediaListStatus {
	if a == nil {
		return anilist.MediaListStatusPlanning
	}
	return *a
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// DownloadAnimeEpisodeImages saves the episode images for the given anime media ID.
// This should be used to update the episode images for an anime, e.g. after a new episode is released.
//
// The episodeImageUrls map should be in the format of {"1": "url1", "2": "url2", ...}, where the key is the episode number (defined in metadata.AnimeMetadata).
// It will download the images to the `<assetsDir>/<mId>` directory and return a map of episode numbers to the downloaded image filenames.
//
//	DownloadAnimeEpisodeImages(logger, "path/to/datadir/local/assets", 123, map[string]string{"1": "url1", "2": "url2"})
//	-> map[string]string{"1": "filename1.jpg", "2": "filename2.jpg"}
func DownloadAnimeEpisodeImages(logger *zerolog.Logger, assetsDir string, mId int, episodeImageUrls map[string]string) (map[string]string, bool) {
	defer util.HandlePanicInModuleThen("sync/DownloadAnimeEpisodeImages", func() {})

	logger.Trace().Msgf("sync: Downloading episode images for anime %d", mId)

	// e.g. /path/to/datadir/local/assets/123
	mediaAssetPath := filepath.Join(assetsDir, fmt.Sprintf("%d", mId))
	imageDownloader := image_downloader.NewImageDownloader(mediaAssetPath, logger)
	// Download the images
	imgUrls := make([]string, 0, len(episodeImageUrls))
	for _, episodeImage := range episodeImageUrls {
		if episodeImage == "" {
			continue
		}
		imgUrls = append(imgUrls, episodeImage)
	}

	err := imageDownloader.DownloadImages(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to download images for anime %d", mId)
		return nil, false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to get image filenames for anime %d", mId)
		return nil, false
	}

	episodeImagePaths := make(map[string]string)
	for episodeNum, episodeImage := range episodeImageUrls {
		episodeImagePaths[episodeNum] = images[episodeImage]
	}

	return episodeImagePaths, true
}

// DownloadAnimeImages saves the banner, cover, and episode images for the given anime entry.
// This should be used to download the images for an anime for the first time.
//
// It will download the images to the `<assetsDir>/<mId>` directory and return the filenames of the banner, cover, and episode images.
//
//	DownloadAnimeImages(logger, "path/to/datadir/local/assets", entry, animeMetadata)
//	-> "banner.jpg", "cover.jpg", map[string]string{"1": "filename1.jpg", "2": "filename2.jpg"}
func DownloadAnimeImages(
	logger *zerolog.Logger,
	assetsDir string,
	entry *anilist.AnimeListEntry,
	animeMetadata *metadata.AnimeMetadata,
	metadataWrapper metadata.AnimeMetadataWrapper,
	lfs []*anime.LocalFile,
) (string, string, map[string]string, bool) {
	defer util.HandlePanicInModuleThen("sync/DownloadAnimeImages", func() {})

	logger.Trace().Msgf("sync: Downloading images for anime %d", entry.Media.ID)
	// e.g. /datadir/local/assets/123
	mediaAssetPath := filepath.Join(assetsDir, fmt.Sprintf("%d", entry.Media.ID))
	imageDownloader := image_downloader.NewImageDownloader(mediaAssetPath, logger)
	// Download the images
	ogBannerImage := entry.GetMedia().GetBannerImageSafe()
	ogCoverImage := entry.GetMedia().GetCoverImageSafe()

	imgUrls := []string{ogBannerImage, ogCoverImage}

	ogEpisodeImages := make(map[string]string)
	for episodeNum, episode := range animeMetadata.Episodes {
		// Check if the episode is in the local files
		if _, found := lo.Find(lfs, func(lf *anime.LocalFile) bool {
			return lf.Metadata.AniDBEpisode == episode.Episode
		}); !found {
			continue
		}

		epMetadata := metadataWrapper.GetEpisodeMetadata(util.StringToIntMust(episodeNum))
		episode = &epMetadata

		if episode.Image == "" {
			continue
		}

		ogEpisodeImages[episodeNum] = episode.Image
		imgUrls = append(imgUrls, episode.Image)
	}

	err := imageDownloader.DownloadImages(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to download images for anime %d", entry.Media.ID)
		return "", "", nil, false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to get image filenames for anime %d", entry.Media.ID)
		return "", "", nil, false
	}

	bannerImage := images[ogBannerImage]
	coverImage := images[ogCoverImage]
	episodeImagePaths := make(map[string]string)
	for episodeNum, episodeImage := range ogEpisodeImages {
		if episodeImage == "" {
			continue
		}
		episodeImagePaths[episodeNum] = images[episodeImage]
	}

	logger.Debug().Msgf("sync: Stored images for anime %d, %+v, %+v, episode images: %+v", entry.Media.ID, bannerImage, coverImage, len(episodeImagePaths))

	return bannerImage, coverImage, episodeImagePaths, true
}

// DownloadMangaImages saves the banner and cover images for the given manga entry.
// This should be used to download the images for a manga for the first time.
//
// It will download the images to the `<assetsDir>/<mId>` directory and return the filenames of the banner and cover images.
//
//	DownloadMangaImages(logger, "path/to/datadir/local/assets", entry)
//	-> "banner.jpg", "cover.jpg"
func DownloadMangaImages(logger *zerolog.Logger, assetsDir string, entry *anilist.MangaListEntry) (string, string, bool) {
	logger.Trace().Msgf("sync: Downloading images for manga %d", entry.Media.ID)

	// e.g. /datadir/local/assets/123
	mediaAssetPath := filepath.Join(assetsDir, fmt.Sprintf("%d", entry.Media.ID))
	imageDownloader := image_downloader.NewImageDownloader(mediaAssetPath, logger)
	// Download the images
	ogBannerImage := entry.GetMedia().GetBannerImageSafe()
	ogCoverImage := entry.GetMedia().GetCoverImageSafe()

	imgUrls := []string{ogBannerImage, ogCoverImage}

	err := imageDownloader.DownloadImages(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to download images for anime %d", entry.Media.ID)
		return "", "", false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("sync: Failed to get image filenames for anime %d", entry.Media.ID)
		return "", "", false
	}

	bannerImage := images[ogBannerImage]
	coverImage := images[ogCoverImage]

	logger.Debug().Msgf("sync: Stored images for manga %d, %+v, %+v", entry.Media.ID, bannerImage, coverImage)

	return bannerImage, coverImage, true
}
