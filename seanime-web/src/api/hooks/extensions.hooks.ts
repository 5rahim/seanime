import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    ExtensionRepo_MangaProviderExtensionItem,
    ExtensionRepo_OnlinestreamProviderExtensionItem,
    ExtensionRepo_TorrentProviderExtensionItem,
} from "@/api/generated/types"

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

export function useListTorrentProviderExtensions() {
    return useServerQuery<Array<ExtensionRepo_TorrentProviderExtensionItem>>({
        endpoint: API_ENDPOINTS.EXTENSIONS.ListTorrentProviderExtensions.endpoint,
        method: API_ENDPOINTS.EXTENSIONS.ListTorrentProviderExtensions.methods[0],
        queryKey: [API_ENDPOINTS.EXTENSIONS.ListTorrentProviderExtensions.key],
        enabled: true,
    })
}
