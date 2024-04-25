import { useServerQuery } from "@/api/client/requests"
import { DirectorySelector_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { DirectorySelectorResponse } from "@/api/generated/types"

export function useDirectorySelector(debouncedInput: string) {
    return useServerQuery<DirectorySelectorResponse, DirectorySelector_Variables>({
        endpoint: API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.endpoint,
        method: API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.methods[0],
        queryKey: [API_ENDPOINTS.DIRECTORY_SELECTOR.DirectorySelector.key, debouncedInput],
        data: { input: debouncedInput },
        enabled: debouncedInput.length > 0,
        muteError: true,
    })
}

