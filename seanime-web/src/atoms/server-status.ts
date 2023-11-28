import { ServerStatus } from "@/lib/server/types"
import { atomWithImmer } from "jotai-immer"

export const serverStatusAtom = atomWithImmer<ServerStatus | undefined>(undefined)