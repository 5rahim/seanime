import { useServerMutation } from "../client/requests"
import { DirectstreamPlayLocalFile_Variables } from "../generated/endpoint.types"
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
