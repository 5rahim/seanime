import { z } from "zod"

export const settingsSchema = z.object({
    libraryPath: z.string().min(1),
    defaultPlayer: z.string(),
    mediaPlayerHost: z.string(),
    vlcUsername: z.string().optional().default(""),
    vlcPassword: z.string().optional().default(""),
    vlcPort: z.number(),
    vlcPath: z.string().optional().default(""),
    mpcPort: z.number(),
    mpcPath: z.string().optional().default(""),
    qbittorrentPath: z.string().optional().default(""),
    qbittorrentHost: z.string(),
    qbittorrentPort: z.number(),
    qbittorrentUsername: z.string().optional().default(""),
    qbittorrentPassword: z.string().optional().default(""),
    hideAudienceScore: z.boolean().optional().default(false)
})