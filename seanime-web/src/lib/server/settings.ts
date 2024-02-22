import { z } from "zod"

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
                    return "/Applications/Client.app/Contents/MacOS/qBittorrent" // Default path for Client on macOS
                default:
                    return "C:/Program Files/qBittorrent/qbittorrent.exe"
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

export const settingsSchema = z.object({
    libraryPath: z.string().min(1),
    defaultPlayer: z.string(),
    torrentProvider: z.string(),
    mediaPlayerHost: z.string(),
    vlcUsername: z.string().optional().default(""),
    vlcPassword: z.string().optional().default(""),
    vlcPort: z.number(),
    vlcPath: z.string().optional().default(""),
    mpcPort: z.number(),
    mpcPath: z.string().optional().default(""),
    mpvSocket: z.string().optional().default(""),
    mpvPath: z.string().optional().default(""),
    qbittorrentPath: z.string().optional().default(""),
    qbittorrentHost: z.string(),
    qbittorrentPort: z.number(),
    qbittorrentUsername: z.string().optional().default(""),
    qbittorrentPassword: z.string().optional().default(""),
    hideAudienceScore: z.boolean().optional().default(false),
    autoUpdateProgress: z.boolean().optional().default(false),
    disableUpdateCheck: z.boolean().optional().default(false),
})
