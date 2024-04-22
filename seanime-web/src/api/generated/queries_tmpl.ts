import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useQueryClient } from "@tanstack/react-query"

export function useTemplateQuery() {
    return useServerQuery({
        endpoint: API_ENDPOINTS.DOCS.GetDocs.endpoint,
        method: API_ENDPOINTS.DOCS.GetDocs.methods[0],
        queryKey: [API_ENDPOINTS.DOCS.GetDocs.key],
    })
}

// type MutationHookProps = {
//     onSuccess
// }

export function useTemplateMutation(props: {}) {
    const queryClient = useQueryClient()
    return useServerMutation({
        endpoint: API_ENDPOINTS.DOCS.GetDocs.endpoint,
        method: API_ENDPOINTS.DOCS.GetDocs.methods[0],
        mutationKey: [API_ENDPOINTS.DOCS.GetDocs.key],
        onSuccess: async () => {

        },
    })
}
