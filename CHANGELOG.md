# Changelog

All notable changes to this project will be documented in this file.

## v2.4.1

- ⚡️ Desktop: Close to minimize to tray
  - The close button no longer exits the app, but minimizes it to the system tray
  - Exit the app by right-clicking the tray icon and selecting 'Quit Seanime'
- ⚡️ Qbittorrent: Custom tag settings #140
- 🦺 Fixed Linux server requiring libc
- 🦺 Desktop: Fixed 'toggle visibility'

## v2.4.0

- 🚀 Desktop app
  - You can now download the new desktop app for Windows, macOS, and Linux
  - The desktop app is a standalone GUI that embeds its own server
- 🦺 Anime library: Fixed toggle lock button
- 🦺 Torrent streaming: Fixed file previews
- 🏗️ Rename 'enhanced scanning'
- 🔨 Updated release workflow

## v2.3.0

- ✨ Real-Debrid support for streaming and downloading
- ⚡️ Manga: Unread chapter count badge
- ⚡️ HTTPS support for qBittorrent and Transmission
- ⚡️ Online streaming: Theater mode
- 🦺 Scanner: Fixed NC false-positive edge case
- 🦺 Fixed pause/resume action for qBittorrent v5 #157
- 🏗️ Added fallback update endpoint & security check
- 🏗️ Fixed update notification reliability
- 🏗️ Fixed cron concurrency issue


## v2.2.3

- 🦺 Offline: Fixed episode images not showing up without an internet connection
  - Remove and add saved series again to fix the issue
- 🦺 Offline: Download only used images
- 🦺 Debrid streaming: Fixed MPV --title flag
- 🦺 Debrid streaming: Fixed stream cancellation
- ⚡️ Media streaming: Custom FFmpeg hardware acceleration options
- 🏗️ Moved filename parser to separate package

## v2.2.2

- ✨ Debrid Streaming: Auto select (Experimental)
- ⚡️ Scanner: Improved episode normalization logic
- ⚡️ Debrid Streaming: Retry mechanism for stream URL checks
- ⚡️ Online streaming: New "Include in library" setting
- ⚡️ Online streaming: Show fetched image & filler metadata on episode cards
- ⚡️ Settings: Torrent client "None" option
- 💄 UI: Integrated online streaming view in anime page
- 🦺 Fixed custom background images not showing up #148
- 🦺 Fixed external player link for downloaded Specials/NC files #139
- 🦺 Fixed "contains" filter for Auto Downloader #149
- 🏗️ Merged "Default to X view" and "Include in library" settings for torrent & debrid streaming
- 🏗️ Made library path optional for onboarding and removable in settings
- 🏗️ Updated empty library screen
- 🏗️ Fix Go toolchain issue #150

## v2.2.1

- ⚡️ New getting started page
- ⚡️ Auto Downloader: Added 'additional terms' filter option
- 🦺 Torrent streaming: Fixed auto-select runtime error
- 🦺 Fixed auto-scanning runtime error
- 🦺 Fixed issue with inexistant log directory
- 🦺 Torrent streaming: Fixed runtime error caused by missing settings
- 🦺 Fixed scan summaries of unresolved files not showing up

## v2.2.0

- 🎉 New offline mode
    - New local data system with granular updates, removing the need for re-downloading metadata each time. Option for automatic local data refreshing. Support for media streaming. Better user interface for offline mode.
- 🎉 Debrid support starting with TorBox integration
    - TorBox is now supported for downloading/auto-downloading and streaming torrents.
    - Automatic local downloading once a torrent is ready
- 🎉 Watch continuity / Resumable playback
    - Resume where you left off across all playback types (downloaded, online streaming, torrent/debrid streaming)
- ✨ Support for multiple library directories
- ✨ Export & import anime library data
- ⚡️ Improved scanner and matcher
    - Matcher now prioritizes distance comparisons to avoid erroneous matches
- ⚡️ Extensions: User configs
- ⚡️ Improved Auto Downloader title comparisons #134
    - New ‘Verify season’ optional setting to improve accuracy if needed
- ⚡️ Online streaming: Manual match
- ⚡️ Torrent streaming: Change default torrent client host #132
- ⚡️ JS Extensions: Torrent data to magnet link global helper function #138
- ⚡️ Media streaming: Direct play only option
- ⚡️ Built-in player: Discrete controls (Hide controls when seeking)
- ⚡️ Built-in player: Auto skip intro, outro
- ⚡️ Support for more video extensions #144
- 🦺 Fixed Semver version comparison implementation (affects migrations)
- 🦺 Fixed Auto Downloader form #133
- 🦺 Fixed ‘continue watching’ button for non-downloaded media #135
- 🦺 Fixed Hianime extension
- 🦺 Fixed specials not working with external player link for torrent streaming #139
- 🦺 Fixed some specials not being streamable
- 🏗️ Refactored metadata provider code
- 🏗️ New documentation website

## v2.1.1

- ✨ Discover: New 'Schedule' and 'Missed sequels' section
- ⚡️ Self update: Replace current process on Linux #114
- ⚡️ Auto play next episode now works for torrent streaming (with auto-select enabled)
- ⚡️ Anime media cards persist list data across pages
- 🦺 Fixed duplicated playback events when 'hide top navbar' is enabled #117
- 🦺 Fixed UI inconsistencies & layout shifts due to scrollbar
- 🦺 Fixed anime media card trailers
- 🦺 Fixed nested popovers not opening on Firefox
- 🏗️ UI: Added desktop-specific components for future desktop app
- 🏗️ Added separate build processes for frontend

## v2.1.0

- ✨ Manage logs from the web interface
- ✨ Extensions: Improved Javascript interpreter
  - New Cheerio-like HTML parsing API
  - New CryptoJS API bindings
- ✨ Extensions: Typescript/Javascript Playground
  - Test your extension code interactively from the web interface
- ✨ AnimeTosho: 'Best release' filter
- ✨ Manga: New built-in "ComicK (Multi)" extension
  - Supports language & scanlator filters
- ✨ Auto play next episode for Desktop media players
  - Enable this in the media player settings
- ✨ Manga extension API now support language & scanlator filters
- ⚡️ Playlist creation filters
- ⚡️ Unmatch select files instead of all
- ⚡️ New option to download files to device #110
- ⚡️ Progress modal key bindings #111
  - 'u' to update progress & 'space' to play next episode
- 🦺 Extensions Fixed JS runtime 'fetch' not working with non-JSON responses
- 🦺 qBittorrent login fix
- 🏗️ Updated extension SDK
  - Breaking changes for third-party extensions

## v2.0.3

- ✨ Settings: Choose default manga source
- 🦺 Fixed 'resolve unmatched' feature
  - Fixed incorrect hydration when manually resolving unmatched files
- 🦺 Torrent streaming: Fixed external player link on Android
- 🦺 UI: Display characters for undownloaded anime
- 🏗️ Updated extension SDK

## v2.0.2

- ✨ Ignore files
- ⚡️ Improved 'resolve unmatched' feature
  - Select individual files to match / ignore
  - Suggestions are fetched faster
- 🦺 Torrent streaming: Fixed MPV cache
- 🦺 Fixed manual match overwriting locked files
- 🦺 Fixed episode summaries

## v2.0.1

- ✨ Torrent streaming: Show previously selected torrent
- ✨ Support for AniList 'repeating' status
- 🦺 Fixed External Player Link not working on Android
- 🦺 Fixed UI inconsistencies
- 🦺 Fixed SeaDex provider

## v2.0.0

- 🎉 Extension System
  - Create or install torrent provider, online streaming, and manga source extensions
  - Support for JavaScript, TypeScript, and Go
  - Easily share extensions by hosting them on GitHub or any public URL
  - Extensions are sandboxed for security and have access only to essential APIs
- 🎉 Windows System Tray App
  - Seanime now runs as a system tray app on Windows, offering quick and easy access
- 🎉 External Media Player Link (Custom scheme)
  - Open media in external player apps like VLC, MX Player, Outplayer, and more, using custom URL schemes
  - Stream both downloaded media and torrents directly to your preferred player that supports custom schemes
- ✨ Torrent Streaming Enhancements
  - Stream torrents to other devices using the external player link settings
  - Manually select files for torrent streaming (#103)
  - View torrent streaming episodes alongside downloaded ones in your library
  - Improved handling of Specials & Adult content (#103)
  - Torrent streaming now passes filenames to media players (#99)
  - Option to switch to torrent streaming view if media isn't in your library
- ⚡️ Enhanced Auto Downloader
  - Improved accuracy with a new option to filter by release group using multiple queries
- ✨ UI Enhancements
  - Customize your experience with new user interface settings
  - Updated design for media cards, episode cards, headers, and more
- ✨ Manga Enhancements
  - Manually match manga series for more accurate results
  - Updated page layout
- ✨ Notifications
  - Stay informed with new in-app notifications
- ⚡️ Smart Search Improvements
  - Enhanced search results for current torrent providers
  - Reduced latency for torrent searches
- ⚡️ Media Streaming Enhancements
  - Defaults to the cache directory for storing video segments, removing the need for a transcode directory
- ⚡️ Library Enhancements
  - Filter by title in the detailed library view (#102)
  - More options for Discord Rich Presence (#104)
- 🦺 Bug Fixes & Stability
  - Fixed incorrect listing on the schedule calendar
  - Resolved runtime error when manually syncing offline progress
  - Resolved runtime error caused by torrent streaming
  - Corrected links on the AniList page's manga cards
- 🏗️ Logging & Output
  - Continuous logging of terminal output to a file in the logs directory
  - FFmpeg crashes are now logged to a file
  - Enforced absolute paths for the `-datadir` flag
- 🏗️ Codebase Improvements
  - Refactored code related to the AniList API for better consistency
  - Enhanced modularity of the codebase for easier maintenance
  - Updated release workflow and dependencies

## v1.7.3

- ⚡️ Perf: Optimized queries
  - Start-up time is reduced
  - Editing list entries has lower latency
  - Fetching larger AniList collections is now up to 5 times faster
- 💄 UI: Updated components
  - Larger media cards
  - Updated episode grid items
  - Use AniList color gradients for scores
  - Improved consistency across components
- ⚡️ Automatically add new media to AniList collection when downloading first episode
- 🦺 Transmission: Escape special characters in password 
- 🦺 UI: Escape parentheses in image filenames

## v1.7.2

- ⚡️ Scanner: Support more file extensions
- ⚡️ Removed third-party app startup check if the application path is not set
- 🦺 Auto update: Fixed update deleting unrelated files in the same directory
- 🦺 Media streaming: Fixed direct play using wrong content type #94
- 🦺 Torrent streaming: Fixed inaccurate file download percentage for batches #96

## v1.7.1

- 🦺 Media streaming: Fixed direct play returning the same file for different episodes
- 🦺 Torrent streaming: Fixed playing individual episode from batch torrents #93
- 🦺 Torrent streaming: Fixed panic caused by torrent file not being found
- 🦺 Fixed crash caused by terminating MPV programmatically / stopping torrent stream
- 🦺 Fixed 'manga feature not enabled' error when opening the web interface #90
- 🦺 Fixed manga list being named 'watching' instead of 'reading'
- 🦺 Media streaming: Fixed 'file already closed' error with direct play
- 🦺 Torrent streaming: Fixed persistent loading bar when torrent stream fails to start
- 🦺 Schedule: Fixed calendar having inaccurate dates for aired episodes
- 🦺 Media streaming: Fixed byte range request failing when video player requests end bytes first (direct play)
- 🏗️ Media streaming: Refactored direct play file cache system
- 🏗️ Scan summaries: Use preferred titles
- 🏗️ Internal refactoring for code consistency

## v1.7.0

- ✨ Improved anime library page
  - New detailed view with stats, filters and sorting options
- ✨ Revamped manga page
  - Updated layout with dynamic header and genre filters
  - Page now only shows current, paused and planned entries
- ✨ Improved 'Schedule' page: New calendar view for upcoming episodes
- ✨ Improved 'Discover' page: Support for manga
- ✨ Improved 'AniList' page
  - Updated layout with new filters, sorting options and support for manga lists
  - New stats section for anime and manga
- ✨ Global search now supports manga
- ✨ Online streaming: Added support for dubs
- ✨ Media streaming: Auto play and auto next #77
- ⚡️ 'None' option for torrent provider #85
	- This option disables torrent-related UI elements and features
- ⚡️ Torrent streaming: Added filler metadata
- ⚡️ Ability to fetch metadata for shows that are not in the library
- ⚡️ MPV: Added retry mechanism for connection errors
- ⚡️ Perf: Improved speed when saving settings
- ⚡️ Perf: Virtualize media lists for better performance if there are many entries
- ⚡️ Transcoding: Option to toggle JASSUB offscreen rendering
- ⚡️ Online streaming: Refactored media player controls
- ⚡️ UI: Improved layout for media streaming & online streaming
- ⚡️ UI: Added indicator for missing episodes on media cards
- 🦺 Media streaming: Fixed direct play #82
- 🦺 Media streaming: Fixed font files not loading properly
- 🦺 Transcoding: Set default hardware accel device to auto on Windows
- 🦺 Torrent streaming: Fixed manual selection not working with batches #86
- 🦺 Online streaming: Fixed episode being changed when switching providers
- 🦺 Playlists: Fixed list not updating when a playlist is started
- 🦺 UI: Make global search bar clickable on mobile
- 🦺 Online streaming: Fixed Zoro provider
- 🦺 Fixed terminal errors from manga requests
- ⬆️ Updated dependencies

## v1.6.0

- 🚀 The web interface is now bundled with the binary
  - Seanime now ships without the `web` directory
  - This should solve issues with auto updates on Windows
- 🎉 Media streaming: Direct play support
  - Seanime will now, automatically attempt to play media files directly without transcoding if the client supports the codecs
- ✨ Metadata: View filler episodes #74
  - Fetch additional metadata to highlight filler episodes
- ✨ Setting: Refresh library on startup
- ⚡️ Scanner: Support for symbolic links
- 🚀 Transcoding: JASSUB files are now embedded in the binary
  - No need to download JASSUB files separately unless you need to support old browsers
- 🦺 Media streaming: Fixed subtitle rendering issues
  - This should solve issues with subtitles not showing up in the media player
- 🦺 Scanner: Fixed runtime error when files aren't matched by the autoscanner
- 🦺 Media streaming: Fixed JASSUB on iOS
- 🦺 Fixed crash caused by concurrent logs
- 🏗️ BREAKING: Media streaming: Metadata extraction done using FFprobe only
- 🔨 Updated release workflow
- ⬆️ Updated dependencies

## v1.5.5

- ⚡️ Manga reader fullscreen mode (hide the bar)
  - You can now toggle the manga reader bar by clicking the middle of the page or pressing `b` on desktop
  - Click the cog icon to toggle the option on mobile
- ⚡️ Changed manga reader defaults based on screen size
  - Clicking `Reset defaults for (mode)` will now take into account the screen size
- 🦺 Fixed list not updating after editing entry on 'My lists' page
- 🦺 Fixed manga list not updating after deleting entry
- 🦺 Fixed score and recommendations not updating when navigating between series

## v1.5.4

- ⚡️ Added episode info to Torrent Streaming view #69
- ⚡️ Custom anime lists are now shown in 'My Lists' page #70
- 🦺 Fixed active playlist info not showing up on the web UI
- 🦺 Torrent streaming: Fixed manual selection not working when episode is already watched
- 🦺 Torrent Streaming: Fixed transition

## v1.5.3

- ✨ Self update (Experimental)
  - Update Seanime to the latest version directly from the web UI
- 🦺 Media streaming: Fixed issue with media player not using JASSUB #65
- 🦺 Online streaming: Fixed progress syncing #66
- 🦺 Fixed .tar.gz decompression error when downloading new releases on macOS
- 🦺 Fixed some layout issues
- 🏗️ Changed default subtitle renderer styles on mobile #65
- 🏗️ Use binary path as working directory variable by default
  - Fixes macOS startup process and other issues
- 🏗️ Added `server.usebinarypath` field to config.toml
  - Enforces the use of binary path as working directory variable
  - Defaults to `true`. Set to `false` to use the system's working directory
- 🏗️ Removed `-truewd` flag
- 🏗️ Disabled Fiber file compression

## v1.5.2

- 🦺 Fixed transcoding not starting (regression in v1.5.1)
- 🦺 Fixed Discover page header opacity issues
- 🦺 Fixed runtime error caused by missing settings
- 🏗️ Reduced latency when reading local files

## v1.5.1

- ⚡️ Reduced memory usage
- ⚡️ Automatic Transcoding cache cleanup on server startup
- 🚀 Added Docker image for Linux arm64 #63
- 🚑️ Fixed occasional runtime error caused by internal module
- 💄 UI: Improved stream page layouts
- 🦺 Fixed Transcode playback error when switching episodes
- 🦺 Fixed MPV regression caused by custom path being ignored
- 🦺 Fixed hanging request when re-enabling Torrent streaming after initialization failure
- 🦺 Fixed error log coming from Torrent streaming internal package
- 🦺 Fixed 'change default AniList client ID' not working
- 🏗️ Moved 'change default AniList client ID' to config.toml
- 🔨 Updated release workflow

## v1.5.0

This release introduces two major features: Transcoding and Torrent streaming.
Thank you to everyone who has supported the project so far.

-  🎉 New: Media streaming / Transcoding (Experimental)
    - Watch your downloaded episodes on any device with a web browser using dynamic transcoding
    - Support for hardware acceleration (QSV, NVENC, VAAPI)
    - Dynamic quality selection based on bandwidth (HLS)
- 🎉 New: Torrent streaming (Experimental)
    - Stream torrents directly from the server to your media player
    - Automatic selection with no input required, click and play
    - Auto-selection of single episodes from batch torrents
    - Support for seeding in the background after streaming
- ✨ Added ability to view studios' other works
  - Click on the studio name to view some of their other works
- ✨ Added settings option to open web UI & torrent client on startup
- ⚡️ Updated terminal logs
- ⚡️ Improved support for AniList score options #51
  - You can now use decimal scores
- ⚡️ Added ability to change default AniList client ID for authentication
- 💄 UI: Moved UI customization page to the settings page
- 💄 UI: Improved data table component on mobile devices
- 🦺 Fixed failed websocket connection due to protocol mismatch #50
- 🏗️ Playback blocked on secondary devices unless media streaming is enabled
- 🏗️ Online streaming is stable
- 🏗️ Refactored MPV integration

## v1.4.3

- ⚡️ Manga: Improved pagination
  - Pagination between chapters downloaded from different sources is now possible
- ⚡️ Manga: Source selection is now unique to each series
- ⚡️ Manga: Added page container width setting for reader
- ⚡️ UI: Improved handling of custom colors
  - Added additional preset color options 
  - Fixes #43
- ⚡️ Missing episodes are now grouped per series to avoid clutter
- 🦺 Fixed slow animation when loading manga page
- 🦺 Fixed some UI inconsistencies
- 🏗️ Removed playback state logs

## v1.4.2

- 🎉 Customize UI colors
  - You can now easily customize the background and accent colors of the UI
- ✨ Docker image
  - Seanime is now available as a Docker image. Check DOCKER.md for more information
- ⚡️ Added '--truewd' flag to force to Seanime use the binary's directory as the working directory
  - This solves issues encountered on macOS
- ⚡️ Environment variables are now read before initializing the config file
	- This solves issues with setting up Docker containers
- 🦺 Fixed episode card size setting being ignored in anime page
- 🦺 Fixed incorrect 'releasing' badge being shown in media cards when hovering

## v1.4.1

- ✨ Play random episode button
- ⚡️ Scanner: Improved absolute episode number detection and normalization
- 🦺 MPV: Fixed multiple instances launching when using 'Play next episode'
- 🦺 Progress tracking: Fixed progress overwriting when viewing already watched episodes with 'Auto update' on
- 🦺 Manga: Fixed disappearing chapter table
- 🦺 Scanner: Fixed panic caused by failed episode normalization
- 🦺 Offline: Disable Auto Downloader when offline
- 🦺 Manga: Fixed download list not updating properly
- 🦺 Offline: Fixed crash when snapshotting entries with missing metadata
- 💄 Removed legacy anime page layout
- 💄 Fixed some design inconsistencies
- 🏗️ Scanner: Generate scan summary after manual match
- 🏗️ Core: Refactored web interface codebase
  - New code structure
  - More maintainable and less bloated code
  - Code generation for API routes and types

## v1.4.0

- 🎉 New feature: Offline mode
    - Watch anime/read manga in the ‘offline view’ with metadata and images
    - Track your progress and manage your lists offline and sync when you’re back online
- 🎉 New feature: Download Chapters (Experimental)
    - Download from multiple sources without hassle
    - Persistent download queue, interruption handling, error detection
- ✨ Manga: Added more sources
    - Mangadex, Mangapill, Manganato
- ✨ Anime: Improved NSFW support
    - Search engine now supports Nyaa Sukebei
    - Hide NSFW media from your library
- ⚡️ Manga: Improved reader
    - Reader settings are now unique to each manga series
    - Automatic reloading of failed pages
    - Progress bar and page selection
    - Support for more image formats
- ⚡️ Added manga to advanced search
- ⚡️ Unified chapter lists with new toggles
- 💄 New settings page layout
- 💄 Added fade effect to media entry banner image
- 🦺 Scanner: Force media ID when resolving unmatched files
- 🦺 Manga: Fixed page indexing for Mangasee
- 🦺 Fixed incorrect start dates
- 🦺 Progress tracking: Fixed incorrect progress number being used when Episode 0 is included
- 🦺 UI: Fixed issues related to scrollbar visibility
- 🏗️ Core: Built-in image proxy
- ⬆️ Updated Next.js & switched to Turbopack

## v1.3.0

- ✨ Discord Rich Presence
    - Anime & Manga activity + options to disable either one #30
    - Enable this in your settings under the ‘Features’ section
- ✨ Command line flags
    - Use `--datadir` to override the default data directory and use multiple Seanime instances
- ✨ Overhauled Manga Reader
    - Added ‘Double Page’ layout
    - Page layout customization
    - Pagination key bindings
    - Fixes spacing issues #31
    - Note: This introduces breaking changes in the cache system, the migration will be handled automatically.
- ⚡️MAL manga progress syncing
- ⚡️Enable/Disable or Blur NSFW search results
- 🦺 Fixed MAL anime progress syncing using wrong IDs
- 🦺 Fixed MAL token refreshing
- 🦺 Fixed error toasts on authentication
- 🏗️ Removed built-in ‘List Sync’ feature
    - Note: Use MAL-Sync instead
- 🏗️ Refactored config code
- 🏗️ Implemented automatic version migration system
    - Some breaking changes will be handled automatically

## v1.2.0

- 🎉 New feature: Manga (Experimental)
	- Read manga chapters and sync your progress
- ✨ Added "Best releases" filter for Smart Search
  - Currently powered by SeaDex with limited results
- ⚡️ Improved TVDB mappings for missing episode images
- ⚡️ Added YouTube embeds for trailers
- 🦺 Fixed TVDB metadata reloading
  - You can now reload TVDB metadata without having to empty the cache first 
- 🏗️ Improved Discover page
  - Reduced number of requests to AniList with caching
  - Faster loading times, lazy loading, more responsive actions
- 🏗️ Improved file cacher (Manga/Online streaming/TVDB metadata)
  - Faster I/O operations by leveraging partitioned buckets
  - Less overhead and memory usage

## v1.1.2

- ✨ Added support for TVDB images
    - Fix missing episode images by fetching complementary TVDB metadata for specific media
- ⚡️ Improved smart search results for AnimeTosho
- ⚡️ Unresolved file manager sends fewer requests
- 🚑️ Fixed runtime error caused by Auto Downloader
- 🚑️ Fixed bug introduced in v1.1.1 making some pages inaccessible
- 🦺 Removed ambiguous "add to collection" button
- 🦺 Fixed start and completion dates not showing when modifying AniList entries on "My Lists" pages
- 🦺 Fixed Auto Downloader skipping last episodes
- 🦺 Fixed smart search torrent previews
- 🦺 Fixed trailers
- 🏗️ Refactored episode metadata code

## v1.1.1

This release introduced a major bug, skip to v1.1.2+

- ✨ Added support for TVDB images
    - Fix missing episode images by fetching complementary TVDB metadata for specific media
- ⚡️ Improved smart search results for AnimeTosho
- ⚡️ Unresolved file manager sends fewer requests
- 🚑️ Fixed runtime error caused by Auto Downloader
- 🦺 Fixed Auto Downloader skipping last episodes
- 🦺 Fixed smart search torrent previews
- 🦺 Fixed trailers
- 🏗️ Refactored episode metadata code

## v1.1.0

- 🎉 New feature: Online streaming
    - Stream (most) anime from online sources without any additional configuration
- ✨ Added “Play next episode” button in progress modal
- ✨ Added trailers
- ⚡️Improved torrent search for AnimeTosho
- ⚡️Improved auto file section for torrent downloads
    - Seanime can now select the right episode files in multi-season batches, and will only fail when it can’t tell seasons apart
    - Feature now available for Transmission v4
- ⚡️ Custom background images are now visible on all pages
- ⚡️ Added ability to un-match in unknown media resolver
- 🦺 Fixed authentication #26
- 🦺 Fixed torrent name parsing edge case #24
- 🦺 Fixed ‘resume torrent’ button for qBittorrent client #23
- 🦺 Fixed files with episode number ‘0’ not appearing in Playlist creation
- 🦺 Fixed panic caused by torrent search for anime with no AniDB metadata
- 🦺 Fixed incorrect in-app settings documentation for assets #21
- 🦺 Fixed anime title text clipping #22
- 🦺 Fixed frontend Playlist UI issues
- 🦺 Added in-app note for auto scan
- 🏗️ Playlists are now stable
- 🏗️ Refactored old/unstable code
- 🏗️ Refactored all tests

## v1.0.0

- 🎉 Updated UI
  - Smoother navigation
  - Completely refactored components
  - Some layout changes
- 🎉 New feature: Transmission v4 support (Experimental)
- 🎉 New feature: UI Customization
  - Customize the main pages to your liking in the new UI settings page
  - Note: More customization options will be added in future releases
- 🎉 New feature: Playlists (Experimental)
  - Create a queue of episodes and play them in order, (almost) seamlessly
- 🎉 New feature: Auto scan
  - Automatically scan your library for new files when they are added or removed
  - You don't need to manually refresh entries anymore
- ⚡️ Refactored progress tracking
  - Progress tracking is now completely server-side, making it more reliable
- ⚡️ Improved MPV support
  - MPV will now play a new file without opening a new instance
- ⚡️ Added ability to remove active torrents
- 🏗️ Updated config file options
  - The logs directory has been moved to the config directory and is now configurable
  - The web directory path is now configurable (though not recommended to change it)
  - Usage of environment variables for paths is now supported
- 🏗️ Updated terminal logs
- 🏗️ Refactored torrent handlers
- 🦺 "Download missing only" now works with AnimeTosho
- 🦺 Fixed client-side crash caused by empty scan summary
- 🦺 Various bug fixes and improvements
- ⬆️ Updated dependencies

## 0.4.0

- 🎉 Added support for **AnimeTosho**
  - Smart search now returns more results with AnimeTosho as a provider
  - You can change the torrent provider for search and auto-download in the in-app settings
  - Not blocked as often by ISPs #16
- ✨ Added ability to silence missing episode notifications for specific media
- ⚡️ Improved scanning accuracy
  - Fixed various issues related to title parsing, matching and metadata hydration 
- ⚡️ Improved runtime error recovery during scanning
  - Scanner will now try to skip problematic files instead of stopping the entire process
  - Stack traces are now logged in scan summaries when runtime errors occur at a file level, making debugging easier
- ⚡️ Auto Downloader will now add queued episode magnets from the server
- 💄 Minor redesign of the empty library page
- 🦺 Fixed issue with static file serving #18
- 🦺 Fixed panic caused by episode normalization #17
- ⬆️ Updated dependencies
- ⬆️ Migrated to Go 1.22
- 🔨 Updated release workflow

## 0.3.0

- 🏗️ **BREAKING:** Unified server and web interface
  - The web interface is now served from the server process instead of a separate one
  - The configuration file is now named `config.toml`
  - This update will reset your config variables (not settings)
- 🏗️ Handle runtime errors gracefully
  - Seanime will now try to recover from runtime errors and display the stack trace
- ⚡️ Support for different server host and port
  - Changing the server host and port will not break the web interface anymore
- ✨ Added update notifications
  - Seanime will now check for updates on startup and notify you if a new version is available (can be disabled in settings)
  - You can also download the update from the Web UI
- ⚡️ Added ability to download ".torrent" files #11
- ⚡️ Improved MPV support
  - Refactored the implementation to be less error-prone
  - You can now specify the MPV binary file path in the settings
- 🦺 Fixed bug causing scanner to keep deleted files in the database
- 🦺 Fixed UI issues related to Auto Downloader notification badge and scanner dialog
- 🦺 Fixed duplicated UI items caused by AniList custom lists
- 🏗️ Refactored web interface code structure
- ⬆️ Updated dependencies

## 0.2.1

- ✨ Added MPV support (Experimental) #5
- 🦺 Fixed issue with local storage key value limit
- 🦺 Fixed crash caused by incorrect title parsing #7
- 🦺 Fixed hanging requests caused by settings update #8

## 0.2.0

- 🎉 New feature: Track progress on MyAnimeList
  - You can now link your MyAnimeList account to Seanime and automatically update your progress
- 🎉 New feature: Sync anime lists between AniList and MyAnimeList (Experimental)
  - New interface to sync your anime lists when you link your MyAnimeList account
- 🎉 New feature: Automatically download new episodes
  - Add rules (filters) that specify which episodes to download based on parameters such as release group, resolution, episode numbers
  - Seanime will automatically parse the Nyaa RSS feed and download new episodes based on your rules
- ✨ Added scan summaries
  - You can now read detailed summaries of your latest scan results, allowing you to see how files were matched
- ✨ Added ability to automatically update progress without confirmation when you finish an episode
- ⚡️ Improved handling of AniList rate limits
  - Seanime will now pause and resume requests when rate limits are reached without throwing errors. This fixes the largest issue pertaining to scanning.
- ⚡️ AniList media with incorrect mapping to AniDB will be accessible in a limited view (without metadata) instead of being hidden
- ⚡️ Enhanced scanning mode is now stable and more accurate
- 💄 UI improvements
- 🦺 Fixed various UX issues
- ⬆️ Updated dependencies

## 0.1.6

- 🦺 Fixed crash caused by custom lists on Anilist

## 0.1.5

- 🚑️ Fixed scanning error caused by non-existent database entries
- ⬆️ Updated dependencies

## 0.1.4

- ⚡️ Added ability to resolve hidden media
  - Before this update, media absent from your Anilist collection would not appear in your library even if they were successfully scanned.
- 🦺 Fixed crash caused by manually matching media
- 🦺 Fixed client-side crash caused by an empty Anilist collection
- 🦺 Fixed rate limit issue when adding media to Anilist collection during scanning
- 🦺 Fixed some UX issues
- ⬆️ Updated dependencies

## 0.1.3

- ✨ Added scanner logs
  - Logs will appear in the `logs` folder in the directory as the executable
- ⚡️ New filename parser
- ⚡️ Improved standard scanning mode accuracy
  - The scanner now takes into account media sequel/prequel relationships when comparing filenames to Anilist entries
- 🦺 Fixed unmatched file manager
- 🏗️ Refactored code and tests
- ⬆️ Updated dependencies
- 🔨 Updated release workflow

## 0.1.2

- 🚑️ Fixed incorrect redirection to non-existent page

## 0.1.1

- ✨ Added ability to hide audience score
- ✨ Added ability to delete Anilist list entries
- ✨ Added ability to delete files and remove empty folders
- 🦺 Fixed issue where the app would crash when opening the torrent list page
- 🦺 Fixed minor issues

## 0.1.0

- 🎉 Alpha release

