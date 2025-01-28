export type PlaybackManager_PlaybackState = {
    filename: string
    mediaTitle: string
    mediaTotalEpisodes: number
    episodeNumber: number
    completionPercentage: number
    canPlayNext: boolean
    progressUpdated: boolean
    mediaId: number
    mediaCoverImage: string
}

export type PlaybackManager_PlaylistState = {
    current: PlaybackManager_PlaylistStateItem | null
    next: PlaybackManager_PlaylistStateItem | null
    remaining: number
}

export type PlaybackManager_PlaylistStateItem = {
    name: string
    mediaImage: string
}
