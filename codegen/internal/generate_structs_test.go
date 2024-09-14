package codegen

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestGetGoStructsFromFile(t *testing.T) {

	testPath := filepath.Join(".", "examples", "structs1.go")

	info, err := os.Stat(testPath)
	require.NoError(t, err)

	goStructs, err := getGoStructsFromFile(testPath, info)
	require.NoError(t, err)

	spew.Dump(goStructs)

}
