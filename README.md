# WORK IN PROGRESS

## Seanime Server

Seanime Server is a self-hosted server and AniList companion for managing your local anime library.

# Features

Seanime focuses on ease of use, this means that (for the most part) you don't need to rename your torrents or have a
specific folder structure.
You're able to just download, scan and that's it.

Here are the main features of Seanime Server:

- Easy integration with AniList.
- Scan local library and automatically match local files with corresponding anime.
- No mandatory folder structure / No need for renaming.
- Support for absolute episode numbers.
- Download missing or new episodes with Nyaa and qBittorrent.
- Automatically update your progress on AniList when you watch an episode using VLC or MPC-HC.

### What it is not

Seanime Server is not a replacement for Plex/Jellyfin. It is not meant to be a media server.
It does not download metadata, transcode or stream your files to your devices.

It is only meant to be used as a companion for AniList, to manage your local anime library, track your progress while
watching with third-party media players and download new episodes.

## Development

`TODO`

### Generate AniList GraphQL

```bash
go get github.com/Yamashou/gqlgenc
cd internal/anilist
go run github.com/Yamashou/gqlgenc
cd ../..
go mod tidy
```