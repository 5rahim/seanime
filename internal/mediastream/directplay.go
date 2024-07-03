package mediastream

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/seanime-app/seanime/internal/events"
	"io"
	"os"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Direct
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ServeFiberDirectPlay(ctx *fiber.Ctx, clientId string) error {

	if !r.IsInitialized() {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Module not initialized")
		return errors.New("module not initialized")
	}

	// Get current media
	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
	if !found {
		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "no file has been loaded")
		return errors.New("no file has been loaded")
	}

	return r.streamVideo(ctx, mediaContainer.Filepath, clientId)
}

type VideoStream struct {
	FilePath string
	File     *os.File
	FileSize int64
}

func (r *Repository) streamVideo(ctx *fiber.Ctx, filePath string, clientId string) error {

	// Check if the file is already open
	videoStream, found := r.directPlayVideoStreamCache.Get(clientId)
	if !found || videoStream.FilePath != filePath {

		// If a file was previously opened by the client, close it
		if videoStream != nil {
			go func(vs *VideoStream) {
				_ = vs.File.Close()
			}(videoStream)
		}

		// Open the video file
		f, err := os.Open(filePath)
		if err != nil {
			r.logger.Error().Err(err).Msgf("mediastream: Error opening video file")
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Get the file information
		info, err := f.Stat()
		if err != nil {
			_ = f.Close()
			r.logger.Error().Err(err).Msgf("mediastream: Error getting file information")
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		videoStream = &VideoStream{
			FilePath: filePath,
			FileSize: info.Size(),
			File:     f,
		}

		r.directPlayVideoStreamCache.Set(clientId, videoStream)
	}

	// Default chunk size for partial content to 5MB
	const defaultChunkSize int64 = 5242880

	// Get the range header from the request
	rangeHeader := ctx.Get("Range")
	if rangeHeader != "" {
		var start, end int64

		ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		if len(ranges) != 2 {
			r.logger.Error().Msg("mediastream: Invalid Range Header")
			return ctx.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid Range Header")
		}

		// Get the start range
		start, err := strconv.ParseInt(ranges[0], 10, 64)
		if err != nil || start >= videoStream.FileSize {
			r.logger.Error().Err(err).Msg("mediastream: Error parsing start byte position or start out of range")
			return ctx.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid Range Start")
		}

		// Calculate the end range if not provided
		if ranges[1] == "" {
			end = start + defaultChunkSize - 1
			if end >= videoStream.FileSize {
				end = videoStream.FileSize - 1
			}
		} else {
			end, err = strconv.ParseInt(ranges[1], 10, 64)
			if err != nil || end >= videoStream.FileSize {
				end = videoStream.FileSize - 1
			}
		}

		// Ensure end is not less than start
		if end < start {
			r.logger.Error().Msgf("End byte is less than start byte: %d, %d", start, end)
			return ctx.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid Range End")
		}

		// If range is well-defined, serve the file using SendFile
		if ranges[1] != "" {
			return ctx.SendFile(filePath)
		}

		// Setting required response headers for partial content
		ctx.Set(fiber.HeaderContentType, "video/webm")
		ctx.Set(fiber.HeaderAcceptRanges, "bytes")
		ctx.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, videoStream.FileSize))
		ctx.Set(fiber.HeaderContentLength, strconv.FormatInt(end-start+1, 10))
		ctx.Status(fiber.StatusPartialContent)

		//fmt.Printf("Bytes: %d-%d/%d\n", start, end, fileSize)

		// Seek to the start position
		_, seekErr := videoStream.File.Seek(start, io.SeekStart)
		if seekErr != nil {
			r.logger.Error().Err(seekErr).Msg("mediastream: Error seeking to start position")
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		//remainingBytes := end - start + 1
		//_, copyErr := io.CopyN(ctx.Response().BodyWriter(), videoStream.File, remainingBytes)
		//if copyErr != nil && copyErr != io.EOF {
		//	r.logger.Error().Err(copyErr).Msg("mediastream: Error copying bytes to response")
		//	return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		//}

		// Copy the specified range of bytes to the response in smaller chunks
		buffer := make([]byte, 4096)
		remainingBytes := end - start + 1
		var totalCopied int64

		for totalCopied < remainingBytes {
			bytesRead, readErr := videoStream.File.Read(buffer)
			if readErr != nil {
				if readErr != io.EOF {
					r.logger.Error().Err(readErr).Msg("mediastream: Error reading file")
					return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				}
				if bytesRead == 0 {
					break
				}
			}

			// Ensure we don't copy more than the specified range
			if totalCopied+int64(bytesRead) > remainingBytes {
				bytesRead = int(remainingBytes - totalCopied)
			}

			// Copy the chunk to the response
			_, copyErr := ctx.Response().BodyWriter().Write(buffer[:bytesRead])
			if copyErr != nil {
				r.logger.Error().Err(copyErr).Msg("mediastream: Error copying bytes to response")
				return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			totalCopied += int64(bytesRead)
		}

	} else {
		return ctx.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid Range Header")
	}

	return nil
}

//func (r *Repository) ServeFiberDirectPlay(ctx *fiber.Ctx, clientId string) error {
//
//	if !r.IsInitialized() {
//		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "Module not initialized")
//		return errors.New("module not initialized")
//	}
//
//	// Get current media
//	mediaContainer, found := r.playbackManager.currentMediaContainer.Get()
//	if !found {
//		r.wsEventManager.SendEvent(events.MediastreamShutdownStream, "no file has been loaded")
//		return errors.New("no file has been loaded")
//	}
//
//	// Open the video file to get its size
//	file, err := os.Open(mediaContainer.Filepath)
//	if err != nil {
//		r.logger.Error().Err(err).Msgf("mediastream: Error opening video file")
//		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
//	}
//	defer file.Close()
//
//	// Get the file size
//	fileInfo, err := file.Stat()
//	if err != nil {
//		r.logger.Error().Err(err).Msgf("mediastream: Error getting file information")
//		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
//	}
//	fileSize := fileInfo.Size()
//
//	// Default chunk size
//	const defaultChunkSize int64 = 1048576 // 1 MB
//
//	// Get the Range header
//	rangeHeader := ctx.Get("Range")
//	var start, end int64
//	var newRangeHeader string
//
//	r.logger.Trace().Msgf("mediastream: Range Header: %s", rangeHeader)
//
//	if rangeHeader != "" {
//		ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
//		if len(ranges[0]) > 0 {
//			start, err = strconv.ParseInt(ranges[0], 10, 64)
//			if err != nil || start >= fileSize {
//				r.logger.Warn().Err(err).Int64("start", start).Int64("fileSize", fileSize).Msgf("mediastream: Error parsing start byte position or start out of range")
//				start = 0 // Default to the start of the file if invalid
//			}
//		} else {
//			start = 0
//		}
//
//		if ranges[1] == "" {
//			end = start + defaultChunkSize - 1
//			if end >= fileSize {
//				end = fileSize - 1
//			}
//		} else {
//			end, err = strconv.ParseInt(ranges[1], 10, 64)
//			if err != nil || end >= fileSize {
//				end = fileSize - 1
//			}
//		}
//		//if len(ranges) == 2 && len(ranges[1]) > 0 {
//		//	end, err = strconv.ParseInt(ranges[1], 10, 64)
//		//	if err != nil || end >= fileSize {
//		//		end = fileSize - 1
//		//	}
//		//} else {
//		//	end = start + defaultChunkSize - 1
//		//	if end >= fileSize {
//		//		end = fileSize - 1
//		//	}
//		//}
//		//
//		//// Ensure end is not less than start
//		//if end < start {
//		//	end = start + defaultChunkSize - 1
//		//	if end >= fileSize {
//		//		end = fileSize - 1
//		//	}
//		//}
//
//		// Create a new Range header
//		newRangeHeader = fmt.Sprintf("bytes=%d-%d", start, end)
//	} else {
//		// If no Range header, set a default range
//		start = 0
//		end = defaultChunkSize - 1
//		if end >= fileSize {
//			end = fileSize - 1
//		}
//		newRangeHeader = fmt.Sprintf("bytes=%d-%d", start, end)
//	}
//
//	// Log the modified range
//	r.logger.Info().Msgf("mediastream: Modified Range: %s", newRangeHeader)
//
//	// Update the request header
//	//ctx.Request().Header.Set("Content-Length", strconv.FormatInt(end-start+1, 10))
//	ctx.Request().Header.Add("Range", newRangeHeader)
//
//	ctx.Set("Content-Type", "video/webm")
//
//	return ctx.SendFile(mediaContainer.Filepath)
//}
