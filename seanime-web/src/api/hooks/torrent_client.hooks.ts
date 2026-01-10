import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    TorrentClientAction_Variables,
    TorrentClientAddMagnetFromRule_Variables,
    TorrentClientDownload_Variables,
    TorrentClientGetFiles_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { HibikeTorrent_AnimeTorrent, Nullish, TorrentClient_Torrent } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"
import { toast } from "sonner"

export function useGetActiveTorrentList(enabled: boolean, category: string, sort: string) {
    const query = React.useMemo(() => {
        if (!category && !sort) return ""
        let q = "?"
        if (category) q += `category=${category}&`
        if (sort) q += `sort=${sort}`
        return q
    }, [category, sort])
    return useServerQuery<Array<TorrentClient_Torrent>>({
        endpoint: API_ENDPOINTS.TORRENT_CLIENT.GetActiveTorrentList.endpoint + query,
        method: API_ENDPOINTS.TORRENT_CLIENT.GetActiveTorrentList.methods[0],
        queryKey: [API_ENDPOINTS.TORRENT_CLIENT.GetActiveTorrentList.key, category, sort],
        refetchInterval: 1500,
        gcTime: 0,
        enabled: enabled,
    })
}

export function useTorrentClientAction(onSuccess?: () => void) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, TorrentClientAction_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAction.endpoint,
        method: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAction.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAction.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.TORRENT_CLIENT.GetActiveTorrentList.key] })
            toast.success("Action performed")
            onSuccess?.()
        },
    })
}

export function useTorrentClientDownload(onSuccess?: () => void) {
    return useServerMutation<boolean, TorrentClientDownload_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientDownload.endpoint,
        method: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientDownload.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENT_CLIENT.TorrentClientDownload.key],
        onSuccess: async () => {
            toast.success("Download started")
            onSuccess?.()
        },
    })
}

export function useTorrentClientAddMagnetFromRule() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, TorrentClientAddMagnetFromRule_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAddMagnetFromRule.endpoint,
        method: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAddMagnetFromRule.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENT_CLIENT.TorrentClientAddMagnetFromRule.key],
        onSuccess: async () => {
            toast.success("Download started")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key] })
        },
    })
}

export function useTorrentClientGetFiles({ torrent, provider }: { torrent: Nullish<HibikeTorrent_AnimeTorrent>, provider: Nullish<string> }) {
    return useServerQuery<Array<string>, TorrentClientGetFiles_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientGetFiles.endpoint,
        method: API_ENDPOINTS.TORRENT_CLIENT.TorrentClientGetFiles.methods[0],
        queryKey: [API_ENDPOINTS.TORRENT_CLIENT.TorrentClientGetFiles.key, torrent, provider],
        enabled: !!torrent && !!provider,
        data: {
            torrent: torrent!,
            provider: provider!,
        },
    })
}
