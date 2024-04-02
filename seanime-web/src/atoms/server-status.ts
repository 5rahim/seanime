import { ServerStatus } from "@/lib/types/server-status.types"
import { atomWithImmer } from "jotai-immer"

export const serverStatusAtom = atomWithImmer<ServerStatus | undefined>(undefined)
