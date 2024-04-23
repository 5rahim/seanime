"use client"

import { AnilistCollectionLists } from "@/app/(main)/anilist/_containers/anilist-collection-lists"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import React from "react"

export const dynamic = "force-static"

export default function Home() {

    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative"
        >
            <AnilistCollectionLists />
        </PageWrapper>
    )
}
