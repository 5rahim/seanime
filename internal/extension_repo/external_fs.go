package extension_repo

import (
	"github.com/goccy/go-json"
	"os"
	"seanime/internal/extension"
)

func extractExtensionFromFile(filepath string) (ext *extension.Extension, err error) {
	// Get the content of the file
	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	err = json.Unmarshal(fileContent, &ext)
	if err != nil {
		// If the manifest data is corrupted or not a valid manifest, skip loading the extension.
		// We don't add it to the InvalidExtensions list because there's not enough information to
		return
	}

	return
}
