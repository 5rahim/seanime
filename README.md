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
  - Modern, simple, responsive UI
- ✨ **Seamless integration with AniList**
  - Manage your AniList collection (add, update, delete entries)
  - Discover new anime, track your progress, see upcoming episodes, and get recommendations
- 🚀 **Powerful scanner**
  - No mandatory folder structure or naming convention
  - Support for torrents with absolute episode numbers
- 🚀 **Download new episodes automatically**
  - Add rules (filters) that specify which torrent to download based on parameters such as release group, resolution, episode numbers
  - Support for Nyaa and AnimeTosho
- 🚀 **Integrated torrent search engine**
  - Manually search and download new episodes with a few clicks without leaving the web interface
- 🚀 **Third-party media players**
  - Launch an episode from the web interface and Seanime will automatically update your progress on AniList (& MAL)
  - MPV, VLC, and MPC-HC are supported
- 🚀 **MyAnimeList integration**
  - Sync your anime lists between AniList and MyAnimeList (Experimental)
  - Automatically update your progress on MyAnimeList
- **No data collection**

### What it is not

🚨Seanime is not a replacement for Plex/Jellyfin, it requires an internet connection to fetch metadata and does not
support transcoding or streaming to other devices (yet).

# Setup

[How to use Seanime.](https://seanime.rahim.app/docs)

# Roadmap

Here are some features that are planned, from most to least probable:

- [ ] Transmission support
- [ ] Manga support
- [ ] Offline support
- [ ] Kitsu integration
- [ ] Docker container
- [ ] Transcoding and streaming
 
### Not planned

The following features are not planned:

- Support for other providers such as Trakt, SIMKL, etc.
- Torrent streaming
- Support for other languages
- Mobile app

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

# Disclaimer

- Seanime relies exclusively on the content stored by individual users on their personal devices. 
The acquisition and legitimacy of this content are external to our control.
  Seanime is not responsible for the actions of its users and does not condone the illegal use of copyrighted material.
- Seanime and its developer do not host, store, or distribute any content found within the application. All anime
  information, as well as images, are sourced from publicly available APIs such as AniList and MyAnimeList.
- Seanime may, at its discretion, provide links to external websites or applications related to torrents on the
  Internet. These external torrent websites are independently maintained by third parties, and Seanime has no control
  over the legitimacy of their content or operations. The inclusion of these links should not be interpreted as an endorsement of any
  third party, their torrent websites, or any information, products, or services offered through them.
- By choosing to use Seanime, you acknowledge that the developer of the app is not responsible for the content displayed
  within it. You are solely responsible for any legal or ethical implications that may arise from the use of this application.
- Seanime does not collect any kind of personal data or information from its users. You are responsible for maintaining the privacy and security of the third-party authentication tokens stored within your device.
