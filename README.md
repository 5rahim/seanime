<p align="center">
<a href="https://seanime.rahim.app/">
<img src="docs/images/logo_2.png" alt="preview" width="200px"/>
</a>
</p>

<h2 align="center"><b>Seanime</b></h2>

<h4 align="center">Self-hosted, user-friendly media server for anime and manga enthusiasts.</h4>

<p align="center">
<img src="docs/images/v1/rec_play.gif" alt="preview" width="49%"/>
<img src="docs/images/v1/seanime-manga1.gif" alt="preview" width="49%"/>
</p>

Feel free to open issues or contribute. Leave a star if you like this project!

# Features

### Server + Web Interface

- 🪶 **Lightweight server**
  - Low resource usage
  - Performant and responsive
- 🖥️ **Cross-platform**
  - Works on Windows, macOS, and Linux 
- 🔥 **User-friendly**
  - Simple and intuitive web interface
  - Easy to set up and use 
- 🌐 **Seamless integration with AniList**
  - Manage your AniList collection (anime and manga)
  - Discover new anime, browse entries, get recommendations
- ✈️ **Offline mode**
  - Consume media in the ‘offline view’ with downloaded metadata and images
  - Record your progress offline and sync when you’re online
- 🔒 **No data collection**

### Anime

- 🗂️ **Local anime library**
  - **Scan** your local anime library effortlessly
  -  **No mandatory folder structure** or naming convention
  - Browse and manage your library with metadata
  - Support for torrents with absolute episode numbers
- ⬇️ **Download new episodes & movies**
    - Smart search engine powered by public trackers (AnimeTosho, Nyaa), SeaDex
    - Integration with **qBittorrent** & **Transmission**
- 🥱 **Automate your library**
  - Download new episodes automatically by using custom filters
- 🎞️ **External Players**
    - **MPV**, **VLC**, and **MPC-HC** support
    - Sync progress to AniList (& MAL) automatically
    - Create **playlists** and queue episodes

### Manga / Manhwa 

- 📕 **Browse and manage** your AniList manga/manhwa collection
- 📖 **Read** chapters from various sources
  - Built-in, customizable **manga reader**
  - Custom reader settings for different series
- ⬇️ **Download chapters** for offline reading
  - Bulk downloads, persistent download queue

### And more

- 📺 **Online streaming**
  - Stream episodes from online sources using the built-in player
  - Support for multiple sources
- **MyAnimeList integration**
  - Automatically update your progress on MyAnimeList
- **Quality of life features**
  - Customizable UI, Scan summaries, advanced anime search, and more

### What it is not

🚨Seanime is not a replacement for Plex/Jellyfin, it requires an internet connection to fetch metadata and does not
support transcoding or streaming to other devices (yet).

# Get started

[How to use Seanime](https://seanime.rahim.app/docs)
 
### Not planned

- Support for other providers such as Trakt, SIMKL, etc.
- Torrent streaming
- Support for other languages
- Mobile app

## Resources

- [Chalk UI](https://chalk.rahim.app) - UI library
- [Fiber](https://gofiber.io/) - Go Web framework
- [GORM](https://gorm.io/) - Go ORM
- [gqlgenc](https://github.com/Yamashou/gqlgenc) - GraphQL codegen
- [Next.js](https://nextjs.org/) - React framework
- [Tailwind CSS](https://tailwindcss.com/) - CSS framework
- [React Query](https://react-query.tanstack.com/) - Data fetching library
- [Seanime Parser](https://github.com/5rahim/seanime/tree/main/seanime-parser) - Filename parser
- [ani.zip](https://api.ani.zip/mappings?anilist_id=124080) - Mappings and episode data
- [anime-lists](https://github.com/Fribb/anime-lists) - Mappings
- [TheTVDB](https://www.thetvdb.com/) - Episode images
- [AniList API](https://anilist.gitbook.io/anilist-apiv2-docs) - User data, anime and manga metadata
- [MAL API](https://myanimelist.net/apiconfig/references/api/v2)

## Acknowledgements

- [Anikki](https://github.com/Kylart/Anikki/) - Inspired GraphQL fragments
- [Lunarr](https://github.com/lunarr-app/lunarr-go/) - Inspired the use of GORM
- [Mangal](https://github.com/metafates/mangal) - Release note script

# Screenshots

<img src="docs/images/v1/rec_scan.gif" alt="preview" width="100%"/>
<img src="docs/images/v1/my-library.png" alt="preview" width="100%"/>
<img src="docs/images/v1/entry.png" alt="preview" width="100%"/>
<img src="docs/images/v1/rec_torrent-search.gif" alt="preview" width="100%"/>

<img src="docs/images/v1/manga-entry_03.png" alt="preview" width="100%"/>
<img src="docs/images/v1/manga-reader.png" alt="preview" width="100%"/>

<img src="docs/images/v1/discover_02.png" alt="preview" width="100%"/>
<img src="docs/images/v1/online-stream.png" alt="preview" width="100%"/>

# Disclaimer

- Seanime relies exclusively on the content stored by individual users on their personal devices. 
The acquisition and legitimacy of this content are external to the developer's control.
- Seanime and its developer do not host, store, or distribute any content found within the application. All anime
  information, as well as images, are sourced from publicly available APIs such as AniList and MyAnimeList.
- Seanime may, at its discretion, provide links or an interface to external websites or applications. These external websites are independently maintained by third parties, and Seanime has no control
  over the legitimacy of their content or operations. Please contact the website's owners for any concerns.
- Seanime does not collect any kind of personal data or information from its users. You are responsible for maintaining the privacy and security of the third-party authentication tokens stored within your device.
