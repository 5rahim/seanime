import { autoDownloaderItemCountAtom } from "@/app/(main)/_atoms/autodownloader.atoms"
import { useAtomValue } from "jotai/react"

export function useAutoDownloaderQueueCount() {
    return useAtomValue(autoDownloaderItemCountAtom)
}

