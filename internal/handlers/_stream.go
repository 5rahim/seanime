package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func HandleServeVideoFiles(c *RouteCtx) error {

	srcFolder := "E:/ANIME_TEST/blue_lock"

	file := c.Fiber.Params("file")
	if file == "" {
		return c.Fiber.Status(fiber.StatusBadRequest).SendString("Path not provided")
	}

	userAgent := c.Fiber.Get("User-Agent")
	if userAgent == "" {
		return c.Fiber.Status(fiber.StatusBadRequest).SendString("User-Agent not provided")
	}

	isApple := false
	if strings.Contains(userAgent, "Apple") {
		isApple = true
	}

	mime := ""
	if strings.HasSuffix(file, ".m3u8") {
		if isApple {
			mime = "application/vnd.apple.mpegurl"
		} else {
			mime = "application/x-mpegurl"
		}
	} else if strings.HasSuffix(file, ".ts") {
		mime = "video/mp2t"
	} else if strings.HasSuffix(file, ".mp4") {
		mime = "video/mp4"
	} else if strings.HasSuffix(file, ".webm") {
		mime = "video/webm"
	} else if strings.HasSuffix(file, ".mkv") {
		mime = "video/x-matroska"
	}

	path := filepath.Join(srcFolder, file)

	c.Fiber.Set("Content-Type", mime)
	c.Fiber.Set("Accept-Ranges", "bytes")
	c.Fiber.Set("Cache-Control", "no-cache")
	c.Fiber.Set("Connection", "keep-alive")

	return c.Fiber.SendFile(path)
}

func HandleDirectPlay(c *RouteCtx) error {

	// Open the video file
	file, err := os.Open("E:/ANIME/[Judas] Blue Lock (Season 1) [1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs]/[Judas] Blue Lock - S01E03v2.mkv")
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()

	// Set response headers for streaming
	c.Fiber.Set("Content-Type", "video/webm")
	c.Fiber.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Fiber.Set("Accept-Ranges", "bytes")

	// Get the range header from the request
	rangeHeader := c.Fiber.Get("Range")
	if rangeHeader != "" {
		var start, end int64

		ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		if len(ranges) != 2 {
			log.Println("Invalid Range Header:", rangeHeader)
			return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// get the start range
		start, err := strconv.ParseInt(ranges[0], 10, 64)
		if err != nil {
			log.Println("Error parsing start byte position:", err)
			return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Calculate the end range
		if ranges[1] != "" {
			end, err = strconv.ParseInt(ranges[1], 10, 64)
			if err != nil {
				log.Println("Error parsing end byte position:", err)
				return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
		} else {
			end = fileSize - 1
		}

		// Setting required response headers
		c.Fiber.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		c.Fiber.Set(fiber.HeaderContentLength, strconv.FormatInt(end-start+1, 10))
		c.Fiber.Set(fiber.HeaderContentType, "video/webm")
		c.Fiber.Set(fiber.HeaderAcceptRanges, "bytes")
		c.Fiber.Status(fiber.StatusPartialContent)

		// Seek to the start position
		_, seekErr := file.Seek(start, io.SeekStart)
		if seekErr != nil {
			log.Println("Error seeking to start position:", seekErr)
			return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Copy the specified range of bytes to the response in smaller chunks
		chunkSize := 4096 //
		buffer := make([]byte, chunkSize)
		var totalCopied int64

		for totalCopied < end-start+1 {
			bytesRead, readErr := file.Read(buffer)
			if readErr != nil {
				if readErr != io.EOF {
					log.Println("Error reading file:", readErr)
					return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				}
			}

			// Ensure we don't copy more than the specified range
			if totalCopied+int64(bytesRead) > end-start+1 {
				bytesRead = int((end - start + 1) - totalCopied)
			}

			// Copy the chunk to the response
			_, copyErr := c.Fiber.Response().BodyWriter().Write(buffer[:bytesRead])
			if copyErr != nil {
				log.Println("Error copying bytes to response:", copyErr)
				return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			totalCopied += int64(bytesRead)
		}

	} else {
		// If no Range header is present, serve the entire video
		c.Fiber.Set("Content-Length", strconv.FormatInt(fileSize, 10))
		_, copyErr := io.Copy(c.Fiber.Response().BodyWriter(), file)
		if copyErr != nil {
			log.Println("Error copying entire file to response:", copyErr)
			return c.Fiber.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
	}

	return nil
}
