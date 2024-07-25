package anime

import (
	"errors"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/platform"
	"sort"
)

type (
	// AnimeEntry is a container for all data related to a media.
	// It is the primary data structure used by the frontend.
	AnimeEntry struct {
		MediaId                int                     `json:"mediaId"`
		Media                  *anilist.BaseAnime      `json:"media"`
		AnimeEntryListData     *AnimeEntryListData     `json:"listData"`
		AnimeEntryLibraryData  *AnimeEntryLibraryData  `json:"libraryData"`
		AnimeEntryDownloadInfo *AnimeEntryDownloadInfo `json:"downloadInfo,omitempty"`
		Episodes               []*AnimeEntryEpisode    `json:"episodes"`
		NextEpisode            *AnimeEntryEpisode      `json:"nextEpisode"`
		LocalFiles             []*LocalFile            `json:"localFiles"`
		AniDBId                int                     `json:"aniDBId"`
		CurrentEpisodeCount    int                     `json:"currentEpisodeCount"`
	}

	// AnimeEntryListData holds the details of the AniList entry.
	AnimeEntryListData struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}
)

type (
	// NewAnimeEntryOptions is a constructor for AnimeEntry.
	NewAnimeEntryOptions struct {
		MediaId          int
		LocalFiles       []*LocalFile // All local files
		AnizipCache      *anizip.Cache
		AnimeCollection  *anilist.AnimeCollection
		Platform         platform.Platform
		MetadataProvider *metadata.Provider
	}
)

// NewAnimeEntry creates a new AnimeEntry based on the media id and a list of local files.
// A AnimeEntry is a container for all data related to a media.
// It is the primary data structure used by the frontend.
//
// It has the following properties:
//   - AnimeEntryListData: Details of the AniList entry (if any)
//   - AnimeEntryLibraryData: Details of the local files (if any)
//   - AnimeEntryDownloadInfo: Details of the download status
//   - Episodes: List of episodes (if any)
//   - NextEpisode: Next episode to watch (if any)
//   - LocalFiles: List of local files (if any)
//   - AniDBId: AniDB id
//   - CurrentEpisodeCount: Current episode count
func NewAnimeEntry(opts *NewAnimeEntryOptions) (*AnimeEntry, error) {

	if opts.AnimeCollection == nil ||
		opts.Platform == nil {
		return nil, errors.New("missing arguments when creating media entry")
	}

	// Create new AnimeEntry
	entry := new(AnimeEntry)
	entry.MediaId = opts.MediaId

	// +---------------------+
	// |   AniList entry     |
	// +---------------------+

	// Get the Anilist List entry
	anilistEntry, found := opts.AnimeCollection.GetListEntryFromAnimeId(opts.MediaId)

	// Set the media
	// If the Anilist List entry does not exist, fetch the media from AniList
	if !found {
		// If the Anilist entry does not exist, instantiate one with zero values
		anilistEntry = &anilist.MediaListEntry{}

		// Fetch the media
		fetchedMedia, err := opts.Platform.GetAnime(opts.MediaId) // DEVNOTE: Maybe cache it?
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

	libraryData, _ := NewAnimeEntryLibraryData(&NewAnimeEntryLibraryDataOptions{
		EntryLocalFiles: lfs,
		MediaId:         entry.Media.ID,
	})
	entry.AnimeEntryLibraryData = libraryData

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

		// If AniZip data is not found, we will still create the AnimeEntry without it
		simpleAnimeEntry, err := NewSimpleAnimeEntry(&NewSimpleAnimeEntryOptions{
			MediaId:         opts.MediaId,
			LocalFiles:      opts.LocalFiles,
			AnimeCollection: opts.AnimeCollection,
			Platform:        opts.Platform,
		})
		if err != nil {
			return nil, err
		}

		return &AnimeEntry{
			MediaId:                simpleAnimeEntry.MediaId,
			Media:                  simpleAnimeEntry.Media,
			AnimeEntryListData:     simpleAnimeEntry.AnimeEntryListData,
			AnimeEntryLibraryData:  simpleAnimeEntry.AnimeEntryLibraryData,
			AnimeEntryDownloadInfo: nil,
			Episodes:               simpleAnimeEntry.Episodes,
			NextEpisode:            simpleAnimeEntry.NextEpisode,
			LocalFiles:             simpleAnimeEntry.LocalFiles,
			AniDBId:                0,
			CurrentEpisodeCount:    simpleAnimeEntry.CurrentEpisodeCount,
		}, nil
		// +--------------- End

	}
	entry.AniDBId = anizipData.GetMappings().AnidbID

	// Instantiate AnimeEntryListData
	// If the media exist in the user's anime list, add the details
	if found {
		entry.AnimeEntryListData = &AnimeEntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.FuzzyDateToString(anilistEntry.StartedAt),
			CompletedAt: anilist.FuzzyDateToString(anilistEntry.CompletedAt),
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
func (e *AnimeEntry) hydrateEntryEpisodeData(
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

	p := pool.NewWithResults[*AnimeEntryEpisode]()
	for _, lf := range e.LocalFiles {
		p.Go(func() *AnimeEntryEpisode {
			return NewAnimeEntryEpisode(&NewAnimeEntryEpisodeOptions{
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

	info, err := NewAnimeEntryDownloadInfo(&NewAnimeEntryDownloadInfoOptions{
		LocalFiles:       e.LocalFiles,
		AnizipMedia:      anizipData,
		Progress:         anilistEntry.Progress,
		Status:           anilistEntry.Status,
		Media:            e.Media,
		MetadataProvider: metadataProvider,
	})
	if err == nil {
		e.AnimeEntryDownloadInfo = info
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
	media *anilist.BaseAnime,
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