package anilist

import (
	"context"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBaseMediaById(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	// Get Anilist client
	acw := TestGetAnilistClientWrapper()

	tests := []struct {
		name    string
		mediaId int
	}{
		{
			name:    "Cowboy Bebop",
			mediaId: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := acw.Client.BaseMediaByID(context.Background(), &test.mediaId)
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}
