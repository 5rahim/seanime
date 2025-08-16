import { GettingStarted_Variables } from "@/api/generated/endpoint.types"
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
    NYAA_NON_ENG = "nyaa-non-eng",
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
    mpvArgs: z.string().optional().default(""),
    iinaSocket: z.string().optional().default(""),
    iinaPath: z.string().optional().default(""),
    iinaArgs: z.string().optional().default(""),
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
    mangaLocalSourceDirectory: z.string().optional().default(""),
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
    richPresenceUseMediaTitleStatus: z.boolean().optional().default(true),
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
    autoSyncToLocalAccount: z.boolean().optional().default(false),
    nakamaIsHost: z.boolean().optional().default(false),
    nakamaHostPassword: z.string().optional().default(""),
    nakamaRemoteServerURL: z.string().optional().default(""),
    nakamaRemoteServerPassword: z.string().optional().default(""),
    nakamaHostShareLocalAnimeLibrary: z.boolean().optional().default(false),
    nakamaEnabled: z.boolean().optional().default(false),
    nakamaHostEnablePortForwarding: z.boolean().optional().default(false),
    nakamaUsername: z.string().optional().default(""),
    includeNakamaAnimeLibrary: z.boolean().optional().default(false),
    nakamaHostUnsharedAnimeIds: z.array(z.number()).optional().default([]),
    autoSaveCurrentMediaOffline: z.boolean().optional().default(false),
})

export const gettingStartedSchema = _gettingStartedSchema.extend(settingsSchema.shape)

export const getDefaultSettings = (data: z.infer<typeof gettingStartedSchema>): GettingStarted_Variables => ({
    library: {
        libraryPath: data.libraryPath,
        autoUpdateProgress: true,
        disableUpdateCheck: false,
        torrentProvider: data.torrentProvider || DEFAULT_TORRENT_PROVIDER,
        autoScan: false,
        disableAnimeCardTrailers: false,
        enableManga: data.enableManga,
        enableOnlinestream: data.enableOnlinestream,
        dohProvider: DEFAULT_DOH_PROVIDER,
        openTorrentClientOnStart: false,
        openWebURLOnStart: false,
        refreshLibraryOnStart: false,
        autoPlayNextEpisode: false,
        enableWatchContinuity: data.enableWatchContinuity,
        libraryPaths: [],
        autoSyncOfflineLocalData: false,
        includeOnlineStreamingInLibrary: false,
        scannerMatchingThreshold: 0,
        scannerMatchingAlgorithm: "",
        autoSyncToLocalAccount: false,
        autoSaveCurrentMediaOffline: false,
    },
    nakama: {
        enabled: false,
        isHost: false,
        hostPassword: "",
        remoteServerURL: "",
        remoteServerPassword: "",
        hostShareLocalAnimeLibrary: false,
        username: data.nakamaUsername,
        includeNakamaAnimeLibrary: false,
        hostUnsharedAnimeIds: [],
        hostEnablePortForwarding: false,
    },
    manga: {
        defaultMangaProvider: "",
        mangaAutoUpdateProgress: false,
        mangaLocalSourceDirectory: "",
    },
    mediaPlayer: {
        host: data.mediaPlayerHost,
        defaultPlayer: data.defaultPlayer,
        vlcPort: data.vlcPort,
        vlcUsername: data.vlcUsername || "",
        vlcPassword: data.vlcPassword,
        vlcPath: data.vlcPath || "",
        mpcPort: data.mpcPort,
        mpcPath: data.mpcPath || "",
        mpvSocket: data.mpvSocket || "",
        mpvPath: data.mpvPath || "",
        mpvArgs: "",
        iinaSocket: data.iinaSocket || "",
        iinaPath: data.iinaPath || "",
        iinaArgs: "",
    },
    discord: {
        enableRichPresence: data.enableRichPresence,
        enableAnimeRichPresence: true,
        enableMangaRichPresence: true,
        richPresenceHideSeanimeRepositoryButton: false,
        richPresenceShowAniListMediaButton: false,
        richPresenceShowAniListProfileButton: false,
        richPresenceUseMediaTitleStatus: true,
    },
    torrent: {
        defaultTorrentClient: data.defaultTorrentClient,
        qbittorrentPath: data.qbittorrentPath,
        qbittorrentHost: data.qbittorrentHost,
        qbittorrentPort: data.qbittorrentPort,
        qbittorrentPassword: data.qbittorrentPassword,
        qbittorrentUsername: data.qbittorrentUsername,
        qbittorrentTags: "",
        transmissionPath: data.transmissionPath,
        transmissionHost: data.transmissionHost,
        transmissionPort: data.transmissionPort,
        transmissionUsername: data.transmissionUsername,
        transmissionPassword: data.transmissionPassword,
        showActiveTorrentCount: false,
        hideTorrentList: false,
    },
    anilist: {
        hideAudienceScore: false,
        enableAdultContent: data.enableAdultContent,
        blurAdultContent: false,
    },
    enableTorrentStreaming: data.enableTorrentStreaming,
    enableTranscode: data.enableTranscode,
    notifications: {
        disableNotifications: false,
        disableAutoDownloaderNotifications: false,
        disableAutoScannerNotifications: false,
    },
    debridProvider: data.debridProvider,
    debridApiKey: data.debridApiKey,
})


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

export function getDefaultMpvSocket(os: string) {
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

export function getDefaultIinaSocket(os: string) {
    return "/tmp/iina_socket"
}
