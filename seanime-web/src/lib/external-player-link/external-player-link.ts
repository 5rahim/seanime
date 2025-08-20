import { getServerBaseUrl } from "@/api/client/server-url"
import { logger } from "@/lib/helpers/debug"
import { __isDesktop__ } from "@/types/constants"

const log = logger("EXTERNAL PLAYER LINK")

export class ExternalPlayerLink {

    private _playerLink = ""
    private _urlToSend = ""
    private _episodeNumber: number | null = null
    private _mediaTitle: string | null = null

    constructor(playerLink: string) {
        this._playerLink = playerLink
    }

    setUrl(url: string) {
        this._urlToSend = url
    }

    setEpisodeNumber(ep: number | undefined) {
        this._episodeNumber = ep ?? null
    }

    setMediaTitle(title: string | undefined) {
        this._mediaTitle = title ?? null
    }

    async to(props: { endpoint: string, onTokenQueryParam?: () => Promise<string> }) {
        let url = getServerBaseUrl() + props.endpoint
        if (props.onTokenQueryParam) {
            url += await props.onTokenQueryParam()
        }
        logger("MEDIALINKS").info("Formatted URL to send", url)
        this._urlToSend = url
    }

    getFullUrl() {
        const urlToSend = this._getUrlToSend()
        log.info("Sending URL to external player", urlToSend)
        return this._formatFinalUrl(urlToSend)
    }

    private _getUrlToSend() {
        if (this._playerLink.includes("?")) {
            return encodeURIComponent(this._urlToSend)
        }
        return this._urlToSend
    }

    private _cleanTitle(title: string) {
        return title.replace(/[\\/:*?"<>|]/g, "")
    }

    private _formatFinalUrl(url: string): string {
        let link = this._playerLink
        link = link.replace("{scheme}", window.location.protocol.replace(":", ""))
        link = link.replace("{host}", window.location.host)
        link = link.replace("{hostname}", window.location.hostname)
        link = link.replace("{mediaTitle}", this._cleanTitle(this._mediaTitle ?? ""))
        link = link.replace("{episodeNumber}", this._episodeNumber?.toString?.() ?? "")
        link = link.replace("{mime}", "video/webm")
        if (link.includes("{formattedTitle}")) {
            let title = this._mediaTitle ?? ""
            if (this._episodeNumber !== null && !!title.length) {
                title = `Episode ${this._episodeNumber} - ${title}`
            }
            link = link.replace("{formattedTitle}", this._cleanTitle(title ?? ""))
        }
        log.info("Formatted external player link", link)
        if (__isDesktop__) {
            let retUrl = link.replace("{url}", url)
            if (link.startsWith("intent://")) {
                // e.g. "intent://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
                retUrl = retUrl.replace("intent://http://", "intent://").replace("intent://https://", "intent://")
            }
            return retUrl
        }

        // e.g. "mpv://http://localhost:43214/stream/..."
        // e.g. "intent://http://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
        let retUrl = link.replace("{url}", url)
            .replace("127.0.0.1", window.location.hostname)
            .replace("localhost", window.location.hostname)


        if (link.startsWith("intent://")) {
            // e.g. "intent://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
            retUrl = retUrl.replace("intent://http://", "intent://").replace("intent://https://", "intent://")
        }

        return retUrl
    }
}
