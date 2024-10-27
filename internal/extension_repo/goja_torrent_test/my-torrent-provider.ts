/// <reference path="./anime-torrent-provider.d.ts" />

class Provider {

    api = "{{api}}"
    withSmartSearch = "{{withSmartSearch}}"
    type = "{{type}}"

    getSettings(): AnimeProviderSettings {
        return {
            canSmartSearch: this.withSmartSearch === "true",
            smartSearchFilters: ["batch", "episodeNumber", "resolution"],
            supportsAdult: false,
            type: this.type as AnimeProviderType,
        }
    }

    async fetchTorrents(url: string): Promise<ToshoTorrent[]> {
        const furl = `${this.api}${url}`

        try {
            const response = await fetch(furl)

            if (!response.ok) {
                throw new Error(`Failed to fetch torrents, ${response.statusText}`)
            }

            const torrents: ToshoTorrent[] = await response.json()

            return torrents.map(t => {
                if (t.seeders > 30000) {
                    t.seeders = 0
                }
                if (t.leechers > 30000) {
                    t.leechers = 0
                }
                return t
            })
        }
        catch (error) {
            throw new Error(`Error fetching torrents: ${error}`)
        }
    }

    async search(opts: AnimeSearchOptions): Promise<AnimeTorrent[]> {
        const query = `?q=${encodeURIComponent(opts.query)}&only_tor=1`
        console.log(query)
        const torrents = await this.fetchTorrents(query)
        return torrents.map(t => this.toAnimeTorrent(t))
    }

    async smartSearch(opts: AnimeSmartSearchOptions): Promise<AnimeTorrent[]> {
        const ret: AnimeTorrent[] = []

        if (opts.batch) {
            if (!opts.anidbAID) return []

            let torrents = await this.searchByAID(opts.anidbAID, opts.resolution)

            if (!(opts.media.format == "MOVIE" || opts.media.episodeCount == 1)) {
                torrents = torrents.filter(t => t.num_files > 1)
            }

            for (const torrent of torrents) {
                const t = this.toAnimeTorrent(torrent)
                t.isBatch = true
                ret.push()
            }

            return ret
        }

        if (!opts.anidbEID) return []

        const torrents = await this.searchByEID(opts.anidbEID, opts.resolution)

        for (const torrent of torrents) {
            ret.push(this.toAnimeTorrent(torrent))
        }

        return ret
    }

    async getTorrentInfoHash(torrent: AnimeTorrent): Promise<string> {
        return torrent.infoHash || ""
    }

    async getTorrentMagnetLink(torrent: AnimeTorrent): Promise<string> {
        return torrent.magnetLink || ""
    }

    async getLatest(): Promise<AnimeTorrent[]> {
        const query = `?q=&only_tor=1`
        const torrents = await this.fetchTorrents(query)
        return torrents.map(t => this.toAnimeTorrent(t))
    }

    async searchByAID(aid: number, quality: string): Promise<ToshoTorrent[]> {
        const q = encodeURIComponent(this.formatCommonQuery(quality))
        const query = `?qx=1&order=size-d&aid=${aid}&q=${q}`
        return this.fetchTorrents(query)
    }

    async searchByEID(eid: number, quality: string): Promise<ToshoTorrent[]> {
        const q = encodeURIComponent(this.formatCommonQuery(quality))
        const query = `?qx=1&eid=${eid}&q=${q}`
        return this.fetchTorrents(query)
    }


    formatCommonQuery(quality: string): string {
        if (quality === "") {
            return ""
        }

        quality = quality.replace(/p$/, "")

        const resolutions = ["480", "540", "720", "1080"]

        const others = resolutions.filter(r => r !== quality)
        const othersStrs = others.map(r => `!"${r}"`)

        return `("${quality}" ${othersStrs.join(" ")})`
    }

    toAnimeTorrent(torrent: ToshoTorrent): AnimeTorrent {
        return {
            name: torrent.title,
            date: new Date(torrent.timestamp * 1000).toISOString(),
            size: torrent.total_size,
            formattedSize: "",
            seeders: torrent.seeders,
            leechers: torrent.leechers,
            downloadCount: torrent.torrent_download_count,
            link: torrent.link,
            downloadUrl: torrent.torrent_url,
            magnetLink: torrent.magnet_uri,
            infoHash: torrent.info_hash,
            resolution: "",
            isBatch: false,
            isBestRelease: false,
            confirmed: true,
        }
    }
}

type ToshoTorrent = {
    id: number
    title: string
    link: string
    timestamp: number
    status: string
    tosho_id?: number
    nyaa_id?: number
    nyaa_subdom?: any
    anidex_id?: number
    torrent_url: string
    info_hash: string
    info_hash_v2?: string
    magnet_uri: string
    seeders: number
    leechers: number
    torrent_download_count: number
    tracker_updated?: any
    nzb_url?: string
    total_size: number
    num_files: number
    anidb_aid: number
    anidb_eid: number
    anidb_fid: number
    article_url: string
    article_title: string
    website_url: string
}
