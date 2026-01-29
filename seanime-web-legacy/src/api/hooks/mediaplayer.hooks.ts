import { useServerMutation } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"

export function useStartDefaultMediaPlayer() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.endpoint,
        method: API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIAPLAYER.StartDefaultMediaPlayer.key],
        onSuccess: async () => {

        },
    })
}

