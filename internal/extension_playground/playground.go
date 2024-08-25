package extension_playground

import (
	"bytes"
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/extension"
	"seanime/internal/extension_repo"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"time"
)

type (
	PlaygroundRepository struct {
		logger           *zerolog.Logger
		platform         platform.Platform
		baseAnimeCache   *result.Cache[int, *anilist.BaseAnime]
		anizipMediaCache *result.Cache[int, *anizip.Media]
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

func NewPlaygroundRepository(logger *zerolog.Logger, platform platform.Platform) *PlaygroundRepository {
	return &PlaygroundRepository{
		logger:           logger,
		platform:         platform,
		baseAnimeCache:   result.NewCache[int, *anilist.BaseAnime](),
		anizipMediaCache: result.NewCache[int, *anizip.Media](),
	}
}

func (r *PlaygroundRepository) RunPlaygroundCode(params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {
	defer util.HandlePanicInModuleWithError("extension_playground/RunPlaygroundCode", &err)

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

	switch params.Type {
	case extension.TypeMangaProvider:
		//return r.runPlaygroundCodeMangaProvider(ext, params)
	case extension.TypeOnlinestreamProvider:
		//return r.runPlaygroundCodeOnlinestreamProvider(params)
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

func (r *PlaygroundRepository) getAnime(mediaId int) (anime *anilist.BaseAnime, am *anizip.Media, err error) {
	var ok bool
	anime, ok = r.baseAnimeCache.Get(mediaId)
	if !ok {
		anime, err = r.platform.GetAnime(mediaId)
		if err != nil {
			return nil, nil, err
		}
		r.baseAnimeCache.SetT(mediaId, anime, 24*time.Hour)
	}

	am, ok = r.anizipMediaCache.Get(mediaId)
	if !ok {
		am, _ = anizip.FetchAniZipMedia("anilist", mediaId)
		r.anizipMediaCache.SetT(mediaId, am, 24*time.Hour)
	}
	return anime, am, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *PlaygroundRepository) runPlaygroundCodeAnimeTorrentProvider(ext *extension.Extension, params *RunPlaygroundCodeParams) (resp *RunPlaygroundCodeResponse, err error) {

	logger := r.newPlaygroundDebugLogger()

	// Inputs
	// - mediaId int
	// - options struct

	r.logger.Debug().Msgf("playground: Inputs: %s", strings.ReplaceAll(spew.Sprint(params.Inputs), "\n", ""))

	mediaId, ok := params.Inputs["mediaId"].(float64)
	if !ok || mediaId <= 0 {
		return nil, fmt.Errorf("invalid mediaId")
	}

	// Fetch the anime
	anime, anizipMedia, err := r.getAnime(int(mediaId))
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
		_, provider, err := extension_repo.NewGojaAnimeTorrentProvider(ext, params.Language, logger.logger)
		if err != nil {
			return nil, err
		}
		defer provider.GetVM().ClearInterrupt()

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

			spew.Dump(params.Inputs["options"])

			anidbAID := 0
			anidbEID := 0

			// Get the AniDB Anime ID and Episode ID
			if anizipMedia != nil {
				// Override absolute offset value of queryMedia
				queryMedia.AbsoluteSeasonOffset = anizipMedia.GetOffset()

				if anizipMedia.GetMappings() != nil {

					anidbAID = anizipMedia.GetMappings().AnidbID
					// Find Anizip Episode based on inputted episode number
					anizipEpisode, found := anizipMedia.FindEpisode(strconv.Itoa(options.EpisodeNumber))
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

	switch params.Language {
	case extension.LanguageGo:
	//...
	case extension.LanguageJavascript, extension.LanguageTypescript:
		_, provider, err := extension_repo.NewGojaAnimeTorrentProvider(ext, params.Language, logger.logger)
		if err != nil {
			return newPlaygroundResponse(logger, err), nil
		}
		defer provider.GetVM().ClearInterrupt()

		_ = provider
	}

	return nil, fmt.Errorf("unknown call")
}
