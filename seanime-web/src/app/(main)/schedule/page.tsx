"use client"

import { useGetMissingEpisodes } from "@/api/hooks/anime_entries.hooks"
import { MissingEpisodes } from "@/app/(main)/schedule/_components/missing-episodes"
import { ComingUpNext } from "@/app/(main)/schedule/_containers/coming-up-next"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { LoadingSpinner } from "@/components/ui/loading-spinner"

export const dynamic = "force-static"

export default function Page() {

    const { data, isLoading } = useGetMissingEpisodes()

    if (isLoading) return <LoadingSpinner />

    return (
        <PageWrapper
            className="p-4 sm:p-8 space-y-10 pb-10"
        >
            <MissingEpisodes data={data} isLoading={isLoading} />
            <ComingUpNext />
            <RecentReleases />
        </PageWrapper>
    )
}
