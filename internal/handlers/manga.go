package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/manga"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	baseMangaCache    = result.NewCache[int, *anilist.BaseManga]()
	mangaDetailsCache = result.NewCache[int, *anilist.MangaDetailsById_Media]()
)

// HandleGetMangaPreferences
//
//	@summary returns server-backed manga source preferences.
//	@route /api/v1/manga/preferences [GET]
//	@returns manga.MangaPreferences
func (h *Handler) HandleGetMangaPreferences(c echo.Context) error {
	preferences, err := h.App.MangaRepository.GetMangaPreferences()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, preferences)
}

// HandleImportMangaPreferences
//
//	@summary imports client manga preferences missing from the server.
//	@route /api/v1/manga/preferences/import [POST]
//	@returns manga.MangaPreferences
func (h *Handler) HandleImportMangaPreferences(c echo.Context) error {
	var body manga.MangaPreferences
	if err := c.Bind(&body); err != nil {
		return h.RespondWithStatusError(c, http.StatusBadRequest, err)
	}
	preferences, err := h.App.MangaRepository.ImportPreferences(&body)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, preferences)
}

// HandlePatchMangaPreference
//
//	@summary updates a manga source preference.
//	@route /api/v1/manga/preferences/{mediaId} [PATCH]
//	@returns manga.MangaEntryPreference
func (h *Handler) HandlePatchMangaPreference(c echo.Context) error {
	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil || mediaId <= 0 {
		return h.RespondWithStatusError(c, http.StatusBadRequest, errors.New("invalid media id"))
	}

	var body manga.MangaPreferencePatch
	if err := c.Bind(&body); err != nil {
		return h.RespondWithStatusError(c, http.StatusBadRequest, err)
	}
	preference, err := h.App.MangaRepository.PatchPreference(mediaId, &body, true)
	if err != nil {
		return h.RespondWithStatusError(c, http.StatusBadRequest, err)
	}
	return h.RespondWithData(c, preference)
}

// HandleStartMangaSourceRefresh
//
//	@summary starts a background manga source refresh.
//	@route /api/v1/manga/source-refresh [POST]
//	@returns manga.MangaSourceRefreshJob
func (h *Handler) HandleStartMangaSourceRefresh(c echo.Context) error {
	type body struct {
		Mode manga.MangaSourceRefreshMode `json:"mode"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithStatusError(c, http.StatusBadRequest, err)
	}
	if !manga.IsMangaSourceRefreshModeValid(b.Mode) {
		return h.RespondWithStatusError(c, http.StatusBadRequest, errors.New("invalid manga source refresh mode"))
	}
	clientId := getContextClientId(c)
	job, err := h.App.MangaRepository.GetActiveMangaSourceRefresh(clientId)
	if err != nil {
		return h.RespondWithStatusError(c, http.StatusConflict, err)
	}
	if job != nil {
		return h.RespondWithData(c, job)
	}
	collection, err := h.App.GetMangaCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	job, err = h.App.MangaRepository.StartMangaSourceRefresh(clientId, b.Mode, collection)
	if err != nil {
		if errors.Is(err, manga.ErrMangaSourceRefreshConflict) {
			return h.RespondWithStatusError(c, http.StatusConflict, err)
		}
		if errors.Is(err, manga.ErrNoMangaProviders) {
			return h.RespondWithStatusError(c, http.StatusBadRequest, err)
		}
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, job)
}

// HandleGetMangaSourceRefresh
//
//	@summary returns the client's active or latest manga source refresh.
//	@route /api/v1/manga/source-refresh [GET]
//	@returns manga.MangaSourceRefreshJob
func (h *Handler) HandleGetMangaSourceRefresh(c echo.Context) error {
	return h.RespondWithData(c, h.App.MangaRepository.GetMangaSourceRefresh(getContextClientId(c)))
}

// HandleStopMangaSourceRefresh
//
//	@summary stops or dismisses the client's manga source refresh.
//	@route /api/v1/manga/source-refresh [DELETE]
//	@returns manga.MangaSourceRefreshJob
func (h *Handler) HandleStopMangaSourceRefresh(c echo.Context) error {
	job, err := h.App.MangaRepository.StopMangaSourceRefresh(getContextClientId(c))
	if err != nil {
		if errors.Is(err, manga.ErrMangaSourceRefreshConflict) {
			return h.RespondWithStatusError(c, http.StatusConflict, err)
		}
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, job)
}

// HandleGetAnilistMangaCollection
//
//	@summary returns the user's AniList manga collection.
//	@route /api/v1/manga/anilist/collection [GET]
//	@returns anilist.MangaCollection
func (h *Handler) HandleGetAnilistMangaCollection(c echo.Context) error {

	type body struct {
		BypassCache bool `json:"bypassCache"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	collection, err := h.App.GetMangaCollection(b.BypassCache)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, collection)
}

// HandleGetRawAnilistMangaCollection
//
//	@summary returns the user's AniList manga collection.
//	@route /api/v1/manga/anilist/collection/raw [GET,POST]
//	@returns anilist.MangaCollection
func (h *Handler) HandleGetRawAnilistMangaCollection(c echo.Context) error {

	bypassCache := c.Request().Method == "POST"

	// Get the user's anilist collection
	mangaCollection, err := h.App.GetRawMangaCollection(bypassCache)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, mangaCollection)
}

var mangaTagsCache *anilist.MediaTagMap

// HandleGetRawAnilistMangaCollectionTags
//
//	@summary returns the AniList tags for the user's raw manga collection.
//	@desc This runs a dedicated AniList tags query used by the lists page filters.
//	@route /api/v1/manga/anilist/collection/raw/tags [GET]
//	@returns anilist.MediaTagMap
func (h *Handler) HandleGetRawAnilistMangaCollectionTags(c echo.Context) error {
	h.App.OnRefreshAnilistCollectionFuncs.Set("HandleGetRawAnilistMangaCollectionTags", func() {
		mangaTagsCache = nil
	})

	if mangaTagsCache != nil {
		return h.RespondWithData(c, *mangaTagsCache)
	}

	userName := h.App.GetUsername()
	if userName == "" || h.App.GetUser().IsSimulated {
		return h.RespondWithData(c, anilist.MediaTagMap{})
	}

	ret, err := h.App.AnilistPlatformRef.Get().GetAnilistClient().MangaCollectionTags(c.Request().Context(), &userName)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	tags := anilist.MediaTagMapFromMangaCollectionTags(ret)
	mangaTagsCache = &tags

	return h.RespondWithData(c, tags)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetMangaCollection
//
//	@summary returns the user's main manga collection.
//	@desc This is an object that contains all the user's manga entries in a structured format.
//	@route /api/v1/manga/collection [GET]
//	@returns manga.Collection
func (h *Handler) HandleGetMangaCollection(c echo.Context) error {

	animeCollection, err := h.App.GetMangaCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	collection, err := manga.NewCollection(&manga.NewCollectionOptions{
		MangaCollection: animeCollection,
		PlatformRef:     h.App.AnilistPlatformRef,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, collection)
}

// HandleGetMangaEntry
//
//	@summary returns a manga entry for the given AniList manga id.
//	@desc This is used by the manga media entry pages to get all the data about the anime. It includes metadata and AniList list data.
//	@route /api/v1/manga/entry/{id} [GET]
//	@param id - int - true - "AniList manga media ID"
//	@returns manga.Entry
func (h *Handler) HandleGetMangaEntry(c echo.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	animeCollection, err := h.App.GetMangaCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	entry, err := manga.NewEntry(c.Request().Context(), &manga.NewEntryOptions{
		MediaId:         id,
		Logger:          h.App.Logger,
		FileCacher:      h.App.FileCacher,
		PlatformRef:     h.App.AnilistPlatformRef,
		MangaCollection: animeCollection,
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if entry != nil {
		baseMangaCache.SetT(entry.MediaId, entry.Media, 1*time.Hour)
	}

	return h.RespondWithData(c, entry)
}

// HandleGetMangaEntryDetails
//
//	@summary returns more details about an AniList manga entry.
//	@desc This fetches more fields omitted from the base queries.
//	@route /api/v1/manga/entry/{id}/details [GET]
//	@param id - int - true - "AniList manga media ID"
//	@returns anilist.MangaDetailsById_Media
func (h *Handler) HandleGetMangaEntryDetails(c echo.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if detailsMedia, found := mangaDetailsCache.Get(id); found {
		return h.RespondWithData(c, detailsMedia)
	}

	details, err := h.App.AnilistPlatformRef.Get().GetMangaDetails(c.Request().Context(), id)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	mangaDetailsCache.SetT(id, details, 1*time.Hour)

	return h.RespondWithData(c, details)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetMangaLatestChapterNumbersMap
//
//	@summary returns the latest chapter number for all manga entries.
//	@route /api/v1/manga/latest-chapter-numbers [GET]
//	@returns map[int][]manga.MangaLatestChapterNumberItem
func (h *Handler) HandleGetMangaLatestChapterNumbersMap(c echo.Context) error {
	ret, err := h.App.MangaRepository.GetMangaLatestChapterNumbersMap()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

// HandleRefetchMangaChapterContainers
//
//	@summary refetches the chapter containers for all manga entries previously cached.
//	@route /api/v1/manga/refetch-chapter-containers [POST]
//	@returns bool
func (h *Handler) HandleRefetchMangaChapterContainers(c echo.Context) error {

	type body struct {
		SelectedProviderMap map[int]string `json:"selectedProviderMap"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	mangaCollection, err := h.App.GetMangaCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	err = h.App.MangaRepository.RefreshChapterContainers(mangaCollection, b.SelectedProviderMap)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return nil
}

// HandleEmptyMangaEntryCache
//
//	@summary empties the cache for a manga entry.
//	@desc This will empty the cache for a manga entry (chapter lists and pages), allowing the client to fetch fresh data.
//	@desc HandleGetMangaEntryChapters should be called after this to fetch the new chapter list.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/manga/entry/cache [DELETE]
//	@returns bool
func (h *Handler) HandleEmptyMangaEntryCache(c echo.Context) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.MangaRepository.EmptyMangaCache(b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleGetMangaEntryChapters
//
//	@summary returns the chapters for a manga entry based on the provider.
//	@route /api/v1/manga/chapters [POST]
//	@returns manga.ChapterContainer
func (h *Handler) HandleGetMangaEntryChapters(c echo.Context) error {

	type body struct {
		MediaId  int    `json:"mediaId"`
		Provider string `json:"provider"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	var titles []*string
	baseManga, found := baseMangaCache.Get(b.MediaId)
	if !found {
		var err error
		baseManga, err = h.App.AnilistPlatformRef.Get().GetManga(c.Request().Context(), b.MediaId)
		if err != nil {
			return h.RespondWithError(c, err)
		}
		titles = baseManga.GetAllTitles()
		baseMangaCache.SetT(b.MediaId, baseManga, 24*time.Hour)
	} else {
		titles = baseManga.GetAllTitles()
	}

	container, err := h.App.MangaRepository.GetMangaChapterContainer(&manga.GetMangaChapterContainerOptions{
		Provider: b.Provider,
		MediaId:  b.MediaId,
		Titles:   titles,
		Year:     baseManga.GetStartYearSafe(),
	})
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, container)
}

// HandleGetMangaEntryPages
//
//	@summary returns the pages for a manga entry based on the provider and chapter id.
//	@desc This will return the pages for a manga chapter.
//	@desc If the app is offline and the chapter is not downloaded, it will return an error.
//	@desc If the app is online and the chapter is not downloaded, it will return the pages from the provider.
//	@desc If the chapter is downloaded, it will return the appropriate struct.
//	@desc If 'double page' is requested, it will fetch image sizes and include the dimensions in the response.
//	@route /api/v1/manga/pages [POST]
//	@returns manga.PageContainer
func (h *Handler) HandleGetMangaEntryPages(c echo.Context) error {

	type body struct {
		MediaId    int    `json:"mediaId"`
		Provider   string `json:"provider"`
		ChapterId  string `json:"chapterId"`
		DoublePage bool   `json:"doublePage"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	container, err := h.App.MangaRepository.GetMangaPageContainer(b.Provider, b.MediaId, b.ChapterId, b.DoublePage, h.App.IsOfflineRef())
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, container)
}

// HandleGetMangaEntryDownloadedChapters
//
//	@summary returns all download chapters for a manga entry,
//	@route /api/v1/manga/downloaded-chapters/{id} [GET]
//	@param id - int - true - "AniList manga media ID"
//	@returns []manga.ChapterContainer
func (h *Handler) HandleGetMangaEntryDownloadedChapters(c echo.Context) error {

	mId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	mangaCollection, err := h.App.GetMangaCollection(false)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	container, err := h.App.MangaRepository.GetDownloadedMangaChapterContainers(mId, mangaCollection)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, container)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	anilistListMangaCache = result.NewCache[string, *anilist.ListManga]()
)

// HandleAnilistListManga
//
//	@summary returns a list of manga based on the search parameters.
//	@desc This is used by "Advanced Search" and search function.
//	@route /api/v1/manga/anilist/list [POST]
//	@returns anilist.ListManga
func (h *Handler) HandleAnilistListManga(c echo.Context) error {

	type body struct {
		Page                *int                   `json:"page,omitempty"`
		Search              *string                `json:"search,omitempty"`
		PerPage             *int                   `json:"perPage,omitempty"`
		Sort                []*anilist.MediaSort   `json:"sort,omitempty"`
		Status              []*anilist.MediaStatus `json:"status,omitempty"`
		Genres              []*string              `json:"genres,omitempty"`
		Tags                []*string              `json:"tags,omitempty"`
		AverageScoreGreater *int                   `json:"averageScore_greater,omitempty"`
		Year                *int                   `json:"year,omitempty"`
		CountryOfOrigin     *string                `json:"countryOfOrigin,omitempty"`
		IsAdult             *bool                  `json:"isAdult,omitempty"`
		Format              *anilist.MediaFormat   `json:"format,omitempty"`
	}

	p := new(body)
	if err := c.Bind(p); err != nil {
		return h.RespondWithError(c, err)
	}

	if p.Page == nil || p.PerPage == nil {
		*p.Page = 1
		*p.PerPage = 20
	}

	var isAdult *bool = nil
	if p.IsAdult != nil {
		isAdult = new(*p.IsAdult && h.App.Settings.GetAnilist().EnableAdultContent)
	}

	cacheKey := anilist.ListMangaCacheKey(
		p.Page,
		p.Search,
		p.PerPage,
		p.Sort,
		p.Status,
		p.Genres,
		p.Tags,
		p.AverageScoreGreater,
		nil,
		p.Year,
		p.Format,
		p.CountryOfOrigin,
		isAdult,
	)

	cached, ok := anilistListMangaCache.Get(cacheKey)
	if ok {
		return h.RespondWithData(c, cached)
	}

	ret, err := anilist.ListMangaM(
		h.App.AnilistPlatformRef.Get().GetAnilistClient(),
		p.Page,
		p.Search,
		p.PerPage,
		p.Sort,
		p.Status,
		p.Genres,
		p.Tags,
		p.AverageScoreGreater,
		p.Year,
		p.Format,
		p.CountryOfOrigin,
		isAdult,
		h.App.Logger,
		h.App.GetUserAnilistToken(),
	)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if ret != nil {
		anilistListMangaCache.SetT(cacheKey, ret, time.Minute*10)
	}

	return h.RespondWithData(c, ret)
}

// HandleUpdateMangaProgress
//
//	@summary updates the progress of a manga entry.
//	@desc Note: MyAnimeList is not supported
//	@route /api/v1/manga/update-progress [POST]
//	@returns bool
func (h *Handler) HandleUpdateMangaProgress(c echo.Context) error {

	type body struct {
		MediaId       int `json:"mediaId"`
		MalId         int `json:"malId,omitempty"`
		ChapterNumber int `json:"chapterNumber"`
		TotalChapters int `json:"totalChapters"`
	}

	b := new(body)
	if err := c.Bind(b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Update the progress on AniList
	err := h.App.AnilistPlatformRef.Get().UpdateEntryProgress(
		c.Request().Context(),
		b.MediaId,
		b.ChapterNumber,
		&b.TotalChapters,
	)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	_, _ = h.App.RefreshMangaCollection() // Refresh the AniList collection

	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleMangaManualSearch
//
//	@summary returns search results for a manual search.
//	@desc Returns search results for a manual search.
//	@route /api/v1/manga/search [POST]
//	@returns []hibikemanga.SearchResult
func (h *Handler) HandleMangaManualSearch(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		Query    string `json:"query"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	ret, err := h.App.MangaRepository.ManualSearch(b.Provider, b.Query)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, ret)
}

// HandleMangaManualMapping
//
//	@summary manually maps a manga entry to a manga ID from the provider.
//	@desc This is used to manually map a manga entry to a manga ID from the provider.
//	@desc The client should re-fetch the chapter container after this.
//	@route /api/v1/manga/manual-mapping [POST]
//	@returns bool
func (h *Handler) HandleMangaManualMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
		MangaId  string `json:"mangaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.MangaRepository.ManualMapping(b.Provider, b.MediaId, b.MangaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleGetMangaMapping
//
//	@summary returns the mapping for a manga entry.
//	@desc This is used to get the mapping for a manga entry.
//	@desc An empty string is returned if there's no manual mapping. If there is, the manga ID will be returned.
//	@route /api/v1/manga/get-mapping [POST]
//	@returns manga.MappingResponse
func (h *Handler) HandleGetMangaMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	mapping := h.App.MangaRepository.GetMapping(b.Provider, b.MediaId)
	return h.RespondWithData(c, mapping)
}

// HandleRemoveMangaMapping
//
//	@summary removes the mapping for a manga entry.
//	@desc This is used to remove the mapping for a manga entry.
//	@desc The client should re-fetch the chapter container after this.
//	@route /api/v1/manga/remove-mapping [POST]
//	@returns bool
func (h *Handler) HandleRemoveMangaMapping(c echo.Context) error {

	type body struct {
		Provider string `json:"provider"`
		MediaId  int    `json:"mediaId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.MangaRepository.RemoveMapping(b.Provider, b.MediaId)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetLocalMangaPage
//
//	@summary returns a local manga page.
//	@route /api/v1/manga/local-page/{path} [GET]
//	@returns manga.PageContainer
func (h *Handler) HandleGetLocalMangaPage(c echo.Context) error {
	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	path := c.Param("path")
	path, err := url.PathUnescape(path)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	path = strings.TrimPrefix(path, manga_providers.LocalServePath)

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](h.App.ExtensionRepository.GetExtensionBank(), manga_providers.LocalProvider)
	if !ok {
		return h.RespondWithError(c, errors.New("manga: Local provider not found"))
	}

	localProvider, ok := providerExtension.GetProvider().(*manga_providers.Local)
	if !ok {
		return h.RespondWithError(c, errors.New("manga: Local provider not found"))
	}

	reader, err := localProvider.ReadPage(path)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	headers := c.Response().Header()
	headers.Set("Access-Control-Allow-Origin", "*")
	headers.Set("Cross-Origin-Resource-Policy", "cross-origin")

	return c.Stream(http.StatusOK, "image/jpeg", reader)
}
