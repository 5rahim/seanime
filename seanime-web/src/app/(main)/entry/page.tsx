"use client"

import { useGetAnilistMediaDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { EntryHeaderBackground } from "@/app/(main)/entry/_components/entry-header-background"
import { EpisodeListGridProvider } from "@/app/(main)/entry/_components/episode-list-grid"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-section/episode-section"
import { LegacyEpisodeSection } from "@/app/(main)/entry/_containers/episode-section/legacy-episode-section"
import { RelationsRecommendationsAccordion } from "@/app/(main)/entry/_containers/meta-section/_components/relations-recommendations-accordion"
import { LegacyMetaSection } from "@/app/(main)/entry/_containers/meta-section/legacy-meta-section"
import { MetaSection } from "@/app/(main)/entry/_containers/meta-section/meta-section"
import { TorrentSearchDrawer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { CustomBackgroundImage } from "@/components/shared/custom-ui/custom-background-image"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "framer-motion"
import { useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const dynamic = "force-static"

export default function Page() {
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: mediaDetails, isLoading: mediaDetailsLoading } = useGetAnilistMediaDetails(mediaId)

    // [CUSTOM UI]
    const ts = useThemeSettings()
    const newDesign = ts.animeEntryScreenLayout === "stacked"

    useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if ((!mediaEntryLoading && !mediaEntry)) {
            router.push("/")
        }
    }, [mediaEntry, mediaEntryLoading])


    if (mediaEntryLoading || mediaDetailsLoading) return <LoadingDisplay />
    if (!mediaEntry) return null

    if (newDesign) {
        return <div>
            {/*[CUSTOM UI]*/}
            <CustomBackgroundImage />

            <MetaSection entry={mediaEntry} details={mediaDetails} />

            <div className="px-4 md:px-8 relative z-[8]">

                <PageWrapper
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, y: 60 },
                        transition: {
                            type: "spring",
                            damping: 10,
                            stiffness: 80,
                            delay: 0.6,
                        },
                    }}
                >
                    <EpisodeListGridProvider container="expanded">
                        <EpisodeSection entry={mediaEntry} details={mediaDetails} />
                    </EpisodeListGridProvider>
                </PageWrapper>
            </div>

            <TorrentSearchDrawer entry={mediaEntry} />
        </div>
    }

    return (
        <div>
            {/*[CUSTOM UI]*/}
            <CustomBackgroundImage />

            <EntryHeaderBackground entry={mediaEntry} />

            <div
                className={cn(
                    "-mt-[8rem] relative z-10 max-w-full px-4 md:px-10 grid grid-cols-1 gap-8 pb-16 2xl:grid-cols-2",
                    { "2xl:grid-cols-[minmax(0,1.2fr),1fr]": !!mediaEntry?.libraryData },
                )}
            >

                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.2 }}
                    className={cn(
                        "w-full z-[0] left-0 absolute top-[8rem] h-[30rem] bg-gradient-to-b from-[--background] via-[--background] via-opacity-50 via-50% to-transparent",
                        !mediaEntry.libraryData && "h-[10rem]",
                    )}
                />

                <div className="-mt-[18rem] h-[fit-content] 2xl:sticky top-[5rem] space-y-8">
                    <div className="backdrop-blur-xl rounded-xl">
                        <LegacyMetaSection entry={mediaEntry} details={mediaDetails} />
                    </div>
                    <RelationsRecommendationsAccordion
                        entry={mediaEntry}
                        details={mediaDetails}
                    />
                </div>
                <PageWrapper className="relative 2xl:order-first pb-10 z-[1]">
                    <LegacyEpisodeSection entry={mediaEntry} />
                </PageWrapper>
            </div>
            <TorrentSearchDrawer entry={mediaEntry} />
        </div>
    )
}

function LoadingDisplay() {
    return (
        <div className="__header h-[30rem]">
            <div
                className="h-[30rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[1] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                <Skeleton className="h-full absolute w-full" />
                <div
                    className="w-full absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-transparent to-transparent"
                />
            </div>
        </div>
    )
}
