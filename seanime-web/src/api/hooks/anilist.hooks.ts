import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    AnilistListAnime_Variables,
    AnilistListRecentAiringAnime_Variables,
    DeleteAnilistListEntry_Variables,
    EditAnilistListEntry_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    AL_AnimeCollection,
    AL_DeleteEntry,
    AL_ListMedia,
    AL_ListRecentMedia,
    AL_MediaDetailsById_Media,
    AL_UpdateMediaListEntry,
    Nullish,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetAnilistCollection() {
    return useServerQuery<AL_AnimeCollection>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnilistCollection.endpoint,
        method: API_ENDPOINTS.ANILIST.GetAnilistCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAnilistCollection.key],
        enabled: true,
    })
}

// export function useGetAnilistCollection() {
//     return useServerQuery<AL_AnimeCollection>({
//         endpoint: API_ENDPOINTS.ANILIST.GetAnilistCollection.endpoint,
//         method: API_ENDPOINTS.ANILIST.GetAnilistCollection.methods[1],
//         queryKey: [API_ENDPOINTS.ANILIST.GetAnilistCollection.key],
//     })
// }

export function useEditAnilistListEntry(id: Nullish<string | number>, type: "anime" | "manga") {
    const queryClient = useQueryClient()

    return useServerMutation<AL_UpdateMediaListEntry, EditAnilistListEntry_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.EditAnilistListEntry.endpoint,
        method: API_ENDPOINTS.ANILIST.EditAnilistListEntry.methods[0],
        mutationKey: [API_ENDPOINTS.ANILIST.EditAnilistListEntry.key, String(id)],
        onSuccess: async () => {
            toast.success("Entry updated")
            if (type === "anime") {
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnilistCollection] })
            } else if (type === "manga") {
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
            }
        },
    })
}

export function useGetAnilistMediaDetails(id: Nullish<number | string>) {
    return useServerQuery<AL_MediaDetailsById_Media>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnilistMediaDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANILIST.GetAnilistMediaDetails.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAnilistMediaDetails.key],
        enabled: true,
    })
}

export function useDeleteAnilistListEntry(id: Nullish<string | number>, type: "anime" | "manga") {
    const queryClient = useQueryClient()

    return useServerMutation<AL_DeleteEntry, DeleteAnilistListEntry_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.endpoint,
        method: API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.methods[0],
        mutationKey: [API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.key],
        onSuccess: async () => {
            toast.success("Entry deleted")
            if (type === "anime") {
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnilistCollection] })
            } else if (type === "manga") {
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
                await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
            }
        },
    })
}

export function useAnilistListAnime(variables: AnilistListAnime_Variables, enabled: boolean) {
    return useServerQuery<AL_ListMedia, AnilistListAnime_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.AnilistListAnime.endpoint,
        method: API_ENDPOINTS.ANILIST.AnilistListAnime.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.AnilistListAnime.key, variables],
        data: variables,
        enabled: enabled ?? true,
    })
}

export function useAnilistListRecentAiringAnime(variables: AnilistListRecentAiringAnime_Variables) {
    return useServerQuery<AL_ListRecentMedia, AnilistListRecentAiringAnime_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.endpoint,
        method: API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.key],
        data: variables,
    })
}

