import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import type { MpvCore_InSightCharacterDetails } from "@/api/generated/types"

export function useMpvCoreInSightGetCharacterDetails(malId: number) {
    return useServerQuery<MpvCore_InSightCharacterDetails>({
        endpoint: API_ENDPOINTS.MPVCORE.MpvCoreInSightGetCharacterDetails.endpoint.replace("{malId}", String(malId)),
        method: API_ENDPOINTS.MPVCORE.MpvCoreInSightGetCharacterDetails.methods[0],
        queryKey: [API_ENDPOINTS.MPVCORE.MpvCoreInSightGetCharacterDetails.key, malId],
    })
}
