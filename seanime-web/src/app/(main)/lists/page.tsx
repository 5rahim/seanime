"use client"

import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AnilistCollectionLists } from "@/app/(main)/lists/_containers/anilist-collection-lists"
import { PageWrapper } from "@/components/shared/page-wrapper"
import React from "react"

export const dynamic = "force-static"

export default function Home() {

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper
                className="p-4 sm:p-8 pt-4 relative"
                data-anilist-page
            >
                <AnilistCollectionLists />
            </PageWrapper>
        </>
    )
}
