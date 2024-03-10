package test_utils

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {

	assert.NotNil(t, ConfigData)

	spew.Dump(ConfigData)
}
