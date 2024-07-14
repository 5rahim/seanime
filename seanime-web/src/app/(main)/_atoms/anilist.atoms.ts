import { AL_BaseAnime } from "@/api/generated/types"
import { atom } from "jotai/index"

export const anilistUserMediaAtom = atom<AL_BaseAnime[] | undefined>(undefined)
