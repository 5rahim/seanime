import { OfflineAssetMap } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { __DEV_SERVER_PORT } from "@/lib/anilist/config"

export function offline_getAssetUrl(url: string | null | undefined, assetMap: OfflineAssetMap | undefined) {
    if (!url) return undefined
    const filename = assetMap?.[url]
    if (!filename) return "/no-cover.png"
    return process.env.NODE_ENV === "development"
        ? `http://${window?.location?.hostname}:${__DEV_SERVER_PORT}/offline-assets/${filename}`
        : `http://${window?.location?.host}/offline-assets/${filename}`
}
