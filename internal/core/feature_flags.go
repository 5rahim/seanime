package core

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type (
	FeatureFlags struct {
		Experimental ExperimentalFeatureFlags `json:"experimental"`
	}

	ExperimentalFeatureFlags struct {
		Mediastream bool `json:"mediastream"`
	}
)

// NewFeatureFlags initializes the feature flags
func NewFeatureFlags(cfg *Config, logger *zerolog.Logger) FeatureFlags {
	ff := FeatureFlags{}

	ff.Experimental.Mediastream = viper.GetBool("experimental.mediastream")

	checkExperimentalFeatureFlags(&ff, cfg, logger)

	return ff
}

func checkExperimentalFeatureFlags(ff *FeatureFlags, cfg *Config, logger *zerolog.Logger) {
	{
		//
		// Mediastream Feature Flag
		//
		if ff.Experimental.Mediastream {
			enabled := true
			// Check that Jassub is in the asset directory
			jassubPath := filepath.Join(cfg.Web.AssetDir, "/jassub/jassub-worker.js")
			if _, err := os.Stat(jassubPath); os.IsNotExist(err) {
				logger.Error().Msgf("app: [Feature flag] 'Media streaming' is enabled but JASSUB is not in the asset directory. Disabling Mediastream feature flag.")
				ff.Experimental.Mediastream = false
				enabled = false
			}

			if enabled {
				logger.Warn().Msg("app: [Feature flag] 'Media streaming' feature flag is enabled")
			}
		}
	}
}

func (ff *FeatureFlags) IsExperimentalMediastreamEnabled() bool {
	return ff.Experimental.Mediastream
}
