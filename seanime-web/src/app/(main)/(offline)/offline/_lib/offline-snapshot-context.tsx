"use client"
import { useHandleOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/handle-offline-snapshot"
import React from "react"
import { undefined } from "zod"


const __offlineSnapshotContext = React.createContext<ReturnType<typeof useHandleOfflineSnapshot>>({
    animeLists: {},
    continueWatchingEpisodeList: [],
    isLoading: true,
    // @ts-expect-error
    snapshot: undefined,
})

export function OfflineSnapshotProvider({ children }: { children?: React.ReactNode }) {

    const opts = useHandleOfflineSnapshot()


    return (
        <__offlineSnapshotContext.Provider
            value={opts}
        >
            {children}
        </__offlineSnapshotContext.Provider>
    )

}

export function useOfflineSnapshot() {
    return React.useContext(__offlineSnapshotContext)
}
