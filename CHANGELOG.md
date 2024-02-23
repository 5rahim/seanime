# Changelog

All notable changes to this project will be documented in this file.

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

