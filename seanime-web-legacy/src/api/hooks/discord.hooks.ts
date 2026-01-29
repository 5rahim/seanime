import { useServerMutation } from "@/api/client/requests"
import {
    SetDiscordAnimeActivityWithProgress_Variables,
    SetDiscordLegacyAnimeActivity_Variables,
    SetDiscordMangaActivity_Variables,
    UpdateDiscordAnimeActivityWithProgress_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function useSetDiscordMangaActivity() {
    return useServerMutation<boolean, SetDiscordMangaActivity_Variables>({
        endpoint: API_ENDPOINTS.DISCORD.SetDiscordMangaActivity.endpoint,
        method: API_ENDPOINTS.DISCORD.SetDiscordMangaActivity.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.SetDiscordMangaActivity.key],
        onSuccess: async () => {

        },
    })
}

export function useSetDiscordLegacyAnimeActivity() {
    return useServerMutation<boolean, SetDiscordLegacyAnimeActivity_Variables>({
        endpoint: API_ENDPOINTS.DISCORD.SetDiscordLegacyAnimeActivity.endpoint,
        method: API_ENDPOINTS.DISCORD.SetDiscordLegacyAnimeActivity.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.SetDiscordLegacyAnimeActivity.key],
        onSuccess: async () => {

        },
    })
}

export function useSetDiscordAnimeActivityWithProgress() {
    return useServerMutation<boolean, SetDiscordAnimeActivityWithProgress_Variables>({
        endpoint: API_ENDPOINTS.DISCORD.SetDiscordAnimeActivityWithProgress.endpoint,
        method: API_ENDPOINTS.DISCORD.SetDiscordAnimeActivityWithProgress.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.SetDiscordAnimeActivityWithProgress.key],
        onSuccess: async () => {

        },
    })
}

export function useUpdateDiscordAnimeActivityWithProgress() {
    return useServerMutation<boolean, UpdateDiscordAnimeActivityWithProgress_Variables>({
        endpoint: API_ENDPOINTS.DISCORD.UpdateDiscordAnimeActivityWithProgress.endpoint,
        method: API_ENDPOINTS.DISCORD.UpdateDiscordAnimeActivityWithProgress.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.UpdateDiscordAnimeActivityWithProgress.key],
        onSuccess: async () => {

        },
    })
}

export function useCancelDiscordActivity() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.DISCORD.CancelDiscordActivity.endpoint,
        method: API_ENDPOINTS.DISCORD.CancelDiscordActivity.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.CancelDiscordActivity.key],
        onSuccess: async () => {

        },
    })
}

