package manga_providers

import "errors"

const (
	LocalProvider string = "local-manga"
)

var (
	ErrNoResults  = errors.New("no results found")
	ErrNoChapters = errors.New("no chapters found")
	ErrNoPages    = errors.New("no pages found")
)
