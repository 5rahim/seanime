"use client"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useMALAuth } from "@/api/hooks/mal.hooks"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import React from "react"

export default function _page() {

    const router = useRouter()
    const qc = useQueryClient()

    const { code, state, challenge } = React.useMemo(() => {
        const urlParams = new URLSearchParams(window?.location?.search || "")
        const code = urlParams.get("code") || undefined
        const state = urlParams.get("state") || undefined
        const challenge = sessionStorage.getItem("mal-" + state) || undefined
        return { code, state, challenge }
    }, [])

    const { data, isError } = useMALAuth({
        code: code,
        state: state,
        code_verifier: challenge,
    }, !!code && !!state && !!challenge)

    React.useEffect(() => {
        if (!!data?.access_token) {
            (async function () {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
                router.push("/mal")
            })()
        }
    }, [data])

    React.useEffect(() => {
        if (isError) router.push("/mal")
    }, [isError])

    if (!state || !code || !challenge) return (
        <div className="p-12 space-y-4 text-center">
            Invalid URL or Challenge
        </div>
    )

    return (
        <div>
            <LoadingOverlay className="fixed w-full h-full z-[80]">
                <h3 className="mt-2">Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
