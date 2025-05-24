package directstream

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// ServeEchoStream is a proxy to the current stream.
// It sits in between the player and the real stream (whether it's a local file, torrent, or http stream).
//
// If this is an EBML stream, it gets the range request from the player, processes it to stream the correct subtitles, and serves the video.
// Otherwise, it just serves the video.
func (m *Manager) ServeEchoStream() http.Handler {
	return m.getStreamHandler()
}

// ServeEchoAttachments serves the attachments loaded into memory from the current stream.
func (m *Manager) ServeEchoAttachments(c echo.Context) error {
	// Get the current stream
	stream, ok := m.currentStream.Get()
	if !ok {
		return errors.New("no stream")
	}

	filename := c.Param("*")

	filename, _ = url.PathUnescape(filename)

	// Get the attachment
	attachment, ok := stream.GetAttachmentByName(filename)
	if !ok {
		return errors.New("attachment not found")
	}

	return c.Blob(200, attachment.Mimetype, attachment.Data)
}
