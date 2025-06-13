package nakama

import (
	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type WatchPartyManager struct {
	logger  *zerolog.Logger
	manager *Manager

	watchParty mo.Option[*WatchParty]
}

type WatchParty struct {
}

func NewWatchPartyManager(manager *Manager) *WatchPartyManager {
	return &WatchPartyManager{logger: manager.logger, manager: manager}
}
