import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { SaveTorrentstreamSettings_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Models_TorrentstreamSettings, Torrentstream_EpisodeCollection } from "@/api/generated/types"
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

export function useGetTorrentstreamEpisodeCollection(id: number) {
    return useServerQuery<Torrentstream_EpisodeCollection>({
        endpoint: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamEpisodeCollection.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamEpisodeCollection.methods[0],
        queryKey: [API_ENDPOINTS.TORRENTSTREAM.GetTorrentstreamEpisodeCollection.key, id],
        enabled: true,
    })
}
