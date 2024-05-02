import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { MediastreamRequestTranscodeStream_Variables, SaveMediastreamSettings_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MediaContainer, Models_MediastreamSettings } from "@/api/generated/types"

export function useGetMediastreamSettings() {
    return useServerQuery<Models_MediastreamSettings>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.methods[0],
        queryKey: [API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.key],
        enabled: true,
    })
}

export function useSaveMediastreamSettings() {
    return useServerMutation<Models_MediastreamSettings, SaveMediastreamSettings_Variables>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.key],
        onSuccess: async () => {

        },
    })
}

export function useMediastreamRequestTranscodeStream() {
    return useServerMutation<MediaContainer, MediastreamRequestTranscodeStream_Variables>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.MediastreamRequestTranscodeStream.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.MediastreamRequestTranscodeStream.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIASTREAM.MediastreamRequestTranscodeStream.key],
        onSuccess: async () => {

        },
    })
}
