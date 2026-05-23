package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSettingsPath(t *testing.T) {
	// updating one path should preserve the rest of the settings.
	settings := &Settings{
		Anilist: &AnilistSettings{DisableCacheLayer: false},
		Library: &LibrarySettings{EnableOnlinestream: true},
	}

	next, err := SetSettingsPath(settings, "anilist.disableCacheLayer", true)
	require.NoError(t, err)
	require.NotNil(t, next)
	require.NotNil(t, next.Anilist)
	require.NotNil(t, next.Library)

	assert.True(t, next.Anilist.DisableCacheLayer)
	assert.True(t, next.Library.EnableOnlinestream)
	assert.False(t, settings.Anilist.DisableCacheLayer)
}

func TestPatchSettings(t *testing.T) {
	// nested patches should merge instead of replacing sibling settings.
	settings := &Settings{
		Anilist: &AnilistSettings{DisableCacheLayer: false, HideAudienceScore: true},
	}

	next, err := PatchSettings(settings, map[string]interface{}{
		"anilist": map[string]interface{}{
			"disableCacheLayer": true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, next)
	require.NotNil(t, next.Anilist)

	assert.True(t, next.Anilist.DisableCacheLayer)
	assert.True(t, next.Anilist.HideAudienceScore)
	assert.False(t, settings.Anilist.DisableCacheLayer)
}
