export const enum SeaEndpoints {
    STATUS = "/status",
    LOGIN = "/auth/login",
    LOGOUT = "/auth/logout",
    SETTINGS = "/settings", // (PATCH)
    LIST_SYNC_SETTINGS = "/settings/list-sync", // (PATCH)
    AUTO_DOWNLOADER_SETTINGS = "/settings/auto-downloader", // (PATCH)
    START_MEDIA_PLAYER = "/media-player/start", // (POST)
    START_MPV_PLAYBACK_DETECTION = "/media-player/mpv-detect-playback", // (POST)
    OPEN_IN_EXPLORER = "/open-in-explorer", // (POST)
    PLAY_VIDEO = "/media-player/play",
    /**
     * AniList
     */
    ANILIST_LIST_ENTRY = "/anilist/list-entry", // (POST)
    ANILIST_LIST_ENTRY_PROGRESS = "/anilist/list-entry/progress", // (POST)
    ANILIST_COLLECTION = "/anilist/collection", // (GET, POST)
    ANILIST_MEDIA_DETAILS = "/anilist/media-details/{id}", // (GET)
    /**
     * MAL
     */
    MAL_AUTH = "/mal/auth",
    MAL_LOGOUT = "/mal/logout",
    MAL_LIST_ENTRY_PROGRESS = "/mal/list-entry/progress", // (POST)
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
}

export const enum WSEvents {
    SCAN_PROGRESS = "scan-progress",
    SCAN_STATUS = "scan-status",
    REFRESHED_ANILIST_COLLECTION = "refreshed-anilist-collection",
    MEDIA_PLAYER_TRACKING_STOPPED = "media-player-tracking-stopped",
    MEDIA_PLAYER_TRACKING_STARTED = "media-player-tracking-started",
    MEDIA_PLAYER_VIDEO_COMPLETED = "media-player-video-completed",
    MEDIA_PLAYER_PROGRESS_UPDATE_REQUEST = "media-player-progress-update-request",
    MEDIA_PLAYER_PLAYBACK_STATUS = "media-player-playback-status",
    LIBRARY_WATCHER_FILE_ADDED = "library-watcher-file-added",
    LIBRARY_WATCHER_FILE_REMOVED = "library-watcher-file-removed",
    AUTO_DOWNLOADER_ITEM_ADDED = "auto-downloader-item-added",
}
