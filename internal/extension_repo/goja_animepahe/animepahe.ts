/// <reference path="../goja_onlinestream_test/onlinestream-provider.d.ts" />
/// <reference path="../goja_plugin_types/core.d.ts" />

type EpisodeData = {
    id: number; episode: number; title: string; snapshot: string; filler: number; session: string; created_at?: string
}

type AnimeData = {
    id: number; title: string; type: string; year: number; poster: string; session: string
}

class Provider {

    api = "https://animepahe.ru"
    headers = { Referer: "https://kwik.si" }

    getSettings(): Settings {
        return {
            episodeServers: ["kwik"],
            supportsDub: false,
        }
    }

    async search(opts: SearchOptions): Promise<SearchResult[]> {
        const req = await fetch(`${this.api}/api?m=search&q=${encodeURIComponent(opts.query)}`, {
            headers: {
                Cookie: "__ddg1_=;__ddg2_=;",
            },
        })

        if (!req.ok) {
            return []
        }
        const data = (await req.json()) as { data: AnimeData[] }
        const results: SearchResult[] = []

        if (!data?.data) {
            return []
        }

        data.data.map((item: AnimeData) => {
            results.push({
                subOrDub: "sub",
                id: item.session,
                title: item.title,
                url: "",
            })
        })

        return results
    }

    async findEpisodes(id: string): Promise<EpisodeDetails[]> {
        let episodes: EpisodeDetails[] = []

        const req =
            await fetch(
                `${this.api}${id.includes("-") ? `/anime/${id}` : `/a/${id}`}`,
                {
                    headers: {
                        Cookie: "__ddg1_=;__ddg2_=;",
                    },
                },
            )

        const html = await req.text()


        function pushData(data: EpisodeData[]) {
            for (const item of data) {
                episodes.push({
                    id: item.session + "$" + id,
                    number: item.episode,
                    title: item.title && item.title.length > 0 ? item.title : "Episode " + item.episode,
                    url: req.url,
                })
            }
        }

        const $ = LoadDoc(html)

        const tempId = $("head > meta[property='og:url']").attr("content")!.split("/").pop()!

        const { last_page, data } = (await (
            await fetch(`${this.api}/api?m=release&id=${tempId}&sort=episode_asc&page=1`, {
                headers: {
                    Cookie: "__ddg1_=;__ddg2_=;",
                },
            })
        ).json()) as {
            last_page: number;
            data: EpisodeData[]
        }

        pushData(data)

        const pageNumbers = Array.from({ length: last_page - 1 }, (_, i) => i + 2)

        const promises = pageNumbers.map((pageNumber) =>
            fetch(`${this.api}/api?m=release&id=${tempId}&sort=episode_asc&page=${pageNumber}`, {
                headers: {
                    Cookie: "__ddg1_=;__ddg2_=;",
                },
            }).then((res) => res.json()),
        )
        const results = (await Promise.all(promises)) as {
            data: EpisodeData[]
        }[]

        results.forEach((showData) => {
            for (const data of showData.data) {
                if (data) {
                    pushData([data])
                }
            }
        });
        (data as any[]).sort((a, b) => a.number - b.number)

        if (episodes.length === 0) {
            throw new Error("No episodes found.")
        }


        const lowest = episodes[0].number
        if (lowest > 1) {
            for (let i = 0; i < episodes.length; i++) {
                episodes[i].number = episodes[i].number - lowest + 1
            }
        }

        // Remove decimal episode numbers
        episodes = episodes.filter((episode) => Number.isInteger(episode.number))

        // for (let i = 0; i < episodes.length; i++) {
        //     // If an episode number is a decimal, round it up to the nearest whole number
        //     if (Number.isInteger(episodes[i].number)) {
        //         continue
        //     }
        //     const original = episodes[i].number
        //     episodes[i].number = Math.floor(episodes[i].number)
        //     episodes[i].title = `Episode ${episodes[i].number} [{${original}}]`
        // }

        return episodes
    }

    async findEpisodeServer(episode: EpisodeDetails, _server: string): Promise<EpisodeServer> {
        const episodeId = episode.id.split("$")[0]
        const animeId = episode.id.split("$")[1]

        console.log(`${this.api}/play/${animeId}/${episodeId}`)

        const req = await fetch(
            `${this.api}/play/${animeId}/${episodeId}`,
            {
                headers: {
                    Cookie: "__ddg1_=;__ddg2_=;",
                },
            },
        )

        const html = await req.text()

        const regex = /https:\/\/kwik\.si\/e\/\w+/g
        const matches = html.match(regex)

        if (matches === null) {
            throw new Error("Failed to fetch episode server.")
        }

        const $ = LoadDoc(html)

        const result: EpisodeServer = {
            videoSources: [],
            headers: this.headers ?? {},
            server: "kwik",
        }

        $("button[data-src]").each(async (_, el) => {
            let videoSource: VideoSource = {
                url: "",
                type: "m3u8",
                quality: "",
                subtitles: [],
            }

            videoSource.url = el.data("src")!
            if (!videoSource.url) {
                return
            }

            const fansub = el.data("fansub")!
            const quality = el.data("resolution")!

            videoSource.quality = `${quality}p - ${fansub}`

            if (el.data("audio") === "eng") {
                videoSource.quality += " (Eng)"
            }

            if (videoSource.url === matches[0]) {
                videoSource.quality += " (default)"
            }

            result.videoSources.push(videoSource)
        })

        const queries = result.videoSources.map(async (videoSource) => {
            try {
                const src_req = await fetch(videoSource.url, {
                    headers: {
                        Referer: this.headers.Referer,
                        "user-agent":
                            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.56",
                    },
                })

                const src_html = await src_req.text()

                const scripts = src_html.match(/eval\(f.+?\}\)\)/g)
                if (!scripts) {
                    return
                }

                for (const _script of scripts) {
                    const scriptMatch = _script.match(/eval(.+)/)
                    if (!scriptMatch || !scriptMatch[1]) {
                        continue
                    }

                    try {
                        const decoded = eval(scriptMatch[1])
                        const link = decoded.match(/source='(.+?)'/)
                        if (!link || !link[1]) {
                            continue
                        }

                        videoSource.url = link[1]

                    }
                    catch (e) {
                        console.error("Failed to extract kwik link", e)
                    }

                }

            }
            catch (e) {
                console.error("Failed to fetch kwik link", e)
            }
        })

        await Promise.all(queries)

        return result
    }
}

