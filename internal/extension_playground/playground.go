package extension_playground

import (
	"bytes"
	"fmt"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/extension_repo"
	"seanime/internal/manga"
	"seanime/internal/onlinestream"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"

	goja_runtime "seanime/internal/goja/goja_runtime"
)

type (
	PlaygroundRepository struct {
		logger             *zerolog.Logger
		platform           platform.Platform
		baseAnimeCache     *result.Cache[int, *anilist.BaseAnime]
		baseMangaCache     *result.Cache[int, *anilist.BaseManga]
		metadataProvider   metadata.Provider
		gojaRuntimeManager *goja_runtime.Manager
	}

	RunPlaygroundCodeResponse struct {
		Logs  string `json:"logs"`
		Value string `json:"value"`
	}

	RunPlaygroundCodeParams struct {
		Type     extension.Type         `json:"type"`
		Language extension.Language     `json:"language"`
		Code     string                 `json:"code"`
		Inputs   map[string]interface{} `json:"inputs"`
		Function string                 `json:"function"`
	}
)

func NewPlaygroundRepository(logger *zerolog.Logger, platform platform.Platform, metadataProvider metadata.Provider) *PlaygroundRepository {
	return &PlaygroundRepository{
		logger:             logger,
		platform:           platform,
		metadataProvider:   metadataProvider,
		baseAnimeCache:     result.NewCache[int, *anilist.BaseAnime](),
		baseMangaCache:     result.NewCache[int, *anilist.BaseManga](),
		gojaRuntimeManager: goja_runtime.NewManager(logger, 10),
	}
}

func (r *PlaygroundRepository) RunPlaygroundCode(params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {
	defer util.HandlePanicInModuleWithError("extension_playground/RunPlaygroundCode", &err)

	if params == nil {
		return nil, fmt.Errorf("no parameters provided")
	}

	ext := &extension.Extension{
		ID:          "playground-extension",
		Name:        "Playground",
		Version:     "0.0.0",
		ManifestURI: "",
		Language:    params.Language,
		Type:        params.Type,
		Description: "",
		Author:      "",
		Icon:        "",
		Website:     "",
		Payload:     params.Code,
	}

	r.logger.Debug().Msgf("playground: Inputs: %s", strings.ReplaceAll(spew.Sprint(params.Inputs), "\n", ""))

	switch params.Type {
	case extension.TypeMangaProvider:
		return r.runPlaygroundCodeMangaProvider(ext, params)
	case extension.TypeOnlinestreamProvider:
		return r.runPlaygroundCodeOnlinestreamProvider(ext, params)
	case extension.TypeAnimeTorrentProvider:
		return r.runPlaygroundCodeAnimeTorrentProvider(ext, params)
	default:
	}

	runtime.GC()

	return nil, fmt.Errorf("invalid extension type")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PlaygroundDebugLogger struct {
	logger *zerolog.Logger
	buff   *bytes.Buffer
}

func (r *PlaygroundRepository) newPlaygroundDebugLogger() *PlaygroundDebugLogger {
	buff := &bytes.Buffer{}
	fileOutput := zerolog.ConsoleWriter{
		Out:           buff,
		TimeFormat:    time.DateTime,
		FormatMessage: util.ZerologFormatMessageSimple,
		FormatLevel:   util.ZerologFormatLevelSimple,
		NoColor:       true, // Needed to prevent color codes from being written to the file
	}

	logger := zerolog.New(fileOutput).With().Timestamp().Logger()

	return &PlaygroundDebugLogger{
		logger: &logger,
		buff:   buff,
	}
}

func newPlaygroundResponse(logger *PlaygroundDebugLogger, value interface{}) *RunPlaygroundCodeResponse {
	v := ""

	switch value.(type) {
	case error:
		v = fmt.Sprintf("ERROR: %+v", value)
	case string:
		v = value.(string)
	default:
		// Pretty print the value to json
		prettyJSON, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			v = fmt.Sprintf("ERROR: Failed to marshal value to JSON: %+v", err)
		} else {
			v = string(prettyJSON)
		}
	}

	logs := logger.buff.String()

	return &RunPlaygroundCodeResponse{
		Logs:  logs,
		Value: v,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *PlaygroundRepository) getAnime(mediaId int) (anime *anilist.BaseAnime, am *metadata.AnimeMetadata, err error) {
	var ok bool
	anime, ok = r.baseAnimeCache.Get(mediaId)
	if !ok {
		anime, err = r.platform.GetAnime(mediaId)
		if err != nil {
			return nil, nil, err
		}
		r.baseAnimeCache.SetT(mediaId, anime, 24*time.Hour)
	}

	am, _ = r.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
	return anime, am, nil
}

func (r *PlaygroundRepository) getManga(mediaId int) (manga *anilist.BaseManga, err error) {
	var ok bool
	manga, ok = r.baseMangaCache.Get(mediaId)
	if !ok {
		manga, err = r.platform.GetManga(mediaId)
		if err != nil {
			return nil, err
		}
		r.baseMangaCache.SetT(mediaId, manga, 24*time.Hour)
	}
	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *PlaygroundRepository) runPlaygroundCodeAnimeTorrentProvider(ext *extension.Extension, params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {

	logger := r.newPlaygroundDebugLogger()

	// Inputs
	// - mediaId int
	// - options struct

	mediaId, ok := params.Inputs["mediaId"].(float64)
	if !ok || mediaId <= 0 {
		return nil, fmt.Errorf("invalid mediaId")
	}

	// Fetch the anime
	anime, animeMetadata, err := r.getAnime(int(mediaId))
	if err != nil {
		return nil, err
	}

	queryMedia := hibiketorrent.Media{
		ID:                   anime.GetID(),
		IDMal:                anime.GetIDMal(),
		Status:               string(*anime.GetStatus()),
		Format:               string(*anime.GetFormat()),
		EnglishTitle:         anime.GetTitle().GetEnglish(),
		RomajiTitle:          anime.GetRomajiTitleSafe(),
		EpisodeCount:         anime.GetTotalEpisodeCount(),
		AbsoluteSeasonOffset: 0,
		Synonyms:             anime.GetSynonymsContainingSeason(),
		IsAdult:              *anime.GetIsAdult(),
		StartDate: &hibiketorrent.FuzzyDate{
			Year:  *anime.GetStartDate().GetYear(),
			Month: anime.GetStartDate().GetMonth(),
			Day:   anime.GetStartDate().GetDay(),
		},
	}

	switch params.Language {
	case extension.LanguageGo:
	//...
	case extension.LanguageJavascript, extension.LanguageTypescript:
		_, provider, err := extension_repo.NewGojaAnimeTorrentProvider(ext, params.Language, logger.logger, r.gojaRuntimeManager)
		if err != nil {
			return nil, err
		}
		// defer provider.GetVM().ClearInterrupt()

		// Run the code
		switch params.Function {
		case "search":
			res, err := provider.Search(hibiketorrent.AnimeSearchOptions{
				Media: queryMedia,
				Query: params.Inputs["query"].(string),
			})
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		case "smartSearch":
			type p struct {
				Query         string `json:"query"`
				Batch         bool   `json:"batch"`
				EpisodeNumber int    `json:"episodeNumber"`
				Resolution    string `json:"resolution"`
				BestReleases  bool   `json:"bestReleases"`
			}
			m, _ := json.Marshal(params.Inputs["options"])
			var options p
			_ = json.Unmarshal(m, &options)

			anidbAID := 0
			anidbEID := 0

			// Get the AniDB Anime ID and Episode ID
			if animeMetadata != nil {
				// Override absolute offset value of queryMedia
				queryMedia.AbsoluteSeasonOffset = animeMetadata.GetOffset()

				if animeMetadata.GetMappings() != nil {

					anidbAID = animeMetadata.GetMappings().AnidbId
					// Find Anizip Episode based on inputted episode number
					anizipEpisode, found := animeMetadata.FindEpisode(strconv.Itoa(options.EpisodeNumber))
					if found {
						anidbEID = anizipEpisode.AnidbEid
					}
				}
			}

			res, err := provider.SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
				Media:         queryMedia,
				Query:         options.Query,
				Batch:         options.Batch,
				EpisodeNumber: options.EpisodeNumber,
				Resolution:    options.Resolution,
				BestReleases:  options.BestReleases,
				AnidbAID:      anidbAID,
				AnidbEID:      anidbEID,
			})
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		case "getTorrentInfoHash":
			var torrent hibiketorrent.AnimeTorrent
			_ = json.Unmarshal([]byte(params.Inputs["torrent"].(string)), &torrent)

			res, err := provider.GetTorrentInfoHash(&torrent)
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		case "getTorrentMagnetLink":
			var torrent hibiketorrent.AnimeTorrent
			_ = json.Unmarshal([]byte(params.Inputs["torrent"].(string)), &torrent)

			res, err := provider.GetTorrentMagnetLink(&torrent)
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		case "getLatest":
			res, err := provider.GetLatest()
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		case "getSettings":
			res := provider.GetSettings()
			return newPlaygroundResponse(logger, res), nil
		}
	}

	return nil, fmt.Errorf("unknown call")
}

func (r *PlaygroundRepository) runPlaygroundCodeMangaProvider(ext *extension.Extension, params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {

	logger := r.newPlaygroundDebugLogger()

	mediaId, ok := params.Inputs["mediaId"].(float64)
	if !ok || mediaId <= 0 {
		return nil, fmt.Errorf("invalid mediaId")
	}

	media, err := r.getManga(int(mediaId))
	if err != nil {
		return nil, err
	}

	titles := media.GetAllTitles()

	switch params.Language {
	case extension.LanguageGo:
	//...
	case extension.LanguageJavascript, extension.LanguageTypescript:
		_, provider, err := extension_repo.NewGojaMangaProvider(ext, params.Language, logger.logger, r.gojaRuntimeManager)
		if err != nil {
			return newPlaygroundResponse(logger, err), nil
		}
		// defer provider.GetVM().ClearInterrupt()

		// Run the code
		switch params.Function {
		case "search":
			// Search
			y := 0
			if media.GetStartDate().GetYear() != nil {
				y = *media.GetStartDate().GetYear()
			}

			ret := make([]*hibikemanga.SearchResult, 0)
			for _, title := range titles {
				res, err := provider.Search(hibikemanga.SearchOptions{
					Query: *title,
					Year:  y,
				})
				if err != nil {
					logger.logger.Error().Err(err).Msgf("playground: Search failed for title \"%s\"", *title)
				}
				manga.HydrateSearchResultSearchRating(res, title)
				ret = append(ret, res...)
			}

			var selected *hibikemanga.SearchResult
			if len(ret) > 0 {
				selected = manga.GetBestSearchResult(ret)
			}

			return newPlaygroundResponse(logger, selected), nil

		case "findChapters":
			res, err := provider.FindChapters(params.Inputs["id"].(string))
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil

		case "findChapterPages":
			res, err := provider.FindChapterPages(params.Inputs["id"].(string))
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		}
	}

	return nil, fmt.Errorf("unknown call")
}

func (r *PlaygroundRepository) runPlaygroundCodeOnlinestreamProvider(ext *extension.Extension, params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {

	logger := r.newPlaygroundDebugLogger()

	mediaId, ok := params.Inputs["mediaId"].(float64)
	if !ok || mediaId <= 0 {
		return nil, fmt.Errorf("invalid mediaId")
	}

	// Fetch the anime
	anime, _, err := r.getAnime(int(mediaId))
	if err != nil {
		return nil, err
	}

	titles := anime.GetAllTitles()

	switch params.Language {
	case extension.LanguageGo:
	//...
	case extension.LanguageJavascript, extension.LanguageTypescript:
		_, provider, err := extension_repo.NewGojaOnlinestreamProvider(ext, params.Language, logger.logger, r.gojaRuntimeManager)
		if err != nil {
			return newPlaygroundResponse(logger, err), nil
		}
		// defer provider.GetVM().ClearInterrupt()

		// Run the code
		switch params.Function {
		case "search":
			// Search - params: dub: boolean
			ret := make([]*hibikeonlinestream.SearchResult, 0)
			for _, title := range titles {
				res, err := provider.Search(hibikeonlinestream.SearchOptions{
					Query: *title,
					Dub:   params.Inputs["dub"].(bool),
					Year:  anime.GetStartYearSafe(),
				})
				if err != nil {
					logger.logger.Error().Err(err).Msgf("playground: Search failed for title \"%s\"", *title)
				}
				ret = append(ret, res...)
			}

			bestRes := onlinestream.GetBestSearchResult(ret, titles)

			return newPlaygroundResponse(logger, bestRes), nil

		case "findEpisodes":
			// FindEpisodes - params: id: string
			res, err := provider.FindEpisodes(params.Inputs["id"].(string))
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil

		case "findEpisodeServer":
			// FindEpisodeServer - params: episode: EpisodeDetails, server: string
			var episode hibikeonlinestream.EpisodeDetails
			_ = json.Unmarshal([]byte(params.Inputs["episode"].(string)), &episode)

			res, err := provider.FindEpisodeServer(&episode, params.Inputs["server"].(string))
			if err != nil {
				return newPlaygroundResponse(logger, err), nil
			}
			return newPlaygroundResponse(logger, res), nil
		}
	}

	return nil, fmt.Errorf("unknown call")
}
