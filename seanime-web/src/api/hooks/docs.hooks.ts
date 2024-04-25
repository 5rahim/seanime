import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function useGetDocs() {
    return useServerQuery<any>({
        endpoint: API_ENDPOINTS.DOCS.GetDocs.endpoint,
        method: API_ENDPOINTS.DOCS.GetDocs.methods[0],
        queryKey: [API_ENDPOINTS.DOCS.GetDocs.key],
        enabled: true,
    })
}

