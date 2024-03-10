package anilist

import (
	"github.com/seanime-app/seanime/internal/test_utils"
)

// This file contains helper functions for testing the anilist package

func TestGetAnilistClientWrapper() *ClientWrapper {
	return NewClientWrapper(test_utils.ConfigData.Provider.AnilistJwt)
}

func TestGetAnilistClientWrapperAndInfo() (*ClientWrapper, *test_utils.Config) {
	cw := NewClientWrapper(test_utils.ConfigData.Provider.AnilistJwt)
	return cw, test_utils.ConfigData
}
