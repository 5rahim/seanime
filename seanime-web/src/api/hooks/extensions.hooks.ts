import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    FetchExternalExtensionData_Variables,
    GetAllExtensions_Variables,
    InstallExternalExtension_Variables,
    RunExtensionPlaygroundCode_Variables,
    SaveExtensionUserConfig_Variables,
    UninstallExternalExtension_Variables,
    UpdateExtensionCode_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    Extension_Extension,
    ExtensionRepo_AllExtensions,
    ExtensionRepo_AnimeTorrentProviderExtensionItem,
    ExtensionRepo_ExtensionInstallResponse,
    ExtensionRepo_ExtensionUserConfig,
    ExtensionRepo_MangaProviderExtensionItem,
    ExtensionRepo_OnlinestreamProviderExtensionItem,
    Nullish,
    RunPlaygroundCodeResponse,
} from "@/api/generated/types"
import { toast } from "sonner"

export function useGetAllExtensions(withUpdates: boolean) {
    return useServerQuery<ExtensionRepo_AllExtensions, GetAllExtensions_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.GetAllExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.GetAllExtensions.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.GetAllExtensions.key],
        data: {
            withUpdates: withUpdates,
        },
        gcTime: 0,
        enabled: true,
    })
}

export function useFetchExternalExtensionData(id: Nullish<string>) {
    return useServerMutation<Extension_Extension, FetchExternalExtensionData_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.FetchExternalExtensionData.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.FetchExternalExtensionData.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.FetchExternalExtensionData.key, id],
        onSuccess: async () => {

        },
    })
}

export function useInstallExternalExtension() {
    return useServerMutation<ExtensionRepo_ExtensionInstallResponse, InstallExternalExtension_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.InstallExternalExtension.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.InstallExternalExtension.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.InstallExternalExtension.key],
        onSuccess: async () => {
            // DEVNOTE: No need to refetch, the websocket listener will do it
        },
    })
}

export function useUninstallExternalExtension() {
    return useServerMutation<boolean, UninstallExternalExtension_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.key],
        onSuccess: async () => {
            // DEVNOTE: No need to refetch, the websocket listener will do it
            toast.success("Extension uninstalled successfully.")
        },
    })
}

export function useUpdateExtensionCode() {
    return useServerMutation<boolean, UpdateExtensionCode_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.UpdateExtensionCode.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.UpdateExtensionCode.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.UpdateExtensionCode.key],
        onSuccess: async () => {
            // DEVNOTE: No need to refetch, the websocket listener will do it
            toast.success("Extension updated successfully.")
        },
    })
}

export function useReloadExternalExtensions() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ReloadExternalExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ReloadExternalExtensions.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.ReloadExternalExtensions.key],
        onSuccess: async () => {

        },
    })
}

export function useListExtensionData() {
    return useServerQuery<Array<Extension_Extension>>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ListExtensionData.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ListExtensionData.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.ListExtensionData.key],
        enabled: true,
    })
}

export function useListMangaProviderExtensions() {
    return useServerQuery<Array<ExtensionRepo_MangaProviderExtensionItem>>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ListMangaProviderExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ListMangaProviderExtensions.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.ListMangaProviderExtensions.key],
        enabled: true,
    })
}

export function useListOnlinestreamProviderExtensions() {
    return useServerQuery<Array<ExtensionRepo_OnlinestreamProviderExtensionItem>>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ListOnlinestreamProviderExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ListOnlinestreamProviderExtensions.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.ListOnlinestreamProviderExtensions.key],
        enabled: true,
    })
}

export function useAnimeListTorrentProviderExtensions() {
    return useServerQuery<Array<ExtensionRepo_AnimeTorrentProviderExtensionItem>>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ListAnimeTorrentProviderExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ListAnimeTorrentProviderExtensions.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.ListAnimeTorrentProviderExtensions.key],
        enabled: true,
    })
}

export function useRunExtensionPlaygroundCode() {
    return useServerMutation<RunPlaygroundCodeResponse, RunExtensionPlaygroundCode_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.RunExtensionPlaygroundCode.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.RunExtensionPlaygroundCode.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.RunExtensionPlaygroundCode.key],
        onSuccess: async () => {

        },
    })
}

export function useGetExtensionUserConfig(id: string) {
    return useServerQuery<ExtensionRepo_ExtensionUserConfig>({
        endpoint: API_ENDPOINTS.EXTENSIONS.GetExtensionUserConfig.endpoint.replace("{id}", id),
        method: API_ENDPOINTS.EXTENSIONS.GetExtensionUserConfig.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.GetExtensionUserConfig.key, id],
        enabled: true,
        gcTime: 0,
    })
}

export function useSaveExtensionUserConfig() {
    return useServerMutation<boolean, SaveExtensionUserConfig_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.SaveExtensionUserConfig.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.SaveExtensionUserConfig.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.SaveExtensionUserConfig.key],
        onSuccess: async () => {
            // DEVNOTE: No need to refetch, the websocket listener will do it
            toast.success("Config saved successfully.")
        },
    })
}
