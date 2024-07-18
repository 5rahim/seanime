import { getServerBaseUrl } from "@/api/client/server-url"
import { Offline_AssetMapImageMap } from "@/api/generated/types"

export function offline_getAssetUrl(url: string | null | undefined, assetMap: Offline_AssetMapImageMap | undefined) {
    if (!url) return undefined
    const filename = assetMap?.[url]
    if (!filename) return "/no-cover.png"
    return `${getServerBaseUrl()}/offline-assets/${filename}`
}
