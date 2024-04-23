import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Updater_Update } from "@/api/generated/types"

export function useGetLatestUpdate() {
    return useServerQuery<Updater_Update>({
        endpoint: API_ENDPOINTS.RELEASES.GetLatestUpdate.endpoint,
        method: API_ENDPOINTS.RELEASES.GetLatestUpdate.methods[0],
        queryKey: [API_ENDPOINTS.RELEASES.GetLatestUpdate.key],
        enabled: true,
    })
}

