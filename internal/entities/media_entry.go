package entities

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
)

type (
	// MediaEntry is a container for all data related to a media.
	// It is the primary data structure used by the frontend.
	MediaEntry struct {
		MediaId int                `json:"mediaId"`
		Media   *anilist.BaseMedia `json:"media"`

		// If the media exist in the user's anime list, instantiate details.
		// It is nil if the media is not in the user's anime list.
		MediaEntryDetails *MediaEntryDetails `json:"listEntry"`

		// Episodes holds the episodes of the media.
		Episodes []*MediaEntryEpisode `json:"episodes"`

		DownloadInfo struct {
		} `json:"downloadInfo"`

		// AnizipData holds data fetched from AniDB.
		AnizipData *anizip.Media `json:"anizipData"`
		// LocalFiles holds the local files associated with the media.
		LocalFiles []*LocalFile `json:"localFiles"`
	}

	// MediaEntryDetails holds the details of the list entry.
	MediaEntryDetails struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}

	// NewMediaEntryOptions is a constructor for MediaEntry.
	NewMediaEntryOptions struct {
		MediaId           int
		LocalFiles        []*LocalFile // All local files
		AnizipCache       *anizip.Cache
		AnilistCollection *anilist.AnimeCollection
		AnilistClient     *anilist.Client
	}
)

// NewMediaEntry creates a new MediaEntry based on the media id and a list of local files.
func NewMediaEntry(opts *NewMediaEntryOptions) (*MediaEntry, error) {

	if opts.AnilistCollection == nil ||
		opts.LocalFiles == nil ||
		opts.AnizipCache == nil ||
		opts.AnilistClient == nil {
		return nil, errors.New("missing arguments when creating media entry")
	}
	// Create new MediaEntry
	entry := new(MediaEntry)
	entry.MediaId = opts.MediaId

	// Get the Anilist List entry
	anilistEntry, found := opts.AnilistCollection.GetListEntryFromMediaId(opts.MediaId)

	// Set the media
	// If the Anilist List entry does not exist, fetch the media from AniList
	if !found {
		// Fetch the media
		fetchedMedia, err := anilist.GetBaseMediaById(opts.AnilistClient, opts.MediaId) // DEVNOTE: Maybe cache it?
		if err != nil {
			return nil, err
		}
		entry.Media = fetchedMedia
	} else {
		entry.Media = anilistEntry.Media
	}

	// Get the entry's local files
	lfs := GetLocalFilesFromMediaId(opts.LocalFiles, opts.MediaId)
	entry.LocalFiles = lfs

	// Fetch AniDB data and cache it for 10 minutes
	anidb, err := anizip.FetchAniZipMediaC("anilist", opts.MediaId, opts.AnizipCache)
	if err != nil {
		return nil, err
	}
	entry.AnizipData = anidb

	// Instantiate MediaEntryDetails
	// If the media exist in the user's anime list, add the details
	if found {
		entry.MediaEntryDetails = &MediaEntryDetails{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryStartDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryCompletionDate(anilistEntry.CompletedAt),
		}
	}

	// Create episode entities
	createEpisodes(entry)

	//entry.Episodes = episodes

	return entry, nil

}

// createEpisodes
// AniZipData, Media and LocalFiles should be defined
func createEpisodes(me *MediaEntry) {

	if me.AnizipData.Episodes == nil && len(me.AnizipData.Episodes) == 0 {
		return
	}

	episodes := make([]*MediaEntryEpisode, 0)

	// Whether downloaded episodes include a special episode "0"
	hasEpisodeZero := lo.SomeBy(me.LocalFiles, func(lf *LocalFile) bool {
		return lf.Metadata.Episode == 0
	})

	// No episode number is equal to the max episode number
	noEpisodeCeiling := lo.EveryBy(me.LocalFiles, func(lf *LocalFile) bool {
		return lf.Metadata.Episode != me.Media.GetCurrentEpisodeCount()
	})

	// [possibleSpecialInclusion] means that there might be a discrepancy between AniList and AniDB
	// We should use this to check.
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,12] //=> true
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,13] //=> false
	possibleSpecialInclusion := hasEpisodeZero && noEpisodeCeiling

	// AniList episode count > AniDB episode count
	// This means that there is a discrepancy and AniList is most likely including episode 0 as part of main episodes
	hasDiscrepancy := me.Media.GetCurrentEpisodeCount() > len(me.AnizipData.Episodes)

	progressOffset := 0

	if possibleSpecialInclusion && hasDiscrepancy {
		progressOffset = 1
	}

	// TODO: Parallelize
	for _, lf := range me.LocalFiles {
		episodes = append(episodes, NewMediaEntryEpisode(&NewMediaEntryEpisodeOptions{
			localFile:      lf,
			anizipMedia:    me.AnizipData,
			media:          me.Media,
			progressOffset: progressOffset,
			isDownloaded:   true,
		}))
	}

}
