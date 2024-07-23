package manga_providers

import "errors"

const (
	MangaseeProvider  string = "mangasee"
	MangadexProvider  string = "mangadex"
	ComickProvider    string = "comick"
	MangapillProvider string = "mangapill"
	ManganatoProvider string = "manganato"
	MangafireProvider string = "mangafire"
)

var (
	ErrNoResults  = errors.New("no results found")
	ErrNoChapters = errors.New("no chapters found")
	ErrNoPages    = errors.New("no pages found")
)
