"use client"
import { AL_MediaFormat, AL_MediaSeason, AL_MediaSort, AL_MediaStatus } from "@/api/generated/types"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AdvancedSearchList } from "@/app/(main)/search/_components/advanced-search-list"
import { AdvancedSearchOptions } from "@/app/(main)/search/_components/advanced-search-options"
import { AdvancedSearchPageTitle } from "@/app/(main)/search/_components/advanced-search-page-title"
import { __advancedSearch_paramsAtom } from "@/app/(main)/search/_lib/advanced-search.atoms"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutGrid } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { useSetAtom } from "jotai/react"
import { useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { useMount } from "react-use"

export default function Page() {

    const urlParams = useSearchParams()
    const sortingUrlParam = urlParams.get("sorting")
    const genreUrlParam = urlParams.get("genre")
    const statusUrlParam = urlParams.get("status")
    const formatUrlParam = urlParams.get("format")
    const seasonUrlParam = urlParams.get("season")
    const yearUrlParam = urlParams.get("year")
    const typeUrlParam = urlParams.get("type")

    const setParams = useSetAtom(__advancedSearch_paramsAtom)

    useMount(() => {
        if (sortingUrlParam || genreUrlParam || statusUrlParam || formatUrlParam || seasonUrlParam || yearUrlParam || typeUrlParam) {
            setParams({
                active: true,
                title: null,
                sorting: sortingUrlParam ? [sortingUrlParam as AL_MediaSort] : null,
                status: statusUrlParam ? [statusUrlParam as AL_MediaStatus] : null,
                genre: genreUrlParam ? [genreUrlParam] : null,
                format: (formatUrlParam as AL_MediaFormat) === "MANGA" ? null : (formatUrlParam as AL_MediaFormat),
                season: (seasonUrlParam as AL_MediaSeason) || null,
                year: yearUrlParam || null,
                minScore: null,
                isAdult: false,
                countryOfOrigin: null,
                type: (formatUrlParam as AL_MediaFormat) === "MANGA" ? "manga" : (typeUrlParam as "anime" | "manga") || "anime",
            })
        }
    })

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper data-search-page-container className="space-y-6 px-4 md:p-8 pt-0 pb-10">
                <div className="flex items-center gap-4">
                    <SeaLink href={`/discover`}>
                        <Button leftIcon={<AiOutlineArrowLeft />} rounded intent="gray-outline" size="md">
                            Discover
                        </Button>
                    </SeaLink>
                    {/*<h3>Discover</h3>*/}
                </div>
                <div data-search-page-title className="text-center xl:text-left">
                    <AdvancedSearchPageTitle />
                </div>
                <AppLayoutGrid cols={6} spacing="lg">
                    <AdvancedSearchOptions />
                    <div data-search-page-list className="col-span-5">
                        <AdvancedSearchList />
                    </div>
                </AppLayoutGrid>
            </PageWrapper>
        </>
    )
}
