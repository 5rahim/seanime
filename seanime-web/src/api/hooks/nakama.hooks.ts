import { useServerMutation, useServerQuery } from "../client/requests"
import { NakamaPlayVideo_Variables, SendNakamaMessage_Variables } from "../generated/endpoint.types"
import { API_ENDPOINTS } from "../generated/endpoints"
import { Anime_LibraryCollection, Nakama_MessageResponse, Nakama_NakamaStatus } from "../generated/types"

export function useNakamaWebSocket() {
    return useServerQuery<boolean>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaWebSocket.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaWebSocket.methods[0],
        queryKey: [API_ENDPOINTS.NAKAMA.NakamaWebSocket.key],
        enabled: true,
    })
}

export function useGetNakamaStatus() {
    return useServerQuery<Nakama_NakamaStatus>({
        endpoint: API_ENDPOINTS.NAKAMA.GetNakamaStatus.endpoint,
        method: API_ENDPOINTS.NAKAMA.GetNakamaStatus.methods[0],
        queryKey: [API_ENDPOINTS.NAKAMA.GetNakamaStatus.key],
        enabled: true,
    })
}

export function useSendNakamaMessage() {
    return useServerMutation<Nakama_MessageResponse, SendNakamaMessage_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.SendNakamaMessage.endpoint,
        method: API_ENDPOINTS.NAKAMA.SendNakamaMessage.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.SendNakamaMessage.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaReconnectToHost() {
    return useServerMutation<Nakama_MessageResponse, {}>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaReconnectToHost.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaReconnectToHost.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaReconnectToHost.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaRemoveStaleConnections() {
    return useServerMutation<Nakama_MessageResponse, {}>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaRemoveStaleConnections.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaRemoveStaleConnections.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaRemoveStaleConnections.key],
        onSuccess: async () => {

        },
    })
}

export function useGetNakamaAnimeLibraryCollection() {
    return useServerQuery<Anime_LibraryCollection>({
        endpoint: API_ENDPOINTS.NAKAMA.GetNakamaAnimeLibraryCollection.endpoint,
        method: API_ENDPOINTS.NAKAMA.GetNakamaAnimeLibraryCollection.methods[0],
        queryKey: [API_ENDPOINTS.NAKAMA.GetNakamaAnimeLibraryCollection.key],
        enabled: true,
    })
}

export function useNakamaPlayVideo() {
    return useServerMutation<boolean, NakamaPlayVideo_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaPlayVideo.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaPlayVideo.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaPlayVideo.key],
        onSuccess: async () => {

        },
    })
}
