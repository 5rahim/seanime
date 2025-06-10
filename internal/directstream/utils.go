package directstream

import (
	"context"
	"net/url"
	"path/filepath"
	"seanime/internal/mkvparser"
)

func getAttachmentByName(ctx context.Context, stream Stream, filename string) (*mkvparser.AttachmentInfo, bool) {
	filename, _ = url.PathUnescape(filename)

	container, err := stream.LoadPlaybackInfo()
	if err != nil {
		return nil, false
	}

	parser, ok := container.MkvMetadataParser.Get()
	if !ok {
		return nil, false
	}

	attachment, ok := parser.GetMetadata(ctx).GetAttachmentByName(filename)
	if !ok {
		return nil, false
	}

	return attachment, true
}

func isEbmlExtension(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".mkv" || ext == ".m4v" || ext == ".mp4"
}

func isEbmlContent(mimeType string) bool {
	return mimeType == "video/x-matroska" || mimeType == "video/webm"
}
