package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

type (
	ApiDocsGroup struct {
		Filename string          `json:"filename"`
		Name     string          `json:"name"`
		Handlers []*RouteHandler `json:"handlers"`
	}

	RouteHandler struct {
		Name        string           `json:"name"`
		TrimmedName string           `json:"trimmedName"`
		Comments    []string         `json:"comments"`
		Filepath    string           `json:"filepath"`
		Filename    string           `json:"filename"`
		Api         *RouteHandlerApi `json:"api"`
	}

	RouteHandlerApi struct {
		Summary              string               `json:"summary"`
		Descriptions         []string             `json:"descriptions"`
		Endpoint             string               `json:"endpoint"`
		Methods              []string             `json:"methods"`
		Params               []*RouteHandlerParam `json:"params"`
		BodyFields           []*RouteHandlerParam `json:"bodyFields"`
		Returns              string               `json:"returns"`
		ReturnGoType         string               `json:"returnGoType"`
		ReturnTypescriptType string               `json:"returnTypescriptType"`
	}

	RouteHandlerParam struct {
		Name           string   `json:"name"`
		JsonName       string   `json:"jsonName"`
		GoType         string   `json:"goType"`         // e.g., []models.User
		UsedStructType string   `json:"usedStructType"` // e.g., models.User
		TypescriptType string   `json:"typescriptType"` // e.g., Array<User>
		Required       bool     `json:"required"`
		Descriptions   []string `json:"descriptions"`
	}
)

var cachedDocs []*ApiDocsGroup

// HandleGetDocs
//
//	@summary returns the API documentation
//	@route /api/v1/internal/docs [GET]
//	@returns []handlers.ApiDocsGroup
func (h *Handler) HandleGetDocs(c echo.Context) error {

	if len(cachedDocs) > 0 {
		return h.RespondWithData(c, cachedDocs)
	}

	// Read the file
	url := "https://raw.githubusercontent.com/5rahim/seanime/refs/heads/main/codegen/generated/handlers.json?t=1"
	res, err := http.Get(url)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	var data []*RouteHandler
	// Unmarshal the data
	err = json.Unmarshal(body, &data)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	// Group the data
	groups := make(map[string]*ApiDocsGroup)
	for _, handler := range data {
		group, ok := groups[handler.Filename]
		if !ok {
			group = &ApiDocsGroup{
				Filename: handler.Filename,
				Name:     strings.TrimPrefix(handler.Filename, ".go"),
			}
			groups[handler.Filename] = group
		}
		group.Handlers = append(group.Handlers, handler)
	}

	cachedDocs = make([]*ApiDocsGroup, 0, len(groups))
	for _, group := range groups {
		cachedDocs = append(cachedDocs, group)
	}

	return h.RespondWithData(c, groups)
}
