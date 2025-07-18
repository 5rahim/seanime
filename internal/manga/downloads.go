package manga

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/hook"
	chapter_downloader "seanime/internal/manga/downloader"
	manga_providers "seanime/internal/manga/providers"
	"slices"

	"github.com/goccy/go-json"
)

// GetDownloadedMangaChapterContainers retrieves downloaded chapter containers for a specific manga ID.
// It filters the complete set of downloaded chapters to return only those matching the provided manga ID.
func (r *Repository) GetDownloadedMangaChapterContainers(mId int, mangaCollection *anilist.MangaCollection) (ret []*ChapterContainer, err error) {

	containers, err := r.GetDownloadedChapterContainers(mangaCollection)
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.MediaId == mId {
			ret = append(ret, container)
		}
	}

	return ret, nil
}

// GetDownloadedChapterContainers retrieves all downloaded manga chapter containers.
// It scans the download directory for chapter folders, matches them with manga collection entries,
// and collects chapter details from file cache or provider API when necessary.
//
// Ideally, the provider API should never be called assuming the chapter details are cached.
func (r *Repository) GetDownloadedChapterContainers(mangaCollection *anilist.MangaCollection) (ret []*ChapterContainer, err error) {
	ret = make([]*ChapterContainer, 0)

	// Trigger hook event
	reqEvent := &MangaDownloadedChapterContainersRequestedEvent{
		MangaCollection:   mangaCollection,
		ChapterContainers: ret,
	}
	err = hook.GlobalHookManager.OnMangaDownloadedChapterContainersRequested().Trigger(reqEvent)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
		return nil, fmt.Errorf("manga: Error in hook, %w", err)
	}
	mangaCollection = reqEvent.MangaCollection

	// Default prevented, return the chapter containers
	if reqEvent.DefaultPrevented {
		ret = reqEvent.ChapterContainers
		if ret == nil {
			return nil, fmt.Errorf("manga: No chapter containers returned by hook event")
		}
		return ret, nil
	}

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
				// Cache the chapter container in the permanent bucket
				go func() {
					chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)
					chapterContainer, found := r.getChapterContainerFromFilecache(provider, mediaId)
					if found {
						// Store the chapter container in the permanent bucket
						permBucket := getPermanentChapterContainerCacheBucket(provider, mediaId)
						_ = r.fileCacher.SetPerm(permBucket, chapterContainerKey, chapterContainer)
					}
				}()
			}
		} else {
			r.logger.Trace().Int("mediaId", mediaId).Msg("manga: Found chapter container in permanent bucket")
		}

		downloadedContainer := &ChapterContainer{
			MediaId:  container.MediaId,
			Provider: container.Provider,
			Chapters: make([]*hibikemanga.ChapterDetails, 0),
		}

		// Now that we have the container, we'll filter out the chapters that are not downloaded
		// Go through each chapter and check if it's downloaded
		for _, chapter := range container.Chapters {
			// For each chapter, check if the chapter directory exists
			for _, dir := range chapterDirs {
				if dir == chapter_downloader.FormatChapterDirName(provider, mediaId, chapter.ID, chapter.Chapter) {
					downloadedContainer.Chapters = append(downloadedContainer.Chapters, chapter)
					break
				}
			}
		}

		if len(downloadedContainer.Chapters) == 0 {
			continue
		}

		ret = append(ret, downloadedContainer)
	}

	// Add chapter containers from local provider
	localProviderB, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, manga_providers.LocalProvider)
	if ok {
		_, ok := localProviderB.GetProvider().(*manga_providers.Local)
		if ok {
			for _, list := range mangaCollection.MediaListCollection.GetLists() {
				for _, entry := range list.GetEntries() {
					media := entry.GetMedia()
					opts := GetMangaChapterContainerOptions{
						Provider: manga_providers.LocalProvider,
						MediaId:  media.GetID(),
						Titles:   media.GetAllTitles(),
						Year:     media.GetStartYearSafe(),
					}
					container, err := r.GetMangaChapterContainer(&opts)
					if err != nil {
						continue
					}
					ret = append(ret, container)
				}
			}
		}
	}

	// Event
	ev := &MangaDownloadedChapterContainersEvent{
		ChapterContainers: ret,
	}
	err = hook.GlobalHookManager.OnMangaDownloadedChapterContainers().Trigger(ev)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
		return nil, fmt.Errorf("manga: Error in hook, %w", err)
	}
	ret = ev.ChapterContainers

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getDownloadedMangaPageContainer retrieves page information for a downloaded manga chapter.
// It reads the chapter directory and parses the registry file to build a PageContainer
// with details about each downloaded page including dimensions and file paths.
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
