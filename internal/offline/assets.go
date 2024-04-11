package offline

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/util/image_downloader"
)

type (
	assetsHandler struct {
		logger          *zerolog.Logger
		imageDownloader *image_downloader.ImageDownloader
	}
)

func newAssetsHandler(logger *zerolog.Logger, imageDownloader *image_downloader.ImageDownloader) *assetsHandler {
	return &assetsHandler{
		logger:          logger,
		imageDownloader: imageDownloader,
	}
}
