import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { SyncAddMedia_Variables, SyncRemoveMedia_Variables, SyncSetHasLocalChanges_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Sync_QueueState, Sync_TrackedMediaItem } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useSyncGetTrackedMediaItems() {
    return useServerQuery<Array<Sync_TrackedMediaItem>>({
        endpoint: API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.endpoint,
        method: API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.methods[0],
        queryKey: [API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.key],
        enabled: true,
        gcTime: 0,
    })
}

export function useSyncAddMedia() {
    const qc = useQueryClient()
    return useServerMutation<boolean, SyncAddMedia_Variables>({
        endpoint: API_ENDPOINTS.SYNC.SyncAddMedia.endpoint,
        method: API_ENDPOINTS.SYNC.SyncAddMedia.methods[0],
        mutationKey: [API_ENDPOINTS.SYNC.SyncAddMedia.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetQueueState.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetIsMediaTracked.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetLocalStorageSize] })
            toast.success("Added media for offline syncing")
        },
    })
}

export function useSyncRemoveMedia() {
    const qc = useQueryClient()
    return useServerMutation<boolean, SyncRemoveMedia_Variables>({
        endpoint: API_ENDPOINTS.SYNC.SyncRemoveMedia.endpoint,
        method: API_ENDPOINTS.SYNC.SyncRemoveMedia.methods[0],
        mutationKey: [API_ENDPOINTS.SYNC.SyncRemoveMedia.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetQueueState.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetIsMediaTracked.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetLocalStorageSize] })
            toast.success("Removed offline data")
        },
    })
}

export function useSyncLocalData() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.SYNC.SyncLocalData.endpoint,
        method: API_ENDPOINTS.SYNC.SyncLocalData.methods[0],
        mutationKey: [API_ENDPOINTS.SYNC.SyncLocalData.key],
        onSuccess: async () => {
            toast.info("Syncing local data...")
        },
    })
}

export function useSyncGetQueueState() {
    return useServerQuery<Sync_QueueState>({
        endpoint: API_ENDPOINTS.SYNC.SyncGetQueueState.endpoint,
        method: API_ENDPOINTS.SYNC.SyncGetQueueState.methods[0],
        queryKey: [API_ENDPOINTS.SYNC.SyncGetQueueState.key],
        enabled: true,
    })
}

export function useSyncGetIsMediaTracked(id: number, type: string) {
    return useServerQuery<boolean>({
        endpoint: API_ENDPOINTS.SYNC.SyncGetIsMediaTracked.endpoint.replace("{id}", String(id)).replace("{type}", String(type)),
        method: API_ENDPOINTS.SYNC.SyncGetIsMediaTracked.methods[0],
        queryKey: [API_ENDPOINTS.SYNC.SyncGetIsMediaTracked.key, id, type],
        enabled: true,
    })
}

export function useSyncAnilistData() {
    const qc = useQueryClient()
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.SYNC.SyncAnilistData.endpoint,
        method: API_ENDPOINTS.SYNC.SyncAnilistData.methods[0],
        mutationKey: [API_ENDPOINTS.SYNC.SyncAnilistData.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetRawAnilistMangaCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetLocalStorageSize] })
            toast.success("Updated Anilist data")
        },
    })
}

export function useSyncSetHasLocalChanges() {
    const qc = useQueryClient()
    return useServerMutation<boolean, SyncSetHasLocalChanges_Variables>({
        endpoint: API_ENDPOINTS.SYNC.SyncSetHasLocalChanges.endpoint,
        method: API_ENDPOINTS.SYNC.SyncSetHasLocalChanges.methods[0],
        mutationKey: [API_ENDPOINTS.SYNC.SyncSetHasLocalChanges.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetHasLocalChanges.key] })
        },
    })
}

export function useSyncGetHasLocalChanges() {
    return useServerQuery<boolean>({
        endpoint: API_ENDPOINTS.SYNC.SyncGetHasLocalChanges.endpoint,
        method: API_ENDPOINTS.SYNC.SyncGetHasLocalChanges.methods[0],
        queryKey: [API_ENDPOINTS.SYNC.SyncGetHasLocalChanges.key],
        enabled: true,
    })
}

export function useSyncGetLocalStorageSize() {
    return useServerQuery<string>({
        endpoint: API_ENDPOINTS.SYNC.SyncGetLocalStorageSize.endpoint,
        method: API_ENDPOINTS.SYNC.SyncGetLocalStorageSize.methods[0],
        queryKey: [API_ENDPOINTS.SYNC.SyncGetLocalStorageSize.key],
        enabled: true,
    })
}
