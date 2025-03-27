/// <reference path="./onlinestream-provider.d.ts" />
/// <reference path="../../goja/goja_bindings/js/core.d.ts" />

class ProviderN {

    api = "https://anitaku.to"
    ajaxURL = "https://ajax.gogocdn.net"

    getSettings(): Settings {
        return {
            episodeServers: ["gogocdn", "vidstreaming", "streamsb"],
            supportsDub: true,
        }
    }

    async search(opts: SearchOptions): Promise<SearchResult[]> {
        const request = await fetch(`${this.api}/search.html?keyword=${encodeURIComponent(opts.query)}`)
        if (!request.ok) {
            return []
        }
        const data = await request.text()
        const results: SearchResult[] = []

        const $ = LoadDoc(data)

        $("ul.items > li").each((_, el) => {
            const title = el.find("p.name a").text().trim()
            const id = el.find("div.img a").attr("href")
            if (!id) {
                return
            }

            results.push({
                id: id,
                title: title,
                url: id,
                subOrDub: "sub",
            })
        })

        return results
    }

    async findEpisodes(id: string): Promise<EpisodeDetails[]> {
        const episodes: EpisodeDetails[] = []

        const data = await (await fetch(`${this.api}${id}`)).text()

        const $ = LoadDoc(data)

        const epStart = $("#episode_page > li").first().find("a").attr("ep_start")
        const epEnd = $("#episode_page > li").last().find("a").attr("ep_end")
        const movieId = $("#movie_id").attr("value")
        const alias = $("#alias_anime").attr("value")

        const req = await (await fetch(`${this.ajaxURL}/ajax/load-list-episode?ep_start=${epStart}&ep_end=${epEnd}&id=${movieId}&default_ep=${0}&alias=${alias}`)).text()

        const $$ = LoadDoc(req)

        $$("#episode_related > li").each((i, el) => {
            episodes?.push({
                id: el.find("a").attr("href")?.trim() ?? "",
                url: el.find("a").attr("href")?.trim() ?? "",
                number: parseFloat(el.find(`div.name`).text().replace("EP ", "")),
                title: el.find(`div.name`).text(),
            })
        })

        return episodes.reverse()
    }

    async findEpisodeServer(episode: EpisodeDetails, _server: string): Promise<EpisodeServer> {
        let server = "gogocdn"
        if (_server !== "default") {
            server = _server
        }

        const episodeServer: EpisodeServer = {
            server: server,
            headers: {},
            videoSources: [],
        }

        if (episode.id.startsWith("http")) {
            const serverURL = episode.id
            try {
                const es = await new Extractor(serverURL, episodeServer).extract(server)
                if (es) {
                    return es
                }
            }
            catch (e) {
                console.error(e)
                return episodeServer
            }
            return episodeServer
        }

        const data = await (await fetch(`${this.api}${episode.id}`)).text()

        const $ = LoadDoc(data)

        let serverURL: string

        switch (server) {
            case "gogocdn":
                serverURL = `${$("#load_anime > div > div > iframe").attr("src")}`
                break
            case "vidstreaming":
                serverURL = `${$("div.anime_video_body > div.anime_muti_link > ul > li.vidcdn > a").attr("data-video")}`
                break
            case "streamsb":
                serverURL = $("div.anime_video_body > div.anime_muti_link > ul > li.streamsb > a").attr("data-video")!
                break
            default:
                serverURL = `${$("#load_anime > div > div > iframe").attr("src")}`
                break
        }

        episode.id = serverURL
        return await this.findEpisodeServer(episode, server)
    }

}


class Extractor {
    private url: string
    private result: EpisodeServer

    constructor(url: string, result: EpisodeServer) {
        this.url = url
        this.result = result
    }

    async extract(server: string): Promise<EpisodeServer | undefined> {
        try {
            switch (server) {
                case "gogocdn":
                    console.log("GogoCDN extraction")
                    return await this.extractGogoCDN(this.url, this.result)
                case "vidstreaming":
                    return await this.extractGogoCDN(this.url, this.result)
                default:
                    return undefined
            }
        }
        catch (e) {
            console.error(e)
            return undefined
        }
    }


    public async extractGogoCDN(url: string, result: EpisodeServer): Promise<EpisodeServer> {
        const keys = {
            key: CryptoJS.enc.Utf8.parse("37911490979715163134003223491201"),
            secondKey: CryptoJS.enc.Utf8.parse("54674138327930866480207815084989"),
            iv: CryptoJS.enc.Utf8.parse("3134003223491201"),
        }

        function generateEncryptedAjaxParams(id: string) {
            const encryptedKey = CryptoJS.AES.encrypt(id, keys.key, {
                iv: keys.iv,
            })

            const scriptValue = $("script[data-name='episode']").data("value")!

            const decryptedToken = CryptoJS.AES.decrypt(scriptValue, keys.key, {
                iv: keys.iv,
            }).toString(CryptoJS.enc.Utf8)

            return `id=${encryptedKey.toString(CryptoJS.enc.Base64)}&alias=${id}&${decryptedToken}`
        }

        function decryptAjaxData(encryptedData: string) {

            const decryptedData = CryptoJS.AES.decrypt(encryptedData, keys.secondKey, {
                iv: keys.iv,
            }).toString(CryptoJS.enc.Utf8)

            return JSON.parse(decryptedData)
        }

        const req = await fetch(url)

        const $ = LoadDoc(await req.text())

        const encryptedParams = generateEncryptedAjaxParams(new URL(url).searchParams.get("id") ?? "")

        const xmlHttpUrl = `${new URL(url).protocol}//${new URL(url).hostname}/encrypt-ajax.php?${encryptedParams}`

        const encryptedData = await fetch(xmlHttpUrl, {
            headers: {
                "X-Requested-With": "XMLHttpRequest",
            },
        })


        const decryptedData = await decryptAjaxData(((await encryptedData.json()) as { data: any })?.data)
        if (!decryptedData.source) throw new Error("No source found. Try a different server.")

        if (decryptedData.source[0].file.includes(".m3u8")) {
            const resResult = await fetch(decryptedData.source[0].file.toString())
            const resolutions = (await resResult.text()).match(/(RESOLUTION=)(.*)(\s*?)(\s*.*)/g)

            resolutions?.forEach((res: string) => {
                const index = decryptedData.source[0].file.lastIndexOf("/")
                const quality = res.split("\n")[0].split("x")[1].split(",")[0]
                const url = decryptedData.source[0].file.slice(0, index)

                result.videoSources.push({
                    url: url + "/" + res.split("\n")[1],
                    quality: quality + "p",
                    subtitles: [],
                    type: "m3u8",
                })
            })

            decryptedData.source.forEach((source: any) => {
                result.videoSources.push({
                    url: source.file,
                    quality: "default",
                    subtitles: [],
                    type: "m3u8",
                })
            })
        } else {
            decryptedData.source.forEach((source: any) => {
                result.videoSources.push({
                    url: source.file,
                    quality: source.label.split(" ")[0] + "p",
                    subtitles: [],
                    type: "m3u8",
                })
            })

            decryptedData.source_bk.forEach((source: any) => {
                result.videoSources.push({
                    url: source.file,
                    quality: "backup",
                    subtitles: [],
                    type: "m3u8",
                })
            })
        }

        return result
    }
}
