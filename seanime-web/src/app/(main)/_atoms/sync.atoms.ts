import { atom } from "jotai"
import { useAtom } from "jotai/react"

const syncIsActiveAtom = atom(false)

export function useSyncIsActive() {
    const [syncIsActive, setSyncIsActive] = useAtom(syncIsActiveAtom)
    return { syncIsActive, setSyncIsActive }
}
