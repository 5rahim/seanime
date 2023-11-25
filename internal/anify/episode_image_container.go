package anify

import "sync"

type (
	// EpisodeImageContainer holds the images that we have stored. It also gives helper functions to easily
	// request more images
	EpisodeImageContainer struct {
		Entries []*MediaEpisodeImagesEntry
		mu      sync.RWMutex
	}
)

func NewEpisodeImageContainer() *EpisodeImageContainer {
	return &EpisodeImageContainer{
		Entries: make([]*MediaEpisodeImagesEntry, 0),
	}
}

func (c *EpisodeImageContainer) GetEpisodeImage(mId int, ep int) (string, bool) {
	if c == nil {
		return "", false
	}
	for _, entry := range c.Entries {
		if entry.MediaId == mId {
			for _, img := range entry.EpisodeImageData {
				if img.EpisodeNumber == ep {
					return img.Image, true
				}
			}
		}
	}

	c.addToRequestQueue(mId)

	return "", false
}

// addToRequestQueue adds the media id to the request queue.
// The request queue is a list of media ids that need to be fetched automatically.
// This is used to prevent multiple requests for the same media id.
func (c *EpisodeImageContainer) addToRequestQueue(mId int) {
	// TODO: Implement
}

//----------------------------------------------------------------------------------------------------------------------

func (c *EpisodeImageContainer) SetEntries(e []*MediaEpisodeImagesEntry) {
	c.Entries = e
}

func (c *EpisodeImageContainer) AddEntry(e *MediaEpisodeImagesEntry) {
	c.mu.Lock()
	c.Entries = append(c.Entries, e)
	c.mu.Unlock()
}
