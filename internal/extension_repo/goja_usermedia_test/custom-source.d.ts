/// <reference path="../goja_plugin_types/app.d.ts" />

declare type Settings = {
    supportsAnime: boolean
    supportsManga: boolean
}

declare type ListResponse<T extends $app.AL_BaseAnime | $app.AL_BaseManga> = {
    media: T[]
    page: number
    totalPages: number
    total: number
}

declare abstract class CustomSource {
    getSettings(): Settings

    async getAnime(ids: number[]): Promise<$app.AL_BaseAnime[]>

    async getAnimeMetadata(id: number): Promise<$app.Metadata_AnimeMetadata | null>

    async getAnimeWithRelations(id: number): Promise<$app.AL_CompleteAnime>

    async getAnimeDetails(id: number): Promise<$app.AL_AnimeDetailsById_Media | null>

    async getManga(ids: number[]): Promise<$app.AL_BaseManga[]>

    async listAnime(search: string, page: number, perPage: number): Promise<ListResponse<$app.AL_BaseAnime>>

    async getMangaDetails(id: number): Promise<$app.AL_MangaDetailsById_Media | null>

    async listManga(search: string, page: number, perPage: number): Promise<ListResponse<$app.AL_BaseManga>>
}
