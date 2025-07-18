import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_EpisodeCollection } from "@/api/generated/types"

export function useGetAnimeEpisodeCollection(id: number) {
    return useServerQuery<Anime_EpisodeCollection>({
        endpoint: API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANIME.GetAnimeEpisodeCollection.key, String(id)],
        enabled: true,
    })
}
