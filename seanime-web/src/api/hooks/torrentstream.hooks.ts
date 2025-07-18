import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    GetTorrentstreamBatchHistory_Variables,
    GetTorrentstreamTorrentFilePreviews_Variables,
    SaveTorrentstreamSettings_Variables,
    TorrentstreamStartStream_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Models_TorrentstreamSettings, Nullish, Torrentstream_BatchHistoryResponse, Torrentstream_FilePreview } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetTorrentstreamSettings() {
    return useServerQuery<Models_TorrentstreamSettings>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamSettings.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamSettings.methods[0],
        queryKey: [API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamSettings.key],
        enabled: true,
    })
}

export function useSaveTorrentstreamSettings() {
    const qc = useQueryClient()
    return useServerMutation<Models_TorrentstreamSettings, SaveTorrentstreamSettings_Variables>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.SaveTorrentstreamSettings.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.SaveTorrentstreamSettings.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENTSTREAM.SaveTorrentstreamSettings.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamSettings.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings saved")
        },
    })
}

export function useTorrentstreamStartStream() {
    return useServerMutation<boolean, TorrentstreamStartStream_Variables>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStartStream.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStartStream.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStartStream.key],
        onSuccess: async () => {
        },
    })
}

export function useTorrentstreamStopStream() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStopStream.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStopStream.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENTSTREAM.TorrentstreamStopStream.key],
        onSuccess: async () => {
            toast.success("Stream stopped")
        },
    })
}

export function useTorrentstreamDropTorrent() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamDropTorrent.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.TorrentstreamDropTorrent.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENTSTREAM.TorrentstreamDropTorrent.key],
        onSuccess: async () => {
            toast.success("Torrent dropped")
        },
    })
}

export function useGetTorrentstreamTorrentFilePreviews(variables: Partial<GetTorrentstreamTorrentFilePreviews_Variables>, enabled: boolean) {
    return useServerQuery<Array<Torrentstream_FilePreview>, GetTorrentstreamTorrentFilePreviews_Variables>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamTorrentFilePreviews.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamTorrentFilePreviews.methods[0],
        queryKey: [API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamTorrentFilePreviews.key, variables],
        data: variables as GetTorrentstreamTorrentFilePreviews_Variables,
        enabled: enabled,
    })
}

export function useGetTorrentstreamBatchHistory(mediaId: Nullish<string | number>, enabled: boolean) {
    return useServerQuery<Torrentstream_BatchHistoryResponse, GetTorrentstreamBatchHistory_Variables>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamBatchHistory.endpoint,
        method: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamBatchHistory.methods[0],
        queryKey: [API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamBatchHistory.key, String(mediaId), enabled],
        data: {
            mediaId: Number(mediaId)!,
        },
        enabled: !!mediaId,
    })
}
