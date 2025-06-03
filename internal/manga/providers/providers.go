package manga_providers

import "errors"

const (
	WeebCentralProvider        = "weebcentral"
	MangadexProvider    string = "mangadex"
	ComickProvider      string = "comick"
	MangapillProvider   string = "mangapill"
	ManganatoProvider   string = "manganato"
	MangafireProvider   string = "mangafire"
	LocalProvider       string = "local-manga"
)

var (
	ErrNoResults  = errors.New("no results found")
	ErrNoChapters = errors.New("no chapters found")
	ErrNoPages    = errors.New("no pages found")
)
