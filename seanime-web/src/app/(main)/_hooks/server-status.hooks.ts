import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { useAtomValue } from "jotai"
import React from "react"

export function useServerStatus() {
    return useAtomValue(serverStatusAtom)
}

export function useCurrentUser() {
    const serverStatus = useServerStatus()
    return React.useMemo(() => serverStatus?.user?.viewer, [serverStatus?.user?.viewer])
}
