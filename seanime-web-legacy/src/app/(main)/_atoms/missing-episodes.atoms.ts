import { Anime_Episode } from "@/api/generated/types"
import { atom } from "jotai"

export const missingEpisodesAtom = atom<Anime_Episode[]>([])

export const missingSilencedEpisodesAtom = atom<Anime_Episode[]>([])

export const missingEpisodeCountAtom = atom(get => get(missingEpisodesAtom).length)

