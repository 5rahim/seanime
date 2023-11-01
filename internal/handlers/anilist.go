package handlers

func HandleGetAnilistCollection(c *RouteCtx) error {

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(anilistCollection)

}
