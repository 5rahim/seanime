"use client"

import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { ExtensionList } from "@/app/(main)/extensions/_containers/extension-list"
import { PageWrapper } from "@/components/shared/page-wrapper"
import React from "react"

export default function Page() {

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper className="p-4 sm:p-8 pt-0 space-y-8 relative z-[4]">
                <ExtensionList />
            </PageWrapper>
        </>
    )

}
