import { useServerMutation } from "../client/requests"
import {
    DirectstreamConvertSubs_Variables,
    DirectstreamFetchAndConvertToASS_Variables,
    DirectstreamPlayLocalFile_Variables,
} from "../generated/endpoint.types"
import { API_ENDPOINTS } from "../generated/endpoints"
import { Mediastream_MediaContainer } from "../generated/types"

export function useDirectstreamPlayLocalFile() {
    return useServerMutation<Mediastream_MediaContainer, DirectstreamPlayLocalFile_Variables>({
        endpoint: API_ENDPOINTS.DIRECTSTREAM.DirectstreamPlayLocalFile.endpoint,
        method: API_ENDPOINTS.DIRECTSTREAM.DirectstreamPlayLocalFile.methods[0],
        mutationKey: [API_ENDPOINTS.DIRECTSTREAM.DirectstreamPlayLocalFile.key],
        onSuccess: async () => {

        },
    })
}

export function useDirectstreamFetchAndConvertToASS({ onSuccess }: { onSuccess: (data: string | undefined) => void }) {
    return useServerMutation<string, DirectstreamFetchAndConvertToASS_Variables>({
        endpoint: API_ENDPOINTS.DIRECTSTREAM.DirectstreamFetchAndConvertToASS.endpoint,
        method: API_ENDPOINTS.DIRECTSTREAM.DirectstreamFetchAndConvertToASS.methods[0],
        mutationKey: [API_ENDPOINTS.DIRECTSTREAM.DirectstreamFetchAndConvertToASS.key],
        onSuccess: async (data) => {
            onSuccess(data)
        },
    })
}

export function useDirectstreamConvertSubs({ onSuccess }: { onSuccess: (data: string | undefined) => void }) {
    return useServerMutation<string, DirectstreamConvertSubs_Variables>({
        endpoint: API_ENDPOINTS.DIRECTSTREAM.DirectstreamConvertSubs.endpoint,
        method: API_ENDPOINTS.DIRECTSTREAM.DirectstreamConvertSubs.methods[0],
        mutationKey: [API_ENDPOINTS.DIRECTSTREAM.DirectstreamConvertSubs.key],
        onSuccess: async (data) => {
            onSuccess(data)
        },
    })
}
