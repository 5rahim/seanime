"use client"
import { useGetOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.hooks"
import React from "react"
import { undefined } from "zod"

type OfflineSnapshotContextProps = ReturnType<typeof useGetOfflineSnapshot>

const __offlineSnapshotContext = React.createContext<OfflineSnapshotContextProps>({
    animeLists: {},
    continueWatchingEpisodeList: [],
    isLoading: true,
    //@ts-expect-error
    snapshot: undefined,
})

export function OfflineSnapshotProvider({ children }: { children?: React.ReactNode }) {

    const opts = useGetOfflineSnapshot()


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
