import { useServerMutation } from "@/api/client/requests"
import { PlaybackStartPlaylist_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function usePlaybackSyncCurrentProgress() {
    return useServerMutation<number>({
        endpoint: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackSyncCurrentProgress.endpoint,
        method: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackSyncCurrentProgress.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackSyncCurrentProgress.key],
        onSuccess: async () => {

        },
    })
}

export function usePlaybackPlayNextEpisode() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlayNextEpisode.endpoint,
        method: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlayNextEpisode.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlayNextEpisode.key],
        onSuccess: async () => {

        },
    })
}

export function usePlaybackStartPlaylist() {
    return useServerMutation<boolean, PlaybackStartPlaylist_Variables>({
        endpoint: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackStartPlaylist.endpoint,
        method: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackStartPlaylist.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackStartPlaylist.key],
        onSuccess: async () => {

        },
    })
}

export function usePlaybackCancelCurrentPlaylist() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackCancelCurrentPlaylist.endpoint,
        method: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackCancelCurrentPlaylist.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackCancelCurrentPlaylist.key],
        onSuccess: async () => {

        },
    })
}

export function usePlaybackPlaylistNext() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlaylistNext.endpoint,
        method: API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlaylistNext.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYBACK_MANAGER.PlaybackPlaylistNext.key],
        onSuccess: async () => {

        },
    })
}

