import { useServerMutation } from "@/api/client/requests"
import { PopulateFillerData_Variables, RemoveFillerData_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function usePopulateFillerData() {
    const queryClient = useQueryClient()

    return useServerMutation<true, PopulateFillerData_Variables>({
        endpoint: API_ENDPOINTS.METADATA.PopulateFillerData.endpoint,
        method: API_ENDPOINTS.METADATA.PopulateFillerData.methods[0],
        mutationKey: [API_ENDPOINTS.METADATA.PopulateFillerData.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            toast.success("Filler data fetched")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.key] })
        },
    })
}

export function useRemoveFillerData() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, RemoveFillerData_Variables>({
        endpoint: API_ENDPOINTS.METADATA.RemoveFillerData.endpoint,
        method: API_ENDPOINTS.METADATA.RemoveFillerData.methods[0],
        mutationKey: [API_ENDPOINTS.METADATA.RemoveFillerData.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            toast.success("Filler data removed")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.key] })
        },
    })
}

