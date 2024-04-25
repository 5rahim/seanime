import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    AnilistListManga_Variables,
    EmptyMangaEntryCache_Variables,
    GetAnilistMangaCollection_Variables,
    GetMangaEntryChapters_Variables,
    GetMangaEntryPages_Variables,
    UpdateMangaProgress_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    AL_ListManga,
    AL_MangaCollection,
    AL_MangaDetailsById_Media,
    Manga_ChapterContainer,
    Manga_Collection,
    Manga_Entry,
    Manga_PageContainer,
    Nullish,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"

export function useGetAnilistMangaCollection() {
    return useServerQuery<AL_MangaCollection, GetAnilistMangaCollection_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetAnilistMangaCollection.endpoint,
        method: API_ENDPOINTS.MANGA.GetAnilistMangaCollection.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key],
        enabled: true,
    })
}

export function useGetMangaCollection() {
    return useServerQuery<Manga_Collection>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaCollection.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaCollection.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key],
        enabled: true,
    })
}

export function useGetMangaEntry(id: Nullish<string | number>) {
    return useServerQuery<Manga_Entry>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntry.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.MANGA.GetMangaEntry.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)],
        enabled: !!id,
    })
}

export function useGetMangaEntryDetails(id: Nullish<string | number>) {
    return useServerQuery<AL_MangaDetailsById_Media>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.MANGA.GetMangaEntryDetails.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryDetails.key, String(id)],
        enabled: !!id,
    })
}

export function useEmptyMangaEntryCache() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, EmptyMangaEntryCache_Variables>({
        endpoint: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.endpoint,
        method: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.EmptyMangaEntryCache.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key] })
        },
    })
}

export function useGetMangaEntryChapters(variables: Partial<GetMangaEntryChapters_Variables>) {
    return useServerQuery<Manga_ChapterContainer, GetMangaEntryChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryChapters.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryChapters.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key, String(variables.mediaId), variables.provider],
        data: variables as GetMangaEntryChapters_Variables,
        enabled: !!variables.mediaId && !!variables.provider,
    })
}

export function useGetMangaEntryPages(variables: Partial<GetMangaEntryPages_Variables>) {
    return useServerQuery<Manga_PageContainer, GetMangaEntryPages_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryPages.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryPages.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key, String(variables.mediaId), variables.provider, variables.chapterId,
            variables.doublePage],
        data: variables as GetMangaEntryPages_Variables,
        enabled: !!variables.mediaId && !!variables.provider && !!variables.chapterId,
    })
}

export function useAnilistListManga(variables: AnilistListManga_Variables) {
    return useServerQuery<AL_ListManga, AnilistListManga_Variables>({
        endpoint: API_ENDPOINTS.MANGA.AnilistListManga.endpoint,
        method: API_ENDPOINTS.MANGA.AnilistListManga.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.AnilistListManga.key, variables],
        data: variables,
    })
}

export function useUpdateMangaProgress(id: Nullish<string | number>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, UpdateMangaProgress_Variables>({
        endpoint: API_ENDPOINTS.MANGA.UpdateMangaProgress.endpoint,
        method: API_ENDPOINTS.MANGA.UpdateMangaProgress.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.UpdateMangaProgress.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
        },
    })
}

