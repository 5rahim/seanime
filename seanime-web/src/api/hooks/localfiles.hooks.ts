import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { DeleteLocalFiles_Variables, LocalFileBulkAction_Variables, UpdateLocalFileData_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_LocalFile, Nullish } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetLocalFiles() {
    return useServerQuery<Array<Anime_LocalFile>>({
        endpoint: API_ENDPOINTS.LOCALFILES.GetLocalFiles.endpoint,
        method: API_ENDPOINTS.LOCALFILES.GetLocalFiles.methods[0],
        queryKey: [API_ENDPOINTS.LOCALFILES.GetLocalFiles.key],
        enabled: true,
    })
}

export function useLocalFileBulkAction() {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>, LocalFileBulkAction_Variables>({
        endpoint: API_ENDPOINTS.LOCALFILES.LocalFileBulkAction.endpoint,
        method: API_ENDPOINTS.LOCALFILES.LocalFileBulkAction.methods[0],
        mutationKey: [API_ENDPOINTS.LOCALFILES.LocalFileBulkAction.key],
        onSuccess: async () => {
            toast.success("Action completed")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry] })
        },
    })
}

export function useUpdateLocalFileData() {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>, UpdateLocalFileData_Variables>({
        endpoint: API_ENDPOINTS.LOCALFILES.UpdateLocalFileData.endpoint,
        method: API_ENDPOINTS.LOCALFILES.UpdateLocalFileData.methods[0],
        mutationKey: [API_ENDPOINTS.LOCALFILES.UpdateLocalFileData.key],
        onSuccess: async () => {
            toast.success("File metadata updated")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry] })
        },
    })
}

export function useDeleteLocalFiles(id: Nullish<number>) {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>, DeleteLocalFiles_Variables>({
        endpoint: API_ENDPOINTS.LOCALFILES.DeleteLocalFiles.endpoint,
        method: API_ENDPOINTS.LOCALFILES.DeleteLocalFiles.methods[0],
        mutationKey: [API_ENDPOINTS.LOCALFILES.DeleteLocalFiles.key],
        onSuccess: async () => {
            toast.success("Files deleted")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection] })
            if (id) {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry, id] })
            }
        },
    })
}

export function useRemoveEmptyDirectories() {
    return useServerMutation<boolean>({
        endpoint: API_ENDPOINTS.LOCALFILES.RemoveEmptyDirectories.endpoint,
        method: API_ENDPOINTS.LOCALFILES.RemoveEmptyDirectories.methods[0],
        mutationKey: [API_ENDPOINTS.LOCALFILES.RemoveEmptyDirectories.key],
        onSuccess: async () => {
            toast.success("Empty directories removed")
        },
    })
}

