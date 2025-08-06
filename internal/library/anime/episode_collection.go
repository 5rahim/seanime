package anime

import (
	"cmp"
	"context"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/hook"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/result"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

var episodeCollectionCache = result.NewCache[int, *EpisodeCollection]()
var EpisodeCollectionFromLocalFilesCache = result.NewCache[int, *EpisodeCollection]()

type (
	// EpisodeCollection represents a collection of episodes.
	EpisodeCollection struct {
		HasMappingError bool                    `json:"hasMappingError"`
		Episodes        []*Episode              `json:"episodes"`
		Metadata        *metadata.AnimeMetadata `json:"metadata"`
	}
)

type NewEpisodeCollectionOptions struct {
	// AnimeMetadata can be nil, if not provided, it will be fetched from the metadata provider.
	AnimeMetadata    *metadata.AnimeMetadata
	Media            *anilist.BaseAnime
	MetadataProvider metadata.Provider
	Logger           *zerolog.Logger
}

// NewEpisodeCollection creates a new episode collection by leveraging EntryDownloadInfo.
// The returned EpisodeCollection is cached for 6 hours.
//
// AnimeMetadata is optional, if not provided, it will be fetched from the metadata provider.
//
// Note: This is used by Torrent and Debrid streaming
func NewEpisodeCollection(opts NewEpisodeCollectionOptions) (ec *EpisodeCollection, err error) {
	if opts.Logger == nil {
		opts.Logger = lo.ToPtr(zerolog.Nop())
	}

	if opts.Media == nil {
		return nil, fmt.Errorf("cannont create episode collectiom, media is nil")
	}

	if opts.MetadataProvider == nil {
		return nil, fmt.Errorf("cannot create episode collection, metadata provider is nil")
	}

	if ec, ok := episodeCollectionCache.Get(opts.Media.ID); ok {
		opts.Logger.Debug().Msg("torrentstream: Using cached episode collection")
		return ec, nil
	}

	if opts.AnimeMetadata == nil {
		// Fetch the metadata
		opts.AnimeMetadata, err = opts.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, opts.Media.ID)
		if err != nil {
			opts.AnimeMetadata = &metadata.AnimeMetadata{
				Titles:       make(map[string]string),
				Episodes:     make(map[string]*metadata.EpisodeMetadata),
				EpisodeCount: 0,
				SpecialCount: 0,
				Mappings: &metadata.AnimeMappings{
					AnilistId: opts.Media.GetID(),
				},
			}
			opts.AnimeMetadata.Titles["en"] = opts.Media.GetTitleSafe()
			opts.AnimeMetadata.Titles["x-jat"] = opts.Media.GetRomajiTitleSafe()
			err = nil
		}
	}

	reqEvent := &AnimeEpisodeCollectionRequestedEvent{
		Media:             opts.Media,
		Metadata:          opts.AnimeMetadata,
		EpisodeCollection: &EpisodeCollection{},
	}
	err = hook.GlobalHookManager.OnAnimEpisodeCollectionRequested().Trigger(reqEvent)
	if err != nil {
		return nil, err
	}
	opts.Media = reqEvent.Media
	opts.AnimeMetadata = reqEvent.Metadata

	if reqEvent.DefaultPrevented {
		return reqEvent.EpisodeCollection, nil
	}

	ec = &EpisodeCollection{
		HasMappingError: false,
		Episodes:        make([]*Episode, 0),
		Metadata:        opts.AnimeMetadata,
	}

	// +---------------------+
	// |    Download Info    |
	// +---------------------+

	info, err := NewEntryDownloadInfo(&NewEntryDownloadInfoOptions{
		LocalFiles:       nil,
		AnimeMetadata:    opts.AnimeMetadata,
		Progress:         lo.ToPtr(0), // Progress is 0 because we want the entire list
		Status:           lo.ToPtr(anilist.MediaListStatusCurrent),
		Media:            opts.Media,
		MetadataProvider: opts.MetadataProvider,
	})
	if err != nil {
		opts.Logger.Error().Err(err).Msg("torrentstream: could not get media entry info")
		return nil, err
	}

	// As of v2.8.0, this should never happen, getMediaInfo always returns an anime metadata struct, even if it's not found
	// causing NewEntryDownloadInfo to return a valid list of episodes to download
	if info == nil || info.EpisodesToDownload == nil {
		opts.Logger.Debug().Msg("torrentstream: no episodes found from AniDB, using AniList")
		for epIdx := range opts.Media.GetCurrentEpisodeCount() {
			episodeNumber := epIdx + 1

			mediaWrapper := opts.MetadataProvider.GetAnimeMetadataWrapper(opts.Media, nil)
			episodeMetadata := mediaWrapper.GetEpisodeMetadata(episodeNumber)

			episode := &Episode{
				Type:                  LocalFileTypeMain,
				DisplayTitle:          fmt.Sprintf("Episode %d", episodeNumber),
				EpisodeTitle:          opts.Media.GetPreferredTitle(),
				EpisodeNumber:         episodeNumber,
				AniDBEpisode:          fmt.Sprintf("%d", episodeNumber),
				AbsoluteEpisodeNumber: episodeNumber,
				ProgressNumber:        episodeNumber,
				LocalFile:             nil,
				IsDownloaded:          false,
				EpisodeMetadata: &EpisodeMetadata{
					AnidbId:  0,
					Image:    episodeMetadata.Image,
					AirDate:  "",
					Length:   0,
					Summary:  "",
					Overview: "",
					IsFiller: false,
				},
				FileMetadata:  nil,
				IsInvalid:     false,
				MetadataIssue: "",
				BaseAnime:     opts.Media,
			}
			ec.Episodes = append(ec.Episodes, episode)
		}
		ec.HasMappingError = true
		return
	}

	if len(info.EpisodesToDownload) == 0 {
		opts.Logger.Error().Msg("torrentstream: no episodes found")
		return nil, fmt.Errorf("no episodes found")
	}

	ec.Episodes = lo.Map(info.EpisodesToDownload, func(episode *EntryDownloadEpisode, i int) *Episode {
		return episode.Episode
	})

	slices.SortStableFunc(ec.Episodes, func(i, j *Episode) int {
		return cmp.Compare(i.EpisodeNumber, j.EpisodeNumber)
	})

	event := &AnimeEpisodeCollectionEvent{
		EpisodeCollection: ec,
	}
	err = hook.GlobalHookManager.OnAnimeEpisodeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}
	ec = event.EpisodeCollection

	episodeCollectionCache.SetT(opts.Media.ID, ec, time.Hour*6)

	return
}

/////////

type NewEpisodeCollectionFromLocalFilesOptions struct {
	LocalFiles       []*LocalFile
	Media            *anilist.BaseAnime
	AnimeCollection  *anilist.AnimeCollection
	Platform         platform.Platform
	MetadataProvider metadata.Provider
	Logger           *zerolog.Logger
}

func NewEpisodeCollectionFromLocalFiles(ctx context.Context, opts NewEpisodeCollectionFromLocalFilesOptions) (*EpisodeCollection, error) {
	if opts.Logger == nil {
		opts.Logger = lo.ToPtr(zerolog.Nop())
	}

	if ec, ok := EpisodeCollectionFromLocalFilesCache.Get(opts.Media.GetID()); ok {
		return ec, nil
	}

	// Make sure to keep the local files from the media only
	opts.LocalFiles = lo.Filter(opts.LocalFiles, func(lf *LocalFile, i int) bool {
		return lf.MediaId == opts.Media.GetID()
	})

	// Create a new media entry
	entry, err := NewEntry(ctx, &NewEntryOptions{
		MediaId:          opts.Media.GetID(),
		LocalFiles:       opts.LocalFiles,
		AnimeCollection:  opts.AnimeCollection,
		Platform:         opts.Platform,
		MetadataProvider: opts.MetadataProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot play local file, could not create entry: %w", err)
	}

	// Should be cached if it exists
	animeMetadata, err := opts.MetadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, opts.Media.ID)
	if err != nil {
		animeMetadata = &metadata.AnimeMetadata{
			Titles:       make(map[string]string),
			Episodes:     make(map[string]*metadata.EpisodeMetadata),
			EpisodeCount: 0,
			SpecialCount: 0,
			Mappings: &metadata.AnimeMappings{
				AnilistId: opts.Media.GetID(),
			},
		}
		animeMetadata.Titles["en"] = opts.Media.GetTitleSafe()
		animeMetadata.Titles["x-jat"] = opts.Media.GetRomajiTitleSafe()
		err = nil
	}

	ec := &EpisodeCollection{
		HasMappingError: false,
		Episodes:        entry.Episodes,
		Metadata:        animeMetadata,
	}

	EpisodeCollectionFromLocalFilesCache.SetT(opts.Media.GetID(), ec, time.Hour*6)

	return ec, nil
}

/////////

func (ec *EpisodeCollection) FindEpisodeByNumber(episodeNumber int) (*Episode, bool) {
	for _, episode := range ec.Episodes {
		if episode.EpisodeNumber == episodeNumber {
			return episode, true
		}
	}
	return nil, false
}

func (ec *EpisodeCollection) FindEpisodeByAniDB(anidbEpisode string) (*Episode, bool) {
	for _, episode := range ec.Episodes {
		if episode.AniDBEpisode == anidbEpisode {
			return episode, true
		}
	}
	return nil, false
}

// GetMainLocalFiles returns the *main* local files.
func (ec *EpisodeCollection) GetMainLocalFiles() ([]*Episode, bool) {
	ret := make([]*Episode, 0)
	for _, episode := range ec.Episodes {
		if episode.LocalFile == nil || episode.LocalFile.IsMain() {
			ret = append(ret, episode)
		}
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}

// FindNextEpisode returns the *main* local file whose episode number is after the given local file.
func (ec *EpisodeCollection) FindNextEpisode(current *Episode) (*Episode, bool) {
	episodes, ok := ec.GetMainLocalFiles()
	if !ok {
		return nil, false
	}
	// Get the local file whose episode number is after the given local file
	var next *Episode
	for _, e := range episodes {
		if e.GetEpisodeNumber() == current.GetEpisodeNumber()+1 {
			next = e
			break
		}
	}
	if next == nil {
		return nil, false
	}
	return next, true
}
