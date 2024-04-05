export const DEFAULT_TORRENT_PROVIDER = "animetosho"

export const DEFAULT_TORRENT_CLIENT = "qbittorrent"

export type AutoDownloaderSettings = {
    provider: string
    interval: number
    enabled: boolean
    downloadAutomatically: boolean
}

export type Settings = {
    library?: LibrarySettings
    mediaPlayer?: MediaPlayerSettings
    torrent?: TorrentSettings
    anilist?: AnilistSettings
    listSync?: ListSyncSettings
    autoDownloader?: AutoDownloaderSettings
    discord?: DiscordSettings
}

export type DiscordSettings = {
    enableRichPresence: boolean
    enableAnimeRichPresence: boolean
    enableMangaRichPresence: boolean
}

export type AnilistSettings = {
    hideAudienceScore: boolean
    enableAdultContent: boolean
    blurAdultContent: boolean
}

export type MediaPlayerSettings = {
    defaultPlayer: string
    host: string
    vlcUsername: string
    vlcPassword: string
    vlcPort: number
    vlcPath: string
    mpcPort: number
    mpcPath: string
    mpvSocket: string
    mpvPath: string
}

export type LibrarySettings = {
    libraryPath: string
    autoUpdateProgress: boolean
    disableUpdateCheck: boolean
    torrentProvider: string
    autoScan: boolean
    enableOnlinestream: boolean
    disableAnimeCardTrailers: boolean
    enableManga: boolean
}

export type TorrentSettings = {
    defaultTorrentClient: string
    qbittorrentPath: string
    qbittorrentHost: string
    qbittorrentPort: number
    qbittorrentUsername: string
    qbittorrentPassword: string
    transmissionPath: string
    transmissionHost: string
    transmissionPort: number
    transmissionUsername: string
    transmissionPassword: string
}

export type ListSyncSettings = {
    origin: string
    automatic: boolean
}

export type ThemeSettings = {
    animeEntryScreenLayout: string
    smallerEpisodeCarouselSize: boolean
    expandSidebarOnHover: boolean
    backgroundColor: string
    sidebarBackgroundColor: string
    libraryScreenBannerType: string
    libraryScreenCustomBannerImage: string
    libraryScreenCustomBannerPosition: string
    libraryScreenCustomBannerOpacity: number
    libraryScreenCustomBackgroundImage: string
    libraryScreenCustomBackgroundOpacity: number
}
