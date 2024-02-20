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

// HandleDirectorySelector is a route handler that returns directory suggestions and content based on the inputted path.
// It is used by the directory selector component.
//
//	POST /v1/directory-selector
func HandleDirectorySelector(c *RouteCtx) error {

	var request struct {
		Input string `json:"input"`
	}

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
