/**
 * Updates / Releases
 */

export type Update = {
    release?: Release
    type: string
}

export type LatestReleaseResponse = {
    release: Release
}

export type Release = {
    url: string
    html_url: string
    node_id: string
    tag_name: string
    name: string
    body: string
    published_at: string
    released: boolean
    version: string
    assets: ReleaseAsset[]
}

export type ReleaseAsset = {
    url: string
    id: number
    node_id: string
    name: string
    content_type: string
    uploaded: boolean
    size: number
    browser_download_url: string
}
