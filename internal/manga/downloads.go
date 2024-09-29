package manga

import (
	"cmp"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/manga/downloader"
	"slices"
)

func (r *Repository) GetDownloadedChapterContainers(mangaCollection *anilist.MangaCollection) (ret []*ChapterContainer, err error) {
	ret = make([]*ChapterContainer, 0)

	// Read download directory
	files, err := os.ReadDir(r.downloadDir)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to read download directory")
		return nil, err
	}

	// Get all chapter directories
	// e.g. manga_comick_123_10010_13
	chapterDirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			_, ok := chapter_downloader.ParseChapterDirName(file.Name())
			if !ok {
				continue
			}
			chapterDirs = append(chapterDirs, file.Name())
		}
	}

	if len(chapterDirs) == 0 {
		return nil, nil
	}

	// Now that we have all the chapter directories, we can get the chapter containers

	keys := make([]*chapter_downloader.DownloadID, 0)
	for _, dir := range chapterDirs {
		downloadId, ok := chapter_downloader.ParseChapterDirName(dir)
		if !ok {
			continue
		}
		keys = append(keys, &downloadId)
	}

	providerAndMediaIdPairs := make(map[struct {
		provider string
		mediaId  int
	}]bool)

	for _, key := range keys {
		providerAndMediaIdPairs[struct {
			provider string
			mediaId  int
		}{
			provider: key.Provider,
			mediaId:  key.MediaId,
		}] = true
	}

	// Get the chapter containers
	for pair := range providerAndMediaIdPairs {
		provider := pair.provider
		mediaId := pair.mediaId

		// Get the manga from the collection
		mangaEntry, ok := mangaCollection.GetListEntryFromMangaId(mediaId)
		if !ok {
			r.logger.Warn().Int("mediaId", mediaId).Msg("manga: [GetDownloadedChapterContainers] Manga not found in collection")
			continue
		}

		// Get the list of chapters for the manga
		// Check the permanent file cache
		container, found := r.getChapterContainerFromPermanentFilecache(provider, mediaId)
		if !found {
			// Check the temporary file cache
			container, found = r.getChapterContainerFromFilecache(provider, mediaId)
			if !found {
				// Get the chapters from the provider
				// This stays here for backwards compatibility, but ideally the method should not require an internet connection
				// so this will fail if the chapters were not cached & with no internet
				opts := GetMangaChapterContainerOptions{
					Provider: provider,
					MediaId:  mediaId,
					Titles:   mangaEntry.GetMedia().GetAllTitles(),
					Year:     mangaEntry.GetMedia().GetStartYearSafe(),
				}
				container, err = r.GetMangaChapterContainer(&opts)
				if err != nil {
					r.logger.Error().Err(err).Int("mediaId", mediaId).Msg("manga: [GetDownloadedChapterContainers] Failed to retrieve cached list of manga chapters")
					continue
				}

			}
		}

		// Now that we have the container, we'll filter out the chapters that are not downloaded
		// Go through each chapter and check if it's downloaded
		for _, chapter := range container.Chapters {
			// For each chapter, check if the chapter directory exists
			for _, dir := range chapterDirs {
				if dir == chapter_downloader.FormatChapterDirName(provider, mediaId, chapter.ID, chapter.Chapter) {
					container.Chapters = append(container.Chapters, chapter)
					break
				}
			}
		}

		if len(container.Chapters) == 0 {
			continue
		}

		ret = append(ret, container)
	}

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getDownloadedMangaPageContainer returns the PageContainer for a downloaded manga chapter based on the provider.
func (r *Repository) getDownloadedMangaPageContainer(
	provider string,
	mediaId int,
	chapterId string,
) (*PageContainer, error) {

	// Check if the chapter is downloaded
	found := false

	// Read download directory
	files, err := os.ReadDir(r.downloadDir)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to read download directory")
		return nil, err
	}

	chapterDir := "" // e.g. manga_comick_123_10010_13
	for _, file := range files {
		if file.IsDir() {

			downloadId, ok := chapter_downloader.ParseChapterDirName(file.Name())
			if !ok {
				continue
			}

			if downloadId.Provider == provider &&
				downloadId.MediaId == mediaId &&
				downloadId.ChapterId == chapterId {
				found = true
				chapterDir = file.Name()
				break
			}
		}
	}

	if !found {
		return nil, ErrChapterNotDownloaded
	}

	r.logger.Debug().Msg("manga: Found downloaded chapter directory")

	// Open registry file
	registryFile, err := os.Open(filepath.Join(r.downloadDir, chapterDir, "registry.json"))
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to open registry file")
		return nil, err
	}
	defer registryFile.Close()

	r.logger.Debug().Str("chapterId", chapterId).Msg("manga: Reading registry file")

	// Read registry file
	var pageRegistry *chapter_downloader.Registry
	err = json.NewDecoder(registryFile).Decode(&pageRegistry)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to decode registry file")
		return nil, err
	}

	pageList := make([]*hibikemanga.ChapterPage, 0)
	pageDimensions := make(map[int]*PageDimension)

	// Get the downloaded pages
	for pageIndex, pageInfo := range *pageRegistry {
		pageList = append(pageList, &hibikemanga.ChapterPage{
			Index:    pageIndex,
			URL:      filepath.Join(chapterDir, pageInfo.Filename),
			Provider: provider,
		})
		pageDimensions[pageIndex] = &PageDimension{
			Width:  pageInfo.Width,
			Height: pageInfo.Height,
		}
	}

	slices.SortStableFunc(pageList, func(i, j *hibikemanga.ChapterPage) int {
		return cmp.Compare(i.Index, j.Index)
	})

	container := &PageContainer{
		MediaId:        mediaId,
		Provider:       provider,
		ChapterId:      chapterId,
		Pages:          pageList,
		PageDimensions: pageDimensions,
		IsDownloaded:   true,
	}

	r.logger.Debug().Str("chapterId", chapterId).Msg("manga: Found downloaded chapter")

	return container, nil
}
