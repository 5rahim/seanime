import { __DEV_SERVER_PORT } from "@/lib/server/config"

export function getAssetUrl(path: string) {
    let p = path.replaceAll("\\", "/")

    if (p.startsWith("/")) {
        p = p.substring(1)
    }

    p = encodeURIComponent(p).replace(/\(/g, "%28").replace(/\)/g, "%29")

    return process.env.NODE_ENV === "development"
        ? `http://${window?.location?.hostname}:${__DEV_SERVER_PORT}/assets/${p}`
        : `${window?.location?.protocol}//${window?.location?.host}/assets/${p}`
}
