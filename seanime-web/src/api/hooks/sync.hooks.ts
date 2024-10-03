import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { SyncAddMedia_Variables, SyncRemoveMedia_Variables } from "@/api/generated/endpoint.types"
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
            toast.success("Added media for syncing")
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
            toast.success("Removed media from syncing")
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
