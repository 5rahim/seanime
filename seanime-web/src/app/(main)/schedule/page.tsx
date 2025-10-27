"use client"

import { useGetMissingEpisodes } from "@/api/hooks/anime_entries.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MissingEpisodes } from "@/app/(main)/schedule/_components/missing-episodes"
import { ComingUpNext } from "@/app/(main)/schedule/_containers/coming-up-next"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
import { ScheduleCalendar } from "./_components/schedule-calendar"

export const dynamic = "force-static"

export default function Page() {

    const { data, isLoading } = useGetMissingEpisodes()

    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper
                className="p-4 sm:p-8 space-y-10 pb-10"
            >
                <MissingEpisodes data={data} isLoading={isLoading} />
                <AppLayoutStack>

                    <div className="hidden lg:block space-y-2">
                        <h2>Release schedule</h2>
                        <p className="text-[--muted]">Based on your anime list</p>
                    </div>

                    <ScheduleCalendar />
                </AppLayoutStack>
                <ComingUpNext />
            </PageWrapper>
        </>
    )
}
