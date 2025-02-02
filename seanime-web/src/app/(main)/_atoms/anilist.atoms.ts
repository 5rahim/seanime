import { AL_BaseAnime, Anime_EntryListData, Manga_EntryListData } from "@/api/generated/types"
import { atom } from "jotai/index"

export const __anilist_userAnimeMediaAtom = atom<AL_BaseAnime[] | undefined>(undefined)

// e.g. { "123": { ... } }
export const __anilist_userAnimeListDataAtom = atom<Record<string, Anime_EntryListData>>({})

export const __anilist_userMangaListDataAtom = atom<Record<string, Manga_EntryListData>>({})
