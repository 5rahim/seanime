package core

import (
	"github.com/rs/zerolog"
)

type (
	FeatureFlags struct {
	}

	ExperimentalFeatureFlags struct {
	}
)

// NewFeatureFlags initializes the feature flags
func NewFeatureFlags(cfg *Config, logger *zerolog.Logger) FeatureFlags {
	ff := FeatureFlags{}

	return ff
}

func checkExperimentalFeatureFlags(ff *FeatureFlags, cfg *Config, logger *zerolog.Logger) {
	{
		//
		// Mediastream Feature Flag
		//
		//if ff.Experimental.Mediastream {
		//	enabled := true
		//	// Check that Jassub is in the asset directory
		//	jassubPath := filepath.Join(cfg.Web.AssetDir, "/jassub/jassub-worker.js")
		//	if _, err := os.Stat(jassubPath); os.IsNotExist(err) {
		//		logger.Error().Msgf("app: [Feature flag] 'Media streaming' is enabled but JASSUB is not in the asset directory. Disabling Mediastream feature flag.")
		//		ff.Experimental.Mediastream = false
		//		enabled = false
		//	}
		//
		//	if enabled {
		//		logger.Warn().Msg("app: [Feature flag] 'Media streaming' feature flag is enabled")
		//	}
		//}
	}
}
