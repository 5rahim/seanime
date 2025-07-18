import { useServerMutation } from "@/api/client/requests"
import {
    EmptyTVDBEpisodes_Variables,
    PopulateFillerData_Variables,
    PopulateTVDBEpisodes_Variables,
    RemoveFillerData_Variables,
} from "@/api/generated/endpoint.types"
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
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            toast.success("Metadata updated")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.key] })
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
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            toast.success("TheTVDB Metadata emptied")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.key] })
        },
    })
}

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

