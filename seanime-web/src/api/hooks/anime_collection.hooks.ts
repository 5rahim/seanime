import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { AddUnknownMedia_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { AL_AnimeCollection, Anime_LibraryCollection, AnimeCollectionScheduleItem } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetLibraryCollection() {
    return useServerQuery<Anime_LibraryCollection>({
        endpoint: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.endpoint,
        method: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key],
        enabled: true,
    })
}

export function useAddUnknownMedia() {
    const queryClient = useQueryClient()

    return useServerMutation<AL_AnimeCollection, AddUnknownMedia_Variables>({
        endpoint: API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.endpoint,
        method: API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.key],
        onSuccess: async () => {
            toast.success("Media added successfully")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
        },
    })
}

export function useGetAnimeCollectionSchedule() {
    return useServerQuery<Array<AnimeCollectionScheduleItem>>({
        endpoint: API_ENDPOINTS.ANIME_COLLECTION.GetAnimeCollectionSchedule.endpoint,
        method: API_ENDPOINTS.ANIME_COLLECTION.GetAnimeCollectionSchedule.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetAnimeCollectionSchedule.key],
        enabled: true,
    })
}
