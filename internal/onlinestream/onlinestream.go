package onlinestream

import "github.com/rs/zerolog"

type (
	OnlineStream struct {
		logger *zerolog.Logger
	}
)

type (
	NewOnlineStreamOptions struct {
		Logger *zerolog.Logger
	}
)

func New(opts *NewOnlineStreamOptions) *OnlineStream {
	return &OnlineStream{
		logger: opts.Logger,
	}
}

func (os *OnlineStream) Start() {

}
