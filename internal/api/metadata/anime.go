package metadata

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/mappings"
	"seanime/internal/api/tvdb"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"
	"time"
)

type (
	AnimeWrapperImpl struct {
		anizipMedia mo.Option[*anizip.Media]
		baseAnime   *anilist.BaseAnime
		fileCacher  *filecache.Cacher
		logger      *zerolog.Logger
		// TVDB
		tvdbEpisodes []*tvdb.Episode
	}

	NewMediaWrapperOptions struct {
		AnizipMedia *anizip.Media
		Logger      *zerolog.Logger
	}

	EpisodeMetadata struct {
		AniDBId       int    `json:"aniDBId,omitempty"` // Episode AniDB ID
		TVDBId        int64  `json:"tvdbId,omitempty"`  // Episode TVDB ID
		Title         string `json:"title,omitempty"`   // Episode title
		Image         string `json:"image,omitempty"`
		AirDate       string `json:"airDate,omitempty"`
		Length        int    `json:"length,omitempty"`
		Summary       string `json:"summary,omitempty"`
		Overview      string `json:"overview,omitempty"`
		EpisodeNumber int    `json:"episodeNumber,omitempty"`
	}
)

// GetAnimeMetadata creates a new anime wrapper.
// Example:
//
//	metadataProvider.GetAnimeMetadata(media, anizipMedia)
//	metadataProvider.GetAnimeMetadata(media, nil)
func (p *ProviderImpl) GetAnimeMetadata(media *anilist.BaseAnime, anizipMedia *anizip.Media) AnimeMetadata {
	aw := &AnimeWrapperImpl{
		anizipMedia:  mo.None[*anizip.Media](),
		baseAnime:    media,
		fileCacher:   p.fileCacher,
		logger:       p.logger,
		tvdbEpisodes: make([]*tvdb.Episode, 0),
	}

	if anizipMedia != nil {
		aw.anizipMedia = mo.Some(anizipMedia)
	}

	episodes, err := aw.GetTVDBEpisodes(false)
	if err == nil {
		aw.tvdbEpisodes = episodes
	}

	return aw
}

func (aw *AnimeWrapperImpl) GetEpisodeMetadata(epNum int) EpisodeMetadata {
	meta := EpisodeMetadata{
		EpisodeNumber: epNum,
	}

	hasTVDBMetadata := aw.tvdbEpisodes != nil && len(aw.tvdbEpisodes) > 0

	anizipEpisode := mo.None[*anizip.Episode]()
	if aw.anizipMedia.IsAbsent() {
		meta.Image = aw.baseAnime.GetBannerImageSafe()
	} else {
		anizipEpisodeF, found := aw.anizipMedia.MustGet().FindEpisode(strconv.Itoa(epNum))
		if found {
			meta.AniDBId = anizipEpisodeF.AnidbEid
			anizipEpisode = mo.Some(anizipEpisodeF)
		}
	}

	// If we don't have AniZip metadata, just return the metadata containing the image
	if anizipEpisode.IsAbsent() {
		return meta
	}

	// TVDB metadata
	if hasTVDBMetadata {
		tvdbEpisode, found := aw.GetTVDBEpisodeByNumber(epNum)
		if found {
			meta.Image = tvdbEpisode.Image
			meta.TVDBId = tvdbEpisode.ID
		}
	}

	if meta.Image == "" {
		// Set AniZip image if TVDB image is not set
		if anizipEpisode.MustGet().Image != "" {
			meta.Image = anizipEpisode.MustGet().Image
		} else {
			// If AniZip image is not set, use the base media image
			meta.Image = aw.baseAnime.GetBannerImageSafe()
		}
	}

	meta.AirDate = anizipEpisode.MustGet().Airdate
	meta.Length = anizipEpisode.MustGet().Length
	if anizipEpisode.MustGet().Runtime > 0 {
		meta.Length = anizipEpisode.MustGet().Runtime
	}
	meta.Summary = strings.ReplaceAll(anizipEpisode.MustGet().Summary, "`", "'")
	meta.Overview = strings.ReplaceAll(anizipEpisode.MustGet().Overview, "`", "'")

	return meta
}

func getTvdbIDFromAnimeLists(anidbID int) (tvdbID int, ok bool) {
	res, err := mappings.GetReducedAnimeLists()
	if err != nil {
		return 0, false
	}
	return res.FindTvdbIDFromAnidbID(anidbID)
}

func (aw *AnimeWrapperImpl) EmptyTVDBEpisodesBucket(mediaId int) error {

	if aw.anizipMedia.IsAbsent() {
		return nil
	}

	// Get TVDB ID
	var tvdbId int
	tvdbId = aw.anizipMedia.MustGet().Mappings.ThetvdbID
	if tvdbId == 0 {
		if aw.anizipMedia.MustGet().Mappings.AnidbID > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(aw.anizipMedia.MustGet().Mappings.AnidbID)
		}
	}

	if tvdbId == 0 {
		return errors.New("metadata: could not find tvdb id")
	}

	return aw.fileCacher.Remove(fmt.Sprintf("tvdb_episodes_%d", mediaId))
}

func (aw *AnimeWrapperImpl) GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error) {
	key := aw.baseAnime.GetID()

	if aw.anizipMedia.IsAbsent() {
		return nil, errors.New("metadata: anizip media is absent")
	}

	// Get TVDB ID
	var tvdbId int
	tvdbId = aw.anizipMedia.MustGet().Mappings.ThetvdbID
	if tvdbId == 0 {
		if aw.anizipMedia.MustGet().Mappings.AnidbID > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(aw.anizipMedia.MustGet().Mappings.AnidbID)
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
			Logger: aw.logger,
		})

		episodes, err = tv.FetchSeriesEpisodes(tvdbId, tvdb.FilterEpisodeMediaInfo{
			Year:           aw.baseAnime.GetStartDate().GetYear(),
			Month:          aw.baseAnime.GetStartDate().GetMonth(),
			TotalEp:        aw.anizipMedia.MustGet().GetMainEpisodeCount(),
			AbsoluteOffset: aw.anizipMedia.MustGet().GetOffset(),
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
