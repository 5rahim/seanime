# WORK IN PROGRESS

## Seanime Server

Seanime Server is a self-hosted server and AniList companion for managing your local anime library, built around.

# Features

## Local library

- Easy integration with AniList.
- Scan local library and automatically match local files with corresponding anime.
- No mandatory naming/folder structures.
- Support for absolute episode numbers.

## Download

- [ ] Support for qBittorrent.

## Development

`TODO`

### Generate AniList GraphQL

```bash
go get github.com/Yamashou/gqlgenc
cd internal/anilist
go run github.com/Yamashou/gqlgenc
```