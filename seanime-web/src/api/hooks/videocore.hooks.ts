import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { VideoCore_InSightCharacterDetails } from "@/api/generated/types"

export function useVideoCoreInSightGetCharacterDetails(malId: number) {
    return useServerQuery<VideoCore_InSightCharacterDetails>({
        endpoint: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.endpoint.replace("{malId}", String(malId)),
        method: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.methods[0],
        queryKey: [API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.key, malId],
    })
}
