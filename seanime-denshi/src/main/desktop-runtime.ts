import { net } from "electron"
import http from "node:http"
import type { AddressInfo } from "node:net"
import { log } from "./logging"

const LOCAL_EMBED_HOST = "127.0.0.1"
const DESKTOP_SERVER_HOST = "127.0.0.1"
const DESKTOP_SERVER_DEFAULT_PORT = 43211
const DESKTOP_SERVER_DEV_PORT = 43000

export const allowedWebviewOrigins = new Set<string>()

let localServerPort: number | undefined

export function getLocalServerPort(): number | undefined {
    return localServerPort
}

export function isAllowedLocalEmbedURL(rawURL: string): boolean {
    if (!localServerPort) return false

    try {
        const parsed = new URL(rawURL)
        return parsed.protocol === "http:"
            && parsed.hostname === LOCAL_EMBED_HOST
            && parsed.port === String(localServerPort)
            && parsed.pathname.startsWith("/player/")
    }
    catch {
        return false
    }
}

export function normalizeUpdateFeedURL(candidate: string, fallbackURL: string): string {
    try {
        const parsed = new URL(candidate)
        if (parsed.protocol !== "https:" || !parsed.host) {
            throw new Error("update feeds must use https")
        }
        return parsed.toString()
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error)
        log.warn(`[Denshi] Ignoring update feed URL ${candidate}: ${message}`)
        return fallbackURL
    }
}

export function getDesktopServerPort(): number {
    if (process.env.NODE_ENV === "development") return DESKTOP_SERVER_DEV_PORT

    const envPort = Number.parseInt(process.env.SEANIME_SERVER_PORT || "", 10)
    if (Number.isInteger(envPort) && envPort > 0) return envPort
    return DESKTOP_SERVER_DEFAULT_PORT
}

export function getDesktopServerBaseUrl(): string {
    return `http://${DESKTOP_SERVER_HOST}:${getDesktopServerPort()}`
}

export async function isDesktopServerReachable(): Promise<boolean> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 1000)
    try {
        const response = await net.fetch(`${getDesktopServerBaseUrl()}/api/v1/status`, {
            signal: controller.signal,
        })
        return response.ok
    }
    catch {
        return false
    }
    finally {
        clearTimeout(timeoutId)
    }
}

export function startLocalServer(): number {
    const server = http.createServer((req, res) => {
        const match = req.url?.match(/^\/player\/([\w-]+)/)
        if (!match) {
            res.writeHead(404)
            res.end("Not found")
            return
        }

        const id = match[1]
        let url = `https://www.youtube-nocookie.com/embed/${id}?autoplay=1&enablejsapi=1&autoplay=1&playsinline=1&modestbranding=1&rel=0e`
        if (id.startsWith("compact_")) {
            const videoID = id.substring(8)
            url = `https://www.youtube-nocookie.com/embed/${videoID}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&vq=medium&playlist=${videoID}&cc_lang_pref=ja&enablejsapi=true`
        }
        if (id.startsWith("banner_")) {
            const videoID = id.substring(7)
            url = `https://www.youtube-nocookie.com/embed/${videoID}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&playlist=${videoID}&cc_lang_pref=ja&enablejsapi=true`
        }

        res.writeHead(200, { "Content-Type": "text/html" })
        res.end(`<!DOCTYPE html>
<html lang="en">
<head><style>html,body{margin:0;height:100%;background:black}iframe{position:absolute;inset:0;width:100%;height:100%;border:0}</style></head>
<body><iframe src="${url}" allow="autoplay; encrypted-media; picture-in-picture" allowfullscreen></iframe></body>
</html>`)
    })

    server.listen(0)
    localServerPort = (server.address() as AddressInfo).port
    console.log(`Local server running at http://${LOCAL_EMBED_HOST}:${localServerPort}`)
    return localServerPort
}
