import { useServerMutation } from "@/api/client/requests"
import { SetDiscordMangaActivity_Variables } from "@/api/generated/endpoint.types"
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

export function useCancelDiscordActivity() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.DISCORD.CancelDiscordActivity.endpoint,
        method: API_ENDPOINTS.DISCORD.CancelDiscordActivity.methods[0],
        mutationKey: [API_ENDPOINTS.DISCORD.CancelDiscordActivity.key],
        onSuccess: async () => {

        },
    })
}

