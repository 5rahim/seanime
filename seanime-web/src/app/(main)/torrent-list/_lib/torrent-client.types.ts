export type TorrentClientTorrent = {
    name: string
    hash: string
    seeds: number
    upSpeed: string
    downSpeed: string
    progress: number
    size: string
    eta: string
    status: "downloading" | "paused" | "seeding"
    contentPath: string
}

export type TorrentClientTorrentActionProps = { hash: string, action: "pause" | "resume" | "remove" | "open", dir: string }
