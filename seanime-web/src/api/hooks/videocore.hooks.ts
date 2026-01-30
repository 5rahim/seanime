import { useServerQuery } from "@/api/client/requests.ts"
import { API_ENDPOINTS } from "@/api/generated/endpoints.ts"
import { VideoCore_InSightCharacterDetails } from "@/api/generated/types.ts"

export function useVideoCoreInSightGetCharacterDetails(malId: number) {
    return useServerQuery<VideoCore_InSightCharacterDetails>({
        endpoint: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.endpoint.replace("{malId}", String(malId)),
        method: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.methods[0],
        queryKey: [API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.key, malId],
    })
}
