import { useServerMutation } from "@/api/client/requests"
import { CustomSourceListAnime_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { HibikeCustomSource_ListAnimeResponse } from "@/api/generated/types"

export function useCustomSourceListAnime() {
    return useServerMutation<HibikeCustomSource_ListAnimeResponse, CustomSourceListAnime_Variables>({
        endpoint: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.endpoint,
        method: API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.methods[0],
        mutationKey: [API_ENDPOINTS.CUSTOM_SOURCE.CustomSourceListAnime.key],
        onSuccess: async () => {

        },
    })
}
