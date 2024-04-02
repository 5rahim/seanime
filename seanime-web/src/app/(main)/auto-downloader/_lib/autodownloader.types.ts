export type AutoDownloaderRule = {
    dbId: number
    enabled: boolean
    mediaId: number
    releaseGroups: string[]
    resolutions: string[]
    comparisonTitle: string
    titleComparisonType: string
    episodeType: string
    episodeNumbers?: number[]
    destination: string
}

export type AutoDownloaderItem = {
    id: number
    createdAt: string
    updatedAt: string
    ruleId: number
    mediaId: number
    episode: number
    link: string
    hash: string
    magnet: string
    torrentName: string
    downloaded: boolean
}
