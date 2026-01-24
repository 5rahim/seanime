"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AutoDownloaderPage } from "@/app/(main)/auto-downloader/_containers/autodownloader-page"
import { PageWrapper } from "@/components/shared/page-wrapper"
import React from "react"

export const dynamic = "force-static"

export default function Page() {

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper className="p-4 sm:p-8 space-y-4">
                <div className="flex justify-between items-center w-full relative">
                    <div>
                        <h2>Auto Downloader</h2>
                        <p className="text-[--muted]">
                            Automatically download new episodes as they are released.
                        </p>
                    </div>
                </div>
                <AutoDownloaderPage />
            </PageWrapper>
        </>
    )

}
