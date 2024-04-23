import { Anime_MediaEntryEpisode } from "@/api/generated/types"
import { atom } from "jotai"

export const missingEpisodesAtom = atom<Anime_MediaEntryEpisode[]>([])

export const missingSilencedEpisodesAtom = atom<Anime_MediaEntryEpisode[]>([])

export const missingEpisodeCountAtom = atom(get => get(missingEpisodesAtom).length)

