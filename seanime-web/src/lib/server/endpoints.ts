export const enum SeaEndpoints {
    STATUS = "/status",
    LOGIN = "/auth/login",
    LOGOUT = "/auth/logout",
    SETTINGS = "/settings", // (PATCH)
    START_MEDIA_PLAYER = "/media-player/start", // (POST)
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
    /**
     * Nyaa
     */
    NYAA_SEARCH = "/nyaa/search",
    /**
     * Download/Torrent
     */
    DOWNLOAD = "/download", // (POST)
    TORRENTS = "/torrents", // (GET)
    TORRENT = "/torrent", // (POST)
}

export const enum WSEvents {
    SCAN_PROGRESS = "scan-progress",
    REFRESHED_ANILIST_COLLECTION = "refreshed-anilist-collection",
    MEDIA_PLAYER_TRACKING_STOPPED = "media-player-tracking-stopped",
    MEDIA_PLAYER_TRACKING_STARTED = "media-player-tracking-started",
    MEDIA_PLAYER_VIDEO_COMPLETED = "media-player-video-completed",
    MEDIA_PLAYER_PLAYBACK_STATUS = "media-player-playback-status",
}