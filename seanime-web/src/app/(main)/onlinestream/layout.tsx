"use client"

import { serverStatusAtom } from "@/atoms/server-status"
import { useAtomValue } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"

export default function Layout({ children }: { children?: React.ReactNode }) {

    const router = useRouter()
    const status = useAtomValue(serverStatusAtom)

    React.useEffect(() => {
        if (!status?.settings?.library?.enableOnlinestream) {
            router.replace("/")
        }
    }, [status?.settings?.library?.enableOnlinestream])

    if (!status?.settings?.library?.enableOnlinestream) return null

    return <>
        {children}
    </>
}

export const dynamic = "force-static"
