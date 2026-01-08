package metadata_provider

import (
	"regexp"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/hook"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	AnimeWrapperImpl struct {
		metadata   mo.Option[*metadata.AnimeMetadata]
		db         *db.Database
		baseAnime  *anilist.BaseAnime
		fileCacher *filecache.Cacher
		logger     *zerolog.Logger

		provider *ProviderImpl

		parentEntry         *metadata.AnimeMetadata
		parentSpecialOffset int
	}
)

func (aw *AnimeWrapperImpl) GetEpisodeMetadata(ep string) (ret metadata.EpisodeMetadata) {
	if aw == nil || aw.baseAnime == nil {
		return
	}

	epNumber, _ := ExtractEpisodeInteger(ep)

	if aw.parentEntry != nil {
		epAsSpecial := ep
		if !strings.HasPrefix(ep, "S") {
			epAsSpecial = "S" + ep
		}
		epAsSpecial = OffsetAnidbEpisode(epAsSpecial, aw.parentSpecialOffset)
		ep = epAsSpecial
	}

	ret = metadata.EpisodeMetadata{
		AnidbId:               0,
		TvdbId:                0,
		Title:                 "",
		Image:                 "",
		AirDate:               "",
		Length:                0,
		Summary:               "",
		Overview:              "",
		EpisodeNumber:         epNumber,
		Episode:               ep,
		SeasonNumber:          0,
		AbsoluteEpisodeNumber: 0,
		AnidbEid:              0,
	}

	defer util.HandlePanicInModuleThen("api/metadata/GetEpisodeMetadata", func() {})

	reqEvent := &metadata.AnimeEpisodeMetadataRequestedEvent{}
	reqEvent.MediaId = aw.baseAnime.GetID()
	reqEvent.Episode = ep
	reqEvent.EpisodeNumber = epNumber
	reqEvent.EpisodeMetadata = &ret
	_ = hook.GlobalHookManager.OnAnimeEpisodeMetadataRequested().Trigger(reqEvent)
	ep = reqEvent.Episode
	epNumber = reqEvent.EpisodeNumber

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

	episode := mo.None[*metadata.EpisodeMetadata]()
	if aw.metadata.IsAbsent() {
		ret.Image = aw.baseAnime.GetBannerImageSafe()
	} else {
		episodeF, found := aw.metadata.MustGet().FindEpisode(ep)
		if found {
			episode = mo.Some(episodeF)
		}
	}

	// If we don't have Animap metadata, just return the metadata containing the image
	if episode.IsAbsent() {
		return ret
	}

	ret = *episode.MustGet()

	// If TVDB image is not set, use Animap image, if that is not set, use the AniList banner image
	if ret.Image == "" {
		// Set Animap image if TVDB image is not set
		if episode.MustGet().Image != "" {
			ret.Image = episode.MustGet().Image
		} else {
			// If Animap image is not set, use the base media image
			ret.Image = aw.baseAnime.GetBannerImageSafe()
		}
	}

	// Event
	event := &metadata.AnimeEpisodeMetadataEvent{
		EpisodeMetadata: &ret,
		Episode:         ep,
		EpisodeNumber:   epNumber,
		MediaId:         aw.baseAnime.GetID(),
	}
	_ = hook.GlobalHookManager.OnAnimeEpisodeMetadata().Trigger(event)
	if event.EpisodeMetadata == nil {
		return ret
	}
	ret = *event.EpisodeMetadata

	return ret
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
	if offset == 0 {
		return s
	}
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
