import { useServerMutation } from "@/api/client/requests"
import { PlayVideo_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function usePlayVideo() {
    return useServerMutation<null, PlayVideo_Variables>({
        endpoint: API_ENDPOINTS.MEDIAPLAYER.PlayVideo.endpoint,
        method: API_ENDPOINTS.MEDIAPLAYER.PlayVideo.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIAPLAYER.PlayVideo.key],
        onSuccess: async () => {

        },
    })
}

export function useStartDefaultMediaPlayer() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.endpoint,
        method: API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.key],
        onSuccess: async () => {

        },
    })
}

