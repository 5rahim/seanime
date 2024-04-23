import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Status } from "@/api/generated/types"

export function useGetStatus() {
    return useServerQuery<Status>({
        endpoint: API_ENDPOINTS.STATUS.GetStatus.endpoint,
        method: API_ENDPOINTS.STATUS.GetStatus.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetStatus.key],
        enabled: true,
    })
}

