package core

import (
	"seanime/internal/constants"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Tours represents all version tours available to highlight changes
var Tours = map[string][]string{
	// Tour version -> [previous version, current version]
	"3.5.0": {"< 3.5.0", "< 3.6.0, >= 3.5.0"},
}

// runMigrations checks the previous version and runs any necessary migrations based on the version difference.
// This is run synchronously on app startup.
func (a *App) runMigrations() {

	//go func() {
	a.Logger.Debug().Msg("app: Checking for version migrations")
	done := false
	defer func() {
		if done {
			a.Logger.Info().Msg("app: Version migration complete")
		}
	}()
	defer util.HandlePanicThen(func() {
		a.Logger.Error().Msg("app: runMigrations failed")
	})

	previousVersion, err := semver.NewVersion(a.previousVersion)
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to parse previous version")
		return
	}

	if a.previousVersion != constants.Version {

		hasUpdated := util.VersionIsOlderThan(a.previousVersion, constants.Version)

		// handle tours
		if hasUpdated {
			currVersion, _ := semver.NewVersion(constants.Version)
			if currVersion != nil {
				for tourVersion, tourConstraints := range Tours {
					from := tourConstraints[0]
					to := tourConstraints[1]
					fromC, err := semver.NewConstraint(from)
					if err != nil {
						continue
					}
					toC, err := semver.NewConstraint(to)
					if err != nil {
						continue
					}
					if fromC.Check(previousVersion) && toC.Check(currVersion) {
						a.Logger.Debug().Msgf("app: Tour for %s", tourVersion)
						a.ShowTour = tourVersion
					}
				}
			}
		}

		//-----------------------------------------------------------------------------------------
		// DEVNOTE: 1.2.0 uses an incorrect manga cache format for MangaSee pages
		// This migration will remove all manga cache files that start with "manga_"
		if a.previousVersion == "1.2.0" && hasUpdated {
			a.Logger.Debug().Msg("app: Executing version migration task")
			err := a.FileCacher.RemoveAllBy(func(filename string) bool {
				return strings.HasPrefix(filename, "manga_")
			})
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
				a.Logger.Error().Msg("app: Failed to remove 'manga' cache files, please clear them manually by going to the settings. Ignore this message if you have no manga cache files.")
			}
			done = true
		}

		//-----------------------------------------------------------------------------------------

		c1, _ := semver.NewConstraint("<= 1.3.0, >= 1.2.0")
		if c1.Check(previousVersion) {
			a.Logger.Debug().Msg("app: Executing version migration task")
			err := a.FileCacher.RemoveAllBy(func(filename string) bool {
				return strings.HasPrefix(filename, "manga_")
			})
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
				a.Logger.Error().Msg("app: Failed to remove 'manga' cache files, please clear them manually by going to the settings. Ignore this message if you have no manga cache files.")
			}
			done = true
		}

		//-----------------------------------------------------------------------------------------

		// DEVNOTE: 1.5.6 uses a different cache format for media streaming info
		// -> Delete the cache files when updated from any version between 1.5.0 and 1.5.5
		c2, _ := semver.NewConstraint("<= 1.5.5, >= 1.5.0")
		if c2.Check(previousVersion) {
			a.Logger.Debug().Msg("app: Executing version migration task")
			err := a.FileCacher.RemoveAllBy(func(filename string) bool {
				return strings.HasPrefix(filename, "mediastream_mediainfo_")
			})
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
				a.Logger.Error().Msg("app: Failed to remove transcoding cache files, please clear them manually by going to the settings. Ignore this message if you have no transcoding cache files.")
			}
			done = true
		}

		//-----------------------------------------------------------------------------------------

		// DEVNOTE: 2.0.0 uses a different cache format for online streaming
		// -> Delete the cache files when updated from a version older than 2.0.0 and newer than 1.5.0
		c3, _ := semver.NewConstraint("< 2.0.0, >= 1.5.0")
		if c3.Check(previousVersion) {
			a.Logger.Debug().Msg("app: Executing version migration task")
			err := a.FileCacher.RemoveAllBy(func(filename string) bool {
				return strings.HasPrefix(filename, "onlinestream_")
			})
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
				a.Logger.Error().Msg("app: Failed to remove online streaming cache files, please clear them manually by going to the settings. Ignore this message if you have no online streaming cache files.")
			}
			done = true
		}

		//-----------------------------------------------------------------------------------------

		// DEVNOTE: 2.1.0 refactored the manga cache format
		// -> Delete the cache files when updated from a version older than 2.1.0
		c4, _ := semver.NewConstraint("< 2.1.0")
		if c4.Check(previousVersion) {
			a.Logger.Debug().Msg("app: Executing version migration task")
			err := a.FileCacher.RemoveAllBy(func(filename string) bool {
				return strings.HasPrefix(filename, "manga_")
			})
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
				a.Logger.Error().Msg("app: Failed to remove 'manga' cache files, please clear them manually by going to the settings. Ignore this message if you have no manga cache files.")
			}
			done = true
		}
		//-----------------------------------------------------------------------------------------

		// DEVNOTE: 3.0.4 refactored custom source IDs and fixed a bug with extension identifiers
		// -> Delete the custom source collections
		c5, _ := semver.NewConstraint("< 3.0.4, >= 3.0.0")
		if c5.Check(previousVersion) {
			a.Logger.Debug().Msg("app: Executing version migration task (deleting custom source collections)")
			err := a.Database.Gorm().Where("1 = 1").Delete(&models.CustomSourceCollection{}).Error
			if err != nil {
				a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED")
			}
			done = true
		}

		//-----------------------------------------------------------------------------------------

		//c6, _ := semver.NewConstraint("< 3.4.0")
		//if c5.Check(previousVersion) {
		//	a.Logger.Debug().Msg("app: Executing version migration task (deleting custom source collections)")
		//	err := a.Database.Gorm().Where("1 = 1").Delete(&models.CustomSourceCollection{}).Error
		//	if err != nil {
		//		a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED")
		//	}
		//	done = true
		//}
	}
	//}()

}
