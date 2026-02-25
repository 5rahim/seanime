package constants

import (
	"seanime/internal/util"
	"time"
)

const (
	Version              = "3.5.1"
	VersionName          = "Hakumei"
	GcTime               = time.Minute * 30
	ConfigFileName       = "config.toml"
	MalClientId          = "51cb4294feb400f3ddc66a30f9b9a00f"
	DiscordApplicationId = "1224777421941899285"
	AnilistApiUrl        = "https://graphql.anilist.co"
	IsRspackFrontend     = true
)

const (
	SeanimeRoomsApiUrl   = "https://seanime.app/api/rooms"
	SeanimeRoomsApiWsUrl = "wss://seanime.app/api/rooms"
	SeanimeRoomsVersion  = "1.0.0"
)

var DefaultExtensionMarketplaceURL = util.Decode("aHR0cHM6Ly9yYXcuZ2l0aHVidXNlcmNvbnRlbnQuY29tLzVyYWhpbS9zZWFuaW1lLWV4dGVuc2lvbnMvcmVmcy9oZWFkcy9tYWluL21hcmtldHBsYWNlLmpzb24=")
var AnnouncementURL = util.Decode("aHR0cHM6Ly9yYXcuZ2l0aHVidXNlcmNvbnRlbnQuY29tLzVyYWhpbS9oaWJpa2UvcmVmcy9oZWFkcy9tYWluL3B1YmxpYy9hbm5vdW5jZW1lbnRzLmpzb24=")
var InternalMetadataURL = util.Decode("aHR0cHM6Ly9hbmltZS5jbGFwLmluZw==")
