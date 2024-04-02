import { MediaEntryEpisode } from "@/app/(main)/(library)/_lib/anime-library.types"

export type TorrentSearchData = {
    previews: TorrentPreview[]
    torrents: AnimeTorrent[]
}

export type AnimeTorrent = {
    name: string
    date: string
    size: number
    formattedSize: string
    seeders: number
    leechers: number
    downloadCount: number
    link: string
    downloadUrl: string
    infoHash: string
    resolution?: string
    isBatch: boolean
    episodeNumber?: number
    releaseGroup?: string
    provider: string
    isBestRelease: boolean
}

export type TorrentPreview = {
    torrent: AnimeTorrent
    episode: MediaEntryEpisode | null
}
