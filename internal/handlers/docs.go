package handlers

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/docs"
	"os"
	"path/filepath"
)

// HandleGetDocs
//
//	@summary returns the API documentation
//	@route /api/v1/internal/docs [GET]
//	@returns docs.Docs
func HandleGetDocs(c *RouteCtx) error {

	// Read the file
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(filepath.Join(wd, "seanime-docs/routes.json"))
	if err != nil {
		return c.RespondWithError(err)
	}

	var data *docs.Docs
	// Unmarshal the data
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(data)
}
