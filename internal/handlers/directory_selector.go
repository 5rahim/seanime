package handlers

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"path/filepath"
	"strings"
)

type DirectoryInfo struct {
	FullPath   string `json:"fullPath"`
	FolderName string `json:"folderName"`
}

// HandleDirectorySelector
//
//	@summary returns directory content based on the input path.
//	@desc This used by the directory selector component to get directory validation and suggestions.
//	@desc It returns subdirectories based on the input path.
//	@desc It returns 500 error if the directory does not exist (or cannot be accessed).
//	@route /api/v1/directory-selector [POST]
//	@returns handlers.DirectoryInfo
func HandleDirectorySelector(c *RouteCtx) error {

	type body struct {
		Input string `json:"input"`
	}
	var request body

	if err := c.Fiber.BodyParser(&request); err != nil {
		return c.Fiber.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	input := request.Input
	directoryExists, err := checkDirectoryExists(input)
	if err != nil {
		return c.Fiber.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error checking directory: " + err.Error(),
		})
	}

	if directoryExists {
		suggestions, err := getAutocompletionSuggestions(input)
		if err != nil {
			return c.Fiber.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error generating suggestions: " + err.Error(),
			})
		}

		content, err := getDirectoryContent(input)
		if err != nil {
			return c.Fiber.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error retrieving directory content: " + err.Error(),
			})
		}

		return c.Fiber.JSON(fiber.Map{
			"exists":      true,
			"suggestions": suggestions,
			"content":     content,
		})
	}

	suggestions, err := getAutocompletionSuggestions(input)
	if err != nil {
		return c.Fiber.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating suggestions: " + err.Error(),
		})
	}

	return c.Fiber.JSON(fiber.Map{
		"exists":      false,
		"suggestions": suggestions,
	})
}

func checkDirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getAutocompletionSuggestions(input string) ([]DirectoryInfo, error) {
	suggestions := []DirectoryInfo{}
	baseDir := filepath.Dir(input)
	prefix := filepath.Base(input)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(prefix)) {
			suggestions = append(suggestions, DirectoryInfo{
				FullPath:   filepath.Join(baseDir, entry.Name()),
				FolderName: entry.Name(),
			})
		}
	}

	return suggestions, nil
}

func getDirectoryContent(path string) ([]DirectoryInfo, error) {
	content := []DirectoryInfo{}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			content = append(content, DirectoryInfo{
				FullPath:   filepath.Join(path, entry.Name()),
				FolderName: entry.Name(),
			})
		}
	}

	return content, nil
}
