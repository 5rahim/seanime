package local

import (
	"fmt"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/image_downloader"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

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

	logger.Trace().Msgf("local manager: Downloading episode images for anime %d", mId)

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
		logger.Error().Err(err).Msgf("local manager: Failed to download images for anime %d", mId)
		return nil, false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("local manager: Failed to get image filenames for anime %d", mId)
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
	animeMetadata *metadata.AnimeMetadata, // This is updated
	metadataWrapper metadata.AnimeMetadataWrapper,
	lfs []*anime.LocalFile,
) (string, string, map[string]string, bool) {
	defer util.HandlePanicInModuleThen("sync/DownloadAnimeImages", func() {})

	logger.Trace().Msgf("local manager: Downloading images for anime %d", entry.Media.ID)
	// e.g. /datadir/local/assets/123
	mediaAssetPath := filepath.Join(assetsDir, fmt.Sprintf("%d", entry.Media.ID))
	imageDownloader := image_downloader.NewImageDownloader(mediaAssetPath, logger)
	// Download the images
	ogBannerImage := entry.GetMedia().GetBannerImageSafe()
	ogCoverImage := entry.GetMedia().GetCoverImageSafe()

	imgUrls := []string{ogBannerImage, ogCoverImage}

	lfMap := make(map[string]*anime.LocalFile)
	for _, lf := range lfs {
		lfMap[lf.Metadata.AniDBEpisode] = lf
	}

	ogEpisodeImages := make(map[string]string)
	for episodeNum, episode := range animeMetadata.Episodes {
		// Check if the episode is in the local files
		if _, ok := lfMap[episodeNum]; !ok {
			continue
		}

		episodeInt, ok := util.StringToInt(episodeNum)
		if !ok {
			ogEpisodeImages[episodeNum] = episode.Image
			imgUrls = append(imgUrls, episode.Image)
			continue
		}

		epMetadata := metadataWrapper.GetEpisodeMetadata(episodeInt)
		episode = &epMetadata

		ogEpisodeImages[episodeNum] = episode.Image
		imgUrls = append(imgUrls, episode.Image)
	}

	err := imageDownloader.DownloadImages(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("local manager: Failed to download images for anime %d", entry.Media.ID)
		return "", "", nil, false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("local manager: Failed to get image filenames for anime %d", entry.Media.ID)
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

	logger.Debug().Msgf("local manager: Stored images for anime %d, %+v, %+v, episode images: %+v", entry.Media.ID, bannerImage, coverImage, len(episodeImagePaths))

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
	logger.Trace().Msgf("local manager: Downloading images for manga %d", entry.Media.ID)

	// e.g. /datadir/local/assets/123
	mediaAssetPath := filepath.Join(assetsDir, fmt.Sprintf("%d", entry.Media.ID))
	imageDownloader := image_downloader.NewImageDownloader(mediaAssetPath, logger)
	// Download the images
	ogBannerImage := entry.GetMedia().GetBannerImageSafe()
	ogCoverImage := entry.GetMedia().GetCoverImageSafe()

	imgUrls := []string{ogBannerImage, ogCoverImage}

	err := imageDownloader.DownloadImages(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("local manager: Failed to download images for anime %d", entry.Media.ID)
		return "", "", false
	}

	images, err := imageDownloader.GetImageFilenamesByUrls(imgUrls)
	if err != nil {
		logger.Error().Err(err).Msgf("local manager: Failed to get image filenames for anime %d", entry.Media.ID)
		return "", "", false
	}

	bannerImage := images[ogBannerImage]
	coverImage := images[ogCoverImage]

	logger.Debug().Msgf("local manager: Stored images for manga %d, %+v, %+v", entry.Media.ID, bannerImage, coverImage)

	return bannerImage, coverImage, true
}
