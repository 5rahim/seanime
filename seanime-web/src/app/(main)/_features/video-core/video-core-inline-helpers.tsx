import { AL_BaseAnime } from "@/api/generated/types"
import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { TbLayoutSidebarRightCollapse, TbLayoutSidebarRightExpand } from "react-icons/tb"
import { useWindowSize } from "react-use"

const vc_inlineHelper_progressUpdateData = atom<{ media: AL_BaseAnime, currentProgress: number, currentEpisodeNumber: number } | null>(null)
const vc_inlineHelper_hasUpdatedProgress = atom<boolean>(false)

type VideoCoreInlineHelpers = {
    playerRef: React.MutableRefObject<HTMLVideoElement | null>,
    media: AL_BaseAnime,
    currentEpisodeNumber: number | null,
    currentProgress: number,
    url: string | null
}

export function VideoCoreInlineHelpers({
    playerRef,
    media,
    currentEpisodeNumber,
    currentProgress,
    url,
}: VideoCoreInlineHelpers) {
    const serverStatus = useServerStatus()

    // Track progress update state
    const [hasUpdatedProgress, setHasUpdateProgress] = useAtom(vc_inlineHelper_hasUpdatedProgress)
    const [progressUpdateData, setProgressUpdateData] = useAtom(vc_inlineHelper_progressUpdateData)

    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: updated } = useUpdateAnimeEntryProgress(
        media?.id,
        currentProgress,
    )

    // Reset state when media, episode, or update completes
    React.useEffect(() => {
        setProgressUpdateData(null)
        setHasUpdateProgress(false)
    }, [media, currentEpisodeNumber, url, updated])

    React.useEffect(() => {
        if (!playerRef.current || !media || currentEpisodeNumber === null || !url) return

        const PROGRESS_THRESHOLD = 0.8
        const CHECK_INTERVAL = 1000 // Check every second
        const MIN_VALID_TIME = 1

        const checkProgress = () => {
            const player = playerRef.current
            if (!player) return

            // Skip if already updated or currently updating
            if (hasUpdatedProgress || isUpdatingProgress) return

            // Skip if progress update data already exists
            if (progressUpdateData !== null) return

            const { currentTime, duration } = player

            if (!(currentTime > MIN_VALID_TIME && duration > MIN_VALID_TIME && isFinite(duration))) return

            if (currentProgress >= currentEpisodeNumber) return

            const watchedRatio = currentTime / duration
            if (watchedRatio < PROGRESS_THRESHOLD) return

            // Handle auto-update or prompt user
            if (serverStatus?.settings?.library?.autoUpdateProgress) {
                setHasUpdateProgress(true)
                updateProgress({
                    episodeNumber: currentEpisodeNumber,
                    mediaId: media.id,
                    totalEpisodes: media.episodes || 0,
                    malId: media.idMal || undefined,
                }, {
                    onSuccess: () => {
                        setHasUpdateProgress(true)
                    },
                    onError: () => {
                        setHasUpdateProgress(false)
                    },
                })
            } else {
                setProgressUpdateData({
                    media,
                    currentProgress,
                    currentEpisodeNumber,
                })
            }
        }

        // Start interval
        const intervalId = setInterval(checkProgress, CHECK_INTERVAL)

        // Cleanup
        return () => {
            clearInterval(intervalId)
        }
    }, [
        currentEpisodeNumber,
        media,
        playerRef,
        hasUpdatedProgress,
        isUpdatingProgress,
        serverStatus?.settings?.library?.autoUpdateProgress,
        currentProgress,
        progressUpdateData,
        url,
    ])

    return null
}

export function VideoCoreInlineHelperUpdateProgressButton() {
    const progressUpdateData = useAtomValue(vc_inlineHelper_progressUpdateData)
    const [hasUpdatedProgress, setHasUpdateProgress] = useAtom(vc_inlineHelper_hasUpdatedProgress)

    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: _ } = useUpdateAnimeEntryProgress(
        progressUpdateData?.media?.id,
        progressUpdateData?.currentEpisodeNumber ?? 0,
    )

    function handleProgressUpdate() {
        if (!progressUpdateData || !progressUpdateData.media) return
        updateProgress({
            episodeNumber: (progressUpdateData?.currentEpisodeNumber || 0),
            mediaId: progressUpdateData?.media?.id,
            totalEpisodes: progressUpdateData?.media?.episodes || 0,
            malId: progressUpdateData?.media?.idMal || undefined,
        }, {
            onSuccess: () => {
                setHasUpdateProgress(true)
            },
        })
    }

    if (progressUpdateData !== null && !hasUpdatedProgress) {
        return <Button
            className="animate-pulse"
            loading={isUpdatingProgress}
            disabled={hasUpdatedProgress}
            onClick={handleProgressUpdate}
        >
            Update progress
        </Button>
    }

    return null
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const vc_inlineLayoutTheaterMode = atomWithStorage("sea-video-core-theater-mode", false)

export type VideoCoreInlineLayoutProps = {
    mediaId?: string | number
    title?: string
    currentEpisodeNumber: number | null
    hideBackButton?: boolean
    leftHeaderActions?: React.ReactNode
    rightHeaderActions?: React.ReactNode
    mediaPlayer: React.ReactNode
    episodeList: React.ReactNode
    episodes: any[] | undefined
    loadingEpisodeList?: boolean
}

export function VideoCoreInlineLayout(props: VideoCoreInlineLayoutProps) {
    const {
        mediaId,
        title,
        hideBackButton,
        leftHeaderActions,
        rightHeaderActions,
        mediaPlayer,
        episodeList,
        episodes,
        loadingEpisodeList,
        currentEpisodeNumber,
    } = props

    const [theaterMode, setTheaterMode] = useAtom(vc_inlineLayoutTheaterMode)
    const { width } = useWindowSize()

    // Scroll to selected episode element when the episode list changes (on mount)
    const episodeListContainerRef = React.useRef<HTMLDivElement>(null)
    const episodeListViewportRef = React.useRef<HTMLDivElement>(null)
    const scrollTimeoutRef = React.useRef<NodeJS.Timeout>()
    const mediaPlayerContainerRef = React.useRef<HTMLDivElement>(null)
    const contentContainerRef = React.useRef<HTMLDivElement>(null)

    // Sync episode list height with media player container height
    React.useEffect(() => {
        if (!mediaPlayerContainerRef.current || !contentContainerRef.current || theaterMode) return
        const updateHeight = () => {
            if (!mediaPlayerContainerRef.current || !contentContainerRef.current) return
            const height = mediaPlayerContainerRef.current.offsetHeight
            contentContainerRef.current.style.setProperty("--player-height", `${height}px`)
        }
        updateHeight()
        const resizeObserver = new ResizeObserver(updateHeight)
        resizeObserver.observe(mediaPlayerContainerRef.current)

        return () => {
            resizeObserver.disconnect()
        }
    }, [theaterMode, width, mediaPlayerContainerRef.current])

    React.useEffect(() => {
        if (!episodeListContainerRef.current || !episodeListViewportRef.current || width <= 1536 || currentEpisodeNumber === null) return

        // Clear any existing timeout
        if (scrollTimeoutRef.current) {
            clearTimeout(scrollTimeoutRef.current)
        }

        scrollTimeoutRef.current = setTimeout(() => {
            const container = episodeListContainerRef.current
            const viewport = episodeListViewportRef.current
            if (!container || !viewport || theaterMode) return

            // Scroll page
            const containerTop = container.getBoundingClientRect().top + window.scrollY
            const padding = 20
            window.scrollTo({
                top: containerTop - padding,
                behavior: "smooth",
            })

            // Then scroll within the episode list viewport
            setTimeout(() => {
                const element = document.getElementById(`episode-${currentEpisodeNumber}`)
                if (element && viewport) {
                    const viewportRect = viewport.getBoundingClientRect()
                    const elementRect = element.getBoundingClientRect()
                    const scrollOffset = elementRect.top - viewportRect.top + viewport.scrollTop - 20

                    viewport.scrollTo({
                        top: scrollOffset,
                        behavior: "smooth",
                    })
                }
            }, 300)
        }, 100)

        // Cleanup
        return () => {
            if (scrollTimeoutRef.current) {
                clearTimeout(scrollTimeoutRef.current)
            }
        }
    }, [width, episodes, loadingEpisodeList, currentEpisodeNumber, theaterMode])


    return (
        <div data-sea-media-player-layout className="space-y-4">
            <div data-sea-media-player-layout-header className="flex flex-col lg:flex-row gap-2 w-full justify-between">
                {!hideBackButton && <div className="flex w-full gap-4 items-center relative">
                    <SeaLink href={`/entry?id=${mediaId}`}>
                        <IconButton icon={<AiOutlineArrowLeft />} rounded intent="gray-outline" size="sm" />
                    </SeaLink>
                    <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">{title}</h3>
                </div>}

                <div data-sea-media-player-layout-header-actions className="flex flex-wrap gap-2 items-center lg:justify-end w-full">
                    {leftHeaderActions}
                    <div className="flex flex-1"></div>
                    {rightHeaderActions}
                    <IconButton
                        onClick={() => setTheaterMode(p => !p)}
                        intent="gray-basic"
                        icon={theaterMode ? <TbLayoutSidebarRightExpand /> : <TbLayoutSidebarRightCollapse />}
                        className="hidden 2xl:flex"
                    />
                </div>
            </div>

            {!loadingEpisodeList ? <div
                ref={contentContainerRef}
                data-sea-media-player-layout-content
                className={cn(
                    "flex gap-4 w-full flex-col 2xl:flex-row",
                    theaterMode && "block space-y-4",
                )}
            >
                <div
                    ref={mediaPlayerContainerRef}
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
                    viewportRef={episodeListViewportRef}
                    data-sea-media-player-layout-content-episode-list
                    className={cn(
                        "2xl:max-w-[450px] w-full relative 2xl:sticky overflow-y-auto pr-4 pt-0",
                        theaterMode ? "2xl:max-w-full h-[75dvh]" : "h-[75dvh] 2xl:h-auto",
                    )}
                    style={!theaterMode ? { height: "var(--player-height, 75dvh)" } as React.CSSProperties : undefined}
                >
                    <div data-sea-media-player-layout-content-episode-list-container className="space-y-3">
                        {episodeList}
                    </div>
                </ScrollArea>
            </div> : <div
                className="grid 2xl:grid-cols-[1fr,450px] gap-4 xl:gap-4"
            >
                <div className="w-full min-h-[70dvh] relative">
                    <Skeleton className="h-full w-full absolute" />
                </div>

                <Skeleton className="hidden 2xl:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />

            </div>}
        </div>
    )
}
