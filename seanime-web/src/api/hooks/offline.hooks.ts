import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { CreateOfflineSnapshot_Variables, UpdateOfflineEntryListData_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Offline_Snapshot, Offline_SnapshotEntry } from "@/api/generated/types"

export function useCreateOfflineSnapshot() {
    return useServerMutation<boolean, CreateOfflineSnapshot_Variables>({
        endpoint: API_ENDPOINTS.OFFLINE.CreateOfflineSnapshot.endpoint,
        method: API_ENDPOINTS.OFFLINE.CreateOfflineSnapshot.methods[0],
        mutationKey: [API_ENDPOINTS.OFFLINE.CreateOfflineSnapshot.key],
        onSuccess: async () => {

        },
    })
}

export function useGetOfflineSnapshot() {
    return useServerQuery<Offline_Snapshot>({
        endpoint: API_ENDPOINTS.OFFLINE.GetOfflineSnapshot.endpoint,
        method: API_ENDPOINTS.OFFLINE.GetOfflineSnapshot.methods[0],
        queryKey: [API_ENDPOINTS.OFFLINE.GetOfflineSnapshot.key],
        enabled: true,
    })
}

export function useGetOfflineSnapshotEntry() {
    return useServerQuery<Offline_SnapshotEntry>({
        endpoint: API_ENDPOINTS.OFFLINE.GetOfflineSnapshotEntry.endpoint,
        method: API_ENDPOINTS.OFFLINE.GetOfflineSnapshotEntry.methods[0],
        queryKey: [API_ENDPOINTS.OFFLINE.GetOfflineSnapshotEntry.key],
        enabled: true,
    })
}

export function useUpdateOfflineEntryListData() {
    return useServerMutation<boolean, UpdateOfflineEntryListData_Variables>({
        endpoint: API_ENDPOINTS.OFFLINE.UpdateOfflineEntryListData.endpoint,
        method: API_ENDPOINTS.OFFLINE.UpdateOfflineEntryListData.methods[0],
        mutationKey: [API_ENDPOINTS.OFFLINE.UpdateOfflineEntryListData.key],
        onSuccess: async () => {

        },
    })
}

export function useSyncOfflineData() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.OFFLINE.SyncOfflineData.endpoint,
        method: API_ENDPOINTS.OFFLINE.SyncOfflineData.methods[0],
        mutationKey: [API_ENDPOINTS.OFFLINE.SyncOfflineData.key],
        onSuccess: async () => {

        },
    })
}

