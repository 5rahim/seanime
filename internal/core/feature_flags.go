package core

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	FeatureFlags struct {
		MainServerTorrentStreaming bool
	}

	ExperimentalFeatureFlags struct {
	}
)

// NewFeatureFlags initializes the feature flags
func NewFeatureFlags(cfg *Config, logger *zerolog.Logger) FeatureFlags {
	ff := FeatureFlags{
		MainServerTorrentStreaming: viper.GetBool("experimental.mainServerTorrentStreaming"),
	}

	checkExperimentalFeatureFlags(&ff, cfg, logger)

	return ff
}

func checkExperimentalFeatureFlags(ff *FeatureFlags, cfg *Config, logger *zerolog.Logger) {
	if ff.MainServerTorrentStreaming {
		logger.Warn().Msg("app: [Feature flag] 'Main Server Torrent Streaming' experimental feature is enabled")
	}
}

func (ff *FeatureFlags) IsMainServerTorrentStreamingEnabled() bool {
	return ff.MainServerTorrentStreaming
}
