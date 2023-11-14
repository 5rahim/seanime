package entities

import (
	"errors"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
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
		MediaEntryDownloadInfo *MediaEntryDownloadInfo `json:"downloadInfo"`
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

	entry.CurrentEpisodeCount = entry.Media.GetCurrentEpisodeCount()

	// Get the entry's local files
	lfs := GetLocalFilesFromMediaId(opts.LocalFiles, opts.MediaId)
	entry.LocalFiles = lfs // Returns empty slice if no local files are found

	libraryData, _ := NewMediaEntryLibraryData(&NewMediaEntryLibraryDataOptions{
		entryLocalFiles: lfs,
		mediaId:         entry.Media.ID,
	})
	entry.MediaEntryLibraryData = libraryData

	// Fetch AniDB data and cache it for 10 minutes
	anizipData, err := anizip.FetchAniZipMediaC("anilist", opts.MediaId, opts.AnizipCache)
	if err != nil {
		return nil, err
	}
	entry.AniDBId = anizipData.GetMappings().AnidbID

	// Instantiate MediaEntryListData
	// If the media exist in the user's anime list, add the details
	if found {
		entry.MediaEntryListData = &MediaEntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryStartDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryCompletionDate(anilistEntry.CompletedAt),
		}
	}

	// Create episode entities
	entry.hydrateEntryEpisodeData(anilistEntry, anizipData)

	return entry, nil

}

// hydrateEntryEpisodeData
// AniZipData, Media and LocalFiles should be defined
func (e *MediaEntry) hydrateEntryEpisodeData(
	anilistEntry *anilist.AnimeCollection_MediaListCollection_Lists_Entries,
	anizipData *anizip.Media,
) {

	if anizipData.Episodes == nil && len(anizipData.Episodes) == 0 {
		return
	}

	possibleSpecialInclusion, hasDiscrepancy := detectDiscrepancy(e.LocalFiles, e.Media, anizipData)

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

	//
	// Episode entities
	//

	p := pool.NewWithResults[*MediaEntryEpisode]()
	for _, lf := range e.LocalFiles {
		lf := lf
		p.Go(func() *MediaEntryEpisode {
			return NewMediaEntryEpisode(&NewMediaEntryEpisodeOptions{
				localFile:            lf,
				optionalAniDBEpisode: "",
				anizipMedia:          anizipData,
				media:                e.Media,
				progressOffset:       progressOffset,
				isDownloaded:         true,
			})
		})
	}
	episodes := p.Wait()
	// Sort by progress number
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].EpisodeNumber < episodes[j].EpisodeNumber
	})

	//
	// Info
	//
	info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
		localFiles:  e.LocalFiles,
		anizipMedia: anizipData,
		progress:    anilistEntry.Progress,
		status:      anilistEntry.Status,
		media:       e.Media,
	})
	if err == nil {
		e.MediaEntryDownloadInfo = info
	}

	e.Episodes = episodes

	nextEp, found := e.FindNextEpisode()
	if found {
		e.NextEpisode = nextEp
	}

}

//----------------------------------------------------------------------------------------------------------------------

// detectDiscrepancy detects whether there is a discrepancy between AniList and AniDB.
// e.g, AniList includes episode 0 as part of main episodes, but AniDB does not.
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

	// [possibleSpecialInclusion] means that there might be a discrepancy between AniList and AniDB
	// We should use this to check.
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,12] //=> true
	// e.g, epCeiling = 13 AND downloaded episodes = [0,...,13] //=> false
	possibleSpecialInclusion = hasEpisodeZero && noEpisodeCeiling

	_, aniDBHasS1 := anizipData.Episodes["S1"]
	// AniList episode count > AniDB episode count
	// This means that there is a discrepancy and AniList is most likely including episode 0 as part of main episodes
	hasDiscrepancy = media.GetCurrentEpisodeCount() > anizipData.GetMainEpisodeCount() && aniDBHasS1

	return

}
