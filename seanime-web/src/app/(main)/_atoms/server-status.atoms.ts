import { Status } from "@/api/generated/types"
import { atom } from "jotai"
import { atomWithImmer } from "jotai-immer"
import { atomWithStorage } from "jotai/utils"

export const serverStatusAtom = atomWithImmer<Status | undefined>(undefined)

export const isLoginModalOpenAtom = atom(false)

export const serverPasswordAtom = atomWithStorage<string | undefined>("sea-server-password", undefined, undefined, { getOnInit: true })
