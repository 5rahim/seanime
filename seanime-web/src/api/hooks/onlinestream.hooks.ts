import { useServerMutation } from "@/api/client/requests"
import {
    GetOnlineStreamEpisodeList_Variables,
    GetOnlineStreamEpisodeSource_Variables,
    OnlineStreamEmptyCache_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Onlinestream_EpisodeListResponse, Onlinestream_EpisodeSource } from "@/api/generated/types"

export function useGetOnlineStreamEpisodeList() {
    return useServerMutation<Onlinestream_EpisodeListResponse, GetOnlineStreamEpisodeList_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key],
        onSuccess: async () => {

        },
    })
}

export function useGetOnlineStreamEpisodeSource() {
    return useServerMutation<Onlinestream_EpisodeSource, GetOnlineStreamEpisodeSource_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.key],
        onSuccess: async () => {

        },
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

