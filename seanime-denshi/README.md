# Seanime Denshi

Electron-based desktop client for Seanime.

## TODO

- [ ] Built-in video player for streaming and local media
  - Subtitle extraction
  - Thumbnail generation

## Development

### Prerequisites

- Node.js 18+
- Yarn or npm

### Setup

1. Place the appropriate Seanime server binaries in the `binaries` folder:
   - For Windows: `seanime-server-windows.exe`
   - For macOS/Intel: `seanime-server-darwin-amd64`
   - For macOS/ARM: `seanime-server-darwin-arm64`
   - For Linux/x86_64: `seanime-server-linux-amd64`
   - For Linux/ARM64: `seanime-server-linux-arm64`

2. Start the development server:
   ```
   npm run dev
   ```

### Building

To build the desktop client for all platforms:

```
npm run build
```

To build for specific platforms:

```
npm run build:mac
npm run build:win
npm run build:linux
```

## Structure

- `src/` - Electron application code
  - `main.js` - Main process entry point
  - `preload.js` - Preload script for renderer processes
- `binaries/` - Contains the Seanime server binaries
- `assets/` - Contains application assets like icons
