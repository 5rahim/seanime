package entities

import (
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const mediaEntryMockDataFile = "./media_entry_test_mock_data.json"

// TestNewMediaEntry tests /library/entry endpoint.
func TestNewMediaEntry(t *testing.T) {

	var mediaId = 154587 // Sousou no Frieren
	lfs, anilistCollection, err := getMediaEntryMockData(t, mediaId)

	if assert.NoErrorf(t, err, "Failed to get mock data") &&
		assert.NotNil(t, lfs) &&
		assert.NotNil(t, anilistCollection) {

		entry, err := NewMediaEntry(&NewMediaEntryOptions{
			MediaId:           mediaId,
			LocalFiles:        lfs,
			AnizipCache:       anizip.NewCache(),
			AnilistCollection: anilistCollection,
			AnilistClient:     anilist.MockGetAnilistClient(),
		})

		if assert.NoError(t, err) {

			assert.Equalf(t, len(lfs), len(entry.Episodes), "Number of episodes mismatch")

			// Mock data progress is 4
			if assert.NotNilf(t, entry.NextEpisode, "Next episode not found") {
				assert.Equal(t, 5, entry.NextEpisode.EpisodeNumber, "Next episode mismatch")
			}

			t.Logf("Success, found %v episodes", len(entry.Episodes))

		}

	}
}

func TestNewMediaEntry2(t *testing.T) {

	var mediaId = 146065 // Mushoku Tensei: Jobless Reincarnation Season 2
	_, anilistCollection, err := getMediaEntryMockData(t, mediaId)

	var lfs []*LocalFile
	for idx, path := range []string{
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 00 (1080p) [9C362DC3].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 01 (1080p) [EC64C8B1].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 02 (1080p) [7EA9E789].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 03 (1080p) [BEF3095D].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 04 (1080p) [FD2285EB].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 05 (1080p) [E691CDB3].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 06 (1080p) [0438103E].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 07 (1080p) [DA6366AD].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 08 (1080p) [A761377D].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 09 (1080p) [DFE9A041].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 10 (1080p) [DFE1B93B].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 11 (1080p) [F70DC34C].mkv",
		"E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 12 (1080p) [BAA0EBAD].mkv",
	} {
		lf := NewLocalFile(path, "E:/Anime")
		// Mock hydration
		lf.MediaId = mediaId
		lf.Metadata = &LocalFileMetadata{
			Type:         LocalFileTypeMain,
			Episode:      idx,
			AniDBEpisode: strconv.Itoa(idx),
		}
		if idx == 0 {
			lf.Metadata.AniDBEpisode = "S1"
		}
		lfs = append(lfs, lf)
	}

	if assert.NoErrorf(t, err, "Failed to get mock data") &&
		assert.NotNil(t, lfs) &&
		assert.NotNil(t, anilistCollection) {

		entry, err := NewMediaEntry(&NewMediaEntryOptions{
			MediaId:           mediaId,
			LocalFiles:        lfs,
			AnizipCache:       anizip.NewCache(),
			AnilistCollection: anilistCollection,
			AnilistClient:     anilist.MockGetAnilistClient(),
		})

		if assert.NoError(t, err) {

			assert.Equalf(t, len(lfs), len(entry.Episodes), "Number of episodes mismatch")

			// Mock data progress is 0 so the next episode should be "0" (S1) with progress number 1
			if assert.NotNilf(t, entry.NextEpisode, "Next episode not found") {
				assert.Equal(t, 0, entry.NextEpisode.EpisodeNumber, "Next episode number mismatch")
				assert.Equal(t, 1, entry.NextEpisode.ProgressNumber, "Next episode progress number mismatch")
			}

			t.Logf("Success, found %v episodes", len(entry.Episodes))

		}

	}
}

func TestNewMissingEpisodes(t *testing.T) {

	var mediaId = 154587 // Sousou no Frieren
	lfs, anilistCollection, err := getMediaEntryMockData(t, mediaId)

	if assert.NoError(t, err) {
		missingData := NewMissingEpisodes(&NewMissingEpisodesOptions{
			AnilistCollection: anilistCollection,
			LocalFiles:        lfs,
			AnizipCache:       anizip.NewCache(),
		})

		// Mock data has 5 files, Current number of episodes is 10, so 5 episodes are missing
		assert.Equal(t, 5, len(missingData.Episodes))
	}

}

// Fetching "Mushoku Tensei: Jobless Reincarnation Season 2" download info
// Anilist: 13 episodes, Anizip: 12 episodes + "S1"
// Info should include "S1" as episode 0
func TestNewMediaEntryDownloadInfo(t *testing.T) {

	var mediaId = 146065 // Mushoku Tensei: Jobless Reincarnation Season 2

	_, anilistCollection, err := getMediaEntryMockData(t, mediaId)
	if err != nil {
		t.Fatal(err)
	}

	anizipData, err := anizip.FetchAniZipMedia("anilist", mediaId)
	if err != nil {
		t.Fatal(err)
	}

	anilistEntry, ok := anilistCollection.GetListEntryFromMediaId(mediaId)

	if assert.Truef(t, ok, "Could not find media entry for %d", mediaId) {

		assert.Equal(t, 13, anilistEntry.Media.GetCurrentEpisodeCount(), "Number of episodes mismatch on Anilist")
		assert.Equal(t, 12, anizipData.GetMainEpisodeCount(), "Number of episodes mismatch on Anizip")

		info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
			localFiles:  nil,
			anizipMedia: anizipData,
			progress:    lo.ToPtr(0),
			status:      lo.ToPtr(anilist.MediaListStatusCurrent),
			media:       anilistEntry.Media,
		})

		if assert.NoError(t, err) && assert.NotNil(t, info) {

			_, found := lo.Find(info.EpisodesToDownload, func(i *MediaEntryDownloadEpisode) bool {
				return i.EpisodeNumber == 0 && i.AniDBEpisode == "S1"
				// && i.Episode.ProgressNumber == 1 DEVNOTE: Progress numbers are always 0 because we don't have local files
			})

			assert.True(t, found)

		}

	}

}

//----------------------------------------------------------------------------------------------------------------------

func getMediaEntryMockData(t *testing.T, mediaId int) ([]*LocalFile, *anilist.AnimeCollection, error) {

	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Open the JSON file
	file, err := os.Open(filepath.Join(path, mediaEntryMockDataFile))
	if err != nil {
		t.Fatal("Error opening file:", err.Error())
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("Error reading file:", err.Error())
	}

	var data map[string]struct {
		LocalFiles        []*LocalFile             `json:"localFiles"`
		AnilistCollection *anilist.AnimeCollection `json:"anilistCollection"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatal("Error unmarshaling JSON:", err.Error())
	}

	ret, _ := data[strconv.Itoa(mediaId)]

	return ret.LocalFiles, ret.AnilistCollection, nil

}
