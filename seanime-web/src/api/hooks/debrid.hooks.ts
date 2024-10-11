import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    DebridAddTorrents_Variables,
    DebridCancelDownload_Variables,
    DebridDeleteTorrent_Variables,
    DebridDownloadTorrent_Variables,
    SaveDebridSettings_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Debrid_TorrentItem, Models_DebridSettings } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetDebridSettings() {
    return useServerQuery<Models_DebridSettings>({
        endpoint: API_ENDPOINTS.DEBRID.GetDebridSettings.endpoint,
        method: API_ENDPOINTS.DEBRID.GetDebridSettings.methods[0],
        queryKey: [API_ENDPOINTS.DEBRID.GetDebridSettings.key],
        enabled: true,
    })
}

export function useSaveDebridSettings() {
    const qc = useQueryClient()
    return useServerMutation<Models_DebridSettings, SaveDebridSettings_Variables>({
        endpoint: API_ENDPOINTS.DEBRID.SaveDebridSettings.endpoint,
        method: API_ENDPOINTS.DEBRID.SaveDebridSettings.methods[0],
        mutationKey: [API_ENDPOINTS.DEBRID.SaveDebridSettings.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.DEBRID.GetDebridSettings.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings saved")
        },
    })
}

export function useDebridAddTorrents(onSuccess: () => void) {
    const qc = useQueryClient()
    return useServerMutation<boolean, DebridAddTorrents_Variables>({
        endpoint: API_ENDPOINTS.DEBRID.DebridAddTorrents.endpoint,
        method: API_ENDPOINTS.DEBRID.DebridAddTorrents.methods[0],
        mutationKey: [API_ENDPOINTS.DEBRID.DebridAddTorrents.key],
        onSuccess: async () => {
            onSuccess()
            toast.success("Torrent added")
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.DEBRID.DebridGetTorrents.key] })
        },
    })
}

export function useDebridDownloadTorrent() {
    return useServerMutation<boolean, DebridDownloadTorrent_Variables>({
        endpoint: API_ENDPOINTS.DEBRID.DebridDownloadTorrent.endpoint,
        method: API_ENDPOINTS.DEBRID.DebridDownloadTorrent.methods[0],
        mutationKey: [API_ENDPOINTS.DEBRID.DebridDownloadTorrent.key],
        onSuccess: async () => {
            toast.info("Download started")
        },
    })
}

export function useDebridCancelDownload() {
    return useServerMutation<boolean, DebridCancelDownload_Variables>({
        endpoint: API_ENDPOINTS.DEBRID.DebridCancelDownload.endpoint,
        method: API_ENDPOINTS.DEBRID.DebridCancelDownload.methods[0],
        mutationKey: [API_ENDPOINTS.DEBRID.DebridCancelDownload.key],
        onSuccess: async () => {
            toast.info("Download cancelled")
        },
    })
}

export function useDebridDeleteTorrent() {
    const qc = useQueryClient()
    return useServerMutation<boolean, DebridDeleteTorrent_Variables>({
        endpoint: API_ENDPOINTS.DEBRID.DebridDeleteTorrent.endpoint,
        method: API_ENDPOINTS.DEBRID.DebridDeleteTorrent.methods[0],
        mutationKey: [API_ENDPOINTS.DEBRID.DebridDeleteTorrent.key],
        onSuccess: async () => {
            toast.success("Torrent deleted")
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.DEBRID.DebridGetTorrents.key] })
        },
    })
}

export function useDebridGetTorrents(enabled: boolean, refetchInterval: number) {
    return useServerQuery<Array<Debrid_TorrentItem>>({
        endpoint: API_ENDPOINTS.DEBRID.DebridGetTorrents.endpoint,
        method: API_ENDPOINTS.DEBRID.DebridGetTorrents.methods[0],
        queryKey: [API_ENDPOINTS.DEBRID.DebridGetTorrents.key],
        enabled: enabled,
        retry: 3,
        refetchInterval: refetchInterval,
    })
}
