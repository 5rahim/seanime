import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    DeleteMangaDownloadedChapters_Variables,
    DownloadMangaChapters_Variables,
    GetMangaDownloadData_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Manga_DownloadListItem, Manga_MediaDownloadData, Models_ChapterDownloadQueueItem, Nullish } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useDownloadMangaChapters(id: Nullish<string | number>, provider: Nullish<string>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, DownloadMangaChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.key, String(id), provider],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key] })
        },
    })
}

export function useGetMangaDownloadData(variables: Partial<GetMangaDownloadData_Variables>) {
    return useServerQuery<Manga_MediaDownloadData, GetMangaDownloadData_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.methods[0],
        queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key, String(variables.mediaId), String(variables.cached)],
        data: variables as GetMangaDownloadData_Variables,
        enabled: !!variables.mediaId,
    })
}

export function useGetMangaDownloadQueue() {
    return useServerQuery<Array<Models_ChapterDownloadQueueItem>>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.methods[0],
        queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key],
        enabled: true,
    })
}

export function useStartMangaDownloadQueue() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key] })
            toast.info("Downloading chapters")
        },
    })
}

export function useStopMangaDownloadQueue() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key] })
            toast.success("Download queue stopped")
        },
    })
}

export function useClearAllChapterDownloadQueue() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key] })
            toast.success("Download queue cleared")
        },
    })
}

export function useResetErroredChapterDownloadQueue() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key] })
            toast.success("Reset errored chapters")
        },
    })
}

export function useDeleteMangaDownloadedChapters(id: Nullish<string | number>, provider: string | null) {
    const queryClient = useQueryClient()
    return useServerMutation<boolean, DeleteMangaDownloadedChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaDownloadedChapters.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaDownloadedChapters.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaDownloadedChapters.key, String(id), provider],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key] })
            toast.success("Chapters deleted")
        },
    })
}

export function useGetMangaDownloadsList() {

    return useServerQuery<Array<Manga_DownloadListItem>>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadsList.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadsList.methods[0],
        queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadsList.key],
        enabled: true,
    })
}

