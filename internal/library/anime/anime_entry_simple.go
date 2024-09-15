package anime

import (
	"errors"
	"github.com/sourcegraph/conc/pool"
	"seanime/internal/api/anilist"
	"seanime/internal/platforms/platform"
	"sort"
)

type (
	SimpleAnimeEntry struct {
		MediaId               int                    `json:"mediaId"`
		Media                 *anilist.BaseAnime     `json:"media"`
		AnimeEntryListData    *AnimeEntryListData    `json:"listData"`
		AnimeEntryLibraryData *AnimeEntryLibraryData `json:"libraryData"`
		Episodes              []*Episode             `json:"episodes"`
		NextEpisode           *Episode               `json:"nextEpisode"`
		LocalFiles            []*LocalFile           `json:"localFiles"`
		CurrentEpisodeCount   int                    `json:"currentEpisodeCount"`
	}

	SimpleAnimeEntryListData struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}

	NewSimpleAnimeEntryOptions struct {
		MediaId         int
		LocalFiles      []*LocalFile // All local files
		AnimeCollection *anilist.AnimeCollection
		Platform        platform.Platform
	}
)

func NewSimpleAnimeEntry(opts *NewSimpleAnimeEntryOptions) (*SimpleAnimeEntry, error) {

	if opts.AnimeCollection == nil ||
		opts.Platform == nil {
		return nil, errors.New("missing arguments when creating simple media entry")
	}
	// Create new AnimeEntry
	entry := new(SimpleAnimeEntry)
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

	// Instantiate AnimeEntryListData
	// If the media exist in the user's anime list, add the details
	if found {
		entry.AnimeEntryListData = &AnimeEntryListData{
			Progress:    *anilistEntry.Progress,
			Score:       *anilistEntry.Score,
			Status:      anilistEntry.Status,
			StartedAt:   anilist.ToEntryStartDate(anilistEntry.StartedAt),
			CompletedAt: anilist.ToEntryCompletionDate(anilistEntry.CompletedAt),
		}
	}

	// +---------------------+
	// |       Episodes      |
	// +---------------------+

	// Create episode entities
	entry.hydrateEntryEpisodeData()

	return entry, nil

}

//----------------------------------------------------------------------------------------------------------------------

// hydrateEntryEpisodeData
// AniZipData, Media and LocalFiles should be defined
func (e *SimpleAnimeEntry) hydrateEntryEpisodeData() {

	// +---------------------+
	// |       Episodes      |
	// +---------------------+

	p := pool.NewWithResults[*Episode]()
	for _, lf := range e.LocalFiles {
		lf := lf
		p.Go(func() *Episode {
			return NewSimpleEpisode(&NewSimpleEpisodeOptions{
				LocalFile:    lf,
				Media:        e.Media,
				IsDownloaded: true,
			})
		})
	}
	episodes := p.Wait()
	// Sort by progress number
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].EpisodeNumber < episodes[j].EpisodeNumber
	})
	e.Episodes = episodes

	nextEp, found := e.FindNextEpisode()
	if found {
		e.NextEpisode = nextEp
	}

}
