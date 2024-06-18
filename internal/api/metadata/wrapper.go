package metadata

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/tvdb"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"strings"
)

type (
	// A MediaWrapper is a service that provides metadata for media.
	// It primarily fetches metadata from AniZip and AniList.
	// The user can request metadata to be fetched from TVDB as well, which will be done and stored in the cache.
	MediaWrapper struct {
		anizipMedia *anizip.Media
		baseMedia   *anilist.BasicMedia // TODO should be basicMedia
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
func (p *Provider) NewMediaWrapper(media *anilist.BasicMedia, anizipMedia *anizip.Media) *MediaWrapper {
	mw := &MediaWrapper{
		anizipMedia:  anizipMedia,
		baseMedia:    media,
		fileCacher:   p.fileCacher,
		logger:       p.logger,
		tvdbEpisodes: make([]*tvdb.Episode, 0),
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

	hasAnizipMetadata := mw.anizipMedia != nil && mw.anizipMedia.Episodes != nil
	hasTVDBMetadata := mw.tvdbEpisodes != nil && len(mw.tvdbEpisodes) > 0

	if !hasAnizipMetadata {
		meta.Image = mw.baseMedia.GetBannerImageSafe()
	}

	anizipEpisode, found := mw.anizipMedia.FindEpisode(strconv.Itoa(epNum))
	meta.AniDBId = anizipEpisode.AnidbEid

	// If we don't have AniZip metadata, just return the metadata containing the image
	if !found {
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
		if anizipEpisode.Image != "" {
			meta.Image = anizipEpisode.Image
		} else {
			// If AniZip image is not set, use the base media image
			meta.Image = mw.baseMedia.GetBannerImageSafe()
		}
	}

	meta.AirDate = anizipEpisode.Airdate
	meta.Length = anizipEpisode.Length
	if anizipEpisode.Runtime > 0 {
		meta.Length = anizipEpisode.Runtime
	}
	meta.Summary = strings.ReplaceAll(anizipEpisode.Summary, "`", "'")
	meta.Overview = strings.ReplaceAll(anizipEpisode.Overview, "`", "'")

	return meta
}
