import { useServerMutation } from "@/api/client/requests"
import { DownloadRelease_Variables, DownloadTorrentFile_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { DownloadReleaseResponse } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { toast } from "sonner"

export function useDownloadTorrentFile(onSuccess?: () => void) {
    return useServerMutation<boolean, DownloadTorrentFile_Variables>({
        endpoint: API_ENDPOINTS.DOWNLOAD.DownloadTorrentFile.endpoint,
        method: API_ENDPOINTS.DOWNLOAD.DownloadTorrentFile.methods[0],
        mutationKey: [API_ENDPOINTS.DOWNLOAD.DownloadTorrentFile.key],
        onSuccess: async () => {
            toast.success("Files downloaded")
            onSuccess?.()
        },
    })
}

export function useDownloadRelease() {
    const { mutate: openInExplorer } = useOpenInExplorer()

    return useServerMutation<DownloadReleaseResponse, DownloadRelease_Variables>({
        endpoint: API_ENDPOINTS.DOWNLOAD.DownloadRelease.endpoint,
        method: API_ENDPOINTS.DOWNLOAD.DownloadRelease.methods[0],
        mutationKey: [API_ENDPOINTS.DOWNLOAD.DownloadRelease.key],
        onSuccess: async data => {
            toast.success("Update downloaded successfully!")
            if (data?.error) {
                toast.error(data.error)
            }
            if (data?.destination) {
                openInExplorer({
                    path: data.destination,
                })
            }
        },
    })
}

