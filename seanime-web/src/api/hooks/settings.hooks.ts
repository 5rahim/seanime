import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    GettingStarted_Variables,
    PatchSetting_Variables,
    SaveAutoDownloaderSettings_Variables,
    SaveMediaPlayerSettings_Variables,
    SaveSettings_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Models_Settings, Status } from "@/api/generated/types"
import { isLoginModalOpenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useQueryClient } from "@tanstack/react-query"
import { useSetAtom } from "jotai/react"
import { toast } from "sonner"

export function useGetSettings(options?: { enabled?: boolean }) {
    return useServerQuery<Models_Settings>({
        endpoint: API_ENDPOINTS.SETTINGS.GetSettings.endpoint,
        method: API_ENDPOINTS.SETTINGS.GetSettings.methods[0],
        queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key],
        enabled: options?.enabled ?? true,
    })
}

export function useGettingStarted() {
    const queryClient = useQueryClient()
    const setLoginModalOpen = useSetAtom(isLoginModalOpenAtom)

    return useServerMutation<Status, GettingStarted_Variables>({
        endpoint: API_ENDPOINTS.SETTINGS.GettingStarted.endpoint,
        method: API_ENDPOINTS.SETTINGS.GettingStarted.methods[0],
        mutationKey: [API_ENDPOINTS.SETTINGS.GettingStarted.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            setLoginModalOpen(true)
        },
    })
}

export function useSaveSettings() {
    const queryClient = useQueryClient()

    return useServerMutation<Status, SaveSettings_Variables>({
        endpoint: API_ENDPOINTS.SETTINGS.SaveSettings.endpoint,
        method: API_ENDPOINTS.SETTINGS.SaveSettings.methods[0],
        mutationKey: [API_ENDPOINTS.SETTINGS.SaveSettings.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings saved")
        },
    })
}

export function usePatchSetting() {
    const queryClient = useQueryClient()
    const setServerStatus = useSetServerStatus()

    return useServerMutation<Status, PatchSetting_Variables>({
        endpoint: API_ENDPOINTS.SETTINGS.PatchSetting.endpoint,
        method: API_ENDPOINTS.SETTINGS.PatchSetting.methods[0],
        mutationKey: [API_ENDPOINTS.SETTINGS.PatchSetting.key],
        onSuccess: async data => {
            if (data) {
                setServerStatus(data)
                queryClient.setQueryData([API_ENDPOINTS.STATUS.GetStatus.key], data)
                queryClient.setQueryData([API_ENDPOINTS.SETTINGS.GetSettings.key], data.settings)
            }

            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings updated")
        },
    })
}

export function useSaveAutoDownloaderSettings() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, SaveAutoDownloaderSettings_Variables>({
        endpoint: API_ENDPOINTS.SETTINGS.SaveAutoDownloaderSettings.endpoint,
        method: API_ENDPOINTS.SETTINGS.SaveAutoDownloaderSettings.methods[0],
        mutationKey: [API_ENDPOINTS.SETTINGS.SaveAutoDownloaderSettings.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.success("Settings saved")
        },
    })
}

export function useSaveMediaPlayerSettings() {
    const queryClient = useQueryClient()
    return useServerMutation<boolean, SaveMediaPlayerSettings_Variables>({
        endpoint: API_ENDPOINTS.SETTINGS.SaveMediaPlayerSettings.endpoint,
        method: API_ENDPOINTS.SETTINGS.SaveMediaPlayerSettings.methods[0],
        mutationKey: [API_ENDPOINTS.SETTINGS.SaveMediaPlayerSettings.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.SETTINGS.GetSettings.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
        },
    })
}
