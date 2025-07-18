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
    AL_AnimeDetailsById_Media,
    AL_BaseAnime,
    AL_ListAnime,
    AL_ListRecentAnime,
    AL_Stats,
    AL_StudioDetails,
    Nullish,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetAnimeCollection() {
    return useServerQuery<AL_AnimeCollection>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnimeCollection.endpoint,
        method: API_ENDPOINTS.ANILIST.GetAnimeCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key],
        enabled: true,
    })
}

export function useGetRawAnimeCollection() {
    return useServerQuery<AL_AnimeCollection>({
        endpoint: API_ENDPOINTS.ANILIST.GetRawAnimeCollection.endpoint,
        method: API_ENDPOINTS.ANILIST.GetRawAnimeCollection.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key],
        enabled: true,
    })
}

export function useRefreshAnimeCollection() {
    const queryClient = useQueryClient()

    return useServerMutation<AL_AnimeCollection>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnimeCollection.endpoint,
        method: API_ENDPOINTS.ANILIST.GetAnimeCollection.methods[1],
        mutationKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key],
        onSuccess: async () => {
            toast.success("AniList is up-to-date")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key] })
        },
    })
}

export function useEditAnilistListEntry(id: Nullish<string | number>, type: "anime" | "manga") {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, EditAnilistListEntry_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.EditAnilistListEntry.endpoint,
        method: API_ENDPOINTS.ANILIST.EditAnilistListEntry.methods[0],
        mutationKey: [API_ENDPOINTS.ANILIST.EditAnilistListEntry.key, String(id)],
        onSuccess: async () => {
            toast.success("Entry updated")
            if (type === "anime") {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
            } else if (type === "manga") {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
            }
        },
    })
}

export function useGetAnilistAnimeDetails(id: Nullish<number | string>) {
    return useServerQuery<AL_AnimeDetailsById_Media>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnilistAnimeDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANILIST.GetAnilistAnimeDetails.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAnilistAnimeDetails.key, String(id)],
        enabled: !!id,
    })
}

export function useDeleteAnilistListEntry(id: Nullish<string | number>, type: "anime" | "manga", onSuccess: () => void) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, DeleteAnilistListEntry_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.endpoint,
        method: API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.methods[0],
        mutationKey: [API_ENDPOINTS.ANILIST.DeleteAnilistListEntry.key],
        onSuccess: async () => {
            toast.success("Entry deleted")
            if (type === "anime") {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
            } else if (type === "manga") {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
            }
            onSuccess()
        },
    })
}

export function useAnilistListAnime(variables: AnilistListAnime_Variables, enabled: boolean) {
    return useServerQuery<AL_ListAnime, AnilistListAnime_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.AnilistListAnime.endpoint,
        method: API_ENDPOINTS.ANILIST.AnilistListAnime.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.AnilistListAnime.key, variables],
        data: variables,
        enabled: enabled ?? true,
    })
}

export function useAnilistListRecentAiringAnime(variables: AnilistListRecentAiringAnime_Variables, enabled: boolean = true) {
    return useServerQuery<AL_ListRecentAnime, AnilistListRecentAiringAnime_Variables>({
        endpoint: API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.endpoint,
        method: API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.AnilistListRecentAiringAnime.key, JSON.stringify(variables)],
        data: variables,
        enabled: enabled,
    })
}

export function useGetAnilistStudioDetails(id: number) {
    return useServerQuery<AL_StudioDetails>({
        endpoint: API_ENDPOINTS.ANILIST.GetAnilistStudioDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANILIST.GetAnilistStudioDetails.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAnilistStudioDetails.key, String(id)],
        enabled: true,
    })
}

export function useGetAniListStats(enabled: boolean = true) {
    return useServerQuery<AL_Stats>({
        endpoint: API_ENDPOINTS.ANILIST.GetAniListStats.endpoint,
        method: API_ENDPOINTS.ANILIST.GetAniListStats.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.GetAniListStats.key],
        enabled: enabled,
    })
}

export function useAnilistListMissedSequels(enabled: boolean) {
    return useServerQuery<Array<AL_BaseAnime>>({
        endpoint: API_ENDPOINTS.ANILIST.AnilistListMissedSequels.endpoint,
        method: API_ENDPOINTS.ANILIST.AnilistListMissedSequels.methods[0],
        queryKey: [API_ENDPOINTS.ANILIST.AnilistListMissedSequels.key],
        enabled: enabled,
    })
}
