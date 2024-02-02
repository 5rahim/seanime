package listsync

const (
	AnimeDiffTypeMissingOrigin    AnimeDiffType         = "missing_in_origin" // Anime is missing in the origin (i.e. Delete from target)
	AnimeDiffTypeMissingTarget    AnimeDiffType         = "missing_in_target" // Anime is missing in the target (i.e. Add to target)
	AnimeDiffTypeMetadata         AnimeDiffType         = "metadata"          // Anime metadata is different in the origin and the target (i.e. Update in target)
	AnimeMetadataDiffTypeScore    AnimeMetadataDiffType = "score"
	AnimeMetadataDiffTypeProgress AnimeMetadataDiffType = "progress"
	AnimeMetadataDiffTypeStatus   AnimeMetadataDiffType = "status"
)

type (
	AnimeDiffType         string
	AnimeMetadataDiffType string
	ListSync              struct {
		Origin  *Provider
		Targets []*Provider
	}

	MissingAnime struct {
		Provider      *Provider
		OriginEntries []*AnimeEntry // Entries that are present in the origin but not in the target
	}
	AnimeDiff struct {
		Provider      *Provider     // The provider that has the diff
		OriginEntries []*AnimeEntry // Entries that are different in the origin and the target
		Type          AnimeDiffType
	}
)

// NewListSync creates a new list sync
func NewListSync(origin *Provider, targets []*Provider) *ListSync {
	return &ListSync{
		Origin:  origin,
		Targets: targets,
	}
}

func (ls *ListSync) CheckDiffs() []*AnimeDiff {
	diff := make([]*AnimeDiff, 0)

	for _, target := range ls.Targets {
		// First, check for missing anime in the target
		missing, ok := checkMissingFrom(ls.Origin, target)
		if ok {
			diff = append(diff, &AnimeDiff{
				Provider:      target,
				OriginEntries: missing.OriginEntries,
				Type:          AnimeDiffTypeMissingTarget,
			})
		}

		// Then, check for missing anime in the origin
		missing, ok = checkMissingFrom(target, ls.Origin)
		if ok {
			diff = append(diff, &AnimeDiff{
				Provider:      target,
				OriginEntries: missing.OriginEntries,
				Type:          AnimeDiffTypeMissingOrigin,
			})
		}

		// Finally, check for different metadata
		for _, entry := range ls.Origin.Entries {
			if targetEntry, ok := target.EntriesMap[entry.MalID]; ok {
				_, found := entry.FindMetadataDiffs(targetEntry)
				if found {
					diff = append(diff, &AnimeDiff{
						Provider:      target,
						OriginEntries: []*AnimeEntry{entry},
						Type:          AnimeDiffTypeMetadata,
					})
				}
			}
		}
	}

	return diff
}

// CheckMissingFrom checks for anime that are present in the origin but not in the target.
func checkMissingFrom(origin *Provider, target *Provider) (*MissingAnime, bool) {
	missing := make([]*AnimeEntry, 0)

	for _, entry := range origin.Entries {
		if _, ok := target.EntriesMap[entry.MalID]; !ok {
			missing = append(missing, entry)
		}
	}

	if len(missing) == 0 {
		return nil, false
	}

	return &MissingAnime{
		Provider:      target,
		OriginEntries: missing,
	}, true
}

// SyncMetadata syncs metadata between the origin and targets when a match is found.
// It does not sync the lists themselves.
func (ls *ListSync) SyncMetadata() {
}

// SyncMedia syncs media between the origin and targets
func (ls *ListSync) SyncMedia() {
}
