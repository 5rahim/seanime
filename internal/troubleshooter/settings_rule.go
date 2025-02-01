package troubleshooter

func mediaPlayerRules(state *AppState) *RuleBuilder {
	return NewRule("Media Player")
	// Desc("Rules for detecting media player issues").
	// ModuleIs(ModuleMediaPlayer).
	// LevelIs(LevelError).
	// // Branch that checks if MPV is configured
	// Branch().
	// When(func(l LogLine) bool {
	// 	mpvPath, ok := state.Settings["mpv_path"].(string)
	// 	return strings.Contains(l.Message, "player not found") && (!ok || mpvPath == "")
	// }, "MPV not configured").
	// Then(
	// 	"MPV player is not configured",
	// 	"Go to settings and configure the MPV player path",
	// ).
	// WithSeverity("error").
}
