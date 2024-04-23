import { serverStatusAtom } from "@/app/(main)/_atoms/server-status"
import { useAtomValue } from "jotai"

export function useServerStatus() {
    return useAtomValue(serverStatusAtom)
}
