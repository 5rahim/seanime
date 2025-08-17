# Changelog

All notable changes to this project will be documented in this file.

## v2.9.6

- 🦺 Fixed server crash caused by navigating to 'Schedule' page

## v2.9.5

- ⚡️ Updated Discord RPC: Media title used as activity name, links
- ⚡️ Offline mode: Option to auto save currently watched/read media locally #376
- ⚡️ Offline mode: Bulk select media to save locally #377
- ⚡️ Metadata: Prefer TVDB title when AniDB isn't up-to-date
- ⚡️ Scan summaries: Search input for filenames
- 🦺 Potential fixes for high memory usage and app slowdowns
- 🦺 Torrent list: Fixed 'Stop seeding all' button pausing downloading torrents
- 🦺 Playground: Fixed UI crash caused by console logs
- 🦺 Scanner: Fixed matching being messed up by "Part" keyword in filenames
- 🦺 Parser: Fixed folder names with single-word titles being ignored
- 🦺 Online streaming: Don't hide button for adult entries
- 🦺 Online streaming: Fixed wrong episode selection when page is loaded #384
- 🦺 Potential fix for auto play not being canceled
- 🦺 Nakama: Fixed host's list data being added to anime that aren't in the collection
- 🦺 External Player Link: Fixed incorrect stream URL when server password is set
- 🦺 Media player: Use filepaths for comparison when loading media instead of filenames
- 🦺 Nakama: Fixed case sensitivity issue when comparing file paths on Windows
- 🦺 Fixed external player links by encoding stream URL if it contains a query parameter #387
- 🦺 Playlists: Fixed playlist deletion
- 🏗️ Slight changes to the online streaming page for more clarity
- 🏗️ Settings: Added memory profiling to 'logs' section
- 🏗️ Anime: Removed (obsolete) manual TVDB metadata fetching option
- 🏗️ Perf(Extensions): Do not download payload when checking for updates

## v2.9.4

- ⚡️ Migrated to Seanime's own anime metadata API
- ⚡️ Release calendar: Watch status is now shown in popovers
- 🦺 Fixed schedule missing some anime entries due to custom lists
- 🦺 Watch history: Fixed resumed playback not working for local files
- 🦺 Fixed streaming anime with no AniList schedule and no episode count
- 🦺 Fixed 'Upload local lists to AniList' button not working
- 🦺 Fixed repeated entries in 'Currently watching' list on the AniList page

## v2.9.3

- ⚡️ Plugins: Added Textarea component, 'onSelect' event for input/textarea
- 🦺 Fixed release calendar missing long-running series
- 🦺 Include in Library: Fixed 'repeating' entries not showing up

## v2.9.2

- ⚡️ Discover: Added 'Top of the Season', genre filters to more sections
- ⚡️ Nakama: Detailed library view now available for shared library
- ⚡️ TorBox: Optimized TorBox file list query - @MidnightKittenCat
- ⚡️ Episode pagination: Bumped number of items per page to 24
- 🦺 Nakama: Fixed dropdown menu not showing up for shared anime
- 🦺 Nakama: Base unwatched count on shared episodes
- 🦺 Scanner: Fixed modal having 'Use anilist data' checked off by default
- 🦺 UI: Revert to modal for AniList entry editor on media cards
- 🦺 Plugins: Allow programmatic tray opening on mobile
- 🦺 Fixed incorrect dates in AniList entry editor #356
- 🦺 UI: Revert incorrect video element CSS causing pixelation #355

## v2.9.1

- 🦺 Server Password: Fixed token validation on public endpoints
- 🦺 Server Password: Fixed login from non-localhost, HTTP clients #350
- ⚡️ Release calendar: Option to disable image transitions
- ⚡️ Manga: Double page offset keybindings - @Ari-03
- 🦺 Plugin: Fixed newMediaCardContextMenuItem and other APIs
- 🦺 Fixed IINA settings not being applied
- 🏗️ Downgraded Next.js and React Compiler
  - Potential solution for client-side rendering errors #349

## v2.9.0

- 🎉 New feature: Nakama - Communication between Seanime instances
  - You can now communicate with other Seanime instances over the internet
- 🎉 Nakama: Watch together (Alpha)
  - Watch (local media, torrent or debrid streams) together with friends with playback syncing
  - Peers will stream from the host with synchronized playback
- 🎉 Nakama: Share your anime library (Alpha)
  - Share your local anime library with other Seanime instances or consume your remote library
- ✨ Local account
  - By default, Seanime no longer requires an AniList account and stores everything locally
- ✨ Server password
  - Lock your exposed Seanime instance by adding a password in your config file
- ✨ Manga: Local source extension (Alpha)
  - New built-in extension for reading your local manga (CBZ, ZIP, Images)
- ✨ New schedule calendar
- ✨ macOS: Support for IINA media player
- ✨ Toggle offline mode without restarting the server
- ✨ New getting started screen
- ⚡️ Discord: Pausing anime does not remove activity anymore
- ⚡️ UI: New setting option to unpin menu items from the sidebar
- ⚡️ UI: Added pagination for long episode lists
- ⚡️ Online streaming: Episode number grid view
- ⚡️ Performance: Plugins: Deduplicate and batch events
- ⚡️ Discord: Added option to show media title in activity status (arRPC only) - @kyoruno
- ⚡️ PWA support (HTTPS only) - @HyperKiko
- ⚡️ MPV/IINA: Pass custom arguments
- ⚡️ Discord: Keep activity when anime is paused
- ⚡️ UI: Updated some animations
- 🦺 Fixed multiple Plugin API issues
- 🦺 Goja: Added OpenSSL support to CryptoJS binding
- 🦺 Fixed filecache EOF error
- 🦺 Fixed offline syncing

## v2.8.5

- 🦺 Fixed scraping for manga extensions
- 🦺 Library: Fixed bulks actions not available for unreleased anime
- 🦺 Auto Downloader: Button not showing up for finished anime
- 🦺 Online streaming: Fixed 'auto next episode' not working for some anime

## v2.8.4

- ⚡️ Plugin development improvements
    - New Discord Rich Presence event hooks
    - New bindings for watch history, torrent client, auto downloader, external player link, filler manager
    - Plugins in development mode that experience a fatal error can now be reloaded multiple times
    - Uncaught exceptions are now correctly logged in the browser devtool console
- 🦺 Fixed macOS/iOS client-side exception caused by 'upath' #238
- 🦺 Removed 'add to list' buttons in manga download modal media cards
- 🦺 Manga: Fixed reader keybinding editing not working on macOS desktop
- 🦺 Fixed AniList page filters not persisting
- 🦺 Fixed 'Advanced Search' input not being emptied when resetting search params
- 🦺 Extensions: Fixed caught exceptions being logged as empty objects
- 🦺 Fixed extension market button disabled by custom background image
- 🦺 Fixed Plugin APIs
    - Fixed DOM manipulation methods not working
    - Correctly remove DOM elements created by plugin when unloaded
    - Fixed incorrectly named hooks
    - Fixed manga bindings for promises
    - Fixed select and radio group tray components
    - Fixed incorrect event object field mapping (Breaking)
- 🏗️ Frontend: Replace 'upath' dependency

## v2.8.3

- ⚡️ Updated Playground 
- ⚡️ Discover page: Play the trailer on hover; carousel buttons 
- 🦺 Playground: Fix online streaming search options missing media object
- 🦺 Discord: Fixed anime rich presence displaying old episodes
- 🦺 Discord: Fixed manga rich presence activity #282
- 🦺 Library: Fixed anime unwatched count for shows not in the library
- 🦺 Library: Fixed filtering for shows not in the library
- 🦺 Library: Fixed 'Show unwatched only' filter
- 🦺 Torrent search: Fixed Nyaa batch search with 'any' resolution
- 🏗️ Torrent Search: Truncate displayed language label number

## v2.8.2

- ✨ UI: Custom CSS support
- ✨ In-app extension marketplace
    - Find extensions to install directly from the interface
- ⚡️ Discord: Rich Presence anime activity with progress track
- ⚡️ Torrent: New 'Nyaa (Non-English)' built-in extension with smart search
- ⚡️ Torrent search: Added labels for audio, video, subtitles, dubs
- ⚡️ Torrent search: Improved non-smart search UI
- ⚡️ Extensions: Built-in extensions now support user preferences
  - API Urls are now configurable for some built-in extensions
- ⚡️ Extensions: Auto check for updates with notification
- ⚡️ Extensions: Added media object to Online streaming search options
- ⚡️ Extensions: User config (preferences) now accessible with '$getUserPreference' global function
- ⚡️ UI Settings: Color scheme live preview #277
- ⚡️ Manga: Fullscreen toggle on mobile (Android) #279
- 🦺 Library: Fixed genre selector making library disappear #275
- 🦺 Online streaming: Fixed search query being altered
- 🦺 Fixed offline mode infinite loading screen (regression from v2.7.2) #278
- 🦺 Extensions: Fixed playground console output #276
- 🦺 Extensions: Fixed JS extension pool memory leak
- 🦺 Extensions: Fixed Plugin Actions API
- 🏗️ Removed Cloudflare bypass from ComicK extension
- 🏗️ Extensions: Deprecated 'getMagnetLinkFromTorrentData' in favor of '$torrentUtils.getMagnetLinkFromTorrentData'
- 🏗️ Plugins: New 'ctx.anime' API
- 🏗️ Server: Use binary (IEC) measurement on Windows and Linux #280
- 🏗️ Extensions: Updated and fixed type declaration files
- 🏗️ Extensions: New 'semverConstraint' field

## v2.8.1

- 🦺 Fixed runtime error when launching the app for the first time
- 🦺 Fixed torrent search episode input
- 🦺 Fixed update popup showing empty "Updates you've missed"

## v2.8.0

- 🎉 Plugins: A powerful new way to extend and customize Seanime
    - Build your own features using a wide range of APIs — all in JavaScript.
- ✨ Playback: Faster media tracking, better responsiveness
    - Faster autoplay, progress tracking, playlists
- ✨ Torrent streaming: Improved performance and responsiveness
    - Streams start up to 2x faster, movies start up to 50x faster
- ✨ Server: DNS over HTTPS support
- ✨ Manga: Refresh all sources at once #233
- ✨ Library/Streaming: Episode list now includes specials included by AniList in main count
- ✨ Torrent search: Sorting options #253
- ✨ Debrid streaming: Improved stream startup time
- ✨ Library: New 'Most/least recent watch' sorting options (w/ watch history enabled) #244
- ✨ Extensions: Ability to edit the code of installed extensions
- ⚡️ Streaming: Added Nyaa as a fallback provider for auto select
- ⚡️ Manga: Unread count badge now takes into account selected scanlator and language
- ⚡️ Torrent list: Stop all completed torrents #250
- ⚡️ Library/Streaming: Improved handling of discrepancies between AniList and AniDB
- ⚡️ Library: Show episode summaries by default #265
- ⚡️ UI: Option to hide episode summaries and episode filename
- ⚡️ AniList: Option to clear date field when editing entry
- ⚡️ Extensions: New 'Update all' button to update all extensions at once 
- ⚡️ Extensions: Added 'payloadURI' as an alternative to pasting extension code
- ⚡️ Extensions: 'Development mode' that allows loading source code from a file in the manifest
- ⚡️ Torrent streaming: Option to change cache directory
- ⚡️ Manga: Selecting a language will now filter scanlator options and vice versa
- ⚡️ Discover page: Context menu for 'Airing Schedule' items #267 - @kyoruno
- ⚡️ Added AniList button to preview modals #264 - @kyoruno
- 🦺 Fixed AnimeTosho smart search #260
- 🦺 AutoPlay: Fixed autoplay starting erroneously
- 🦺 Scanner: Fixed local file parsing with multiple directories
- 🦺 Scanner: Fixed resolved symlinks being ignored #251
- 🦺 Scanner: Removed post-matching validation causing some files to be unmatched #246
- 🦺 Library: Fixed 'unwatched episode count' not showing with 'repeating' status
- 🦺 Library: Fixed incorrect episode offset for some anime
- 🦺 Torrent search: Fixed excessive API requests being sent during search query typing
- 🦺 Parser: Fixed crash caused by parsing 'SxExxx-SxExxx'
- 🦺 Video Proxy: Fixed streaming .mp4 media files - @kRYstall9
- 🦺 Extensions: Fixed bug causing invalid extensions to be uninstallable from UI
- 🦺 Extensions: Fixed concurrent fetch requests and concurrent executions
- 🏗️ Debrid streaming changes
    - Added visual feedback when video is being sent to media player
    - Removed stream integrity check for faster startup
- 🏗️ Refactored websocket system
    - New bidirectional communication between client and server
    - Better handling of silent websocket connection closure
- 🏗️ Refactored extension system
    - Usage of runtime pools for better performance and concurrency
    - Improved JS bindings/bridges
- 🏗️ Web UI: Added data attributes to HTML elements
- 🏗️ Offline mode: Syncing now caches downloaded chapters if refetching
- 🏗️ BREAKING(Extensions): Content provider extension methods are now run in separate runtimes
    - State sharing across methods no longer works but concurrent execution is now possible
- ⬆️ Migrated to Go 1.24.1
- ⬆️ Updated dependencies

## v2.7.5

- 🦺 Extensions: Fixed runtime errors caused by concurrent requests
- 🦺 Manga: Removed light novels from manga library #234
- 🦺 Fixed torrent stream overlay blocking UI #243
- 🏗️ Server: Removed DNS resolver fallback

## v2.7.4

- 🚑️ Fixed infinite loading screen when launching app for the first time
- ⚡️ External player link: Option to encode file path to Base64 (v2.7.3) 
- 🦺 Desktop: Fixed startup failing due to long AniList request (v2.7.3) 
- 🦺 Debrid: Fixed downloading to nonexistent destination (v2.7.3) 
- 🦺 Anime library: Fixed external player link not working due to incorrect un-escaping (v2.7.3) 
- 🦺 Small UI fixes (v2.7.3) 
- 🏗️ Server: Support serving Base64 encoded file paths (v2.7.3)

## v2.7.3

- ⚡️ External player link: Option to encode file path to Base64 
- 🦺 Desktop: Fixed startup failing due to long AniList request #232
- 🦺 Debrid: Fixed downloading to nonexistent destination #237
- 🦺 Anime library: Fixed external player link not working due to incorrect un-escaping #240
- 🦺 Small UI fixes
- 🏗️ Server: Support serving Base64 encoded file paths

## v2.7.2

- 🦺 Fixed error alert regression
- 🦺 Anime library: Fixed downloading to library root #231
- 🦺 Fixed getting log file contents on Linux
- 🏗️ Use library for 'copy to clipboard' feature

## v2.7.1

- ⚡️ Transcoding: Support for Apple VideoToolbox hardware acceleration
- ⚡️ Manga: New built-in extension
- 🦺 Fixed hardware acceleration regression
- 🦺 Fixed client cookie regression causing external player links to fail
- 🦺 Fixed Direct Play regression #224
- 🦺 Anime library: Fixed selecting multiple episodes to download at once #223
- 🦺 Desktop: Fixed copy to clipboard
- 🦺 Fixed UI inconsistencies
- 🏗️ Extensions: Removed non-working manga extension
- 🏗️ Improved logging in some areas
- 🏗️ Desktop: Refactored macOS fullscreen

## v2.7.0

- ✨ Updated design
- ✨ Command palette (Experimental)
  - Quickly browse, search, perform actions, with more options to come
  - Allows navigation with keyboard only #46
- ✨ Preview cards
  - Preview an anime/manga by right-clicking on a media card
- ✨ Library: Filtering options #210
  - Filter to see only anime with unseen episodes and manga with unread chapters #175 (Works if chapters are cached)
  - New sorting options: Aired recently, Highest unwatched count, ...
- ✨ New UI Settings
  - 'Continue watching' sorting, card customization
  - Show unseen count for anime cards #209
- ⚡️ Torrent/Debrid streaming: 'Auto play next episode' now works with manually selected batches #211
  - This works only if the user did not select the file manually
- ⚡️ Server: Reduced memory usage, improved performance
- ⚡️ Discord Rich Presence now works with online & media streaming
- ⚡️ 'Continue watching' UI setting options, defaults to 'Aired recently'
  - BREAKING: Manga unread count badge needs to be reactivated in settings
- ⚡️ Torrent streaming: Slow seeding mode #200
- ⚡️ Debrid streaming: Auto-select file option
- ⚡️ Quick action menu #197
  - Open preview cards, more options to come
- ⚡️ Revamped Settings page
- ⚡️ Anime library: Improved Direct Play performance
- ⚡️ Quickly add media to AniList from its card
- 🦺 Torrent streaming: Fixed auto-selected file from batches not being downloaded #215
  - Fixed piece prioritization
- 🦺 Debrid streaming: Fixed streaming shows with no AniDB mapping 
- 🦺 Anime library: 'Remove empty directories' now works for other library folders
- 🦺 Anime library: Download destination check now takes all library paths into account
- 🦺 Online streaming: Fixed 'auto next' not playing the last episode
- 🦺 Server: Fixed empty user agent header leading to some failed requests 
- 🦺 Anime library: Ignore AppleDouble files on macOS #208
- 🦺 Manga: Fixed synonyms not being taken into account for auto matching
- 🦺 Manga: Fixed genre link opening anime in advanced search
- 🦺 Extension Playground: Fixed anime torrent provider search input empty value
- 🦺 Continuity: Ignore watch history above a certain threshold
- 🦺 Online streaming: Fixed selecting highest quality by default
- 🦺 Fixed Auto Downloader queuing same items
- 🦺 Manga: Fixed pagination when filtering by language/scanlator #217
- 🦺 Manga: Fixed page layout overflowing on mobile
- 🦺 Torrent streaming: Fixed incorrect download/upload speeds
- 🦺 Anime library: Fixed special episode sorting
- 🏗️ Server: Migrated API from Fiber (FastHTTP) to Echo (HTTP)
- 🏗 External media players: Increased retries when streaming
- 🏗 Torrent streaming: Serve stream from main server
- 🏗 Watch history: Bumped limit from 50 to 100 
- 🏗 Integrated player: Merged both online & media streaming players
  - BREAKING: Auto play, Auto next, Auto skip player settings have been reset to 'off'
- 🏗 Renaming and Removals
  - Scanner: Renamed 'matching data' checkbox
  - Torrent/Debrid streaming: Renamed 'Manually select file' to 'Auto select file'
  - Removed 'Use legacy episode cards' option
  - 'Fluid' media page header layout is now the default
- ⬆️ Migrated to Go 1.23.5
- ⬆️ Updated dependencies

## v2.6.2

- ⚡️ Advanced search: Maintain search params during navigation #195
- 🦺 Torrent streaming: Fixed playback issue
- 🦺 Auto Downloader: Fixed list not updating correctly after batch creation
- 🔧 Torrent streaming: Reverted to using separate streaming server

## v2.6.1

- ⚡️ Anime library: Filtering by year now takes into account the season year
- ⚡️ Torrent streaming: Custom stream URL address setting #182
- 🦺 Scanner: Fixed duplicated files due to incorrect path comparison
- 🦺 Use AniList season year instead of start year for media cards #193
- 🏗️ Issue recorder: Increase data cap limit

## v2.6.0

- ✨ In-app issue log recorder
  - Record browser, network and server logs from an issue you encounter in the app and generate an anonymized file to send for bug reports
- ⚡️ Auto Downloader: Added support for batch creation of rules #180
- ⚡️ Scanner: Improved default matching algorithm
- ⚡️ Scanner: Option to choose different matching algorithms
- ⚡️ Scanner: Improved filename parser, support for SxPx format
- ⚡️ Scanner: Reduced log file sizes and forced logging to single file per scan
- ⚡️ Improved Discover manga page
- ⚡️ New manga filters for country and format #191
- ⚡️ Torrent streaming: Serve streams from main server (Experimental)
  - Lower memory usage, removes need for separate server
- ⚡️ Auto deletion of log files older than 14 days #184
- ⚡️ Online streaming: Added 'f' keybinding to restore fullscreen #186
- 💄 Media page banner image customization #185
- 💄 Media banner layout customization
- 💄 Updated user interface settings page
- 💄 Updated some styles
- 💄 Added 'Fix border rendering artifacts' option to UI settings
- 🦺 Fixed Auto Downloader form #187
- 🦺 Streaming: Fixed auto-select for media with very long titles
- 🦺 Fixed torrent streaming on VLC
- 🦺 Fixed MPV resumed playback with watch continuity enabled
- 🦺 Desktop: Fixed sidebar menu item selection
- 🏗️ Auto Downloader: Set minimum refresh interval to 15 minutes (BREAKING)
  - If your refresh interval less than 15 minutes, it will be force set to 20 minutes. Update the settings accordingly.
- 🏗️ Moved 'watch continuity' setting to 'Seanime' tab

## v2.5.2

- 🦺 Fixed SeaDex extension #179
- 🦺 Fixed Auto Downloader title comparison
- 🦺 Fixed m3u8 proxy HTTP/2 runtime error on Linux
- 🦺 Fixed Auto Downloader array fields
- 🦺 Fixed online streaming error caused by decimals
- 🦺 Fixed manual progress tracking cancellation
- 🦺 Fixed playback manager deadlock
- 🦺 Desktop: Fixed external player links
- 🦺 Desktop: Fixed local file downloading (macOS)
- 🦺 Desktop: Fixed 'open in browser' links (macOS)
- 🦺 Desktop: Fixed torrent list UI glitches (macOS)
- 🏗️ Desktop: Added 'reload' button to loading screen
- ⬆️ Updated filename parser
  - Fixes aggressive episode number parsing in rare cases
- ⬆️ Updated dependencies
- 🔑 Updated license to GPL-3.0

## v2.5.1

- 💄 Updated built-in media player theme
- 🦺 Fixed Auto Downloader form fields (regression)
- 🦺 Fixed online streaming extension API url (regression)
- ⬆️ Migrated to Go 1.23.4
- ⬆️ Updated dependencies

## v2.5.0

- ⚡️ UI: Improved rendering performance
- ⚡️ Online streaming: Built-in Animepahe extension (Experimental)
- ⚡️ Desktop: Automatically restart server process when it crashes/exits
- ⚡️ Desktop: Added 'Restart server' button when server process is terminated
- ⚡️ Auto progress update now works for built-in media player
- ⚡️ Desktop: Back/Forward navigation buttons #171
- ⚡️ Open search page by clicking on media genres and ranks #172
- ⚡️ Support for AniList 'repeat' field #169
- ⚡️ Ignore dropped anime in missing episodes #170
- ⚡️ Improved media player error logging
- ⚡️ Online streaming: m3u8 video proxy support
- ⚡️ Ability to add to AniList individually in 'Resolve unknown media'
- 🦺 Fixed TorBox failed archive extraction
- 🦺 Fixed incorrect 'user-preferred' title languages
- 🦺 Fixed One Piece streaming episode list
- 🦺 Added workaround for macOS video player fullscreen issue #168
  - Clicking 'Hide from Dock' from the tray will solve the issue
- 🦺 Fixed torrent streaming runtime error edge case
- 🦺 Fixed scanner 'Do not use AniList data' runtime error
- 🦺 Fixed Transmission host setting not being applied
- 🦺 Javascript VM: Fixed runtime panics caused by 'fetch' data races
- 🦺 Online streaming: Fixed scroll to current episode
- 🦺 Online streaming: Fixed selecting highest/default quality by default
- 🦺 Fixed UI inconsistencies
- 🏗️ Removed 'Hianime' online streaming extension
- 🏗️ Real Debrid: Select all files by default
- 🏗️ UI: Improved media card virtualized grid performance
- 🏗️ Javascript VM: Added 'url' property to fetch binding
- 🏗️ Reduced online streaming cache duration
- 🏗️ Core: Do not print stack traces concurrently
- 🏗️ UI: Use React Compiler (Experimental)
- ⬆️ Updated dependencies

## v2.4.2

- ⚡️ 'Include in library' will keep displaying shows when caught up
- ⚡️ Settings: Open data directory button
- 🦺 Desktop: Fixed authentication issue on macOS
- ⚡️ Desktop: Force single instance
- ⚡️ Desktop: Try to shut down server on force exit
- ⚡️ Desktop: Disallow update from Web UI
- 🦺 Desktop: Fixed 'toggle visibility'
- 🦺 Desktop: Fixed 'server process terminated' issue

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

