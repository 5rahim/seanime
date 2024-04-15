"use client"

import { ComingUpNext } from "@/app/(main)/schedule/_containers/coming-up-next/coming-up-next"
import { MissingEpisodes } from "@/app/(main)/schedule/_containers/missing-episodes/missing-episodes"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases/recent-releases"
import { useMissingEpisodes } from "@/app/(main)/schedule/_lib/missing-episodes"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { LoadingSpinner } from "@/components/ui/loading-spinner"

export const dynamic = "force-static"

export default function Page() {

    const { missingEpisodes, silencedEpisodes, isLoading } = useMissingEpisodes()

    if (isLoading) return <LoadingSpinner />

    return (
        <PageWrapper
            className="p-4 sm:p-8 space-y-10 pb-10"
        >
            <MissingEpisodes missingEpisodes={missingEpisodes} silencedEpisodes={silencedEpisodes} isLoading={isLoading} />
            <ComingUpNext />
            <RecentReleases />
        </PageWrapper>
    )
}
