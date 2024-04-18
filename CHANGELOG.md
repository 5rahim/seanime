# Changelog

All notable changes to this project will be documented in this file.

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

