package platform

import "github.com/rs/zerolog"

// LocalPlatformManager takes care of storing, retrieving, updating, syncing data between the local database and the Anilist API
type LocalPlatformManager struct {
	logger *zerolog.Logger
}
