package nakama

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/playbackmanager"
	"strconv"
	"strings"
)

type (
	HydrateHostAnimeLibraryOptions struct {
		AnimeCollection   *anilist.AnimeCollection
		LibraryCollection *anime.LibraryCollection
		MetadataProvider  metadata.Provider
	}
)

// HydrateHostAnimeLibrary hydrates the anime collection object with the host's anime library
func (m *Manager) HydrateHostAnimeLibrary(opts *HydrateHostAnimeLibraryOptions) {
	if !m.settings.Enabled || !m.settings.IncludeNakamaAnimeLibrary || !m.IsConnectedToHost() {
		return
	}

	// Send a HTTP request to the host to get the anime library
	response, err := m.reqClient.R().
		SetHeader("X-Seanime-Nakama-Password", m.settings.RemoteServerPassword).
		Get(m.GetHostBaseServerURL() + "/api/v1/nakama/host/anime/library/collection")
	if err != nil {
		return
	}

	if !response.IsSuccessState() {
		return
	}

	body := response.Bytes()

	var libraryCollectionResponse struct {
		Data *anime.LibraryCollection `json:"data"`
	}
	err = json.Unmarshal(body, &libraryCollectionResponse)
	if err != nil {
		return
	}

	for _, ep := range libraryCollectionResponse.Data.ContinueWatchingList {
		ep.IsNakamaEpisode = true
	}

	*opts.LibraryCollection = *libraryCollectionResponse.Data
}

func (m *Manager) GetHostAnimeLibraryFiles(mId int) (lfs []*anime.LocalFile, hydrated bool) {
	if !m.settings.Enabled || !m.settings.IncludeNakamaAnimeLibrary || !m.IsConnectedToHost() {
		return nil, false
	}

	response, err := m.reqClient.R().
		SetHeader("X-Seanime-Nakama-Password", m.settings.RemoteServerPassword).
		Get(m.GetHostBaseServerURL() + "/api/v1/nakama/host/anime/library/files/" + strconv.Itoa(mId))
	if err != nil {
		return nil, false
	}

	if !response.IsSuccessState() {
		return nil, false
	}

	body := response.Bytes()

	var entryResponse struct {
		Data []*anime.LocalFile `json:"data"`
	}
	err = json.Unmarshal(body, &entryResponse)
	if err != nil {
		return nil, false
	}

	return entryResponse.Data, true
}

func (m *Manager) PlayHostAnimeLibraryFile(path string, userAgent string, media *anilist.BaseAnime, aniDBEpisode string) error {
	if !m.settings.Enabled || !m.settings.IncludeNakamaAnimeLibrary || !m.IsConnectedToHost() {
		return errors.New("not connected to host")
	}

	m.wsEventManager.SendEvent(events.ShowIndefiniteLoader, "nakama-file")
	m.wsEventManager.SendEvent(events.InfoToast, "Sending stream to player...")

	// Send a HTTP request to the host to get the anime library
	// If we can access it then the host is sharing its anime library
	response, err := m.reqClient.R().
		SetHeader("X-Seanime-Nakama-Password", m.settings.RemoteServerPassword).
		Get(m.GetHostBaseServerURL() + "/api/v1/nakama/host/anime/library/collection")
	if err != nil {
		return fmt.Errorf("cannot access host's anime library: %w", err)
	}

	if !response.IsSuccessState() {
		return fmt.Errorf("cannot access host's anime library: %w", err)
	}

	host := m.serverHost
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	address := fmt.Sprintf("%s:%d", host, m.serverPort)
	ret := fmt.Sprintf("http://%s/api/v1/nakama/stream?type=file&path=%s", address, base64.StdEncoding.EncodeToString([]byte(path)))
	if strings.HasPrefix(ret, "http://http") {
		ret = strings.Replace(ret, "http://http", "http", 1)
	}

	playbackSubscriber := m.playbackManager.SubscribeToPlaybackStatus("nakama-file")

	err = m.playbackManager.StartStreamingUsingMediaPlayer("", &playbackmanager.StartPlayingOptions{
		Payload:   ret,
		UserAgent: userAgent,
		ClientId:  "",
	}, media, aniDBEpisode)
	if err != nil {
		go m.playbackManager.UnsubscribeFromPlaybackStatus("nakama-file")
		return err
	}

	go func(playbackSubscriber *playbackmanager.PlaybackStatusSubscriber) {
		for {
			select {
			case event := <-playbackSubscriber.EventCh:
				switch event.(type) {
				case playbackmanager.StreamStartedEvent:
					m.wsEventManager.SendEvent(events.HideIndefiniteLoader, "nakama-file")
					go m.playbackManager.UnsubscribeFromPlaybackStatus("nakama-file")
				}
			}
		}
	}(playbackSubscriber)

	return nil
}
