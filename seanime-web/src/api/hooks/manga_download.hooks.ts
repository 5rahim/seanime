import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { DeleteMangaChapterDownload_Variables, DownloadMangaChapters_Variables, GetMangaDownloadData_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Manga_DownloadListItem, Manga_MediaDownloadData, Models_ChapterDownloadQueueItem } from "@/api/generated/types"

export function useDownloadMangaChapters() {
    return useServerMutation<boolean, DownloadMangaChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.DownloadMangaChapters.key],
        onSuccess: async () => {

        },
    })
}

export function useGetMangaDownloadData() {
    return useServerMutation<Manga_MediaDownloadData, GetMangaDownloadData_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key],
        onSuccess: async () => {

        },
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
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.StartMangaDownloadQueue.key],
        onSuccess: async () => {

        },
    })
}

export function useStopMangaDownloadQueue() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.StopMangaDownloadQueue.key],
        onSuccess: async () => {

        },
    })
}

export function useClearAllChapterDownloadQueue() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.ClearAllChapterDownloadQueue.key],
        onSuccess: async () => {

        },
    })
}

export function useResetErroredChapterDownloadQueue() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.ResetErroredChapterDownloadQueue.key],
        onSuccess: async () => {

        },
    })
}

export function useDeleteMangaChapterDownload() {
    return useServerMutation<boolean, DeleteMangaChapterDownload_Variables>({
        endpoint: API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaChapterDownload.endpoint,
        method: API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaChapterDownload.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA_DOWNLOAD.DeleteMangaChapterDownload.key],
        onSuccess: async () => {

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

