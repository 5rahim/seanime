"use client"

import { ComingUpNext } from "@/app/(main)/schedule/_containers/coming-up-next/coming-up-next"
import { MissingEpisodes } from "@/app/(main)/schedule/_containers/missing-episodes/missing-episodes"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases/recent-releases"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useMissingEpisodes } from "@/lib/server/hooks/library"

export default function Page() {

    const { missingEpisodes, isLoading } = useMissingEpisodes()

    if (isLoading) return <LoadingSpinner />

    return (
        <div className={"p-8 space-y-10 pb-10"}>
            <MissingEpisodes missingEpisodes={missingEpisodes} isLoading={isLoading} />
            <ComingUpNext/>
            <RecentReleases/>
        </div>
    )
}
