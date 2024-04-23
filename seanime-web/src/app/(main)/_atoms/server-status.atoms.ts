import { Status } from "@/api/generated/types"
import { atomWithImmer } from "jotai-immer"

export const serverStatusAtom = atomWithImmer<Status | undefined>(undefined)
