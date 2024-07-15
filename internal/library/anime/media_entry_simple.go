package anime

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/sourcegraph/conc/pool"
	"sort"
)

type (
	SimpleMediaEntry struct {
		MediaId               int                    `json:"mediaId"`
		Media                 *anilist.BaseAnime     `json:"media"`
		MediaEntryListData    *MediaEntryListData    `json:"listData"`
		MediaEntryLibraryData *MediaEntryLibraryData `json:"libraryData"`
		Episodes              []*MediaEntryEpisode   `json:"episodes"`
		NextEpisode           *MediaEntryEpisode     `json:"nextEpisode"`
		LocalFiles            []*LocalFile           `json:"localFiles"`
		CurrentEpisodeCount   int                    `json:"currentEpisodeCount"`
	}

	SimpleMediaEntryListData struct {
		Progress    int                      `json:"progress,omitempty"`
		Score       float64                  `json:"score,omitempty"`
		Status      *anilist.MediaListStatus `json:"status,omitempty"`
		StartedAt   string                   `json:"startedAt,omitempty"`
		CompletedAt string                   `json:"completedAt,omitempty"`
	}

	NewSimpleMediaEntryOptions struct {
		MediaId         int
		LocalFiles      []*LocalFile // All local files
		AnimeCollection *anilist.AnimeCollection
		Platform        platform.Platform
	}
)

func NewSimpleMediaEntry(opts *NewSimpleMediaEntryOptions) (*SimpleMediaEntry, error) {

	if opts.AnimeCollection == nil ||
		opts.Platform == nil {
		return nil, errors.New("missing arguments when creating simple media entry")
	}
	// Create new MediaEntry
	entry := new(SimpleMediaEntry)
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

	libraryData, _ := NewMediaEntryLibraryData(&NewMediaEntryLibraryDataOptions{
		EntryLocalFiles: lfs,
		MediaId:         entry.Media.ID,
	})
	entry.MediaEntryLibraryData = libraryData

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
func (e *SimpleMediaEntry) hydrateEntryEpisodeData() {

	// +---------------------+
	// |       Episodes      |
	// +---------------------+

	p := pool.NewWithResults[*MediaEntryEpisode]()
	for _, lf := range e.LocalFiles {
		lf := lf
		p.Go(func() *MediaEntryEpisode {
			return NewSimpleMediaEntryEpisode(&NewSimpleMediaEntryEpisodeOptions{
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
