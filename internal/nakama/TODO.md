# TODO

- Make a generic interface that combines PlaybackManager, DirectstreamManager/NativePlayer, OnlinestreamPlayer
  - Generic events & calls so watch party manager doesn't need to know implementation details
  - Generic player state

```go
package main

func main()  {
	 wpm.playback.listenToPlayerEvents()
	 
	 wpm.playback.StartLocalFileStream(...)
	 wpm.playback.StartTorrentStream(...)
	 wpm.playback.StartOnlineStream(...)
	 wpm.playback.StartDebridStream(...)
	 
	 type PlaybackStatus struct {
		ID             string  `json:"id"` // path or url
		CompletionPercentage float64 `json:"completionPercentage"`
		Playing              bool    `json:"playing"`
		CurrentTime float64 `json:"currentTimeInSeconds"` // in seconds
		Duration    float64 `json:"durationInSeconds"`    // in seconds
		PlaybackType PlaybackType `json:"playbackType"` // file, torrentstream, onlinestream, debridstream
	 }
	 type PlaybackState struct {
		EpisodeNumber        int     `json:"episodeNumber"`        // The episode number
		AniDbEpisode         string  `json:"aniDbEpisode"`         // The AniDB episode number
		MediaTitle           string  `json:"mediaTitle"`           // The title of the media
		MediaCoverImage      string  `json:"mediaCoverImage"`      // The cover image of the media
		MediaTotalEpisodes   int     `json:"mediaTotalEpisodes"`   // The total number of episodes
		MediaId              int     `json:"mediaId"`              // The media ID
	 }
}

```
