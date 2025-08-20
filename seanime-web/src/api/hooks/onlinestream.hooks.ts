import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    GetOnlineStreamEpisodeList_Variables,
    GetOnlineStreamEpisodeSource_Variables,
    GetOnlinestreamMapping_Variables,
    OnlineStreamEmptyCache_Variables,
    OnlinestreamManualMapping_Variables,
    OnlinestreamManualSearch_Variables,
    RemoveOnlinestreamMapping_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    HibikeOnlinestream_SearchResult,
    Nullish,
    Onlinestream_EpisodeListResponse,
    Onlinestream_EpisodeSource,
    Onlinestream_MappingResponse,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetOnlineStreamEpisodeList(id: Nullish<string | number>, provider: Nullish<string>, dubbed: boolean) {
    return useServerQuery<Onlinestream_EpisodeListResponse, GetOnlineStreamEpisodeList_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.methods[0],
        queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key, String(id), provider, dubbed],
        data: {
            mediaId: Number(id),
            provider: provider!,
            dubbed,
        },
        enabled: !!id,
        muteError: true,
    })
}

export function useGetOnlineStreamEpisodeSource(id: Nullish<string | number>,
    provider: Nullish<string>,
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
            provider: provider!,
        },
        enabled: enabled && !!provider,
        muteError: true,
    })
}

export function useOnlineStreamEmptyCache() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, OnlineStreamEmptyCache_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.OnlineStreamEmptyCache.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.key] })
            toast.info("Stream cache emptied")
        },
    })
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useOnlinestreamManualSearch(mId: number, provider: Nullish<string>) {
    return useServerMutation<Array<HibikeOnlinestream_SearchResult>, OnlinestreamManualSearch_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualSearch.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualSearch.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualSearch.key],
        onSuccess: async () => {

        },
    })
}

export function useOnlinestreamManualMapping() {
    const qc = useQueryClient()
    return useServerMutation<boolean, OnlinestreamManualMapping_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualMapping.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualMapping.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.OnlinestreamManualMapping.key],
        onSuccess: async () => {
            toast.success("Mapping added")
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlinestreamMapping.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.key] })
        },
    })
}

export function useGetOnlinestreamMapping(variables: Partial<GetOnlinestreamMapping_Variables>) {
    return useServerQuery<Onlinestream_MappingResponse, GetOnlinestreamMapping_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.GetOnlinestreamMapping.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.GetOnlinestreamMapping.methods[0],
        queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlinestreamMapping.key, String(variables.mediaId), variables.provider],
        data: variables as GetOnlinestreamMapping_Variables,
        enabled: !!variables.provider && !!variables.mediaId,
    })
}

export function useRemoveOnlinestreamMapping() {
    const qc = useQueryClient()

    return useServerMutation<boolean, RemoveOnlinestreamMapping_Variables>({
        endpoint: API_ENDPOINTS.ONLINESTREAM.RemoveOnlinestreamMapping.endpoint,
        method: API_ENDPOINTS.ONLINESTREAM.RemoveOnlinestreamMapping.methods[0],
        mutationKey: [API_ENDPOINTS.ONLINESTREAM.RemoveOnlinestreamMapping.key],
        onSuccess: async () => {
            toast.info("Mapping removed")
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlinestreamMapping.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeList.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ONLINESTREAM.GetOnlineStreamEpisodeSource.key] })
        },
    })
}
