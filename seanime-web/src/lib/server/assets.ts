import { getServerBaseUrl } from "@/api/client/server-url"

export function getAssetUrl(path: string) {
    let p = path.replaceAll("\\", "/")

    if (p.startsWith("/")) {
        p = p.substring(1)
    }

    p = encodeURIComponent(p).replace(/\(/g, "%28").replace(/\)/g, "%29")

    return `${getServerBaseUrl()}/assets/${p}`
}
