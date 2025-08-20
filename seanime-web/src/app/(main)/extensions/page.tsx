"use client"

import { useUnauthorizedPluginCount } from "@/api/hooks/extensions.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { __extensions_currentPageAtom, ExtensionList } from "@/app/(main)/extensions/_containers/extension-list"
import { MarketplaceExtensions } from "@/app/(main)/extensions/_containers/marketplace-extensions"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { StaticTabs } from "@/components/ui/tabs"
import { useAtom } from "jotai"
import { AnimatePresence } from "motion/react"
import React from "react"
import { FaExclamation } from "react-icons/fa"
import { LuDownload, LuGlobe } from "react-icons/lu"

export default function Page() {

    const [page, setPage] = useAtom(__extensions_currentPageAtom)
    const unauthorizedPluginCount = useUnauthorizedPluginCount()

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper className="p-4 sm:p-8 space-y-4">

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
                            addon: unauthorizedPluginCount > 0 && (
                                <span className="ml-2 bottom-1 right-1 rounded-full relative">
                                    <FaExclamation className="text-[--orange] animate-bounce size-6" />
                                </span>
                            ),
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
        </>
    )

}
