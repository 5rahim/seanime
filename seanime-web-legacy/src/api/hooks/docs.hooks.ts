import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { ApiDocsGroup } from "@/api/generated/types"

export function useGetDocs() {
    return useServerQuery<ApiDocsGroup[]>({
        endpoint: API_ENDPOINTS.DOCS.GetDocs.endpoint,
        method: API_ENDPOINTS.DOCS.GetDocs.methods[0],
        queryKey: [API_ENDPOINTS.DOCS.GetDocs.key],
        enabled: true,
    })
}

