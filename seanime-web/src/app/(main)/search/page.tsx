"use client"
import { AL_MediaFormat, AL_MediaSeason, AL_MediaSort, AL_MediaStatus } from "@/api/generated/types"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AdvancedSearchList } from "@/app/(main)/search/_components/advanced-search-list"
import { AdvancedSearchOptions } from "@/app/(main)/search/_components/advanced-search-options"
import { AdvancedSearchPageTitle } from "@/app/(main)/search/_components/advanced-search-page-title"
import { __advancedSearch_paramsAtom } from "@/app/(main)/search/_lib/advanced-search.atoms"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AppLayoutGrid } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { useSetAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { useMount } from "react-use"

export const dynamic = "force-static"

export default function Page({ params: urlParams }: {
    params: {
        sorting?: AL_MediaSort,
        genre?: string,
        format?: AL_MediaFormat,
        season?: AL_MediaSeason,
        status?: AL_MediaStatus,
        year?: string
    }
}) {

    const setParams = useSetAtom(__advancedSearch_paramsAtom)

    useMount(() => {
        setParams({
            active: true,
            title: null,
            sorting: urlParams.sorting ? [urlParams.sorting] : null,
            status: urlParams.status ? [urlParams.status] : null,
            genre: urlParams.genre ? [urlParams.genre] : null,
            format: urlParams.format || null,
            season: urlParams.season || null,
            year: urlParams.year || null,
            minScore: null,
            isAdult: false,
            type: "anime",
        })
    })

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper className="space-y-6 px-4 md:p-8 pt-0 pb-10">
                <div className="flex items-center gap-4">
                    <Link href={`/discover`}>
                        <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="sm" />
                    </Link>
                    <h3>Discover</h3>
                </div>
                <div className="text-center xl:text-left">
                    <AdvancedSearchPageTitle />
                </div>
                <AppLayoutGrid cols={6} spacing="lg">
                    <AdvancedSearchOptions />
                    <div className="col-span-5">
                        <AdvancedSearchList />
                    </div>
                </AppLayoutGrid>
            </PageWrapper>
        </>
    )
}
