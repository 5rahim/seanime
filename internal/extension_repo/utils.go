package extension_repo

import (
	"errors"
	"fmt"
	"regexp"
	"seanime/internal/extension"
	"seanime/internal/util"
)

func manifestSanityCheck(ext *extension.Extension) error {
	if ext.ID == "" || ext.Name == "" || ext.Version == "" || ext.Language == "" || ext.Type == "" || ext.Author == "" || ext.Payload == "" {
		return fmt.Errorf("extension is missing required fields, ID: %v, Name: %v, Version: %v, Language: %v, Type: %v, Author: %v, Payload: %v",
			ext.ID, ext.Name, ext.Version, ext.Language, ext.Type, ext.Author, len(ext.Payload))
	}

	// Check the ID
	if err := isValidExtensionID(ext.ID); err != nil {
		return err
	}

	// Check name length
	if len(ext.Name) > 50 {
		return fmt.Errorf("extension name is too long")
	}

	// Check author length
	if len(ext.Author) > 25 {
		return fmt.Errorf("extension author is too long")
	}

	if !util.IsValidVersion(ext.Version) {
		return fmt.Errorf("invalid version: %v", ext.Version)
	}

	// Check language
	if ext.Language != extension.LanguageGo &&
		ext.Language != extension.LanguageJavascript &&
		ext.Language != extension.LanguageTypescript {
		return fmt.Errorf("unsupported language: %v", ext.Language)
	}

	// Check type
	if ext.Type != extension.TypeMangaProvider &&
		ext.Type != extension.TypeOnlinestreamProvider &&
		ext.Type != extension.TypeAnimeTorrentProvider {
		return fmt.Errorf("unsupported extension type: %v", ext.Type)
	}

	return nil
}

// extensionSanityCheck checks if the extension has all the required fields in the manifest.
func (r *Repository) extensionSanityCheck(ext *extension.Extension) error {

	if err := manifestSanityCheck(ext); err != nil {
		return err
	}

	// Check that the ID is unique
	if err := r.isUniqueExtensionID(ext.ID); err != nil {
		return err
	}

	return nil
}

// checks if the extension ID is valid
// Note: The ID must start with a letter and contain only alphanumeric characters
// because it can either be used as a package name or appear in a filename
func isValidExtensionID(id string) error {
	if id == "" {
		return errors.New("extension ID is empty")
	}
	if len(id) > 40 {
		return errors.New("extension ID is too long")
	}
	if len(id) < 3 {
		return errors.New("extension ID is too short")
	}

	if !isValidExtensionIDString(id) {
		return errors.New("extension ID contains invalid characters")
	}

	return nil
}

func isValidExtensionIDString(id string) bool {
	// Check if the ID starts with a letter and contains only alphanumeric characters
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$`)
	ok := re.MatchString(id)

	if !ok {
		return false
	}
	return true
}

func (r *Repository) isUniqueExtensionID(id string) error {
	// Check if the ID is not a reserved built-in extension ID
	_, found := r.extensionBank.Get(id)
	if found {
		return errors.New("extension ID is already in use")
	}
	return nil
}
