export const enum SeaEndpoints {
    STATUS = "/status",
    LOGIN = "/auth/login",
    LOGOUT = "/auth/logout",
    SETTINGS = "/settings", // (PATCH)
    LIST_SYNC_SETTINGS = "/settings/list-sync", // (PATCH)
    AUTO_DOWNLOADER_SETTINGS = "/settings/auto-downloader", // (PATCH)
    START_MEDIA_PLAYER = "/media-player/start", // (POST)
    OPEN_IN_EXPLORER = "/open-in-explorer", // (POST)
    PLAY_VIDEO = "/media-player/play",
    /**
     * AniList
     */
    ANILIST_LIST_ENTRY = "/anilist/list-entry", // (POST)
    ANILIST_COLLECTION = "/anilist/collection", // (GET, POST)
    ANILIST_MEDIA_DETAILS = "/anilist/media-details/{id}", // (GET)
    ANILIST_LIST_ANIME = "/anilist/list-anime", // (POST)
    ANILIST_LIST_RECENT_ANIME = "/anilist/list-recent-anime", // (POST)
    /**
     * MAL
     */
    MAL_AUTH = "/mal/auth",
    MAL_LOGOUT = "/mal/logout",
    /**
     * List Sync
     */
    LIST_SYNC_ANIME = "/list-sync/anime", // (POST)
    LIST_SYNC_ANIME_DIFFS = "/list-sync/anime-diffs", // (GET)
    LIST_SYNC_CACHE = "/list-sync/cache", // (POST)
    /**
     * Library
     */
    EMPTY_DIRECTORIES = "/library/empty-directories", // (POST)
    LOCAL_FILES = "/library/local-files", // (GET, POST)
    LIBRARY_COLLECTION = "/library/collection", // (GET)
    MISSING_EPISODES = "/library/missing-episodes", // (GET)
    SCAN_LIBRARY = "/library/scan", // (POST)
    LOCAL_FILE = "/library/local-file", // (PATCH)
    MEDIA_ENTRY = "/library/media-entry/{id}", // (GET)
    MEDIA_ENTRY_SUGGESTIONS = "/library/media-entry/suggestions", // (POST)
    MEDIA_ENTRY_MANUAL_MATCH = "/library/media-entry/manual-match", // (POST)
    MEDIA_ENTRY_BULK_ACTION = "/library/media-entry/bulk-action", // (PATCH)
    OPEN_MEDIA_ENTRY_IN_EXPLORER = "/library/media-entry/open-in-explorer", // (POST)
    MEDIA_ENTRY_UNKNOWN_MEDIA = "/library/media-entry/unknown-media", // (POST)
    MEDIA_ENTRY_SILENCE_STATUS = "/library/media-entry/silence/{id}", // (GET)
    UPDATE_PROGRESS = "/library/media-entry/update-progress", // (POST)
    MEDIA_ENTRY_SILENCE = "/library/media-entry/silence", // (POST)
    SCAN_SUMMARIES = "/library/scan-summaries", // (GET)
    /**
     * Download/Torrent
     */
    DOWNLOAD_TORRENT_FILE = "/download-torrent-file", // (POST)
    TORRENT_CLIENT_DOWNLOAD = "/torrent-client/download", // (POST)
    TORRENT_CLIENT_LIST = "/torrent-client/list", // (GET)
    TORRENT_CLIENT_ACTION = "/torrent-client/action", // (POST)
    TORRENT_CLIENT_RULE_MAGNET = "/torrent-client/rule-magnet", // (POST)
    TORRENT_SEARCH = "/torrent-search", // (POST)
    TORRENT_NSFW_SEARCH = "/torrent-nsfw-search", // (POST)
    /**
     * Auto downloader
     */
    AUTO_DOWNLOADER_RULES = "/auto-downloader/rules", // (GET)
    AUTO_DOWNLOADER_RULE = "/auto-downloader/rule", // (POST, PATCH)
    AUTO_DOWNLOADER_RULE_DETAILS = "/auto-downloader/rule/{id}", // (GET, DELETE)
    AUTO_DOWNLOADER_ITEMS = "/auto-downloader/items", // (GET)
    AUTO_DOWNLOADER_ITEM = "/auto-downloader/item", // (DELETE)
    RUN_AUTO_DOWNLOADER = "/auto-downloader/run", // (POST)
    /**
     * Updates
     */
    LATEST_UPDATE = "/latest-update", // (GET)
    DOWNLOAD_RELEASE = "/download-release", // (POST)
    /**
     * Playback Manager
     */
    PLAYBACK_MANAGER_SYNC_CURRENT_PROGRESS = "/playback-manager/sync-current-progress", // (POST)
    PLAYBACK_MANAGER_START_PLAYLIST = "/playback-manager/start-playlist", // (POST)
    PLAYBACK_MANAGER_CANCEL_PLAYLIST = "/playback-manager/cancel-playlist", // (POST)
    PLAYBACK_MANAGER_PLAYLIST_NEXT = "/playback-manager/playlist-next", // (GET)
    PLAYBACK_MANAGER_NEXT_EPISODE = "/playback-manager/next-episode", // (POST)
    /**
     * Playlist
     */
    PLAYLISTS = "/playlists", // (GET)
    PLAYLIST = "/playlist", // (POST, PATCH, DELETE)
    PLAYLIST_EPISODES = "/playlist/episodes/{id}/{progress}", // (POST)
    /**
     * OnlineStream
     */
    ONLINESTREAM_EPISODE_LIST = "/onlinestream/episode-list", // (POST)
    ONLINESTREAM_EPISODE_SOURCE = "/onlinestream/episode-source", // (POST)
    ONLINESTREAM_CACHE = "/onlinestream/cache", // (DELETE)
    /**
     * Metadata Provider
     */
    METADATA_PROVIDER_TVDB_EPISODES = "/metadata-provider/tvdb-episodes", // (POST, DELETE)
    /**
     * Manga
     */
    MANGA_ANILIST_COLLECTION = "/manga/anilist/collection", // (POST)
    MANGA_COLLECTION = "/manga/collection", // (GET)
    MANGA_ANILIST_LIST_MANGA = "/manga/anilist/list", // (POST)
    MANGA_ENTRY = "/manga/entry/{id}", // (GET)
    MANGA_ENTRY_DETAILS = "/manga/entry/{id}/details", // (GET)
    MANGA_ENTRY_CACHE = "/manga/entry/cache", // (DELETE)
    MANGA_CHAPTERS = "/manga/chapters", // (POST)
    MANGA_PAGES = "/manga/pages", // (POST)
    MANGA_ENTRY_BACKUPS = "/manga/entry/backups", // (POST)
    DOWNLOAD_MANGA_CHAPTER = "/manga/download-chapter", // (POST)
    UPDATE_MANGA_PROGRESS = "/manga/update-progress", // (POST)
    /**
     * File Cache
     */
    FILECACHE_TOTAL_SIZE = "/filecache/total-size", // (GET)
    FILECACHE_BUCKET = "/filecache/bucket", // (DELETE)
    /**
     * Discord
     */
    DISCORD_PRESENCE_MANGA = "/discord/presence/manga", // (POST)
    DISCORD_PRESENCE_CANCEL = "/discord/presence/cancel", // (POST)
    /**
     * Theme
     */
    THEME = "/theme", // (GET, PATCH)
}

export const enum WSEvents {
    SCAN_PROGRESS = "scan-progress",
    SCAN_STATUS = "scan-status",
    REFRESHED_ANILIST_COLLECTION = "refreshed-anilist-collection",
    REFRESHED_ANILIST_MANGA_COLLECTION = "refreshed-anilist-manga-collection",
    LIBRARY_WATCHER_FILE_ADDED = "library-watcher-file-added",
    LIBRARY_WATCHER_FILE_REMOVED = "library-watcher-file-removed",
    AUTO_DOWNLOADER_ITEM_ADDED = "auto-downloader-item-added",
    AUTO_SCAN_STARTED = "auto-scan-started",
    AUTO_SCAN_COMPLETED = "auto-scan-completed",
    PLAYBACK_MANAGER_PROGRESS_TRACKING_STARTED = "playback-manager-progress-tracking-started",
    PLAYBACK_MANAGER_PROGRESS_TRACKING_STOPPED = "playback-manager-progress-tracking-stopped",
    PLAYBACK_MANAGER_PROGRESS_VIDEO_COMPLETED = "playback-manager-progress-video-completed",
    PLAYBACK_MANAGER_PROGRESS_PLAYBACK_STATE = "playback-manager-progress-playback-state",
    PLAYBACK_MANAGER_PROGRESS_UPDATED = "playback-manager-progress-updated",
    PLAYBACK_MANAGER_PLAYLIST_STATE = "playback-manager-playlist-state",
    MANGA_DOWNLOADER_DOWNLOADING_PROGRESS = "manga-downloader-downloading-progress",
    ERROR_TOAST = "error-toast",
    SUCCESS_TOAST = "success-toast",
    INFO_TOAST = "info-toast",
    WARNING_TOAST = "warning-toast",
}
