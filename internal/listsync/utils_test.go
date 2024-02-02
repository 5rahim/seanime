package listsync

import (
	"slices"
	"testing"
)

func TestAnimeEntry_FindMetadataDiffs(t *testing.T) {

	originEntry := &AnimeEntry{
		Source:       SourceAniList,
		SourceID:     1,
		MalID:        1,
		DisplayTitle: "Title",
		Url:          "",
		TotalEpisode: 12,
		Image:        "",
		Status:       AnimeStatusWatching,
		Progress:     6,
		Score:        8,
	}

	otherEntry := &AnimeEntry{
		Source:       SourceMAL,
		SourceID:     1,
		MalID:        1,
		DisplayTitle: "Title",
		Url:          "",
		TotalEpisode: 12,
		Image:        "",
		Status:       AnimeStatusPlanning,
		Progress:     0,
		Score:        0,
	}

	diffs, found := originEntry.FindMetadataDiffs(otherEntry)
	if !found {
		t.Fatalf("expected diffs to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffTypeStatus) {
		t.Fatalf("expected status diff to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffTypeProgress) {
		t.Fatalf("expected progress diff to be found")
	}

	if !slices.Contains(diffs, AnimeMetadataDiffTypeScore) {
		t.Fatalf("expected score diff to be found")
	}

}
