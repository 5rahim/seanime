import { useServerMutation } from "@/api/client/requests"
import { EditMALListEntryProgress_Variables, MALAuth_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MalAuthResponse } from "@/api/generated/types"

export function useMALAuth() {
    return useServerMutation<MalAuthResponse, MALAuth_Variables>({
        endpoint: API_ENDPOINTS.MAL.MALAuth.endpoint,
        method: API_ENDPOINTS.MAL.MALAuth.methods[0],
        mutationKey: [API_ENDPOINTS.MAL.MALAuth.key],
        onSuccess: async () => {

        },
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
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MAL.MALLogout.endpoint,
        method: API_ENDPOINTS.MAL.MALLogout.methods[0],
        mutationKey: [API_ENDPOINTS.MAL.MALLogout.key],
        onSuccess: async () => {

        },
    })
}

