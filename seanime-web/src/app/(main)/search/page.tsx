"use client"
import { AppLayoutGrid, AppLayoutStack } from "@/components/ui/app-layout"
import React from "react"
import { MediaFormat, MediaSeason, MediaSort, MediaStatus } from "@/lib/anilist/gql/graphql"
import { useMount } from "react-use"
import { useSetAtom } from "jotai/react"
import { IconButton } from "@/components/ui/button"
import { AiOutlineArrowLeft } from "@react-icons/all-files/ai/AiOutlineArrowLeft"
import Link from "next/link"
import { __advancedSearch_paramsAtom } from "@/app/(main)/discover/_containers/advanced-search/_lib/parameters"
import {
    AdvancedSearchPageTitle,
} from "@/app/(main)/discover/_containers/advanced-search/_components/advanced-search-page-title"
import {
    AdvancedSearchOptions,
} from "@/app/(main)/discover/_containers/advanced-search/_components/advanced-search-options"
import { AdvancedSearchList } from "@/app/(main)/discover/_containers/advanced-search/_components/advanced-search-list"


export default function Page({ params: urlParams }: {
    params: {
        sorting?: MediaSort,
        genre?: string,
        format?: MediaFormat,
        season?: MediaSeason,
        status?: MediaStatus,
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
        })
    })

    return (
        <AppLayoutStack spacing={"xl"} className={"mt-8 px-4 pb-10"}>
            <div className={"flex items-center gap-4"}>
                <Link href={`/discover`}>
                    <IconButton icon={<AiOutlineArrowLeft/>} rounded intent={"white-outline"} size={"sm"}/>
                </Link>
                <h3>Discover</h3>
            </div>
            <div className={"text-center xl:text-left"}>
                <AdvancedSearchPageTitle/>
            </div>
            <AppLayoutGrid cols={6} spacing={"lg"}>
                <AdvancedSearchOptions/>
                <div className={"col-span-5"}>
                    <AdvancedSearchList/>
                </div>
            </AppLayoutGrid>
        </AppLayoutStack>
    )
}