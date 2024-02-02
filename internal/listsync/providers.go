package listsync

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/mal"
)

type (
	Provider struct {
		Source     Source
		Entries    []*AnimeEntry
		EntriesMap map[int]*AnimeEntry
	}
)

// NewAnilistProvider creates a new provider for Anilist
func NewAnilistProvider(collection *anilist.AnimeCollection) *Provider {
	entries := FromAnilistCollection(collection)
	entriesMap := make(map[int]*AnimeEntry)
	for _, entry := range entries {
		entriesMap[entry.MalID] = entry
	}
	return &Provider{
		Source:     SourceAniList,
		Entries:    entries,
		EntriesMap: entriesMap,
	}
}

// NewMALProvider creates a new provider for MyAnimeList
func NewMALProvider(collection []*mal.AnimeListEntry) *Provider {
	entries := FromMALCollection(collection)
	entriesMap := make(map[int]*AnimeEntry)
	for _, entry := range entries {
		entriesMap[entry.MalID] = entry
	}
	return &Provider{
		Source:     SourceMAL,
		Entries:    entries,
		EntriesMap: entriesMap,
	}
}
