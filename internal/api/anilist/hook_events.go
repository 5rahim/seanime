package anilist

import "seanime/internal/hook_resolver"

// ListMissedSequelsRequestedEvent is triggered when the list missed sequels request is requested.
// Prevent default to skip the default behavior and return your own data.
type ListMissedSequelsRequestedEvent struct {
	hook_resolver.Event
	AnimeCollectionWithRelations *AnimeCollectionWithRelations `json:"animeCollectionWithRelations"`
	Variables                    map[string]interface{}        `json:"variables"`
	Query                        string                        `json:"query"`
	// Empty data object, will be used if the hook prevents the default behavior
	List []*BaseAnime `json:"list"`
}

type ListMissedSequelsEvent struct {
	hook_resolver.Event
	List []*BaseAnime `json:"list"`
}
