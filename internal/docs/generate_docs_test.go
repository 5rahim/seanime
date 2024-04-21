package docs

import (
	"path/filepath"
	"testing"
)

func TestGenerateDocs(t *testing.T) {

	GenerateDocsFile(filepath.Join("../", "handlers"))

}
