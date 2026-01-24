import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    CreateAutoDownloaderRule_Variables,
    DeleteAutoDownloaderItem_Variables,
    RunAutoDownloaderSimulation_Variables,
    UpdateAutoDownloaderRule_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    Anime_AutoDownloaderProfile,
    Anime_AutoDownloaderRule,
    AutoDownloader_SimulationResult,
    Models_AutoDownloaderItem,
    Nullish,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useRunAutoDownloader() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloader.key],
        onSuccess: async () => {
            toast.success("Auto downloader started")
            setTimeout(() => {
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
            }, 1000)
        },
    })
}

export function useGetAutoDownloaderRule(id: number) {
    return useServerQuery<Anime_AutoDownloaderRule>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRule.endpoint.replace("{id}", String(id)),
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

    return useServerMutation<Anime_AutoDownloaderRule, CreateAutoDownloaderRule_Variables>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderRule.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.key] })
            toast.success("Rule created")
        },
    })
}

export function useUpdateAutoDownloaderRule() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoDownloaderRule, UpdateAutoDownloaderRule_Variables>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderRule.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.key] })
            toast.success("Rule updated")
        },
    })
}

export function useDeleteAutoDownloaderRule(id: Nullish<number>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderRule.key, String(id)],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRules.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.key] })
            toast.success("Rule deleted")
        },
    })
}

export function useGetAutoDownloaderItems(enabled: boolean = true) {
    return useServerQuery<Array<Models_AutoDownloaderItem>>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key],
        enabled: enabled,
    })
}

export function useDeleteAutoDownloaderItem() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, DeleteAutoDownloaderItem_Variables>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderItem.key],
        onSuccess: async () => {
            toast.success("Item deleted")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.key] })
        },
    })
}

export function useGetAutoDownloaderRulesByAnime(id: number, enabled: boolean) {
    return useServerQuery<Array<Anime_AutoDownloaderRule>>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderRulesByAnime.key, String(id)],
        enabled: enabled,
    })
}

export function useGetAutoDownloaderProfiles() {
    return useServerQuery<Array<Anime_AutoDownloaderProfile>>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.key],
        enabled: true,
    })
}

export function useGetAutoDownloaderProfile(id: Nullish<number>) {
    return useServerQuery<Anime_AutoDownloaderProfile>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfile.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfile.methods[0],
        queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfile.key, String(id)],
        enabled: !!id,
    })
}

export function useCreateAutoDownloaderProfile() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoDownloaderProfile, Anime_AutoDownloaderProfile>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderProfile.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderProfile.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.CreateAutoDownloaderProfile.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.key] })
            toast.success("Profile created")
        },
    })
}

export function useUpdateAutoDownloaderProfile() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_AutoDownloaderProfile, Anime_AutoDownloaderProfile>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderProfile.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderProfile.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.UpdateAutoDownloaderProfile.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.key] })
            toast.success("Profile updated")
        },
    })
}

export function useDeleteAutoDownloaderProfile(id: Nullish<number>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderProfile.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderProfile.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.DeleteAutoDownloaderProfile.key, String(id)],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderProfiles.key] })
            toast.success("Profile deleted")
        },
    })
}

export function useRunAutoDownloaderSimulation() {
    return useServerMutation<Array<AutoDownloader_SimulationResult>, RunAutoDownloaderSimulation_Variables>({
        endpoint: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloaderSimulation.endpoint,
        method: API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloaderSimulation.methods[0],
        mutationKey: [API_ENDPOINTS.AUTO_DOWNLOADER.RunAutoDownloaderSimulation.key],
        onSuccess: async () => {

        },
    })
}
