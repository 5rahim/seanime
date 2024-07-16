import { Anime_AnimeEntryEpisode } from "@/api/generated/types"
import { atom } from "jotai"

export const missingEpisodesAtom = atom<Anime_AnimeEntryEpisode[]>([])

export const missingSilencedEpisodesAtom = atom<Anime_AnimeEntryEpisode[]>([])

export const missingEpisodeCountAtom = atom(get => get(missingEpisodesAtom).length)

