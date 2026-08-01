package torrentstream

import (
	"context"
	"encoding/json"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/models"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/platform"
	"seanime/internal/testmocks"
	"seanime/internal/testutil"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func TestGetMediaInfoFromOptionsUsesProvidedMedia(t *testing.T) {
	repo, _, _ := newTorrentstreamTestRepository(t)
	media := testmocks.NewBaseAnime(990748023463256, "Custom Source Anime")

	opts := &StartStreamOptions{MediaId: media.ID}
	opts.SetMedia(media)

	got, _, err := repo.GetMediaInfoFromOptions(context.Background(), opts)

	require.NoError(t, err)
	require.Equal(t, media.ID, got.GetID())
	require.Equal(t, media.GetPreferredTitle(), got.GetPreferredTitle())
}

type recordedWSEvent struct {
	clientID string
	event    string
	payload  interface{}
}

type recordingWSEventManager struct {
	*events.MockWSEventManager
	mu     sync.Mutex
	events []recordedWSEvent
}

func newRecordingWSEventManager(t *testing.T) *recordingWSEventManager {
	t.Helper()
	return &recordingWSEventManager{
		MockWSEventManager: events.NewMockWSEventManager(util.NewLogger()),
		events:             make([]recordedWSEvent, 0),
	}
}

func (m *recordingWSEventManager) SendEvent(eventType string, payload interface{}) {
	m.mu.Lock()
	m.events = append(m.events, recordedWSEvent{event: eventType, payload: payload})
	m.mu.Unlock()
	m.MockWSEventManager.SendEvent(eventType, payload)
}

func (m *recordingWSEventManager) SendEventTo(clientID string, eventType string, payload interface{}, noLog ...bool) {
	m.mu.Lock()
	m.events = append(m.events, recordedWSEvent{clientID: clientID, event: eventType, payload: payload})
	m.mu.Unlock()
	m.MockWSEventManager.SendEventTo(clientID, eventType, payload, noLog...)
}

func (m *recordingWSEventManager) snapshot() []recordedWSEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	ret := make([]recordedWSEvent, len(m.events))
	copy(ret, m.events)
	return ret
}

func newTorrentstreamTestRepository(t *testing.T) (*Repository, *testutil.TestEnv, *recordingWSEventManager) {
	t.Helper()
	metadataProvider := testmocks.NewFakeMetadataProviderBuilder().Build()
	return newTorrentstreamTestRepositoryWithMetadataProvider(t, metadataProvider)
}

func newTorrentstreamTestRepositoryWithMetadataProvider(t *testing.T, metadataProvider metadata_provider.Provider) (*Repository, *testutil.TestEnv, *recordingWSEventManager) {
	t.Helper()
	env := testutil.NewTestEnv(t)
	ws := newRecordingWSEventManager(t)

	repo := NewRepository(&NewRepositoryOptions{
		Logger:              env.Logger(),
		BaseAnimeCache:      anilist.NewBaseAnimeCache(),
		CompleteAnimeCache:  anilist.NewCompleteAnimeCache(),
		PlatformRef:         util.NewRef[platform.Platform](nil),
		MetadataProviderRef: util.NewRef[metadata_provider.Provider](metadataProvider),
		WSEventManager:      ws,
		Database:            env.NewDatabase(""),
	})

	t.Cleanup(func() {
		repo.Shutdown()
	})

	return repo, env, ws
}

func TestHydrateStreamCollectionMergesAniListAndLibraryState(t *testing.T) {
	mediaInLibrary := testmocks.NewBaseAnimeBuilder(1, "Library Show").WithEpisodes(12).Build()
	mediaAlreadyQueued := testmocks.NewBaseAnimeBuilder(2, "Queued Show").WithEpisodes(12).Build()
	unreleasedMedia := testmocks.NewBaseAnimeBuilder(3, "Unreleased Show").WithEpisodes(12).WithStatus(anilist.MediaStatusNotYetReleased).Build()

	fakeMetadata := testmocks.NewFakeMetadataProviderBuilder().
		WithAnimeMetadata(mediaInLibrary.ID, anime.NewAnimeMetadataFromEpisodeCount(mediaInLibrary, []int{1, 2, 3})).
		Build()

	repo, _, _ := newTorrentstreamTestRepositoryWithMetadataProvider(t, fakeMetadata)
	libraryCollection := &anime.LibraryCollection{
		ContinueWatchingList: []*anime.Episode{{
			BaseAnime:       mediaAlreadyQueued,
			ProgressNumber:  1,
			EpisodeNumber:   1,
			DisplayTitle:    "Episode 1",
			EpisodeMetadata: &anime.EpisodeMetadata{},
		}},
		Lists: []*anime.LibraryCollectionList{{
			Status: anilist.MediaListStatusCurrent,
			Entries: []*anime.LibraryCollectionEntry{{
				Media:   mediaInLibrary,
				MediaId: mediaInLibrary.ID,
			}},
		}},
	}

	repo.HydrateStreamCollection(&HydrateStreamCollectionOptions{
		AnimeCollection: &anilist.AnimeCollection{
			MediaListCollection: &anilist.AnimeCollection_MediaListCollection{
				Lists: []*anilist.AnimeCollection_MediaListCollection_Lists{
					newAnimeCollectionList(anilist.MediaListStatusCurrent,
						newAnimeCollectionEntry(mediaInLibrary, 1, anilist.MediaListStatusCurrent),
						newAnimeCollectionEntry(mediaAlreadyQueued, 0, anilist.MediaListStatusCurrent),
						newAnimeCollectionEntry(unreleasedMedia, 0, anilist.MediaListStatusCurrent),
					),
					newAnimeCollectionList(anilist.MediaListStatusRepeating,
						newAnimeCollectionEntry(mediaInLibrary, 1, anilist.MediaListStatusRepeating),
					),
				},
			},
		},
		LibraryCollection:   libraryCollection,
		MetadataProviderRef: util.NewRef[metadata_provider.Provider](fakeMetadata),
	})

	require.NotNil(t, libraryCollection.Stream)
	require.Len(t, libraryCollection.Stream.ContinueWatchingList, 1)
	require.Len(t, libraryCollection.Stream.Anime, 1)
	require.Equal(t, mediaAlreadyQueued.ID, libraryCollection.Stream.Anime[0].ID)
	require.Contains(t, libraryCollection.Stream.ListData, mediaAlreadyQueued.ID)
	require.Equal(t, 0, libraryCollection.Stream.ListData[mediaAlreadyQueued.ID].Progress)
	require.NotContains(t, libraryCollection.Stream.ListData, unreleasedMedia.ID)

	nextEpisode := libraryCollection.Stream.ContinueWatchingList[0]
	require.Equal(t, mediaInLibrary.ID, nextEpisode.BaseAnime.ID)
	require.Equal(t, 2, nextEpisode.EpisodeNumber)
	require.Equal(t, 2, nextEpisode.GetProgressNumber())
	require.Equal(t, "Episode 2", nextEpisode.DisplayTitle)

	require.Equal(t, 1, fakeMetadata.MetadataCalls(mediaInLibrary.ID))
	require.Equal(t, 0, fakeMetadata.MetadataCalls(mediaAlreadyQueued.ID))
	require.Equal(t, 0, fakeMetadata.MetadataCalls(unreleasedMedia.ID))
}

func TestHydrateStreamCollectionFallsBackWhenEpisodeMetadataMissing(t *testing.T) {
	media := testmocks.NewBaseAnimeBuilder(44, "Fallback Show").WithEpisodes(12).WithBannerImage("https://example.com/banner.jpg").Build()
	fakeMetadata := testmocks.NewFakeMetadataProviderBuilder().
		WithAnimeMetadata(media.ID, anime.NewAnimeMetadataFromEpisodeCount(media, []int{1})).
		Build()

	repo, _, _ := newTorrentstreamTestRepositoryWithMetadataProvider(t, fakeMetadata)
	libraryCollection := &anime.LibraryCollection{}

	repo.HydrateStreamCollection(&HydrateStreamCollectionOptions{
		AnimeCollection: &anilist.AnimeCollection{
			MediaListCollection: &anilist.AnimeCollection_MediaListCollection{
				Lists: []*anilist.AnimeCollection_MediaListCollection_Lists{
					newAnimeCollectionList(anilist.MediaListStatusCurrent,
						newAnimeCollectionEntry(media, 1, anilist.MediaListStatusCurrent),
					),
				},
			},
		},
		LibraryCollection:   libraryCollection,
		MetadataProviderRef: util.NewRef[metadata_provider.Provider](fakeMetadata),
	})

	require.NotNil(t, libraryCollection.Stream)
	require.Len(t, libraryCollection.Stream.ContinueWatchingList, 1)
	require.Len(t, libraryCollection.Stream.Anime, 1)

	episode := libraryCollection.Stream.ContinueWatchingList[0]
	require.Equal(t, media.ID, episode.BaseAnime.ID)
	require.Equal(t, 2, episode.EpisodeNumber)
	require.Equal(t, 2, episode.GetProgressNumber())
	require.Equal(t, "Episode 2", episode.DisplayTitle)
	require.Equal(t, media.GetPreferredTitle(), episode.EpisodeTitle)
	require.NotNil(t, episode.EpisodeMetadata)
	require.Equal(t, media.GetBannerImageSafe(), episode.EpisodeMetadata.Image)
	require.True(t, episode.IsInvalid)

	require.Contains(t, libraryCollection.Stream.ListData, media.ID)
	require.Equal(t, 1, libraryCollection.Stream.ListData[media.ID].Progress)
	require.Equal(t, 1, fakeMetadata.MetadataCalls(media.ID))
}

func TestRepositoryDefaultsAndSettingsGuards(t *testing.T) {
	repo, _, _ := newTorrentstreamTestRepository(t)

	require.False(t, repo.IsEnabled())
	require.EqualError(t, repo.FailIfNoSettings(), "torrentstream: no settings provided, the module is dormant")
	require.Equal(t, repo.getDefaultDownloadPath(), repo.GetDownloadDir())

	previous, ok := repo.GetPreviousStreamOptions()
	require.False(t, ok)
	require.Nil(t, previous)

	expected := &StartStreamOptions{MediaId: 10, EpisodeNumber: 3, AniDBEpisode: "3"}
	repo.previousStreamOptions = mo.Some(expected)

	previous, ok = repo.GetPreviousStreamOptions()
	require.True(t, ok)
	require.Same(t, expected, previous)

	repo.settings = mo.Some(Settings{
		TorrentstreamSettings: models.TorrentstreamSettings{
			Enabled:     true,
			DownloadDir: "/custom/downloads",
		},
		Host: "127.0.0.1",
		Port: 43210,
	})

	require.True(t, repo.IsEnabled())
	require.NoError(t, repo.FailIfNoSettings())
	require.Equal(t, "/custom/downloads", repo.GetDownloadDir())
}

func TestStreamOptionsMatchAllowsImplicitFileIndex(t *testing.T) {
	fileIndex := 3
	prepared := &StartStreamOptions{MediaId: 10, EpisodeNumber: 4, AniDBEpisode: "4", FileIndex: &fileIndex}
	request := &StartStreamOptions{MediaId: 10, EpisodeNumber: 4, AniDBEpisode: "4"}

	require.True(t, streamOptionsMatch(request, prepared))
	require.True(t, streamOptionsMatch(prepared, request))
}

func TestStreamOptionsMatchRejectsExplicitDifferentFileIndex(t *testing.T) {
	preparedIndex := 3
	requestIndex := 4
	prepared := &StartStreamOptions{MediaId: 10, EpisodeNumber: 4, AniDBEpisode: "4", FileIndex: &preparedIndex}
	request := &StartStreamOptions{MediaId: 10, EpisodeNumber: 4, AniDBEpisode: "4", FileIndex: &requestIndex}

	require.False(t, streamOptionsMatch(request, prepared))
}

func TestInitModulesRejectsNilSettingsAndDisablesModule(t *testing.T) {
	repo, _, _ := newTorrentstreamTestRepository(t)

	err := repo.InitModules(nil, "127.0.0.1", 8080)
	require.EqualError(t, err, "torrentstream: Cannot initialize module, no settings provided")
	require.True(t, repo.settings.IsAbsent())

	err = repo.InitModules(&models.TorrentstreamSettings{Enabled: false}, "127.0.0.1", 8080)
	require.NoError(t, err)
	require.True(t, repo.settings.IsAbsent())
	require.False(t, repo.IsEnabled())
}

func TestSendStateEvent(t *testing.T) {
	repo, _, ws := newTorrentstreamTestRepository(t)

	repo.sendStateEvent(eventLoading, TLSStateSearchingTorrents)
	repo.sendStateEvent(eventTorrentLoaded)

	eventsSnapshot := ws.snapshot()
	require.Len(t, eventsSnapshot, 2)
	require.Equal(t, events.TorrentStreamState, eventsSnapshot[0].event)
	require.Equal(t, events.TorrentStreamState, eventsSnapshot[1].event)

	firstPayload := decodePayloadMap(t, eventsSnapshot[0].payload)
	require.Equal(t, eventLoading, firstPayload["state"])
	require.Equal(t, string(TLSStateSearchingTorrents), firstPayload["data"])

	secondPayload := decodePayloadMap(t, eventsSnapshot[1].payload)
	require.Equal(t, eventTorrentLoaded, secondPayload["state"])
	require.Nil(t, secondPayload["data"])
}

func TestGetBatchHistoryReturnsEmptyWhenMissing(t *testing.T) {
	repo, _, _ := newTorrentstreamTestRepository(t)

	history := repo.GetBatchHistory(404)

	require.NotNil(t, history)
	require.Nil(t, history.Torrent)
	require.Nil(t, history.Metadata)
	require.Nil(t, history.BatchEpisodeFiles)
}

func TestAddBatchHistoryPersistsAndInvalidatesQueries(t *testing.T) {
	repo, _, ws := newTorrentstreamTestRepository(t)
	torrent := &hibiketorrent.AnimeTorrent{
		Provider: "provider",
		Name:     "[Seanime] Example Show - 01-12 (1080p).mkv",
		InfoHash: "hash-1",
		IsBatch:  true,
	}
	files := &hibiketorrent.BatchEpisodeFiles{
		Current:              0,
		CurrentEpisodeNumber: 1,
		CurrentAniDBEpisode:  "1",
		Files: []*hibiketorrent.AnimeTorrentFile{{
			Index: 0,
			Path:  "/downloads/[Seanime] Example Show - 01.mkv",
			Name:  "[Seanime] Example Show - 01.mkv",
		}},
	}

	repo.AddBatchHistory(100, torrent, files)

	require.Eventually(t, func() bool {
		got := repo.GetBatchHistory(100)
		return got.Torrent != nil && got.Torrent.InfoHash == "hash-1"
	}, 2*time.Second, 20*time.Millisecond)

	history := repo.GetBatchHistory(100)
	require.NotNil(t, history.Torrent)
	require.Equal(t, torrent.InfoHash, history.Torrent.InfoHash)
	require.NotNil(t, history.Metadata)
	require.Equal(t, "Seanime", history.Metadata.ReleaseGroup)
	require.NotNil(t, history.BatchEpisodeFiles)
	require.Equal(t, 1, history.BatchEpisodeFiles.CurrentEpisodeNumber)

	require.Eventually(t, func() bool {
		for _, event := range ws.snapshot() {
			if event.event != events.InvalidateQueries {
				continue
			}
			payload, ok := event.payload.([]string)
			if ok && len(payload) == 1 && payload[0] == events.GetTorrentstreamBatchHistoryEndpoint {
				return true
			}
		}
		return false
	}, 2*time.Second, 20*time.Millisecond)
}

func TestAddBatchHistoryUpdatesExistingRecord(t *testing.T) {
	repo, _, _ := newTorrentstreamTestRepository(t)

	repo.AddBatchHistory(101, &hibiketorrent.AnimeTorrent{
		Name:     "[Seanime] Example Show - 01-12 (1080p).mkv",
		InfoHash: "old-hash",
	}, nil)
	require.Eventually(t, func() bool {
		got := repo.GetBatchHistory(101)
		return got.Torrent != nil && got.Torrent.InfoHash == "old-hash"
	}, 2*time.Second, 20*time.Millisecond)

	updatedFiles := &hibiketorrent.BatchEpisodeFiles{CurrentEpisodeNumber: 5, CurrentAniDBEpisode: "5"}
	repo.AddBatchHistory(101, &hibiketorrent.AnimeTorrent{
		Name:     "[Seanime] Example Show - 05 (1080p).mkv",
		InfoHash: "new-hash",
	}, updatedFiles)

	require.Eventually(t, func() bool {
		got := repo.GetBatchHistory(101)
		return got.Torrent != nil && got.Torrent.InfoHash == "new-hash"
	}, 2*time.Second, 20*time.Millisecond)

	history := repo.GetBatchHistory(101)
	require.Equal(t, "new-hash", history.Torrent.InfoHash)
	require.NotNil(t, history.BatchEpisodeFiles)
	require.Equal(t, 5, history.BatchEpisodeFiles.CurrentEpisodeNumber)
	require.Equal(t, "5", history.BatchEpisodeFiles.CurrentAniDBEpisode)
}

func TestDeleteBatchHistoryRemovesRecordAndInvalidatesQueries(t *testing.T) {
	repo, _, ws := newTorrentstreamTestRepository(t)

	repo.AddBatchHistory(102, &hibiketorrent.AnimeTorrent{
		Name:     "[Seanime] Example Show - 01-12 (1080p).mkv",
		InfoHash: "hash-delete",
		IsBatch:  true,
	}, &hibiketorrent.BatchEpisodeFiles{CurrentEpisodeNumber: 1, CurrentAniDBEpisode: "1"})

	require.Eventually(t, func() bool {
		got := repo.GetBatchHistory(102)
		return got.Torrent != nil && got.Torrent.InfoHash == "hash-delete"
	}, 2*time.Second, 20*time.Millisecond)

	repo.DeleteBatchHistory(102)

	require.Eventually(t, func() bool {
		got := repo.GetBatchHistory(102)
		return got.Torrent == nil && got.Metadata == nil && got.BatchEpisodeFiles == nil
	}, 2*time.Second, 20*time.Millisecond)

	require.Eventually(t, func() bool {
		count := 0
		for _, event := range ws.snapshot() {
			if event.event != events.InvalidateQueries {
				continue
			}
			payload, ok := event.payload.([]string)
			if ok && len(payload) == 1 && payload[0] == events.GetTorrentstreamBatchHistoryEndpoint {
				count++
			}
		}
		return count >= 2
	}, 2*time.Second, 20*time.Millisecond)
}

func decodePayloadMap(t *testing.T, payload interface{}) map[string]interface{} {
	t.Helper()
	bytes, err := json.Marshal(payload)
	require.NoError(t, err)

	ret := make(map[string]interface{})
	require.NoError(t, json.Unmarshal(bytes, &ret))
	return ret
}

func newAnimeCollectionList(status anilist.MediaListStatus, entries ...*anilist.AnimeCollection_MediaListCollection_Lists_Entries) *anilist.AnimeCollection_MediaListCollection_Lists {
	return &anilist.AnimeCollection_MediaListCollection_Lists{
		Status:       &status,
		Name:         new(string(status)),
		IsCustomList: new(false),
		Entries:      entries,
	}
}

func newAnimeCollectionEntry(media *anilist.BaseAnime, progress int, status anilist.MediaListStatus) *anilist.AnimeCollection_MediaListCollection_Lists_Entries {
	return &anilist.AnimeCollection_MediaListCollection_Lists_Entries{
		Media:    media,
		Progress: &progress,
		Score:    new(8.5),
		Repeat:   new(0),
		Status:   &status,
	}
}
