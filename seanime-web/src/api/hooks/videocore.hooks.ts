import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { VideoCoreSaveScreenshot_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { VideoCore_InSightCharacterDetails } from "@/api/generated/types"

export function useVideoCoreInSightGetCharacterDetails(malId: number) {
    return useServerQuery<VideoCore_InSightCharacterDetails>({
        endpoint: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.endpoint.replace("{malId}", String(malId)),
        method: API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.methods[0],
        queryKey: [API_ENDPOINTS.VIDEOCORE.VideoCoreInSightGetCharacterDetails.key, malId],
    })
}

export function useVideoCoreSaveScreenshot() {
    return useServerMutation<boolean, VideoCoreSaveScreenshot_Variables>({
        endpoint: API_ENDPOINTS.VIDEOCORE.VideoCoreSaveScreenshot.endpoint,
        method: API_ENDPOINTS.VIDEOCORE.VideoCoreSaveScreenshot.methods[0],
        mutationKey: [API_ENDPOINTS.VIDEOCORE.VideoCoreSaveScreenshot.key],
    })
}
