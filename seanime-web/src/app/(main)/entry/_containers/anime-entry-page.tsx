import { useGetAnilistMediaDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { EpisodeListGridProvider } from "@/app/(main)/entry/_components/episode-list-grid"
import { MetaSection } from "@/app/(main)/entry/_components/meta-section"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-list/episode-section"
import { TorrentSearchDrawer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Skeleton } from "@/components/ui/skeleton"
import { AnimatePresence } from "framer-motion"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"

export const __anime_wantStreamingAtom = atom(false)

export function AnimeEntryPage() {

    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: mediaDetails, isLoading: mediaDetailsLoading } = useGetAnilistMediaDetails(mediaId)

    const [wantStreaming, setWantStreaming] = useAtom(__anime_wantStreamingAtom)

    React.useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if ((!mediaEntryLoading && !mediaEntry)) {
            router.push("/")
        }
    }, [mediaEntry, mediaEntryLoading])

    if (mediaEntryLoading || mediaDetailsLoading) return <LoadingDisplay />
    if (!mediaEntry) return null

    return (
        <div>
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
                    <AnimatePresence mode="wait" initial={false}>
                        {!wantStreaming && <PageWrapper
                            key="episode-list"
                            className="relative 2xl:order-first pb-10 pt-4"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, y: 20 },
                                transition: {
                                    type: "spring",
                                    damping: 10,
                                    stiffness: 80,
                                },
                            }}
                        >
                            <EpisodeListGridProvider container="expanded">
                                <EpisodeSection entry={mediaEntry} details={mediaDetails} />
                            </EpisodeListGridProvider>
                        </PageWrapper>}

                        {wantStreaming && <PageWrapper
                            key="torrent-streaming-episodes"
                            className="relative 2xl:order-first pb-10 pt-4"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, y: 20 },
                                transition: {
                                    type: "spring",
                                    damping: 10,
                                    stiffness: 80,
                                },
                            }}
                        > <EpisodeListGridProvider container="expanded">
                            Lorem ipsum dolor sit amet, consectetur adipisicing elit.
                        </EpisodeListGridProvider>
                        </PageWrapper>}

                    </AnimatePresence>
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
