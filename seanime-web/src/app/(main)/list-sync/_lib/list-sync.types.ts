export const enum ListSyncOrigin {
    ANILIST = "anilist",
    MAL = "mal"
}

export const enum ListSyncAnimeDiffKind {
    MISSING_IN_ORIGIN = "missing_in_origin",
    MISSING_IN_TARGET = "missing_in_target",
    METADATA = "metadata",
}

export const enum ListSyncAnimeMetadataDiffKind {
    SCORE = "score",
    PROGRESS = "progress",
    STATUS = "status",
}

export type ListSyncAnimeEntry = {
    source: ListSyncOrigin
    sourceID: number
    malID: number
    displayTitle: string
    url: string
    progress: number
    totalEpisodes: number
    status: string
    image: string
    score: string
}

export type ListSyncAnimeDiff = {
    id: string
    targetSource: string
    originEntry?: ListSyncAnimeEntry
    targetEntry?: ListSyncAnimeEntry
    kind: ListSyncAnimeDiffKind
    metadataDiffKinds: ListSyncAnimeMetadataDiffKind[]
}
