<p align="center">
<img src="../docs/images/logo_2.png" alt="preview" width="150px"/>
</p>

<h2 align="center"><b>Seanime Server</b></h2>

- `api`: Third-party metadata API
  - `anilist`: AniList structs and methods
  - `anizip`: Metadata API
  - `listsync`
  - `mal`: MyAnimeList API
  - `mappings`: Mapping API
  - `tvdb`: TheTVDB API
  - `metadata`: **Metadata module**
- `constants`: Self-explanatory
- `core`: App struct and instantiation of modules
- `cron`: Cron jobs
- `database`
  - `db`: **Database module**
  - `models`: Database models
- `discordrpc`: Discord RPC
  - `client`
  - `ipc`
  - `presence`: **Discord Rich Presence module**
- `events`: **Websocket Event Manager module** and constants
- `handlers`: API handlers
- `library`
  - `anime`: Library structs and methods
  - `autodownloader` **Auto downloader module**
  - `autoscanner`: **Auto scanner module**
  - `filesystem`: File system methods
  - `playbackmanager`: **Playback Manager module** for progress tracking
  - `scanner`: **Scanner module**
  - `summary`: Scan summary
- `manga`: Manga structs and **Manga Downloader module**
  - `downloader`: Chapter downloader structs and methods
  - `providers`: Online provider structs and methods
- `mediaplayers`
  - `mediaplayer`: **Media Player Repository** module
  - `mpchc`
  - `mpv`
  - `vlc` 
- `mediastream`: **Media Stream Repository** module
  - `transcoder`
  - `videofile`
- `offline`: **Offline hub module**
- `onlinestream`: **Onlinestream module**
  - `providers`: Stream providers
  - `sources`: Video server sources
- `test_utils`: Test methods
- `torrents`
  - `analyzer`: Scan and identify torrent files
  - `animetosho`
  - `nyaa`
  - `qbittorrent`
  - `seadex`
  - `torrent`: Torrent search structs and methods
  - `torrent_client`: **Torrent client repository module**
  - `transmission`
