package handlers

import "github.com/seanime-app/seanime-server/internal/anify"

// Holds logic to fetch Anify episode covers

// fetchMediaImagesEntry fetches the Anify episode covers for a given media ID.
// If the entry does not exist in the database, it will be fetched from Anify and stored in the database.
func fetchMediaImagesEntry(c *RouteCtx, mId int) {

	go func() {

		_, err := c.App.Database.GetAnifyMediaEpisodeImages(mId)
		if err == nil {
			return
		}

		// Fetch from Anify
		entry, err := anify.FetchMediaEpisodeImagesEntry(mId)
		if err != nil {
			c.App.Logger.Err(err).Msg("handlers: Could not fetch Anify media images entry")
			return
		}

		// Store in the database
		err = c.App.Database.UpsertAnifyMediaEpisodeImages(entry)
		if err != nil {
			c.App.Logger.Err(err).Msg("handlers: Could not store Anify media images entry in database")
			return
		}

		c.App.AnifyEpisodeImageContainer.AddEntry(entry)

	}()

}
