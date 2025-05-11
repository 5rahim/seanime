import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import {
    __seaMediaPlayer_scopedCurrentProgressAtom,
    __seaMediaPlayer_scopedProgressItemAtom,
    useSeaMediaPlayer,
} from "@/app/(main)/_features/sea-media-player/sea-media-player-provider"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { TbLayoutSidebarRightCollapse, TbLayoutSidebarRightExpand } from "react-icons/tb"
import { useWindowSize } from "react-use"

const theaterModeAtom = atomWithStorage("sea-media-theater-mode", false)

export type SeaMediaPlayerLayoutProps = {
    mediaId?: string | number
    title?: string
    hideBackButton?: boolean
    leftHeaderActions?: React.ReactNode
    rightHeaderActions?: React.ReactNode
    mediaPlayer: React.ReactNode
    episodeList: React.ReactNode
    episodes: any[] | undefined
}

export function SeaMediaPlayerLayout(props: SeaMediaPlayerLayoutProps) {
    const {
        mediaId,
        title,
        hideBackButton,
        leftHeaderActions,
        rightHeaderActions,
        mediaPlayer,
        episodeList,
        episodes,
    } = props

    const [theaterMode, setTheaterMode] = useAtom(theaterModeAtom)
    const { media, progress } = useSeaMediaPlayer()
    const [currentProgress, setCurrentProgress] = useAtom(__seaMediaPlayer_scopedCurrentProgressAtom)
    const [progressItem, setProgressItem] = useAtom(__seaMediaPlayer_scopedProgressItemAtom)

    /** Progress update **/
    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: hasUpdatedProgress } = useUpdateAnimeEntryProgress(
        media?.id,
        currentProgress,
    )

    const { width } = useWindowSize()

    /** Scroll to selected episode element when the episode list changes (on mount) **/
    const episodeListContainerRef = React.useRef<HTMLDivElement>(null)
    const scrollTimeoutRef = React.useRef<NodeJS.Timeout>()

    React.useEffect(() => {
        if (!episodeListContainerRef.current || width <= 1536 || !progress.currentEpisodeNumber) return

        // Clear any existing timeout
        if (scrollTimeoutRef.current) {
            clearTimeout(scrollTimeoutRef.current)
        }

        // Set a new timeout to scroll after a brief delay
        scrollTimeoutRef.current = setTimeout(() => {
            let element = document.getElementById(`episode-${progress.currentEpisodeNumber}`)
            if (theaterMode) {
                element = document.getElementById("sea-media-player-container")
            }
            if (element) {
                element.scrollIntoView({ behavior: "smooth" })
            }
        }, 100)

        // Cleanup
        return () => {
            if (scrollTimeoutRef.current) {
                clearTimeout(scrollTimeoutRef.current)
            }
        }
    }, [width, progress.currentEpisodeNumber, theaterMode])

    const handleProgressUpdate = React.useCallback(() => {
        if (!media || !progressItem || isUpdatingProgress || hasUpdatedProgress) return

        updateProgress({
            episodeNumber: progressItem.episodeNumber,
            mediaId: media.id,
            totalEpisodes: media.episodes || 0,
            malId: media.idMal || undefined,
        }, {
            onSuccess: () => {
                setProgressItem(null)
                setCurrentProgress(progressItem.episodeNumber)
            },
        })
    }, [media, progressItem, isUpdatingProgress, hasUpdatedProgress])

    return (
        <div data-sea-media-player-layout className="space-y-4">
            <div data-sea-media-player-layout-header className="flex flex-col lg:flex-row gap-2 w-full justify-between">
                {!hideBackButton && <div className="flex w-full gap-4 items-center relative">
                    <SeaLink href={`/entry?id=${mediaId}`}>
                        <IconButton icon={<AiOutlineArrowLeft />} rounded intent="gray-outline" size="sm" />
                    </SeaLink>
                    <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">{title}</h3>
                </div>}

                <div data-sea-media-player-layout-header-actions className="flex gap-2 items-center justify-end w-full">
                    {leftHeaderActions}
                    <div className="hidden lg:flex flex-1"></div>
                    {(!!progressItem && progressItem.episodeNumber > currentProgress) && (
                        <Button
                            className="animate-pulse"
                            loading={isUpdatingProgress}
                            disabled={hasUpdatedProgress}
                            onClick={handleProgressUpdate}
                        >
                            Update progress
                        </Button>
                    )}
                    {rightHeaderActions}
                    <IconButton
                        onClick={() => setTheaterMode(p => !p)}
                        intent="gray-basic"
                        icon={theaterMode ? <TbLayoutSidebarRightExpand /> : <TbLayoutSidebarRightCollapse />}
                    />
                </div>
            </div>

            <div
                data-sea-media-player-layout-content
                className={cn(
                    "flex gap-4 w-full flex-col 2xl:flex-row",
                    theaterMode && "block space-y-4",
                )}
            >
                <div
                    id="sea-media-player-container"
                    data-sea-media-player-layout-content-player
                    className={cn(
                        "aspect-video relative w-full self-start mx-auto",
                        theaterMode && "max-h-[90vh] !w-auto aspect-video mx-auto",
                    )}
                >
                    {mediaPlayer}
                </div>

                <ScrollArea
                    ref={episodeListContainerRef}
                    data-sea-media-player-layout-content-episode-list
                    className={cn(
                        "2xl:max-w-[450px] w-full relative 2xl:sticky h-[75dvh] overflow-y-auto pr-4 pt-0 -mt-3",
                        theaterMode && "2xl:max-w-full",
                    )}
                >
                    <div data-sea-media-player-layout-content-episode-list-container className="space-y-3">
                        {episodeList}
                    </div>
                    <div
                        data-sea-media-player-layout-content-episode-list-bottom-gradient
                        className={"z-[5] absolute bottom-0 w-full h-[2rem] bg-gradient-to-t from-[--background] to-transparent"}
                    />
                </ScrollArea>
            </div>
        </div>
    )
}
