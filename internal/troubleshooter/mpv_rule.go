package troubleshooter

import (
	"strings"
)

func mpvRules() *RuleBuilder {
	return NewRule("MPV Player").
		Desc("Rules for detecting MPV player related issues").
		ModuleIs(ModuleMediaPlayer).
		LevelIs(LevelError).
		Branch().
		When(func(l LogLine) bool {
			return strings.Contains(l.Message, "Could not open and play video using MPV")
		}, "MPV player failed to open video").
		Then(
			"Seanime cannot communicate with MPV",
			"Go to the settings and set the correct application path for MPV",
		).
		WithSeverity("error").
		Branch().
		When(func(l LogLine) bool {
			return strings.Contains(l.Message, "fork/exec")
		}, "MPV player process failed to start").
		Then(
			"The MPV player process failed to start",
			"Check if MPV is installed correctly and the application path is valid",
		).
		WithSeverity("error")

}
