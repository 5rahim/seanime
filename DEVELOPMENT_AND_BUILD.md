# Development

To get started, you will need to be familiar with Go and React.

I recommend creating a dummy AniList account for testing purposes.

## Setup

1. Install Go 1.22, Node.js 18+ and npm.
2. Use a modern IDE

## Server

1. Run this command to start the server:
	```bash
	go run cmd/main.go --datadir="path/to/datadir"
	```
	I recommend passing the `--datadir` flag to specify a test data directory. This will prevent the server from writing to your actual data directory.

2. The server will start on `http://127.0.0.1:43211`.

## Web

1. Run the web interface:

	```bash
	# get to web directory
	cd seanime-web
	# install dependencies
	npm install
	# run the next.js dev server
	npm run dev
	```
	
	Go to `http://127.0.0.1:43210` to access the web interface.

Since in development mode, the web interface cannot be served by the server, it will run on a different port. This only impacts MyAnimeList OAuth.

# Build

## Server

1. Build the server using the following command:

	```bash
	go build -o seanime cmd/main.go
	```
	
	This will create a `seanime` binary in the root of the project.
	
	You can keep it there or move it to a more appropriate location.

## Web

1. Build the web interface using the following command:

	```bash
	npm run build
	```

2. After the build process is complete, a new `out` directory will be created under the `seanime-web` folder.
	- This is essentially what the server serves when you access the web interface in a normal setup.
3. You now need to copy the contents of the `out` directory to a `web` directory in the root of the project or wherever
   the built server binary is located.

# Overview

## Workflow

Although the codebase is very large, it is modular and relatively easy to navigate and reason about.

`internal/core` creates an instance of `App` , the main struct containing all “modules” a.k.a pointers to instantiated structs (MediaPlayerRepository, MangaDownloader, …). This `App` instance is passed to all route handlers (API endpoints) so that any route can execute the business logic by using whatever struct it needs.

```go
package core

type App struct {
	Config                  *Config
	Database                *db.Database
	Logger                  *zerolog.Logger
	TorrentClientRepository *torrent_client.Repository
	PlaybackManager         *playback_manager.Manager
	// ...
}
```

Speaking of route handlers, those are defined in `internal/handlers` . Any given under that directory will have a number of handlers to execute business logic.

For example, `internal/handlers/playback_manager.go` has a handler named `HandlePlaybackPlayNextEpisode` , defined like this:

```go
package handlers

// HandlePlaybackPlayNextEpisode
//
//	@summary plays the next episode of the currently playing media.
//	@desc This will play the next episode of the currently playing media.
//	@desc This is non-blocking so the client should prevent multiple calls until the next status is received.
//	@route /api/v1/playback-manager/play-next [POST]
//	@returns bool
func HandlePlaybackPlayNextEpisode(c *RouteCtx) error {

	err := c.App.PlaybackManager.PlayNextEpisode()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
```

The comments are a form of documentation similar to OpenAPI, they’ll allow the internal code generator to generate endpoint objects in Typescript, that will be used by the client to make requests, for example:

```tsx
// seanime-web/api/generated/endpoints.ts
// auto generated
export const API_ENDPOINTS = {
	PLAYBACK_MANAGER: {
        /**
         *  @description
         *  Route plays the next episode of the currently playing media.
         *  This will play the next episode of the currently playing media.
         *  This is non-blocking so the client should prevent multiple calls until the next status is received.
         */
        PlaybackPlayNextEpisode: {
            key: "PLAYBACK-MANAGER-playback-play-next-episode",
            methods: ["POST"],
            endpoint: "/api/v1/playback-manager/play-next",
        },
   },
   // ...
}
```

Note that types are also auto-generated in `seanime-web/api/generated/types.ts` based on the `@returns` directive and request body variables of each handler. The code generator will analyze the entire Go codebase and convert returned public structs into Typescript types.

All the routes are declared in `internal/handlers/routes.go` where a handler a passed to each route.

## AniList GraphQL API

Anilist queries are defined in `internal/anilist/queries/*.graphql` and generated using `gqlgenc`.

Run this when you make changes to the GraphQL schema.

```bash
go get github.com/Yamashou/gqlgenc
```
```bash
cd internal/api/anilist
```
```bash
go run github.com/Yamashou/gqlgenc
```
```bash
cd ../../..
```
```bash
go mod tidy
```

`gqlgenc` will generate the different queries, mutations and structs associated with the AniList API and the queries we defined.
These are located in `internal/api/anilist/clieng_gen.go`.


In `internal/api/anilist/client.go`, we manually reimplements the different queries and mutations into a `ClientWrapper` struct and a `MockClientWrapper` for testing.
This is done to avoid using the generated code directly in the business logic. It also allows us to mock the AniList API for testing.

## Tests

**Do not** run the tests all at once. Run them **individually** if you have to.

You should:
- Create a dummy AniList account and grab the access token (from the browser).
- Edit the `test/config.toml` file with the access token and the username. Use `config.example.toml` as a template.

Tests are run using the `test_utils` package. It provides a `InitTestProvider` method that initializes the test config global variable.
This global variable contains the values from the `test/config.toml` file.

```go
func (t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())
}
```

As you can see, it also takes functions as arguments. Those functions merely skip the test if the corresponding flag is not enabled in the `test/config.toml` file.

### Testing with AniList API

The `anilist` package exports a `MockClientWrapper` that you can use to test different packages that depend on the AniList API. (you can access it by calling `anilist.TestGetMockAnilistClientWrapper()`)

When testing a package that requires the user's **AnimeCollection**, the mock client will return a dummy collection stored in `test/daa/BoilerplateAnimeCollection.json` when 
`anilistClientWrapper.AnimeCollection` is called with a nil username.

```go
package test

import (
	"testing"
	"internal/test_utils"
	"internal/api/anilist"
)

func Test(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())
	
	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	// Called with a nil username
    // `anilistCollection` will contain the dummy collection
    anilistCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)
    if err != nil {
    	t.Fatal(err)
    }
}
```

Not all methods are implemented in the mock client. You can add more methods to the mock client if you need them.

When you pass a username, let's say your dummy account's, using `test_utils.ConfigData.Provider.AnilistUsername`,
the mock client will fetch it using a real request and store it in `test/testdata/AnimeCollection.json` and return it. This file will be used for subsequent calls.
(This is just to avoid making too many requests to the AniList API).

Same goes for `anilistClientWrapper.GetBaseMediaById`, `anilistClientWrapper.GetBasicMediaById` and `anilistClientWrapper.GetBasicMediaByMalId`.

### Testing third-party apps

Some tests will directly interact with third-party apps such as Transmission and qBittorrent. You should have them installed and running on your machine.
Edit the `test/config.toml` file and individual tests to match your setup (e.g. port, password, files to open, etc.)
