# Development

To get started, you will need to be familiar with Go and React.

I recommend creating a dummy AniList account for testing purposes.

## Setup

1. Install Go 1.22, Node.js 18+ and npm.
2. Use a modern IDE

## Server

1. Choose a data directory for testing. This is where the server will store all its data.
2. Create a dummy `web` folder if you want to develop the web interface too or build the web interface first and move the contents to the `web` directory at the root of the project.
Either way, a `web` directory should be present at the root of the project.
3. Run this command to start the server:
	```bash
	go run main.go --datadir="path/to/datadir"
	```
	I recommend passing the `--datadir` flag to specify a test data directory. This will prevent the server from writing to your actual data directory.

4. Change the port in the `config.toml` file to `43000` to avoid conflicts with the release version. And change the host to `0.0.0.0` to allow connections from other devices.

5. The server will start on `http://127.0.0.1:43000`.

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

During development, the web interface is served by the Next.js development server instead of the Go server,
leading to different ports.

# Build

## 1. Web

1. Build the web interface using the following command:

	```bash
	npm run build
	```

2. After the build process is complete, a new `out` directory will be created inside `seanime-web`.

3. Move the contents of the `out` directory to a new `web` directory at the root of the project.


## 2. Server

1. Build the server using the following command:

	```bash
	# windows 
	go build -o seanime.exe main.go
 
	# darwin, linux
	go build -o seanime main.go
	```
 
Note that the web interface should be built first before building the server.

# Overview

## Codegen

All routes are declared in `internal/handlers/routes.go` where a handler a passed to each route.
Route handlers are defined in `internal/handlers`.
The comments above each route handler declaration is a form of documentation similar to OpenAPI, they’ll allow the internal code generator (`codegen/main.go`) to generate endpoint objects & types for the client.
Types are also auto-generated in `seanime-web/api/generated/types.ts` based on the `@returns` directive and request body variables of each route handler. The code generator will analyze the entire Go codebase and convert _returned_ public structs into Typescript types.


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
Functions passed to `InitTestProvider` skip the test if the corresponding flag is not enabled in the `test/config.toml` file.

```go
func (t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())
}
```


### Testing third-party apps

Some tests will directly interact with third-party apps such as Transmission and qBittorrent. You should have them installed and running on your machine.
Edit the `test/config.toml` file and individual tests to match your setup (e.g. port, password, files to open, etc.)
