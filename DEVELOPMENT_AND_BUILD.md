# Seanime Development and Build Guide

- [Seanime Denshi Guide](https://github.com/5rahim/seanime/blob/main/seanime-denshi/README.md)

## Tech stack

* Server: Built with [Go](https://go.dev/)
    * API: [Echo](https://echo.labstack.com/)
    * Plugin system: [Goja](https://github.com/dop251/goja) for Javascript runtimes
    * Database: [SQLite](https://github.com/glebarez/sqlite) handled via [GORM](https://gorm.io/)
    * File scanner: [Habari](https://github.com/5rahim/habari) for filename parsing
    * Torrent streaming: [anacrolix/torrent](https://github.com/anacrolix/torrent) for Bittorrent client
    * OS Integration: [Fyne](https://github.com/fyne-io/systray) for Windows system tray management
    * MKV Parser: Fork of [matroska-go](https://github.com/luispater/matroska-go)
* Frontend: Built with [React](https://reactjs.org/), [Vite](https://vitejs.dev/), and [Tanstack Router](https://tanstack.com/router)
	* UI Library: Custom components built with [Tailwind](https://tailwindcss.com/) and [Radix UI](https://www.radix-ui.com/)
	* Data Fetching: [React Query](https://tanstack.com/query/latest)
	* State Management: [Jotai](https://jotai.org/) for global state
	* Built-in Player: Custom-made (VideoCore)
* Seanime Denshi: Built with a custom [Electron](https://www.electronjs.org/)/Chromium to support more codecs

## Prerequisites

- Go 1.23+
- Node.js 18+ and npm

## Build Process

### 1. Building the Web Interface

1. Build the web interface:
   ```bash
   npm run build
   ```

2. After the build completes, a new `out` directory will be created inside `seanime-web`.

3. Move the contents of the `out` directory to a new `web` directory at the root of the project.

### 2. Building the Server

Choose the appropriate command based on your target platform:

1. **Windows (System Tray)**:
   ```bash
   set CGO_ENABLED=1
   go build -o seanime.exe -trimpath -ldflags="-s -w -H=windowsgui -extldflags '-static'"
   ```

2. **Windows (No System Tray)** - Used by the desktop app:
   ```bash
   go build -o seanime.exe -trimpath -ldflags="-s -w" -tags=nosystray
   ```

3. **Linux/macOS**:
   ```bash
   go build -o seanime -trimpath -ldflags="-s -w"
   ```

**Important**: The web interface must be built first before building the server.

---

## Development Guide

### Getting Started

The project is built with:
- Backend: Go server with REST API endpoints
- Frontend: React + Vite + Tanstack Router

For development, you should be familiar with both Go and React.

### Setting Up the Development Environment

#### Server Development

1. **Development environment**:
   - Create a dummy directory that will be used as the data directory during development.
   - Create a dummy `web` folder at the root containing at least one file, or simply do the _Building the Web Interface_ step of the build process. (This is required for the server to start.)

2. **Run the server**:
   ```bash
   go run main.go --datadir="path/to/datadir"
   ```
   
	- This will generate all the files needed in the `path/to/datadir` directory.
   
3. **Configure the development server**:
   - Change the port in the `config.toml` located in the development data directory to `43000`. The web interface will connect to this port during development. Change the host to `0.0.0.0` to allow connections from other devices.
   - Re-run the server with the updated configuration.

   The server will be available at `http://127.0.0.1:43000`.

#### Web Interface Development

1. **Navigate to the web directory**:
   ```bash
   cd seanime-web
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Start the development server**:
   ```bash
   npm run dev
   ```

   The development web interface will be accessible at `http://127.0.0.1:43210`.

**Note**: During development, the web interface is served by the Next.js development server on port `43210`.
The Next.js development environment is configured such that all requests are made to the Go server running on port `43000`.

### Understanding the Codebase Architecture

#### API and Route Handlers

The backend follows a well-defined structure:

1. **Routes Declaration**: 
   - All routes are registered in `internal/handlers/routes.go`
   - Each route is associated with a specific handler method

2. **Handler Implementation**:
   - Handler methods are defined in `internal/handlers/` directory
   - Handlers are documented with comments above each declaration (similar to OpenAPI)

3. **Automated Type Generation**:
   - The comments above route handlers serve as documentation for automatic type generation
   - Types for the frontend are generated in:
     - `seanime-web/api/generated/types.ts`
     - `seanime-web/api/generated/endpoint.types.ts`
     - `seanime-web/api/generated/hooks_template.ts`

#### Updating API Types

After modifying route handlers or structs used by the frontend, you must regenerate the TypeScript types:

```bash
# Run the code generator
go generate ./codegen/main.go
```

#### AniList GraphQL API Integration

The project integrates with the AniList GraphQL API:

1. **GraphQL Queries**:
   - Queries are defined in `internal/anilist/queries/*.graphql`
   - Generated using `gqlgenc`

2. **Updating GraphQL Schema**:
   If you modify the GraphQL schema, run these commands:

```bash
go get github.com/gqlgo/gqlgenc@v0.33.1
```
```bash
cd internal/api/anilist
```
```bash
go run github.com/gqlgo/gqlgenc
```
```bash
cd ../../..
```
```bash
go mod tidy
```

3. **Client Implementation**:
   - Generated queries and types are in `internal/api/anilist/client_gen.go`
   - A wrapper implementation in `internal/api/anilist/client.go` provides a cleaner interface
   - The wrapper also includes a mock client for testing

### Running Tests

**Important**: Run tests individually rather than all at once.

#### Test Configuration

1. Create a dummy AniList account for testing
2. Obtain an access token (from browser)
3. Create/edit `test/config.toml` using `config.example.toml` as a template

#### Writing Tests

Tests use the `test_utils` package which provides:
- `InitTestProvider` method to initialize the test configuration
- Flags to enable/disable specific test categories

Example:
```go
func TestSomething(t *testing.T) {
    test_utils.InitTestProvider(t, test_utils.Anilist())
    // Test code here
}
```

#### Testing with Third-Party Apps

Some tests interact with applications like Transmission and qBittorrent:
- Ensure these applications are installed and running
- Configure `test/config.toml` with appropriate connection details

## Notes and Warnings

- hls.js versions 1.6.0 and above may cause appendBuffer fatal errors
