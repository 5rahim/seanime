import { useServerMutation } from "@/api/client/requests"
import { EmptyTVDBEpisodes_Variables, PopulateTVDBEpisodes_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { TVDB_Episode } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function usePopulateTVDBEpisodes() {
    const queryClient = useQueryClient()

    return useServerMutation<Array<TVDB_Episode>, PopulateTVDBEpisodes_Variables>({
        endpoint: API_ENDPOINTS.METADATA.PopulateTVDBEpisodes.endpoint,
        method: API_ENDPOINTS.METADATA.PopulateTVDBEpisodes.methods[0],
        mutationKey: [API_ENDPOINTS.METADATA.PopulateTVDBEpisodes.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry] })
            toast.success("Metadata updated")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection] })
        },
    })
}

export function useEmptyTVDBEpisodes() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, EmptyTVDBEpisodes_Variables>({
        endpoint: API_ENDPOINTS.METADATA.EmptyTVDBEpisodes.endpoint,
        method: API_ENDPOINTS.METADATA.EmptyTVDBEpisodes.methods[0],
        mutationKey: [API_ENDPOINTS.METADATA.EmptyTVDBEpisodes.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry] })
            toast.success("TheTVDB Metadata emptied")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection] })
        },
    })
}

