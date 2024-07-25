package extension_repo

import "regexp"

// checks if the extension ID is valid
// Note: The ID must start with a letter and contain only alphanumeric characters
// because it can either be used as a package name or appear in a filename
func (r *Repository) isValidExtensionID(id string) bool {
	// Check if the ID starts with a letter and contains only alphanumeric characters
	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	ok := re.MatchString(id)

	if !ok {
		return false
	}

	// Check if the ID is not a reserved built-in extension ID
	_, found := r.mangaProviderExtensionBank.Get(id)
	if found {
		return false
	}

	// TODO...

	return true
}
