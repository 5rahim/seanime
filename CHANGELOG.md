# Changelog

All notable changes to this project will be documented in this file.

## 0.2.0

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

