import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    FetchExternalExtensionData_Variables,
    GetAllExtensions_Variables,
    InstallExternalExtension_Variables,
    UninstallExternalExtension_Variables,
    UpdateExtensionCode_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    Extension_Extension,
    ExtensionRepo_AllExtensions,
    ExtensionRepo_AnimeTorrentProviderExtensionItem,
    ExtensionRepo_ExtensionInstallResponse,
    ExtensionRepo_MangaProviderExtensionItem,
    ExtensionRepo_OnlinestreamProviderExtensionItem,
    Nullish,
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

        },
    })
}

export function useUninstallExternalExtension() {
    return useServerMutation<boolean, UninstallExternalExtension_Variables>({
        endpoint: API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.methods[0],
        mutationKey: [API_ENDPOINTS.EXTENSIONS.UninstallExternalExtension.key],
        onSuccess: async () => {
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
