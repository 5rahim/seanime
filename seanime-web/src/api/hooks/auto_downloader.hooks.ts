import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { DeleteAutoDownloaderItem_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_AutoDownloaderRule, Models_AutoDownloaderItem } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"

export function useRunAutoDownloader() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.key],
        onSuccess: async () => {

        },
    })
}

export function useGetAutoDownloaderRule(id: number) {
    return useServerQuery<Anime_AutoDownloaderRule>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRule.endpoint.replace("id", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRule.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRule.key],
        enabled: true,
    })
}

export function useGetAutoDownloaderRules() {
    return useServerQuery<Array<Anime_AutoDownloaderRule>>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key],
        enabled: true,
    })
}

export function useCreateAutoDownloaderRule() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoDownloaderRule>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
        },
    })
}

export function useUpdateAutoDownloaderRule() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoDownloaderRule>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
        },
    })
}

export function useDeleteAutoDownloaderRule(id: number) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.endpoint.replace("id", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
        },
    })
}

export function useGetAutoDownloaderItems() {
    return useServerQuery<Array<Models_AutoDownloaderItem>>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key],
        enabled: true,
    })
}

export function useDeleteAutoDownloaderItem(id: number) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, DeleteAutoDownloaderItem_Variables>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.endpoint.replace("id", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key] })
        },
    })
}

