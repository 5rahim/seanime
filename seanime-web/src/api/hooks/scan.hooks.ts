import { useServerMutation } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_LocalFile } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useScanLocalFiles() {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>>({
        endpoint: API_ENDPOINTS.SCAN.ScanLocalFiles.endpoint,
        method: API_ENDPOINTS.SCAN.ScanLocalFiles.methods[0],
        mutationKey: [API_ENDPOINTS.SCAN.ScanLocalFiles.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            toast.success("Library scanned")
        },
    })
}


