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
    AL_MangaCollection,
    AL_MangaDetailsById_Media,
    Manga_ChapterContainer,
    Manga_Collection,
    Manga_Entry,
    Manga_PageContainer,
    Nullish,
} from "@/api/generated/types"

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
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key],
        enabled: true,
    })
}

export function useGetMangaEntryDetails(id: Nullish<string | number>) {
    return useServerQuery<AL_MangaDetailsById_Media>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.MANGA.GetMangaEntryDetails.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryDetails.key],
        enabled: true,
    })
}

export function useEmptyMangaEntryCache() {
    return useServerMutation<boolean, EmptyMangaEntryCache_Variables>({
        endpoint: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.endpoint,
        method: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.EmptyMangaEntryCache.key],
        onSuccess: async () => {

        },
    })
}

export function useGetMangaEntryChapters() {
    return useServerMutation<Manga_ChapterContainer, GetMangaEntryChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryChapters.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryChapters.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key],
        onSuccess: async () => {

        },
    })
}

export function useGetMangaEntryPages() {
    return useServerMutation<Manga_PageContainer, GetMangaEntryPages_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryPages.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryPages.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key],
        onSuccess: async () => {

        },
    })
}

export function useAnilistListManga() {
    return useServerMutation<boolean, AnilistListManga_Variables>({
        endpoint: API_ENDPOINTS.MANGA.AnilistListManga.endpoint,
        method: API_ENDPOINTS.MANGA.AnilistListManga.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.AnilistListManga.key],
        onSuccess: async () => {

        },
    })
}

export function useUpdateMangaProgress() {
    return useServerMutation<boolean, UpdateMangaProgress_Variables>({
        endpoint: API_ENDPOINTS.MANGA.UpdateMangaProgress.endpoint,
        method: API_ENDPOINTS.MANGA.UpdateMangaProgress.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.UpdateMangaProgress.key],
        onSuccess: async () => {

        },
    })
}

