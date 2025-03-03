package plugin_ui

type ActionManager struct {
	ctx *Context

	animeActionButtons []*AnimeActionButton
}

// AnimeActionButton is a button that appears on the anime page.
// It can be created by a plugin.
type AnimeActionButton struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Intent  string `json:"intent"`
	OnClick string `json:"onClick"` // Event handler name
}

type AnimeActionDropdownItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	OnClick string `json:"onClick"` // Event handler name
}

type AnimeLibraryActionDropdownItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	OnClick string `json:"onClick"` // Event handler name
}

type MangaActionButton struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Intent  string `json:"intent"`
	OnClick string `json:"onClick"` // Event handler name
}
