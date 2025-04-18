"use client"

import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { ExtensionList } from "@/app/(main)/extensions/_containers/extension-list"
import { MarketplaceExtensions } from "@/app/(main)/extensions/_containers/marketplace-extensions"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { StaticTabs } from "@/components/ui/tabs"
import { AnimatePresence } from "framer-motion"
import React from "react"
import { LuDownload, LuGlobe } from "react-icons/lu"

export default function Page() {

    const [page, setPage] = React.useState<"installed" | "marketplace">("installed")

    return (
        <PageWrapper className="p-4 sm:p-8 space-y-4">
            <CustomLibraryBanner discrete />

            {/*<div className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-lg">*/}
            <StaticTabs
                data-anilist-collection-lists-tabs
                className="h-10 w-fit border rounded-full"
                triggerClass="px-4 py-1 text-md"
                items={[
                    {
                        name: "Installed",
                        isCurrent: page === "installed",
                        onClick: () => setPage("installed"),
                        iconType: LuDownload,
                    },
                    {
                        name: "Marketplace",
                        isCurrent: page === "marketplace",
                        onClick: () => setPage("marketplace"),
                        iconType: LuGlobe,
                    },
                ]}
            />
            {/*</div>*/}

            <AnimatePresence mode="wait">
                {page === "installed" && (
                    <PageWrapper
                        {...{
                            initial: { opacity: 0, y: 60 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, scale: 0.99 },
                            transition: {
                                duration: 0.35,
                            },
                        }}
                        key="installed" className="pt-0 space-y-8 relative z-[4]"
                    >
                        <ExtensionList />
                    </PageWrapper>
                )}
                {page === "marketplace" && (
                    <PageWrapper
                        {...{
                            initial: { opacity: 0, y: 60 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, scale: 0.99 },
                            transition: {
                                duration: 0.35,
                            },
                        }}
                        key="marketplace" className="pt-0 space-y-8 relative z-[4]"
                    >
                        <MarketplaceExtensions />
                    </PageWrapper>
                )}
            </AnimatePresence>
        </PageWrapper>
    )

}
