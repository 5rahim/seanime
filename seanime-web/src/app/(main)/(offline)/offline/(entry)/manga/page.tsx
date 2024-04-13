"use client"

import { OfflineMetaSection } from "@/app/(main)/(offline)/offline/(entry)/_components/offline-meta-section"
import { OfflineChapterList } from "@/app/(main)/(offline)/offline/(entry)/manga/_components/offline-chapter-list"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { LuffyError } from "@/components/shared/luffy-error"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"

export default function Page() {

    const router = useRouter()
    const mediaId = useSearchParams().get("id")
    const { snapshot, isLoading } = useOfflineSnapshot()

    const entry = React.useMemo(() => {
        return snapshot?.entries?.mangaEntries?.find(n => n?.mediaId === Number(mediaId))
    }, [snapshot, mediaId])

    if (isLoading) return <LoadingSpinner />

    if (!entry) return <LuffyError title="Not found" />

    return (
        <>
            <OfflineMetaSection type="manga" entry={entry} assetMap={snapshot?.assetMap} />
            <OfflineChapterList entry={entry} />
        </>
    )

}
