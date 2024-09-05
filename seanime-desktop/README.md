<p align="center">
<img src="../seanime-web/public/logo_2.png" alt="preview" width="150px"/>
</p>

<h2 align="center"><b>Seanime Desktop (WIP)</b></h2>

<p align="center">
Desktop app for Seanime. Embeds server and web interface.
</p>

## Roadmap

- [x] Custom title bar
- [x] Fix Windows fullscreen
- [x] Fix macOS overlay title bar drag
- [x] Check Getting started
- [x] Dedicated log-in
- [ ] Dedicated update process
  - [x] Dedicated update modal
  - [ ] Test update locally (doesn't work with localhost in release build)

## Development

1. Run the web interface

	```shell
	# /seanime-web
	npm run dev:desktop
	```
 
2. Sidecar

	- Build the server
    - Place the binary in `./seanime-desktop/src-tauri/binaries`
    - Platform-specific: Rename the binary to `seanime-{TARGET_TRIPLE}` [Ref](https://v2.tauri.app/develop/sidecar/)
      - e.g. `seanime-x86_64-pc-windows-msvc.exe` for Windows

3. Run Tauri

    `TEST_DATADIR` is needed.

    ```shell
    # /seanime-desktop
    TEST_DATADIR="/path/to/data/dir" npm run start
	# or
	TEST_DATADIR="/path/to/data/dir" tauri dev
   ```

## Build in Development

1. Build the web interface

	```shell
	# /seanime-web
	npm run build:developement:desktop
	# outputs in ./web-desktop
	```

2. Build Tauri

   - `TAURI_SIGNING_PRIVATE_KEY`
   - `TAURI_SIGNING_PRIVATE_KEY_PASSWORD`

	```shell
	# /seanime-desktop
	TAURI_SIGNING_PRIVATE_KEY="" TAURI_SIGNING_PRIVATE_KEY_PASSWORD="" npm run tauri build
	```


## Build

1. Build the desktop version of the web interface

	Uses `.env.desktop` and outputs to `/seanime-web/out-desktop`

	```shell
	# /seanime-web
	npm run build:desktop
	```
 
2. Move the output

	```shell
	mv ./seanime-web/out-desktop ./web-desktop
	```
 
3. Build Tauri

	```shell
	# /seanime-desktop
	npm run tauri build
	```
