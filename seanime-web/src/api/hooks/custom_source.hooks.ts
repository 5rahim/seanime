import { useServerQuery } from "@/api/client/requests"
import { CustomSourceListAnime_Variables, CustomSourceListManga_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { HibikeCustomSource_ListAnimeResponse, HibikeCustomSource_ListMangaResponse } from "@/api/generated/types"

export function useCustomSourceListAnime(variables: CustomSourceListAnime_Variables, { enabled }: { enabled: boolean }) {
    return useServerQuery<HibikeCustomSource_ListAnimeResponse, CustomSourceListAnime_Variables>({
        endpoint: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.endpoint,
        method: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.methods[0],
        queryKey: [API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.key, variables],
        data: variables,
        enabled: enabled,
    })
}

export function useCustomSourceListManga(variables: CustomSourceListManga_Variables, { enabled }: { enabled: boolean }) {
    return useServerQuery<HibikeCustomSource_ListMangaResponse, CustomSourceListManga_Variables>({
        endpoint: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListManga.endpoint,
        method: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListManga.methods[0],
        queryKey: [API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListManga.key, variables],
        data: variables,
        enabled: enabled,
    })
}
