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
- [ ] Fix macOS overlay title bar drag
- [ ] Check Getting started
- [ ] Dedicated log-in
- [ ] Dedicated update process

## Development

1. Run the web interface

	```shell
	# /seanime-web
	npm run dev:desktop
	```

2. Run Tauri

	`TEST_DATADIR` is needed.

	```shell
	# /seanime-desktop
	TEST_DATADIR="/path/to/data/dir" npm run dev
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
