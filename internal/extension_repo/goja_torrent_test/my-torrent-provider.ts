/// <reference path="./anime-torrent-provider.d.ts" />

class Provider {

    api = "https://nyaa.si/?page=rss"

    getSettings(): AnimeProviderSettings {
        return {
            canSmartSearch: false,
            smartSearchFilters: [],
            supportsAdult: false,
            type: "main",
        }
    }

    async fetchTorrents(url: string): Promise<NyaaTorrent[]> {

        const furl = `${this.api}&q=+${encodeURIComponent(url)}&c=1_3`

        try {
            console.log(furl)
            const response = await fetch(furl)

            if (!response.ok) {
                throw new Error(`Failed to fetch torrents, ${response.statusText}`)
            }

            const xmlText = await response.text()
            const torrents = this.parseXML(xmlText)
            console.log(torrents)

            return torrents
        }
        catch (error) {
            throw new Error(`Error fetching torrents: ${error}`)
        }
    }

    async search(opts: AnimeSearchOptions): Promise<AnimeTorrent[]> {
        console.log(opts)
        const torrents = await this.fetchTorrents(opts.query)
        return torrents.map(t => this.toAnimeTorrent(t))
    }

    toAnimeTorrent(torrent: NyaaTorrent): AnimeTorrent {
        return {
            name: torrent.title,
            date: new Date(torrent.timestamp * 1000).toISOString(),
            size: torrent.total_size,
            formattedSize: torrent.size,
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
            confirmed: false,
        }
    }

    async smartSearch(opts: AnimeSmartSearchOptions): Promise<AnimeTorrent[]> {
        const ret: AnimeTorrent[] = []
        return ret
    }

    private parseXML(xmlText: string): NyaaTorrent[] {
        const torrents: NyaaTorrent[] = []

        // Helper to extract content between XML tags
        const getTagContent = (xml: string, tag: string): string => {
            const regex = new RegExp(`<${tag}[^>]*>([^<]*)</${tag}>`)
            const match = xml.match(regex)
            return match ? match[1].trim() : ""
        }

        // Helper to extract content from nyaa namespace tags
        const getNyaaTagContent = (xml: string, tag: string): string => {
            const regex = new RegExp(`<nyaa:${tag}[^>]*>([^<]*)</nyaa:${tag}>`)
            const match = xml.match(regex)
            return match ? match[1].trim() : ""
        }

        // Split XML into items
        const itemRegex = /<item>([\s\S]*?)<\/item>/g
        let match

        let id = 1
        while ((match = itemRegex.exec(xmlText)) !== null) {
            const itemXml = match[1]

            const title = getTagContent(itemXml, "title")
            const link = getTagContent(itemXml, "link")
            const pubDate = getTagContent(itemXml, "pubDate")
            const seeders = parseInt(getNyaaTagContent(itemXml, "seeders")) || 0
            const leechers = parseInt(getNyaaTagContent(itemXml, "leechers")) || 0
            const downloads = parseInt(getNyaaTagContent(itemXml, "downloads")) || 0
            const infoHash = getNyaaTagContent(itemXml, "infoHash")
            const size = getNyaaTagContent(itemXml, "size")

            // Convert size string (e.g., "571.3 MiB") to bytes
            const sizeInBytes = (() => {
                const match = size.match(/^([\d.]+)\s*([KMGT]iB)$/)
                if (!match) return 0
                const [, num, unit] = match
                const multipliers: { [key: string]: number } = {
                    "KiB": 1024,
                    "MiB": 1024 * 1024,
                    "GiB": 1024 * 1024 * 1024,
                    "TiB": 1024 * 1024 * 1024 * 1024,
                }
                return Math.round(parseFloat(num) * multipliers[unit])
            })()

            const torrent: NyaaTorrent = {
                id: id++,
                title,
                link,
                timestamp: Math.floor(new Date(pubDate).getTime() / 1000),
                status: "success",
                torrent_url: link,
                info_hash: infoHash,
                magnet_uri: `magnet:?xt=urn:btih:${infoHash}`,
                seeders,
                leechers,
                torrent_download_count: downloads,
                total_size: sizeInBytes,
                size,
                num_files: 1,
                anidb_aid: 0,
                anidb_eid: 0,
                anidb_fid: 0,
                article_url: link,
                article_title: title,
                website_url: "https://nyaa.si",
            }

            torrents.push(torrent)
        }

        return torrents
    }
}

type NyaaTorrent = {
    id: number
    title: string
    link: string
    timestamp: number
    status: string
    size: string
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
