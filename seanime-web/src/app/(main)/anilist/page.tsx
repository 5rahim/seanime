"use client"

import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AnilistCollectionLists } from "@/app/(main)/anilist/_containers/anilist-collection-lists"
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
                {...{
                    initial: { opacity: 0, y: 10 },
                    animate: { opacity: 1, y: 0 },
                    exit: { opacity: 0, y: 10 },
                    transition: {
                        type: "spring",
                        damping: 20,
                        stiffness: 100,
                    },
                }}
            >
                <AnilistCollectionLists />
            </PageWrapper>
        </>
    )
}
