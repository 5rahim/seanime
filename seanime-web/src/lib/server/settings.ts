import { z } from "zod"

export const DEFAULT_TORRENT_PROVIDER = "animetosho"

export const DEFAULT_TORRENT_CLIENT = "qbittorrent"

export const DEFAULT_DOH_PROVIDER = ""

export const DEFAULT_MPV_TYPE = "socket"

export const enum TORRENT_CLIENT {
    QBITTORRENT = "qbittorrent",
    TRANSMISSION = "transmission",
    NONE = "none",
}

export const enum TORRENT_PROVIDER {
    ANIMETOSHO = "animetosho",
    NYAA = "nyaa",
    NONE = "none",
}

export const _gettingStartedSchema = z.object({
    enableTranscode: z.boolean().optional().default(false),
    enableTorrentStreaming: z.boolean().optional().default(false),
    debridProvider: z.string().optional().default("none"),
    debridApiKey: z.string().optional().default(""),
})

export const settingsSchema = z.object({
    libraryPath: z.string().optional().default(""),
    defaultPlayer: z.string(),
    torrentProvider: z.string().default(DEFAULT_TORRENT_PROVIDER),
    autoScan: z.boolean().optional().default(false),
    mediaPlayerHost: z.string(),
    vlcUsername: z.string().optional().default(""),
    vlcPassword: z.string().optional().default(""),
    vlcPort: z.number(),
    vlcPath: z.string().optional().default(""),
    mpcPort: z.number(),
    mpcPath: z.string().optional().default(""),
    mpvSocket: z.string().optional().default(""),
    mpvPath: z.string().optional().default(""),
    defaultTorrentClient: z.string().optional().default(DEFAULT_TORRENT_CLIENT),
    hideTorrentList: z.boolean().optional().default(false),
    qbittorrentPath: z.string().optional().default(""),
    qbittorrentHost: z.string().optional().default(""),
    qbittorrentPort: z.number(),
    qbittorrentUsername: z.string().optional().default(""),
    qbittorrentPassword: z.string().optional().default(""),
    qbittorrentTags: z.string().optional().default(""),
    transmissionPath: z.string().optional().default(""),
    transmissionHost: z.string().optional().default(""),
    transmissionPort: z.number().optional().default(9091),
    transmissionUsername: z.string().optional().default(""),
    transmissionPassword: z.string().optional().default(""),
    hideAudienceScore: z.boolean().optional().default(false),
    autoUpdateProgress: z.boolean().optional().default(false),
    disableUpdateCheck: z.boolean().optional().default(false),
    enableOnlinestream: z.boolean().optional().default(false),
    includeOnlineStreamingInLibrary: z.boolean().optional().default(false),
    disableAnimeCardTrailers: z.boolean().optional().default(false),
    enableManga: z.boolean().optional().default(true),
    enableRichPresence: z.boolean().optional().default(false),
    enableAnimeRichPresence: z.boolean().optional().default(false),
    enableMangaRichPresence: z.boolean().optional().default(false),
    enableAdultContent: z.boolean().optional().default(false),
    blurAdultContent: z.boolean().optional().default(false),
    dohProvider: z.string().optional().default(""),
    openTorrentClientOnStart: z.boolean().optional().default(false),
    openWebURLOnStart: z.boolean().optional().default(false),
    refreshLibraryOnStart: z.boolean().optional().default(false),
    richPresenceHideSeanimeRepositoryButton: z.boolean().optional().default(false),
    richPresenceShowAniListMediaButton: z.boolean().optional().default(false),
    richPresenceShowAniListProfileButton: z.boolean().optional().default(false),
    disableNotifications: z.boolean().optional().default(false),
    disableAutoDownloaderNotifications: z.boolean().optional().default(false),
    disableAutoScannerNotifications: z.boolean().optional().default(false),
    defaultMangaProvider: z.string().optional().default(""),
    mangaAutoUpdateProgress: z.boolean().optional().default(false),
    autoPlayNextEpisode: z.boolean().optional().default(false),
    showActiveTorrentCount: z.boolean().optional().default(false),
    enableWatchContinuity: z.boolean().optional().default(false),
    libraryPaths: z.array(z.string()).optional().default([]),
    autoSyncOfflineLocalData: z.boolean().optional().default(false),
    scannerMatchingThreshold: z.number().optional().default(0.5),
    scannerMatchingAlgorithm: z.string().optional().default(""),
})

export const gettingStartedSchema = _gettingStartedSchema.extend(settingsSchema.shape)

export function useDefaultSettingsPaths() {

    return {
        getDefaultVlcPath: (os: string) => {
            switch (os) {
                case "windows":
                    return "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
                case "linux":
                    return "/usr/bin/vlc" // Default path for VLC on most Linux distributions
                case "darwin":
                    return "/Applications/VLC.app/Contents/MacOS/VLC" // Default path for VLC on macOS
                default:
                    return "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
            }
        },
        getDefaultQBittorrentPath: (os: string) => {
            switch (os) {
                case "windows":
                    return "C:/Program Files/qBittorrent/qbittorrent.exe"
                case "linux":
                    return "/usr/bin/qbittorrent" // Default path for Client on most Linux distributions
                case "darwin":
                    return "/Applications/qbittorrent.app/Contents/MacOS/qbittorrent" // Default path for Client on macOS
                default:
                    return "C:/Program Files/qBittorrent/qbittorrent.exe"
            }
        },
        getDefaultTransmissionPath: (os: string) => {
            switch (os) {
                case "windows":
                    return "C:/Program Files/Transmission/transmission-qt.exe"
                case "linux":
                    return "/usr/bin/transmission-gtk"
                case "darwin":
                    return "/Applications/Transmission.app/Contents/MacOS/Transmission"
                default:
                    return "C:/Program Files/Transmission/transmission-qt.exe"
            }
        },
    }

}

export function getDefaultMpcSocket(os: string) {
    switch (os) {
        case "windows":
            return "\\\\.\\pipe\\mpv_ipc"
        case "linux":
            return "/tmp/mpv_socket" // Default socket for VLC on most Linux distributions
        case "darwin":
            return "/tmp/mpv_socket" // Default socket for VLC on macOS
        default:
            return "/tmp/mpv_socket"
    }

}
