"use client"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { MalAuthResponse } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import React from "react"

export default function Page() {

    const router = useRouter()
    const qc = useQueryClient()

    const { code, state, challenge } = React.useMemo(() => {
        const urlParams = new URLSearchParams(window.location.search)
        const code = urlParams.get("code")
        const state = urlParams.get("state")
        const challenge = sessionStorage.getItem("mal-" + state)
        return { code, state, challenge }
    }, [])

    const { data, isError } = useSeaQuery<MalAuthResponse>({
        queryKey: ["mal-auth"],
        endpoint: SeaEndpoints.MAL_AUTH,
        method: "post",
        data: {
            code: code,
            state: state,
            code_verifier: challenge,
        },
        enabled: !!code && !!state && !!challenge,
    })

    React.useEffect(() => {
        if (!!data?.access_token) {
            (async function () {
                await qc.refetchQueries({ queryKey: ["status"] })
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
            <LoadingOverlay className={"fixed w-full h-full z-[80]"}>
                <h3 className={"mt-2"}>Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
