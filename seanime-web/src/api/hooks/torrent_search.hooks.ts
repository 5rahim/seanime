import { useServerQuery } from "@/api/client/requests"
import { SearchTorrent_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Torrent_SearchData } from "@/api/generated/types"

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
