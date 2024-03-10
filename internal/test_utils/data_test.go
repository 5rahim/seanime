package test_utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {
	assert.Empty(t, ConfigData)

	InitTestProvider(t)

	assert.NotEmpty(t, ConfigData)
}
