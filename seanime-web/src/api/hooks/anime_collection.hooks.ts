import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { AddUnknownMedia_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { AL_AnimeCollection, Anime_LibraryCollection } from "@/api/generated/types"

export function useGetLibraryCollection() {
    return useServerQuery<Anime_LibraryCollection>({
        endpoint: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.endpoint,
        method: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key],
        enabled: true,
    })
}

// export function useGetLibraryCollection() {
//     return useServerMutation<Anime_LibraryCollection>({
//         endpoint: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.endpoint,
//         method: API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.methods[1],
//         mutationKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key],
//         onSuccess: async () => {
//
//         },
//     })
// }

export function useAddUnknownMedia() {
    return useServerMutation<AL_AnimeCollection, AddUnknownMedia_Variables>({
        endpoint: API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.endpoint,
        method: API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_COLLECTION.AddUnknownMedia.key],
        onSuccess: async () => {

        },
    })
}

