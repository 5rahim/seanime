package listsync

import (
	"slices"
	"testing"
)

func TestAnimeEntry_FindMetadataDiffs(t *testing.T) {

	originEntry := &AnimeEntry{
		Source:        SourceAniList,
		SourceID:      1,
		MalID:         1,
		DisplayTitle:  "Title",
		Url:           "",
		TotalEpisodes: 12,
		Image:         "",
		Status:        AnimeStatusWatching,
		Progress:      6,
		Score:         8,
	}

	otherEntry := &AnimeEntry{
		Source:        SourceMAL,
		SourceID:      1,
		MalID:         1,
		DisplayTitle:  "Title",
		Url:           "",
		TotalEpisodes: 12,
		Image:         "",
		Status:        AnimeStatusPlanning,
		Progress:      0,
		Score:         0,
	}

	diffs, found := originEntry.FindMetadataDiffs(otherEntry)
	if !found {
		t.Fatalf("expected diffs to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffStatus) {
		t.Fatalf("expected status diff to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffKindProgress) {
		t.Fatalf("expected progress diff to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffKindScore) {
		t.Fatalf("expected score diff to be found")
	}

}
