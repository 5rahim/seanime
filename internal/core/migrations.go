package core

import (
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/util"
	"strings"
)

func (a *App) runMigrations() {

	go func() {
		defer a.Logger.Info().Msg("app: Version migration complete")
		defer util.HandlePanicThen(func() {
			a.Logger.Error().Msg("app: runMigrations failed")
		})
		if a.previousVersion != a.Version {
			versionComp, _ := util.CompareVersion(a.previousVersion, constants.Version)

			// DEVNOTE: 1.2.0 uses an incorrect manga cache format for MangaSee pages
			// This migration will remove all manga cache files that start with "manga_"
			if a.previousVersion == "1.2.0" && versionComp > 0 {
				a.Logger.Debug().Msg("app: Executing version migration task")
				err := a.FileCacher.RemoveAllBy(func(filename string) bool {
					return strings.HasPrefix(filename, "manga_")
				})
				if err != nil {
					a.Logger.Error().Err(err).Msg("app: MIGRATION FAILED; READ THIS")
					a.Logger.Error().Msg("app: Failed to remove 'manga' cache files, please clear them manually by going to the settings. Ignore this message if you have no manga cache files.")
				}
			}
		}
	}()

}
