package constants

import (
	"seanime/internal/util"
	"time"
)

const (
	Version              = "2.8.4"
	VersionName          = "Hachi"
	GcTime               = time.Minute * 30
	ConfigFileName       = "config.toml"
	MalClientId          = "51cb4294feb400f3ddc66a30f9b9a00f"
	DiscordApplicationId = "1224777421941899285"
)

var DefaultExtensionMarketplaceURL = util.Decode("aHR0cHM6Ly9yYXcuZ2l0aHVidXNlcmNvbnRlbnQuY29tLzVyYWhpbS9zZWFuaW1lLWV4dGVuc2lvbnMvcmVmcy9oZWFkcy9tYWluL21hcmtldHBsYWNlLmpzb24=")
