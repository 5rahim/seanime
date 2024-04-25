import { AL_BaseMedia } from "@/api/generated/types"
import { atom } from "jotai/index"

export const anilistUserMediaAtom = atom<AL_BaseMedia[] | undefined>(undefined)
