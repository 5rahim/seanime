import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { EditMALListEntryProgress_Variables, MALAuth_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MalAuthResponse } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useMALAuth(variables: Partial<MALAuth_Variables>, enabled: boolean) {
    return useServerQuery<MalAuthResponse, MALAuth_Variables>({
        endpoint: API_ENDPOINTS.MAL.MALAuth.endpoint,
        method: API_ENDPOINTS.MAL.MALAuth.methods[0],
        queryKey: [API_ENDPOINTS.MAL.MALAuth.key],
        data: variables as MALAuth_Variables,
        enabled: enabled,
    })
}

export function useEditMALListEntryProgress() {
    return useServerMutation<boolean, EditMALListEntryProgress_Variables>({
        endpoint: API_ENDPOINTS.MAL.EditMALListEntryProgress.endpoint,
        method: API_ENDPOINTS.MAL.EditMALListEntryProgress.methods[0],
        mutationKey: [API_ENDPOINTS.MAL.EditMALListEntryProgress.key],
        onSuccess: async () => {

        },
    })
}

export function useMALLogout() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MAL.MALLogout.endpoint,
        method: API_ENDPOINTS.MAL.MALLogout.methods[0],
        mutationKey: [API_ENDPOINTS.MAL.MALLogout.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Successfully logged out of MyAnimeList")
        },
    })
}

