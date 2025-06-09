package debrid_client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/util"
	"strings"
)

func (s *StreamManager) getMediaInfo(ctx context.Context, mediaId int) (media *anilist.CompleteAnime, animeMetadata *metadata.AnimeMetadata, err error) {
	// Get the media
	var found bool
	media, found = s.repository.completeAnimeCache.Get(mediaId)
	if !found {
		// Fetch the media
		media, err = s.repository.platform.GetAnimeWithRelations(ctx, mediaId)
		if err != nil {
			return nil, nil, fmt.Errorf("torrentstream: Failed to fetch media: %w", err)
		}
	}

	// Get the media
	animeMetadata, err = s.repository.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
	if err != nil {
		//return nil, nil, fmt.Errorf("torrentstream: Could not fetch AniDB media: %w", err)
		animeMetadata = &metadata.AnimeMetadata{
			Titles:       make(map[string]string),
			Episodes:     make(map[string]*metadata.EpisodeMetadata),
			EpisodeCount: 0,
			SpecialCount: 0,
			Mappings: &metadata.AnimeMappings{
				AnilistId: media.GetID(),
			},
		}
		animeMetadata.Titles["en"] = media.GetTitleSafe()
		animeMetadata.Titles["x-jat"] = media.GetRomajiTitleSafe()
		err = nil
	}

	return
}

func CanStream(streamUrl string) (bool, string) {
	hasExtension, isArchive := IsArchive(streamUrl)

	// If we were able to verify that the stream URL is an archive, we can't stream it
	if isArchive {
		return false, "Stream URL is an archive"
	}

	// If the stream URL has an extension, we can stream it
	if hasExtension {
		ext := filepath.Ext(streamUrl)
		if util.IsValidVideoExtension(ext) {
			return true, ""
		}
		// If the extension is not a valid video extension, we can't stream it
		return false, "Stream URL is not a valid video extension"
	}

	// If the stream URL doesn't have an extension, we'll get the headers to check if it's a video
	// If the headers are not available, we can't stream it

	contentType, err := GetContentType(streamUrl)
	if err != nil {
		return false, "Failed to get content type"
	}

	if strings.HasPrefix(contentType, "video/") {
		return true, ""
	}

	return false, fmt.Sprintf("Stream URL of type %q is not a video", contentType)
}

func IsArchive(streamUrl string) (hasExtension bool, isArchive bool) {
	ext := filepath.Ext(streamUrl)
	if ext == ".zip" || ext == ".rar" {
		return true, true
	}

	if ext != "" {
		return true, false
	}

	return false, false
}

func GetContentTypeHead(url string) string {
	resp, err := http.Head(url)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	return resp.Header.Get("Content-Type")
}

func GetContentType(url string) (string, error) {
	// Try using HEAD request
	if cType := GetContentTypeHead(url); cType != "" {
		return cType, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Only read a small amount of data to determine the content type.
	req.Header.Set("Range", "bytes=0-511")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the first 512 bytes
	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Detect content type based on the read bytes
	contentType := http.DetectContentType(buf[:n])

	return contentType, nil
}
