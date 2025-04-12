import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { InstallLatestUpdate_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Status, Updater_Update } from "@/api/generated/types"
import { useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { toast } from "sonner"

export function useGetLatestUpdate(enabled: boolean) {
    return useServerQuery<Updater_Update>({
        endpoint: API_ENDPOINTS.RELEASES.GetLatestUpdate.endpoint,
        method: API_ENDPOINTS.RELEASES.GetLatestUpdate.methods[0],
        queryKey: [API_ENDPOINTS.RELEASES.GetLatestUpdate.key],
        enabled: enabled,
    })
}

export function useInstallLatestUpdate() {
    const setServerStatus = useSetServerStatus()
    return useServerMutation<Status, InstallLatestUpdate_Variables>({
        endpoint: API_ENDPOINTS.RELEASES.InstallLatestUpdate.endpoint,
        method: API_ENDPOINTS.RELEASES.InstallLatestUpdate.methods[0],
        mutationKey: [API_ENDPOINTS.RELEASES.InstallLatestUpdate.key],
        onSuccess: async (data) => {
            setServerStatus(data) // Update server status
            toast.info("Installing update...")
        },
    })
}

export function useGetChangelog(before: string, after: string, enabled: boolean) {
    return useServerQuery<{ version: string, lines: string[] }[]>({
        endpoint: API_ENDPOINTS.RELEASES.GetChangelog.endpoint + `?before=${before}&after=${after}`,
        method: API_ENDPOINTS.RELEASES.GetChangelog.methods[0],
        queryKey: [API_ENDPOINTS.RELEASES.GetChangelog.key],
        enabled: enabled,
    })
}