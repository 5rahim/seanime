import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import React from "react"

export function useServerStatus() {
    return useAtomValue(serverStatusAtom)
}

export function useSetServerStatus() {
    return useSetAtom(serverStatusAtom)
}

export function useCurrentUser() {
    const serverStatus = useServerStatus()
    return React.useMemo(() => serverStatus?.user?.viewer, [serverStatus?.user?.viewer])
}

export function useHasTorrentProvider() {
    const serverStatus = useServerStatus()
    return {
        hasTorrentProvider: React.useMemo(() => !!serverStatus?.settings?.library?.torrentProvider && serverStatus?.settings?.library?.torrentProvider !== TORRENT_PROVIDER.NONE,
            [serverStatus?.settings?.library?.torrentProvider]),
    }
}

export function useHasDebridService() {
    const serverStatus = useServerStatus()
    return {
        hasDebridService: React.useMemo(() => !!serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider,
            [serverStatus?.debridSettings]),
    }
}
