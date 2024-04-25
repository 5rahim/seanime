import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    GetOnlineStreamEpisodeList_Variables,
    GetOnlineStreamEpisodeSource_Variables,
    OnlineStreamEmptyCache_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Nullish, Onlinestream_EpisodeListResponse, Onlinestream_EpisodeSource } from "@/api/generated/types"

export function useGetOnlineStreamEpisodeList(id: Nullish<string | number>, provider: string, dubbed: boolean) {
    return useServerQuery<Onlinestream_EpisodeListResponse, GetOnlineStreamEpisodeList_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.methods[0],
        queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key, String(id), provider, dubbed],
        data: {
            mediaId: Number(id),
            provider,
            dubbed,
        },
        enabled: !!id,
    })
}

export function useGetOnlineStreamEpisodeSource(id: Nullish<string | number>,
    provider: string,
    episodeNumber: Nullish<number>,
    dubbed: boolean,
    enabled: boolean,
) {
    return useServerQuery<Onlinestream_EpisodeSource, GetOnlineStreamEpisodeSource_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.methods[0],
        queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.key, String(id), provider, episodeNumber, dubbed],
        data: {
            mediaId: Number(id),
            episodeNumber: episodeNumber!,
            dubbed: dubbed,
            provider: provider,
        },
        enabled: enabled,
    })
}

export function useOnlineStreamEmptyCache() {
    return useServerMutation<boolean, OnlineStreamEmptyCache_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.key],
        onSuccess: async () => {

        },
    })
}

