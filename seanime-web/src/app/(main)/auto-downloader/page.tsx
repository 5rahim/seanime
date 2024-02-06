"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { BetaBadge } from "@/components/application/beta-badge"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import React from "react"


export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()


    // const { data, isLoading } = useSeaQuery<ScanSummary[] | null>({
    //     queryKey: ["scan-summaries"],
    //     endpoint: SeaEndpoints.SCAN_SUMMARIES,
    // })


    return (
        <div className="p-12 space-y-4">
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>Auto Downloader <BetaBadge /></h2>
                    <p className="text-[--muted]">
                        Add and manage auto-downloading rules for your favorite anime.
                    </p>
                </div>
            </div>

            <div className="border border-[--border] rounded-[--radius] bg-[--paper] text-lg space-y-2 p-4">
                {/*{isLoading && <LoadingSpinner />}*/}
                {/*{(!isLoading && !data?.length) && <div className="p-4 text-[--muted] text-center">No scan summaries available</div>}*/}

            </div>
        </div>
    )

}
