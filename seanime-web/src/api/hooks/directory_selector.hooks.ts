import { useServerMutation } from "@/api/client/requests"
import { DirectorySelector_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { DirectoryInfo } from "@/api/generated/types"

export function useDirectorySelector() {
    return useServerMutation<DirectoryInfo, DirectorySelector_Variables>({
        endpoint: API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.endpoint,
        method: API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.methods[0],
        mutationKey: [API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.key],
        onSuccess: async () => {

        },
    })
}

