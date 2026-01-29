import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    PreloadMediastreamMediaContainer_Variables,
    RequestMediastreamMediaContainer_Variables,
    SaveMediastreamSettings_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Mediastream_MediaContainer, Models_MediastreamSettings } from "@/api/generated/types"
import { logger } from "@/lib/helpers/debug"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetMediastreamSettings(enabled?: boolean) {
    return useServerQuery<Models_MediastreamSettings>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.methods[0],
        queryKey: [API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.key],
        enabled: enabled,
    })
}

export function useSaveMediastreamSettings() {
    const qc = useQueryClient()
    return useServerMutation<Models_MediastreamSettings, SaveMediastreamSettings_Variables>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIASTREAM.SaveMediastreamSettings.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MEDIASTREAM.GetMediastreamSettings.key] })
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings saved")
        },
    })
}

export function useRequestMediastreamMediaContainer(variables: Partial<RequestMediastreamMediaContainer_Variables>, enabled: boolean) {
    return useServerQuery<Mediastream_MediaContainer, RequestMediastreamMediaContainer_Variables>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.RequestMediastreamMediaContainer.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.RequestMediastreamMediaContainer.methods[0],
        queryKey: [API_ENDPOINTS.MEDIASTREAM.RequestMediastreamMediaContainer.key, variables?.path, variables?.streamType],
        data: variables as RequestMediastreamMediaContainer_Variables,
        enabled: !!variables.path && !!variables.streamType && enabled,
    })
}

export function usePreloadMediastreamMediaContainer() {
    return useServerMutation<boolean, PreloadMediastreamMediaContainer_Variables>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.PreloadMediastreamMediaContainer.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.PreloadMediastreamMediaContainer.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIASTREAM.PreloadMediastreamMediaContainer.key],
        onSuccess: async () => {
            logger("MEDIASTREAM").success("Preloaded mediastream media container")
        },
    })
}

export function useMediastreamShutdownTranscodeStream() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MEDIASTREAM.MediastreamShutdownTranscodeStream.endpoint,
        method: API_ENDPOINTS.MEDIASTREAM.MediastreamShutdownTranscodeStream.methods[0],
        mutationKey: [API_ENDPOINTS.MEDIASTREAM.MediastreamShutdownTranscodeStream.key],
        onSuccess: async () => {

        },
    })
}
