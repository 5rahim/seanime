package fillermanager

import (
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
	"seanime/internal/onlinestream"
)

// HydrateFillerDataRequestedEvent is triggered when the filler manager requests to hydrate the filler data for an entry.
// This is used by the local file episode list.
// Prevent default to skip the default behavior and return your own data.
type HydrateFillerDataRequestedEvent struct {
	hook_resolver.Event
	Entry *anime.Entry `json:"entry"`
}

// HydrateOnlinestreamFillerDataRequestedEvent is triggered when the filler manager requests to hydrate the filler data for online streaming episodes.
// This is used by the online streaming episode list.
// Prevent default to skip the default behavior and return your own data.
type HydrateOnlinestreamFillerDataRequestedEvent struct {
	hook_resolver.Event
	Episodes []*onlinestream.Episode `json:"episodes"`
}

// HydrateEpisodeFillerDataRequestedEvent is triggered when the filler manager requests to hydrate the filler data for specific episodes.
// This is used by the torrent and debrid streaming episode list.
// Prevent default to skip the default behavior and return your own data.
type HydrateEpisodeFillerDataRequestedEvent struct {
	hook_resolver.Event
	Episodes []*anime.Episode `json:"episodes"`
}
