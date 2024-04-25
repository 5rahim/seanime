import { useServerMutation } from "@/api/client/requests"
import { OpenInExplorer_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function useOpenInExplorer() {
    return useServerMutation<boolean, OpenInExplorer_Variables>({
        endpoint: API_ENDPOINTS.EXPLORER.OpenInExplorer.endpoint,
        method: API_ENDPOINTS.EXPLORER.OpenInExplorer.methods[0],
        mutationKey: [API_ENDPOINTS.EXPLORER.OpenInExplorer.key],
        onSuccess: async () => {

        },
    })
}

