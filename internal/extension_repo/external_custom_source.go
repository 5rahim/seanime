package extension_repo

import (
	"fmt"
	"math/rand"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
)

const (
	CustomSourceIdentifierKey    = "1"
	CustomSourceIdentifierBucket = "customer-source-identifier"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom source
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalCustomSourceProviderExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadExternalCustomSourceProviderExtension", &err)

	switch ext.Language {
	case extension.LanguageJavascript, extension.LanguageTypescript:
		err = r.loadExternalCustomSourceExtensionJS(ext, ext.Language)
	default:
		err = fmt.Errorf("unsupported language: %v", ext.Language)
	}

	if err != nil {
		return
	}

	return
}

// generateExtensionIdentifier generates a unique extension identifier for a custom source provider extension
// it ensures that the extension identifier is unique across all custom source provider extensions
func (r *Repository) generateExtensionIdentifier(extId string) int {
	bucket := filecache.NewPermanentBucket(CustomSourceIdentifierBucket)

	identifiers := make(map[string]int)
	found, _ := r.fileCacher.GetPerm(bucket, CustomSourceIdentifierKey, &identifiers)
	if !found {
		_ = r.fileCacher.SetPerm(bucket, CustomSourceIdentifierKey, identifiers)
	}

	// Clean up old entries for extensions that no longer exist
	customSourceExtensions := r.ListCustomSourceExtensions()
	existingExtIds := make(map[string]bool)
	for _, ext := range customSourceExtensions {
		existingExtIds[ext.ID] = true
	}

	if identifier, ok := identifiers[extId]; ok {
		return identifier
	}

	// Get all existing extension identifiers to avoid conflicts
	usedIdentifiers := make(map[int]bool)
	for _, identifier := range identifiers {
		usedIdentifiers[identifier] = true
	}

	// Generate a new unique identifier (1-1023)
	var newIdentifier int
	for {
		newIdentifier = rand.Intn(1023) + 1
		if !usedIdentifiers[newIdentifier] {
			break
		}
	}

	// Store the new identifier
	identifiers[extId] = newIdentifier
	_ = r.fileCacher.SetPerm(bucket, CustomSourceIdentifierKey, identifiers)

	return newIdentifier
}

func (r *Repository) loadExternalCustomSourceExtensionJS(ext *extension.Extension, language extension.Language) error {
	provider, gojaExt, err := NewGojaCustomSource(ext, language, r.logger, r.gojaRuntimeManager, r.wsEventManager)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewCustomSourceExtension(ext, provider)
	retExt.SetExtensionIdentifier(r.generateExtensionIdentifier(ext.ID))
	gojaExt.extensionIdentifier = retExt.GetExtensionIdentifier()
	r.extensionBankRef.Get().Set(ext.ID, retExt)
	r.gojaExtensions.Set(ext.ID, gojaExt)

	r.logger.Trace().Str("id", ext.ID).Int("identifier", gojaExt.extensionIdentifier).Msg("extensions: Loaded external custom source extension")

	return nil
}
