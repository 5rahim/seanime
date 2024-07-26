package metadata

import (
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/tvdb"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"
)

type (
	// A MediaWrapper is a service that provides metadata for media.
	// It primarily fetches metadata from AniZip and AniList.
	// The user can request metadata to be fetched from TVDB as well, which will be done and stored in the cache.
	MediaWrapper struct {
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

	MediaWrapperEpisodeMetadata struct {
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

// NewMediaWrapper creates a new media wrapper.
// Anizip Media can be nil.
func (p *Provider) NewMediaWrapper(media *anilist.BaseAnime, anizipMedia *anizip.Media) *MediaWrapper {
	mw := &MediaWrapper{
		anizipMedia:  mo.None[*anizip.Media](),
		baseAnime:    media,
		fileCacher:   p.fileCacher,
		logger:       p.logger,
		tvdbEpisodes: make([]*tvdb.Episode, 0),
	}

	if anizipMedia != nil {
		mw.anizipMedia = mo.Some(anizipMedia)
	}

	episodes, err := mw.GetTVDBEpisodes(false)
	if err == nil {
		mw.tvdbEpisodes = episodes
	}

	return mw
}

func (mw *MediaWrapper) GetEpisodeMetadata(epNum int) MediaWrapperEpisodeMetadata {
	meta := MediaWrapperEpisodeMetadata{
		EpisodeNumber: epNum,
	}

	hasTVDBMetadata := mw.tvdbEpisodes != nil && len(mw.tvdbEpisodes) > 0

	anizipEpisode := mo.None[*anizip.Episode]()
	if mw.anizipMedia.IsAbsent() {
		meta.Image = mw.baseAnime.GetBannerImageSafe()
	} else {
		anizipEpisodeF, found := mw.anizipMedia.MustGet().FindEpisode(strconv.Itoa(epNum))
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
		tvdbEpisode, found := mw.GetTVDBEpisodeByNumber(epNum)
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
			meta.Image = mw.baseAnime.GetBannerImageSafe()
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
