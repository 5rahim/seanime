package metadata

import (
	"errors"
	"fmt"
	"regexp"
	"seanime/internal/api/anilist"
	"seanime/internal/api/mappings"
	"seanime/internal/api/tvdb"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	AnimeWrapperImpl struct {
		metadata   mo.Option[*AnimeMetadata]
		baseAnime  *anilist.BaseAnime
		fileCacher *filecache.Cacher
		logger     *zerolog.Logger
		// TVDB
		tvdbEpisodes []*tvdb.Episode
	}
)

func (aw *AnimeWrapperImpl) GetEpisodeMetadata(epNum int) (ret EpisodeMetadata) {
	if aw == nil || aw.baseAnime == nil {
		return
	}

	ret = EpisodeMetadata{
		AnidbId:               0,
		TvdbId:                0,
		Title:                 "",
		Image:                 "",
		AirDate:               "",
		Length:                0,
		Summary:               "",
		Overview:              "",
		EpisodeNumber:         epNum,
		Episode:               strconv.Itoa(epNum),
		SeasonNumber:          0,
		AbsoluteEpisodeNumber: 0,
		AnidbEid:              0,
	}

	defer util.HandlePanicInModuleThen("api/metadata/GetEpisodeMetadata", func() {})

	reqEvent := &AnimeEpisodeMetadataRequestedEvent{}
	reqEvent.MediaId = aw.baseAnime.GetID()
	reqEvent.EpisodeNumber = epNum
	reqEvent.EpisodeMetadata = &ret
	hook.GlobalHookManager.OnAnimeEpisodeMetadataRequested().Trigger(reqEvent)
	epNum = reqEvent.EpisodeNumber

	// Default prevented by hook, return the metadata
	if reqEvent.DefaultPrevented {
		if reqEvent.EpisodeMetadata == nil {
			return ret
		}
		return *reqEvent.EpisodeMetadata
	}

	//
	// Process
	//

	// Get TVDB metadata
	hasTVDBMetadata := aw.tvdbEpisodes != nil && len(aw.tvdbEpisodes) > 0

	episode := mo.None[*EpisodeMetadata]()
	if aw.metadata.IsAbsent() {
		ret.Image = aw.baseAnime.GetBannerImageSafe()
	} else {
		episodeF, found := aw.metadata.MustGet().FindEpisode(strconv.Itoa(epNum))
		if found {
			episode = mo.Some(episodeF)
		}
	}

	// If we don't have AniZip metadata, just return the metadata containing the image
	if episode.IsAbsent() {
		return ret
	}

	ret = *episode.MustGet()

	// If TVDB metadata is available, use it to populate the image
	if hasTVDBMetadata {
		tvdbEpisode, found := aw.GetTVDBEpisodeByNumber(epNum)
		if found {
			ret.Image = tvdbEpisode.Image
			ret.TvdbId = int(tvdbEpisode.ID)
		}
	}

	// If TVDB image is not set, use AniZip image, if that is not set, use the AniList banner image
	if ret.Image == "" {
		// Set AniZip image if TVDB image is not set
		if episode.MustGet().Image != "" {
			ret.Image = episode.MustGet().Image
		} else {
			// If AniZip image is not set, use the base media image
			ret.Image = aw.baseAnime.GetBannerImageSafe()
		}
	}

	// Event
	event := &AnimeEpisodeMetadataEvent{}
	event.EpisodeMetadata = &ret
	event.EpisodeNumber = epNum
	event.MediaId = aw.baseAnime.GetID()
	hook.GlobalHookManager.OnAnimeEpisodeMetadataEvent().Trigger(event)
	if event.EpisodeMetadata == nil {
		return ret
	}
	ret = *event.EpisodeMetadata

	return ret
}

func getTvdbIDFromAnimeLists(anidbID int) (tvdbID int, ok bool) {
	res, err := mappings.GetReducedAnimeLists()
	if err != nil {
		return 0, false
	}
	return res.FindTvdbIDFromAnidbID(anidbID)
}

func (aw *AnimeWrapperImpl) EmptyTVDBEpisodesBucket(mediaId int) error {

	if aw.metadata.IsAbsent() {
		return nil
	}

	// Get TVDB ID
	var tvdbId int
	tvdbId = aw.metadata.MustGet().Mappings.ThetvdbId
	if tvdbId == 0 {
		if aw.metadata.MustGet().Mappings.AnidbId > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(aw.metadata.MustGet().Mappings.AnidbId)
		}
	}

	if tvdbId == 0 {
		return errors.New("metadata: could not find tvdb id")
	}

	return aw.fileCacher.Remove(fmt.Sprintf("tvdb_episodes_%d", mediaId))
}

func (aw *AnimeWrapperImpl) GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error) {
	key := aw.baseAnime.GetID()

	if aw.metadata.IsAbsent() {
		return nil, errors.New("metadata: anime metadata is absent")
	}

	// Get TVDB ID
	var tvdbId int
	tvdbId = aw.metadata.MustGet().Mappings.ThetvdbId
	if tvdbId == 0 {
		if aw.metadata.MustGet().Mappings.AnidbId > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(aw.metadata.MustGet().Mappings.AnidbId)
		}
	}

	if tvdbId == 0 {
		return nil, errors.New("metadata: could not find tvdb id")
	}

	bucket := filecache.NewBucket(fmt.Sprintf("tvdb_episodes_%d", aw.baseAnime.GetID()), time.Hour*24*7*365)

	// Find episodes in cache
	var episodes []*tvdb.Episode
	found, _ := aw.fileCacher.Get(bucket, strconv.Itoa(key), &episodes)
	if !populate && found && episodes != nil {
		return episodes, nil
	}

	// Fetch episodes only if we need to populate
	if populate {
		var err error

		tv := tvdb.NewTVDB(&tvdb.NewTVDBOptions{
			ApiKey: "", // Empty
			Logger: aw.logger,
		})

		episodes, err = tv.FetchSeriesEpisodes(tvdbId, tvdb.FilterEpisodeMediaInfo{
			Year:           aw.baseAnime.GetStartDate().GetYear(),
			Month:          aw.baseAnime.GetStartDate().GetMonth(),
			TotalEp:        aw.metadata.MustGet().GetMainEpisodeCount(),
			AbsoluteOffset: aw.metadata.MustGet().GetOffset(),
		})
		if err != nil {
			return nil, err
		}

		err = aw.fileCacher.Set(bucket, strconv.Itoa(key), episodes)

		if err != nil {
			return nil, err
		}
	}

	return episodes, nil
}

func (aw *AnimeWrapperImpl) GetTVDBEpisodeByNumber(number int) (*tvdb.Episode, bool) {
	if aw == nil || aw.tvdbEpisodes == nil {
		return nil, false
	}

	for _, e := range aw.tvdbEpisodes {
		if e.Number == number {
			return e, true
		}
	}

	return nil, false
}

func ExtractEpisodeInteger(s string) (int, bool) {
	pattern := "[0-9]+"
	regex := regexp.MustCompile(pattern)

	// Find the first match in the input string.
	match := regex.FindString(s)

	if match != "" {
		// Convert the matched string to an integer.
		num, err := strconv.Atoi(match)
		if err != nil {
			return 0, false
		}
		return num, true
	}

	return 0, false
}

func OffsetAnidbEpisode(s string, offset int) string {
	pattern := "([0-9]+)"
	regex := regexp.MustCompile(pattern)

	// Replace the first matched integer with the incremented value.
	result := regex.ReplaceAllStringFunc(s, func(matched string) string {
		num, err := strconv.Atoi(matched)
		if err == nil {
			num = num + offset
			return strconv.Itoa(num)
		} else {
			return matched
		}
	})

	return result
}
