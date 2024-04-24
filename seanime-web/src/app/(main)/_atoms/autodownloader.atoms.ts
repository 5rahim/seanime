import { Models_AutoDownloaderItem } from "@/api/generated/types"
import { atom } from "jotai/index"

export const autoDownloaderItemsAtom = atom<Models_AutoDownloaderItem[]>([])
export const autoDownloaderItemCountAtom = atom(get => get(autoDownloaderItemsAtom).length)
