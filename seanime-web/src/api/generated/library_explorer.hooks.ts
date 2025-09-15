import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { LibraryExplorer_FileTreeJSON } from "@/api/generated/types"

export function useGetLibraryExplorerFileTree() {
    return useServerQuery<LibraryExplorer_FileTreeJSON>({
        endpoint: API_ENDPOINTS.LIBRARY_EXPLORER.GetLibraryExplorerFileTree.endpoint,
        method: API_ENDPOINTS.LIBRARY_EXPLORER.GetLibraryExplorerFileTree.methods[0],
        queryKey: [API_ENDPOINTS.LIBRARY_EXPLORER.GetLibraryExplorerFileTree.key],
        enabled: true,
    })
}

export function useRefreshLibraryExplorerFileTree() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.LIBRARY_EXPLORER.RefreshLibraryExplorerFileTree.endpoint,
        method: API_ENDPOINTS.LIBRARY_EXPLORER.RefreshLibraryExplorerFileTree.methods[0],
        mutationKey: [API_ENDPOINTS.LIBRARY_EXPLORER.RefreshLibraryExplorerFileTree.key],
        onSuccess: async () => {

        },
    })
}
