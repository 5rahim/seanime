import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { SaveAutoSelectProfile_Variables, SearchTorrent_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_AutoSelectProfile, Torrent_SearchData } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"

export function useSearchTorrent(variables: SearchTorrent_Variables, enabled: boolean) {
    return useServerQuery<Torrent_SearchData, SearchTorrent_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_SEARCH.SearchTorrent.endpoint,
        method: API_ENDPOINTS.TORRENT_SEARCH.SearchTorrent.methods[0],
        data: variables,
        queryKey: [API_ENDPOINTS.TORRENT_SEARCH.SearchTorrent.key, JSON.stringify(variables)],
        enabled: enabled,
        gcTime: variables.episodeNumber === 0 ? 0 : undefined,
    })
}

export function useGetAutoSelectProfile() {
    return useServerQuery<Anime_AutoSelectProfile>({
        endpoint: API_ENDPOINTS.TORRENT_SEARCH.GetAutoSelectProfile.endpoint,
        method: API_ENDPOINTS.TORRENT_SEARCH.GetAutoSelectProfile.methods[0],
        queryKey: [API_ENDPOINTS.TORRENT_SEARCH.GetAutoSelectProfile.key],
        enabled: true,
    })
}

export function useSaveAutoSelectProfile() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoSelectProfile, SaveAutoSelectProfile_Variables>({
        endpoint: API_ENDPOINTS.TORRENT_SEARCH.SaveAutoSelectProfile.endpoint,
        method: API_ENDPOINTS.TORRENT_SEARCH.SaveAutoSelectProfile.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENT_SEARCH.SaveAutoSelectProfile.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.TORRENT_SEARCH.GetAutoSelectProfile.key] })
        },
    })
}

export function useDeleteAutoSelectProfile() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.TORRENT_SEARCH.DeleteAutoSelectProfile.endpoint,
        method: API_ENDPOINTS.TORRENT_SEARCH.DeleteAutoSelectProfile.methods[0],
        mutationKey: [API_ENDPOINTS.TORRENT_SEARCH.DeleteAutoSelectProfile.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.TORRENT_SEARCH.GetAutoSelectProfile.key] })
        },
    })
}
