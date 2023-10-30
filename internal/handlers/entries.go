package handlers

type libraryEntryQuery struct {
	MediaId int `query:"mediaId" json:"mediaId"`
}

func HandleGetLibraryEntry(c *RouteCtx) error {

	p := new(libraryEntryQuery)
	if err := c.Fiber.QueryParser(p); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(p)
}
