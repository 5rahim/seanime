import { useServerMutation, useServerQuery } from "../client/requests"
import {
    NakamaCreateWatchParty_Variables,
    NakamaJoinWatchParty_Variables,
    NakamaPlayVideo_Variables,
    NakamaSendChatMessage_Variables,
    SendNakamaMessage_Variables,
} from "../generated/endpoint.types"
import { API_ENDPOINTS } from "../generated/endpoints"
import { Nakama_MessageResponse } from "../generated/types"

export function useNakamaWebSocket() {
    return useServerQuery<boolean>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaWebSocket.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaWebSocket.methods[0],
        queryKey: [API_ENDPOINTS.NAKAMA.NakamaWebSocket.key],
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

export function useNakamaPlayVideo() {
    return useServerMutation<boolean, NakamaPlayVideo_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaPlayVideo.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaPlayVideo.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaPlayVideo.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaCreateWatchParty() {
    return useServerMutation<any, NakamaCreateWatchParty_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaCreateWatchParty.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaCreateWatchParty.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaCreateWatchParty.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaJoinWatchParty() {
    return useServerMutation<Nakama_MessageResponse, NakamaJoinWatchParty_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaJoinWatchParty.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaJoinWatchParty.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaJoinWatchParty.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaLeaveWatchParty() {
    return useServerMutation<Nakama_MessageResponse>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaLeaveWatchParty.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaLeaveWatchParty.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaLeaveWatchParty.key],
        onSuccess: async () => {

        },
    })
}

export function useNakamaSendChatMessage() {
    return useServerMutation<boolean, NakamaSendChatMessage_Variables>({
        endpoint: API_ENDPOINTS.NAKAMA.NakamaSendChatMessage.endpoint,
        method: API_ENDPOINTS.NAKAMA.NakamaSendChatMessage.methods[0],
        mutationKey: [API_ENDPOINTS.NAKAMA.NakamaSendChatMessage.key],
        onSuccess: async () => {

        },
    })
}
