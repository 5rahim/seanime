"use client"

import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useRouter } from "next/navigation"
import React from "react"

export default function Layout({ children }: { children?: React.ReactNode }) {

    const router = useRouter()
    const status = useServerStatus()

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
