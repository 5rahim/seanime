package entities

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/sourcegraph/conc/pool"
	"sort"
)

type (
	// MediaEntry is a container for all data related to a media.
	// It is the primary data structure used by the frontend.
	MediaEntry struct {
		MediaId                int                     `json:"mediaId"`
		Media                  *anilist.BaseMedia      `json:"media"`
		MediaEntryListData     *MediaEntryListData     `json:"listData"`
		MediaEntryLibraryData  *MediaEntryLibraryData  `json:"libraryData"`
		MediaEntryDownloadInfo *MediaEntryDownloadInfo `json:"downloadInfo,omitempty"`
		Episodes               []*MediaEntryEpisode    `json:"episodes"`
		NextEpisode            *MediaEntryEpisode      `json:"nextEpisode"`
		LocalFiles             []*LocalFile            `json:"localFiles"`
		AniDBId                int                     `json:"aniDBId"`
		CurrentEpisodeCount    int                     `json:"currentEpisodeCount"`
	}

	// MediaEntryListData holds the details of the AniList entry.
	MediaEntryListData struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}
)

type (
	// NewMediaEntryOptions is a constructor for MediaEntry.
	NewMediaEntryOptions struct {
		MediaId              int
		LocalFiles           []*LocalFile // All local files
		AnizipCache          *anizip.Cache
		AnilistCollection    *anilist.AnimeCollection
		AnilistClientWrapper anilist.ClientWrapperInterface
		MetadataProvider     *metadata.Provider
	}
)

// NewMediaEntry creates a new MediaEntry based on the media id and a list of local files.
// A MediaEntry is a container for all data related to a media.
// It is the primary data structure used by the frontend.
//
// It has the following properties:
//   - MediaEntryListData: Details of the AniList entry (if any)
//   - MediaEntryLibraryData: Details of the local files (if any)
//   - MediaEntryDownloadInfo: Details of the download status
//   - Episodes: List of episodes (if any)
//   - NextEpisode: Next episode to watch (if any)
//   - LocalFiles: List of local files (if any)
//   - AniDBId: AniDB id
//   - CurrentEpisodeCount: Current episode count
func NewMediaEntry(opts *NewMediaEntryOptions) (*MediaEntry, error) {

	if opts.AnilistCollection == nil ||
		opts.AnilistClientWrapper == nil {
		return nil, errors.New("missing arguments when creating media entry")
	}

	// Create new MediaEntry
	entry := new(MediaEntry)
	entry.MediaId = opts.MediaId

	// +---------------------+
	// |   AniList entry     |
	// +---------------------+

	// Get the Anilist List entry
	anilistEntry, found := opts.AnilistCollection.GetListEntryFromMediaId(opts.MediaId)

	// Set the media
	// If the Anilist List entry does not exist, fetch the media from AniList
	if !found {
		// If the Anilist entry does not exist, instantiate one with zero values
		anilistEntry = &anilist.MediaListEntry{}

		// Fetch the media
		fetchedMedia, err := anilist.GetBaseMediaById(opts.AnilistClientWrapper, opts.MediaId) // DEVNOTE: Maybe cache it?
		if err != nil {
			return nil, err
		}
		entry.Media = fetchedMedia
	} else {
		entry.Media = anilistEntry.Media
	}

	entry.CurrentEpisodeCount = entry.Media.GetCurrentEpisodeCount()

	// +---------------------+
	// |   Local files       |
	// +---------------------+

	// Get the entry's local files
	lfs := GetLocalFilesFromMediaId(opts.LocalFiles, opts.MediaId)
	entry.LocalFiles = lfs // Returns empty slice if no local files are found

	libraryData, _ := NewMediaEntryLibraryData(&NewMediaEntryLibraryDataOptions{
		EntryLocalFiles: lfs,
		MediaId:         entry.Media.ID,
	})
	entry.MediaEntryLibraryData = libraryData

	// +---------------------+
	// |       AniZip        |
	// +---------------------+

	// Fetch AniDB data and cache it for 30 minutes
	anizipData, err := anizip.FetchAniZipMediaC("anilist", opts.MediaId, opts.AnizipCache)
	if err != nil {

		// +---------------- Start
		// +---------------------+
		// |   Without AniZip    |
		// +---------------------+

		// If AniZip data is not found, we will still create the MediaEntry without it
		simpleMediaEntry, err := NewSimpleMediaEntry(&NewSimpleMediaEntryOptions{
			MediaId:              opts.MediaId,
			LocalFiles:           opts.LocalFiles,
			AnilistCollection:    opts.AnilistCollection,
			AnilistClientWrapper: opts.AnilistClientWrapper,
		})
		if err != nil {
			return nil, err
		}

		return &MediaEntry{
			MediaId:                simpleMediaEntry.MediaId,
			Media:                  simpleMediaEntry.Media,
			MediaEntryListData:     simpleMediaEntry.MediaEntryListData,
			MediaEntryLibraryData:  simpleMediaEntry.MediaEntryLibraryData,
			MediaEntryDownloadInfo: nil,
			Episodes:               simpleMediaEntry.Episodes,
			NextEpisode:            simpleMediaEntry.NextEpisode,
			LocalFiles:             simpleMediaEntry.LocalFiles,
			AniDBId:                0,
			CurrentEpisodeCount:    simpleMediaEntry.CurrentEpisodeCount,
		}, nil
		// +--------------- End

	}
	entry.AniDBId = anizipData.GetMappings().AnidbID

	// Instantiate MediaEntryListData
	// If the media exist in the user's anime list, add the details
	if found {
		entry.MediaEntryListData = &MediaEntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryDate(anilistEntry.CompletedAt),
		}
	}

	// +---------------------+
	// |       Episodes      |
	// +---------------------+

	// Create episode entities
	entry.hydrateEntryEpisodeData(anilistEntry, anizipData, opts.MetadataProvider)

	return entry, nil

}

//----------------------------------------------------------------------------------------------------------------------

// hydrateEntryEpisodeData
// AniZipData, Media and LocalFiles should be defined
func (e *MediaEntry) hydrateEntryEpisodeData(
	anilistEntry *anilist.MediaListEntry,
	anizipData *anizip.Media,
	metadataProvider *metadata.Provider,
) {

	if anizipData.Episodes == nil && len(anizipData.Episodes) == 0 {
		return
	}

	possibleSpecialInclusion, hasDiscrepancy := detectDiscrepancy(e.LocalFiles, e.Media, anizipData)

	// +---------------------+
	// |     Discrepancy     |
	// +---------------------+

	// We offset the progress number by 1 if there is a discrepancy
	progressOffset := 0
	if possibleSpecialInclusion && hasDiscrepancy {
		progressOffset = 1

	} else if possibleSpecialInclusion && !hasDiscrepancy {
		// Check if the Episode 0 is set to "S1"
		epZero, ok := lo.Find(e.LocalFiles, func(lf *LocalFile) bool {
			return lf.Metadata.Episode == 0
		})
		// If there is no discrepancy, but episode 0 is set to "S1", this means that the hydrator made a mistake (due to torrent name)
		// We will remap "S1" to "1" and offset other AniDB episodes by 1
		if ok && epZero.Metadata.AniDBEpisode == "S1" {
			progressOffset = -1 // Signal that the hydrator mistakenly set AniDB episode to "S1"
		}
	}

	// +---------------------+
	// |       Episodes      |
	// +---------------------+

	p := pool.NewWithResults[*MediaEntryEpisode]()
	for _, lf := range e.LocalFiles {
		p.Go(func() *MediaEntryEpisode {
			return NewMediaEntryEpisode(&NewMediaEntryEpisodeOptions{
				LocalFile:            lf,
				OptionalAniDBEpisode: "",
				AnizipMedia:          anizipData,
				Media:                e.Media,
				ProgressOffset:       progressOffset,
				IsDownloaded:         true,
				MetadataProvider:     metadataProvider,
			})
		})
	}
	episodes := p.Wait()
	// Sort by progress number
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].EpisodeNumber < episodes[j].EpisodeNumber
	})
	e.Episodes = episodes

	// +---------------------+
	// |    Download Info    |
	// +---------------------+

	info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
		LocalFiles:       e.LocalFiles,
		AnizipMedia:      anizipData,
		Progress:         anilistEntry.Progress,
		Status:           anilistEntry.Status,
		Media:            e.Media,
		MetadataProvider: metadataProvider,
	})
	if err == nil {
		e.MediaEntryDownloadInfo = info
	}

	nextEp, found := e.FindNextEpisode()
	if found {
		e.NextEpisode = nextEp
	}

}

//----------------------------------------------------------------------------------------------------------------------

// detectDiscrepancy detects whether there is a discrepancy between AniList and AniDB.
//   - AniList includes episode 0 as part of main episodes, but Anizip does not.
//   - Anizip has "S1"
func detectDiscrepancy(
	mediaLfs []*LocalFile, // Media's local files
	media *anilist.BaseMedia,
	anizipData *anizip.Media,
) (possibleSpecialInclusion bool, hasDiscrepancy bool) {

	if anizipData.Episodes == nil && len(anizipData.Episodes) == 0 {
		return false, false
	}

	// Whether downloaded episodes include a special episode "0"
	hasEpisodeZero := lo.SomeBy(mediaLfs, func(lf *LocalFile) bool {
		return lf.Metadata.Episode == 0
	})

	// No episode number is equal to the max episode number
	noEpisodeCeiling := lo.EveryBy(mediaLfs, func(lf *LocalFile) bool {
		return lf.Metadata.Episode != media.GetCurrentEpisodeCount()
	})

	// [possibleSpecialInclusion] means that there might be a discrepancy between AniList and Anizip
	// We should use this to check.
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,12] //=> true
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,13] //=> false
	possibleSpecialInclusion = hasEpisodeZero && noEpisodeCeiling

	_, aniDBHasS1 := anizipData.Episodes["S1"]
	// AniList episode count > Anizip episode count
	// This means that there is a discrepancy and AniList is most likely including episode 0 as part of main episodes
	hasDiscrepancy = media.GetCurrentEpisodeCount() > anizipData.GetMainEpisodeCount() && aniDBHasS1

	return

}
