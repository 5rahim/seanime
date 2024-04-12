"use client"
import { useGetOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.hooks"
import { OfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import React from "react"

const __offlineSnapshotContext = React.createContext<OfflineSnapshot | undefined>(undefined)

export function OfflineSnapshotProvider({ children }: { children?: React.ReactNode }) {

    const { snapshot } = useGetOfflineSnapshot()


    return (
        <__offlineSnapshotContext.Provider
            value={snapshot}
        >
            {children}
        </__offlineSnapshotContext.Provider>
    )

}

export function useOfflineSnapshot() {
    return React.useContext(__offlineSnapshotContext)
}
