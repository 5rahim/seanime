import { serverAuthTokenAtom, serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { createNakamaHMACAuth, createServerPasswordHMACAuth } from "@/lib/server/hmac-auth"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/index"
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
    return React.useMemo(() => serverStatus?.user, [serverStatus?.user])
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

export function useServerPassword() {
    const serverStatus = useServerStatus()
    const [password] = useAtom(serverAuthTokenAtom)
    return {
        getServerPasswordQueryParam: (symbol?: string) => {
            if (!serverStatus?.serverHasPassword) return ""
            return `${symbol ? `${symbol}` : "?"}password=${password ?? ""}`
        },
    }
}

export function useServerHMACAuth() {
    const serverStatus = useServerStatus()
    const [password] = useAtom(serverAuthTokenAtom)

    return {
        getHMACTokenQueryParam: async (endpoint: string, symbol?: string) => {
            if (!serverStatus?.serverHasPassword || !password) return ""

            try {
                const hmacAuth = createServerPasswordHMACAuth(password)
                return await hmacAuth.generateQueryParam(endpoint, symbol)
            }
            catch (error) {
                console.error("Failed to generate HMAC token:", error)
                return ""
            }
        },
        generateHMACToken: async (endpoint: string) => {
            if (!serverStatus?.serverHasPassword || !password) return ""

            try {
                const hmacAuth = createServerPasswordHMACAuth(password)
                return await hmacAuth.generateToken(endpoint)
            }
            catch (error) {
                console.error("Failed to generate HMAC token:", error)
                return ""
            }
        },
    }
}

export function useNakamaHMACAuth() {
    const serverStatus = useServerStatus()

    const nakamaPassword = serverStatus?.settings?.nakama?.isHost
        ? serverStatus?.settings?.nakama?.hostPassword
        : serverStatus?.settings?.nakama?.remoteServerPassword

    return {
        getHMACTokenQueryParam: async (endpoint: string, symbol?: string) => {
            if (!serverStatus?.settings?.nakama?.enabled) return ""

            if (!nakamaPassword) return ""

            try {
                const hmacAuth = createNakamaHMACAuth(nakamaPassword)
                return await hmacAuth.generateQueryParam(endpoint, symbol)
            }
            catch (error) {
                console.error("Failed to generate Nakama HMAC token:", error)
                return ""
            }
        },
        generateHMACToken: async (endpoint: string) => {
            if (!serverStatus?.settings?.nakama?.enabled) return ""

            if (!nakamaPassword) return ""

            try {
                const hmacAuth = createNakamaHMACAuth(nakamaPassword)
                return await hmacAuth.generateToken(endpoint)
            }
            catch (error) {
                console.error("Failed to generate Nakama HMAC token:", error)
                return ""
            }
        },
    }
}
