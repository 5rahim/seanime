"use client"
import { CustomBackgroundImage } from "@/components/shared/custom-ui/custom-background-image"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import React from "react"


export default function Layout({ children }: { children: React.ReactNode }) {

    return (
        <>
            {/*[CUSTOM UI]*/}
            <CustomBackgroundImage />
            <PageWrapper className="p-4 sm:p-8 space-y-4">
                <div className="flex justify-between items-center w-full relative">
                    <div>
                        <h2>Auto Downloader</h2>
                        <p className="text-[--muted]">
                            Add and manage auto-downloading rules for your favorite anime.
                        </p>
                    </div>
                </div>

                {children}
            </PageWrapper>
        </>
    )

}
