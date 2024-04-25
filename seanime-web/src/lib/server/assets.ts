import { __DEV_SERVER_PORT } from "@/lib/server/config"

export function getAssetUrl(path: string) {
    let p = path.replaceAll("\\", "/")

    if (p.startsWith("/")) {
        p = p.substring(1)
    }

    return process.env.NODE_ENV === "development"
        ? `http://${window?.location?.hostname}:${__DEV_SERVER_PORT}/assets/${p}`
        : `http://${window?.location?.host}/assets/${p}`
}
