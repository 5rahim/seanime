package playlist

import (
	"context"
	"seanime/internal/directstream"
	"seanime/internal/library/playbackmanager"
	"sync"

	"github.com/rs/zerolog"
)

type PlayerType string

const (
	PlayerTypeDesktop            PlayerType = "desktop"
	PlayerTypeExternalPlayerLink PlayerType = "externalPlayerLink"
	PlayerTypeNativePlayer       PlayerType = "nativeplayer"
)

type PlaylistStateItem struct {
	Name       string `json:"name"`
	MediaImage string `json:"mediaImage"`
}

type Manager struct {
	playerType PlayerType

	directstreamManager *directstream.Manager
	playbackManager     *playbackmanager.PlaybackManager

	mu     sync.Mutex
	logger *zerolog.Logger
	cancel context.CancelFunc
}
