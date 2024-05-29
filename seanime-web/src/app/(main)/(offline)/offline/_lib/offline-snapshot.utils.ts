import { Offline_AssetMapImageMap } from "@/api/generated/types"
import { __DEV_SERVER_PORT } from "@/lib/server/config"

export function offline_getAssetUrl(url: string | null | undefined, assetMap: Offline_AssetMapImageMap | undefined) {
    if (!url) return undefined
    const filename = assetMap?.[url]
    if (!filename) return "/no-cover.png"
    return process.env.NODE_ENV === "development"
        ? `http://${window?.location?.hostname}:${__DEV_SERVER_PORT}/offline-assets/${filename}`
        : `${window?.location.protocol}//${window?.location?.host}/offline-assets/${filename}`
}
