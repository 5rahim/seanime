package anilist

import (
	"seanime/internal/util/filecache"
	"sync/atomic"
)

type LocalCache struct {
	fileCacher *filecache.Cacher
	// Whether the API client is currently working
	// If not, the client will use the cached data
	isWorking atomic.Bool
}
