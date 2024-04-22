package codegen

import (
	"path/filepath"
	"testing"
)

func TestGenerateDocs(t *testing.T) {

	GenerateHandlers(filepath.Join("../../internal", "handlers"), "./")

}
