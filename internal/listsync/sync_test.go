package listsync

import "testing"

func TestListSync_CheckDiff(t *testing.T) {
	originEntries := []*AnimeEntry{
		{
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
		},
		{
			Source:       SourceAniList,
			SourceID:     21,
			MalID:        2,
			DisplayTitle: "Title2",
			Url:          "",
			TotalEpisode: 12,
			Image:        "",
			Status:       AnimeStatusPlanning,
			Progress:     0,
			Score:        0,
		},
	}

	originEntriesMap := make(map[int]*AnimeEntry)
	for _, entry := range originEntries {
		originEntriesMap[entry.MalID] = entry
	}

	origin := &Provider{
		Source:     SourceAniList,
		Entries:    originEntries,
		EntriesMap: originEntriesMap,
	}

	targetEntries := []*AnimeEntry{
		{
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
		},
	}

	targetEntriesMap := make(map[int]*AnimeEntry)
	for _, entry := range targetEntries {
		targetEntriesMap[entry.MalID] = entry
	}

	target := &Provider{
		Source:     SourceMAL,
		Entries:    targetEntries,
		EntriesMap: targetEntriesMap,
	}

	ls := NewListSync(origin, []*Provider{
		target,
	})

	diffs := ls.CheckDiffs()
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(diffs))
	}

	if diffs[0].Type != AnimeDiffKindMissingTarget {
		t.Fatalf("expected first diff to be missing_in_target, got %s", diffs[0].Type)
	}

	if diffs[1].Type != AnimeDiffKindMetadata {
		t.Fatalf("expected second diff to be metadata, got %s", diffs[1].Type)
	}

}
