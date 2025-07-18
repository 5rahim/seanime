import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { DeleteLogs_Variables, GetAnnouncements_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Status, Updater_Announcement } from "@/api/generated/types"
import { copyToClipboard } from "@/lib/helpers/browser"
import { __isDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetStatus() {
    return useServerQuery<Status>({
        endpoint: API_ENDPOINTS.STATUS.GetStatus.endpoint,
        method: API_ENDPOINTS.STATUS.GetStatus.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetStatus.key],
        enabled: true,
        retryDelay: 1000,
        // Fixes macOS desktop app startup issue
        retry: 6,
        // Mute error if the platform is desktop
        muteError: __isDesktop__,
    })
}

export function useGetLogFilenames() {
    return useServerQuery<Array<string>>({
        endpoint: API_ENDPOINTS.STATUS.GetLogFilenames.endpoint,
        method: API_ENDPOINTS.STATUS.GetLogFilenames.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetLogFilenames.key],
        enabled: true,
    })
}

export function useDeleteLogs() {
    const qc = useQueryClient()
    return useServerMutation<boolean, DeleteLogs_Variables>({
        endpoint: API_ENDPOINTS.STATUS.DeleteLogs.endpoint,
        method: API_ENDPOINTS.STATUS.DeleteLogs.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.DeleteLogs.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetLogFilenames.key] })
            toast.success("Logs deleted")
        },
    })
}

export function useGetLatestLogContent() {
    const qc = useQueryClient()
    return useServerMutation<string>({
        endpoint: API_ENDPOINTS.STATUS.GetLatestLogContent.endpoint,
        method: API_ENDPOINTS.STATUS.GetLatestLogContent.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetLatestLogContent.key],
        onSuccess: async data => {
            if (!data) return toast.error("Couldn't fetch logs")
            try {
                await copyToClipboard(data)
                toast.success("Copied to clipboard")
            }
            catch (err: any) {
                console.error("Clipboard write error:", err)
                toast.error("Failed to copy logs: " + err.message)
            }
        },
    })
}

export function useGetAnnouncements() {
    return useServerMutation<Array<Updater_Announcement>, GetAnnouncements_Variables>({
        endpoint: API_ENDPOINTS.STATUS.GetAnnouncements.endpoint,
        method: API_ENDPOINTS.STATUS.GetAnnouncements.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetAnnouncements.key],
    })
}