<p align="center">
<img src="docs/images/logo.png" alt="preview" width="75px"/>
</p>

<h2 align="center"><b>Seanime</b></h2>

<h4 align="center">User-friendly, self-hosted server that brings you a Netflix-like experience for your local
anime library.</h4>

<h1 align="center">
<a href="https://seanime.rahim.app/">
<img src="docs/images/rec_main.gif" alt="preview" width="100%"/>
</a>
</h1>

Feel free to open issues or contribute. Leave a star if you like this project!

# Features

- ✨ **User-friendly web interface**
  - Set up Seanime with a few clicks
  - Netflix-like experience for your local anime library
- ✨ **Seamless integration with AniList**
  - Manage your AniList collection (add, update, delete entries)
- 🎉 **Scan your local library**
  - Seanime does not require a mandatory folder structure or naming convention
  - Seanime also supports torrents with absolute episode numbers
- 🎉 **Download new episodes automatically**
  - Add rules (filters) that specify which torrent to download based on parameters such as release group, resolution, episode numbers
  - Seanime will check RSS feeds for new episodes and download them automatically via qBittorrent
- 🎉 **Integrated torrent search engine**
  - You can manually search and download new episodes with a few clicks without leaving the web interface
  - Seanime will notify you when new episodes are available in the schedule page
- 🎉 **Automatically track your progress**
  - Launch an episode from the web interface and Seanime will automatically update your progress on AniList
  - VLC, MPC-HC, and MPV are supported
- 🎉 **MyAnimeList integration**
  - Sync your anime lists between AniList and MyAnimeList (Experimental)
  - Automatically update your progress on MyAnimeList
- **No data collection**

### What it is not

🚨Seanime is not a replacement for Plex/Jellyfin, it requires an internet connection to fetch metadata and does not
support transcoding or streaming to other devices (yet).

# Setup

[How to use Seanime.](https://seanime.rahim.app/docs)

## Resources

- AniList API
- MAL API
- [Chalk UI](https://chalk.rahim.app)
- [Fiber](https://gofiber.io/)
- [GORM](https://gorm.io/)
- [gqlgenc](https://github.com/Yamashou/gqlgenc)
- [Next.js](https://nextjs.org/)
- [Tailwind CSS](https://tailwindcss.com/)
- [React Query](https://react-query.tanstack.com/)
- [Seanime Parser](https://github.com/5rahim/seanime/tree/main/seanime-parser)

## Acknowledgements

- [Anikki](https://github.com/Kylart/Anikki/) - Inspired GraphQL fragments
- [Lunarr](https://github.com/lunarr-app/lunarr-go/) - Inspired the use of GORM
- [Mangal](https://github.com/metafates/mangal) - Release note script

# Screenshots

<h1 align="center">
<a href="https://seanime.rahim.app/">
<img src="docs/images/rec_fresh-scan.gif" alt="preview" width="100%"/>
</a>
</h1>

## Library

<img src="docs/images/my-library_02.png" alt="preview" width="100%"/>

## Anime Entry

<img src="docs/images/entry_2.png" alt="preview" width="100%"/>

## Torrent search & download

<img src="docs/images/torrent-search.png" alt="preview" width="100%"/>

## Discover

<img src="docs/images/discover.png" alt="preview" width="100%"/>

## Schedule

<img src="docs/images/schedule.png" alt="preview" width="100%"/>
