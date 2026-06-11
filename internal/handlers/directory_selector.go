package handlers

import (
	"os"
	"path/filepath"
	"seanime/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

type DirectoryInfo struct {
	FullPath   string `json:"fullPath"`
	FolderName string `json:"folderName"`
}

type DirectorySelectorResponse struct {
	FullPath    string          `json:"fullPath"`
	Exists      bool            `json:"exists"`
	BasePath    string          `json:"basePath"`
	Suggestions []DirectoryInfo `json:"suggestions"`
	Content     []DirectoryInfo `json:"content"`
}

// HandleDirectorySelector
//
//	@summary returns directory content based on the input path.
//	@desc This used by the directory selector component to get directory validation and suggestions.
//	@desc It returns subdirectories based on the input path.
//	@desc It returns 500 error if the directory does not exist (or cannot be accessed).
//	@route /api/v1/directory-selector [POST]
//	@returns handlers.DirectorySelectorResponse
func (h *Handler) HandleDirectorySelector(c echo.Context) error {

	type body struct {
		Input string `json:"input"`
	}
	var request body

	if err := c.Bind(&request); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	input := filepath.ToSlash(filepath.Clean(request.Input))
	actualInput := util.ResolvePhysicalPath(input)
	directoryExists, err := checkDirectoryExists(actualInput)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if directoryExists {
		suggestions, err := getAutocompletionSuggestions(actualInput)
		if err != nil {
			return h.RespondWithError(c, err)
		}

		content, err := getDirectoryContent(actualInput)
		if err != nil {
			return h.RespondWithError(c, err)
		}

		virtualSuggestions := make([]DirectoryInfo, len(suggestions))
		for i, s := range suggestions {
			virtualSuggestions[i] = DirectoryInfo{
				FullPath:   util.ResolveVirtualPath(s.FullPath),
				FolderName: s.FolderName,
			}
		}
		virtualContent := make([]DirectoryInfo, len(content))
		for i, co := range content {
			virtualContent[i] = DirectoryInfo{
				FullPath:   util.ResolveVirtualPath(co.FullPath),
				FolderName: co.FolderName,
			}
		}

		return h.RespondWithData(c, DirectorySelectorResponse{
			FullPath:    util.ResolveVirtualPath(actualInput),
			BasePath:    util.ResolveVirtualPath(filepath.ToSlash(filepath.Dir(actualInput))),
			Exists:      true,
			Suggestions: virtualSuggestions,
			Content:     virtualContent,
		})
	}

	suggestions, err := getAutocompletionSuggestions(actualInput)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	virtualSuggestions := make([]DirectoryInfo, len(suggestions))
	for i, s := range suggestions {
		virtualSuggestions[i] = DirectoryInfo{
			FullPath:   util.ResolveVirtualPath(s.FullPath),
			FolderName: s.FolderName,
		}
	}

	return h.RespondWithData(c, DirectorySelectorResponse{
		FullPath:    util.ResolveVirtualPath(actualInput),
		BasePath:    util.ResolveVirtualPath(filepath.ToSlash(filepath.Dir(actualInput))),
		Exists:      false,
		Suggestions: virtualSuggestions,
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
	var suggestions []DirectoryInfo
	baseDir := filepath.Dir(input)
	prefix := filepath.Base(input)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if util.IsMobile() {
			return nil, nil
		}
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
	var content []DirectoryInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		if util.IsMobile() {
			return nil, nil
		}
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
