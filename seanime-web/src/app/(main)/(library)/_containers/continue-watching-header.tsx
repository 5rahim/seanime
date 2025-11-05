"use client"

import { Anime_Episode } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { SeaLink } from "@/components/shared/sea-link"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { getAssetUrl } from "@/lib/server/assets"
import { ThemeLibraryScreenBannerType, ThemeMediaPageBannerSize, ThemeMediaPageBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { __isDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { AnimatePresence, motion } from "motion/react"
import { useRouter } from "next/navigation"
import React from "react"
import { RiSignalTowerLine } from "react-icons/ri"

export const __continueWatching_hoveringHeaderAtom = atom(false)
export const __continueWatching_currentEpisodeIndexAtom = atom(0)
export const __continueWatching_headerIsTransitioningAtom = atom(false)
export const __continueWatching_setCurrentEpisodeIndexAtom = atom(
    null,
    (get, set, newIndex: number) => {
        const currentIndex = get(__continueWatching_currentEpisodeIndexAtom)
        if (currentIndex !== newIndex) {
            set(__continueWatching_headerIsTransitioningAtom, true)
            setTimeout(() => {
                set(__continueWatching_currentEpisodeIndexAtom, newIndex)
                set(__continueWatching_headerIsTransitioningAtom, false)
            }, 300)
        }
    },
)

const MotionImage = motion.create(SeaImage)

interface ContinueWatchingHeaderProps {
    episodes: Anime_Episode[]
    className?: string
}

interface HeaderCarouselDotsProps {
    totalEpisodes: number
    currentIndex: number
    onIndexChange: (index: number) => void
    className?: string
}

function HeaderCarouselDots({ totalEpisodes, currentIndex, onIndexChange, className }: HeaderCarouselDotsProps) {
    const ts = useThemeSettings()

    // Don't render if there are no episodes or only one episode
    if (totalEpisodes <= 1) return null

    const maxDots = Math.min(totalEpisodes, 99)

    return (
        <div
            className={cn(
                "hidden lg:flex items-center gap-2 z-[10] pl-8 max-w-[20rem] flex-wrap top-[4.5rem]",
                // ts.hideTopNavbar && !__isDesktop__ && "top-[4rem]",
                // (ts.hideTopNavbar || __isDesktop__) && "top-[1.5rem]",
                // ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && (__isDesktop__ || ts.hideTopNavbar) && "top-[0.6rem]",
                className,
            )}
        >
            {Array.from({ length: maxDots }).map((_, index) => (
                <button
                    key={index}
                    className={cn(
                        "h-1.5 rounded-sm transition-all duration-300 cursor-pointer",
                        index === currentIndex ? "w-6 bg-[--muted]" : "w-3 bg-[--subtle] hover:bg-gray-300",
                    )}
                    onClick={() => onIndexChange(index)}
                    aria-label={`Go to episode ${index + 1}`}
                />
            ))}
        </div>
    )
}

interface MediaMetadataProps {
    episode: Anime_Episode
    episodes: Anime_Episode[]
    onHoverChange: (hovering: boolean) => void
}

function MediaMetadata({ episode, episodes, onHoverChange }: MediaMetadataProps) {
    const ts = useThemeSettings()
    const { setPreviewModalMediaId } = useMediaPreviewModal()
    const anime = episode.baseAnime

    const currentEpisodeIndex = useAtomValue(__continueWatching_currentEpisodeIndexAtom)
    const isTransitioning = useAtomValue(__continueWatching_headerIsTransitioningAtom)
    const setCurrentEpisodeIndex = useSetAtom(__continueWatching_setCurrentEpisodeIndexAtom)
    const [isHoveringHeader, setHoveringHeader] = useAtom(__continueWatching_hoveringHeaderAtom)

    const currentEpisode = episodes[currentEpisodeIndex] || null
    const shouldBlurBanner = ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable &&
        !currentEpisode?.baseAnime?.bannerImage

    React.useEffect(() => {
        if (episodes.length <= 1) return

        const interval = setInterval(() => {
            if (!isHoveringHeader) {
                setCurrentEpisodeIndex((currentEpisodeIndex + 1) % episodes.length)
            }
        }, 8000)

        return () => clearInterval(interval)
    }, [currentEpisodeIndex, episodes.length, isHoveringHeader, setCurrentEpisodeIndex])

    if (!anime) return null

    return (
        <div
            className={cn(
                "absolute left-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block",
                "top-[5rem]",
                ts.hideTopNavbar && "top-[4rem]",
                // ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "top-[4rem]",
                // (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "top-[2rem]",
                (__isDesktop__ && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "top-[0rem]",
                (__isDesktop__ && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "top-[2rem]",
            )}
            data-media-id={anime.id}
            data-media-mal-id={anime.idMal}
        >
            <motion.div
                className="flex items-center relative gap-6 p-6 pr-3 w-fit overflow-hidden"
                onMouseEnter={() => onHoverChange(true)}
                onMouseLeave={() => onHoverChange(false)}
                {...{
                    initial: { opacity: 0, x: -40 },
                    animate: { opacity: 1, x: 0 },
                    exit: { opacity: 0, x: -20 },
                    transition: {
                        type: "spring",
                        damping: 20,
                        stiffness: 100,
                    },
                }}
            >
                <motion.div
                    className="flex-none"
                    initial={{ opacity: 0, scale: 0.7, skew: 5 } as any}
                    animate={{ opacity: 1, scale: 1, skew: 0 } as any}
                    exit={{ opacity: 1, scale: 1, skewY: 1 }}
                    transition={{ duration: 0.5 }}
                >
                    <SeaLink href={`/entry?id=${anime.id}`}>
                        {anime.coverImage?.large && (
                            <div className="w-[180px] h-[280px] relative rounded-[--radius-md] overflow-hidden bg-[--background] shadow-md">
                                <SeaImage
                                    src={anime.coverImage.large}
                                    alt="cover image"
                                    fill
                                    priority
                                    placeholder={imageShimmer(700, 475)}
                                    className={cn(
                                        "object-cover object-center transition-opacity duration-1000",
                                        isTransitioning && "opacity-30",
                                        !isTransitioning && "opacity-100",
                                    )}
                                />
                            </div>
                        )}
                    </SeaLink>
                </motion.div>

                <motion.div
                    className="flex-auto space-y-2 z-[1]"
                    initial={{ opacity: 0, x: 10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.5, delay: 0.6 }}
                >
                    <SeaLink href={`/entry?id=${anime.id}`}>
                        <TextGenerateEffect
                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white leading-8 line-clamp-2 pb-1 max-w-md text-pretty text-3xl overflow-ellipsis"
                            words={anime.title?.userPreferred || ""}
                        />
                    </SeaLink>

                    <div className="flex flex-wrap gap-2">
                        {anime.genres?.slice(0, 3).map((genre) => (
                            <div key={genre} className="text-sm font-semibold px-1 text-gray-300">
                                {genre}
                            </div>
                        ))}
                    </div>

                    <div className="flex items-center max-w-lg gap-4">
                        {anime.meanScore && (
                            <div className="rounded-full w-fit inline-block">
                                <MediaEntryAudienceScore meanScore={anime.meanScore} />
                            </div>
                        )}

                        {anime.nextAiringEpisode?.airingAt && (
                            <p className="text-base text-brand-200 inline-flex items-center gap-1.5">
                                <RiSignalTowerLine /> Releasing now
                            </p>
                        )}
                    </div>

                    <motion.div
                        className="pt-0 left-0"
                        initial={{ opacity: 0, x: 10 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.5, delay: 0.7 }}
                    >
                        <ScrollArea className="max-w-lg leading-3 h-[77px] mb-4 p-0 text-sm">
                            {anime.description?.replace(/(<([^>]+)>)/ig, "")}
                        </ScrollArea>

                        <Button
                            size="sm"
                            intent="gray-outline"
                            className="rounded-full"
                            onClick={() => setPreviewModalMediaId(anime.id, "anime")}
                        >
                            Preview
                        </Button>
                    </motion.div>
                </motion.div>
            </motion.div>

            <HeaderCarouselDots
                totalEpisodes={episodes.length}
                currentIndex={currentEpisodeIndex}
                onIndexChange={setCurrentEpisodeIndex}
            />
        </div>
    )
}

interface EpisodeCardSidebarProps {
    episode: Anime_Episode
    isTransitioning: boolean
}

function EpisodeCardSidebar({ episode, isTransitioning }: EpisodeCardSidebarProps) {
    const ts = useThemeSettings()
    const router = useRouter()
    const serverStatus = useServerStatus()
    const { setPlayNext } = usePlayNext()
    const { data: watchHistory } = useGetContinuityWatchHistory()

    const handleEpisodeClick = () => {
        setPlayNext(episode.baseAnime?.id, () => {
            if (!serverStatus?.isOffline) {
                router.push(`/entry?id=${episode.baseAnime?.id}`)
            } else {
                router.push(`/offline/entry/anime?id=${episode.baseAnime?.id}`)
            }
        })
    }

    return (
        <motion.div
            className={cn(
                "absolute right-6 w-fit h-[25rem] z-[3] hidden lg:block overflow-hidden",
                "top-[5rem]",
                ts.hideTopNavbar && "top-[4rem]",
                // ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "top-[4rem]",
                // (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "top-[2rem]",
                (__isDesktop__ && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "top-[1rem]",
                (__isDesktop__ && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "top-[3rem]",
            )}
        >
            <div className="p-6 w-fit">
                <motion.div
                    // initial={{ opacity: 0, scale: 0.9 }}
                    // animate={{ opacity: 1, scale: 1 }}
                    // transition={{ duration: 0.5, delay: 0.3 }}
                    {...{
                        initial: { opacity: 0, x: 40 },
                        animate: { opacity: 1, x: 0 },
                        exit: { opacity: 0, x: 20 },
                        transition: {
                            type: "spring",
                            damping: 20,
                            stiffness: 100,
                        },
                    }}
                    className="2xl:w-[500px] xl:w-[400px] lg:w-[300px] rounded-xl overflow-hidden"
                >
                    {/* <div className="w-[160%] h-[120%] -left-[30%] -top-0 opacity-50 absolute z-[1]">
                     <img src="/radial-shadow.png" alt="radial shadow" className="w-full h-full object-contain" />
                     </div> */}
                    <EpisodeCard
                        episode={episode}
                        image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
                        topTitle={episode.episodeTitle || episode?.baseAnime?.title?.userPreferred}
                        title={episode.displayTitle}
                        isInvalid={episode.isInvalid}
                        progressTotal={episode.baseAnime?.episodes}
                        progressNumber={episode.progressNumber}
                        episodeNumber={episode.episodeNumber}
                        length={episode.episodeMetadata?.length}
                        hasDiscrepancy={episode.episodeNumber !== episode.progressNumber}
                        percentageComplete={getEpisodePercentageComplete(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
                        minutesRemaining={getEpisodeMinutesRemaining(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
                        anime={{
                            id: episode?.baseAnime?.id || 0,
                            image: episode?.baseAnime?.coverImage?.medium,
                            title: episode?.baseAnime?.title?.userPreferred,
                        }}
                        forceSingleContainer
                        onClick={handleEpisodeClick}
                        className={cn(
                            "transition-opacity duration-1000",
                            isTransitioning && "opacity-50",
                        )}
                    />
                </motion.div>
            </div>
        </motion.div>
    )
}

interface BannerImageProps {
    episode: Anime_Episode | null
    isTransitioning: boolean
    shouldBlurBanner: boolean
}

function BannerImage({ episode, isTransitioning, shouldBlurBanner }: BannerImageProps) {
    const ts = useThemeSettings()
    const bannerImage = (!!ts.libraryScreenCustomBannerImage
        && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom) ? getAssetUrl(ts.libraryScreenCustomBannerImage) :
        episode?.baseAnime?.bannerImage || episode?.baseAnime?.coverImage?.extraLarge

    return (
        <div
            className={cn(
                "lg:h-[35rem] w-full flex-none object-cover object-center absolute top-0 bg-[--background] overflow-hidden",
                !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                __isDesktop__ && "top-[-2rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[30rem]",
            )}
        >
            <div className="w-full z-[2] absolute bottom-[-10rem] h-[10rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent" />
            <div className="w-full absolute z-[2] top-0 h-[10rem] opacity-50 bg-gradient-to-b from-[--background] to-transparent" />
            <div
                className={cn(
                    "opacity-0 duration-1000 bg-[var(--background)] w-full h-full absolute z-[2]",
                    isTransitioning && "opacity-70",
                )}
            />

            <AnimatePresence>
                <div className="w-full h-full absolute z-[1] overflow-hidden">
                    {bannerImage && (
                        <MotionImage
                            src={bannerImage}
                            alt="banner image"
                            fill
                            quality={100}
                            priority
                            className={cn(
                                "object-cover object-center z-[1] transition-all duration-1000",
                                isTransitioning && "scale-[1.01] -translate-x-0.5",
                                !isTransitioning && "scale-100 translate-x-0",
                                !episode?.baseAnime?.bannerImage && "opacity-35",
                            )}
                        />
                    )}
                </div>
            </AnimatePresence>

            {shouldBlurBanner && (
                <div className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2]" />
            )}

            <div
                className={cn(
                    "hidden lg:block max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-5% via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
                    "opacity-100 duration-500",
                )}
            />

            <div
                className={cn(
                    "hidden lg:block max-w-[60rem] w-full z-[2] h-full absolute -right-[10rem] &-right-[25rem] &-bottom-[10rem] bg-gradient-to-l from-[--background] &rotate-45 via-[--background] via-opacity-50 via-5% transition-opacity to-transparent",
                    "opacity-100 duration-500",
                )}
            />

            {!ts.disableSidebarTransparency && (
                <div
                    className={cn(
                        "hidden lg:block max-w-[10rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
                        "opacity-70 duration-500",
                    )}
                />
            )}

            <div className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent" />
        </div>
    )
}

export function ContinueWatchingHeader({ episodes, className }: ContinueWatchingHeaderProps) {
    const ts = useThemeSettings()

    const currentEpisodeIndex = useAtomValue(__continueWatching_currentEpisodeIndexAtom)
    const isTransitioning = useAtomValue(__continueWatching_headerIsTransitioningAtom)
    const setCurrentEpisodeIndex = useSetAtom(__continueWatching_setCurrentEpisodeIndexAtom)
    const [isHoveringHeader, setHoveringHeader] = useAtom(__continueWatching_hoveringHeaderAtom)

    const currentEpisode = episodes[currentEpisodeIndex] || null
    const shouldBlurBanner = ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable &&
        !currentEpisode?.baseAnime?.bannerImage

    React.useEffect(() => {
        if (episodes.length <= 1) return

        const interval = setInterval(() => {
            if (!isHoveringHeader) {
                setCurrentEpisodeIndex((currentEpisodeIndex + 1) % episodes.length)
            }
        }, 8000)

        return () => clearInterval(interval)
    }, [currentEpisodeIndex, episodes.length, isHoveringHeader, setCurrentEpisodeIndex])

    if (!episodes.length) return null

    return (
        <motion.div
            className={cn(
                "__header lg:h-[28rem] max-w-full overflow-hidden",
                ts.hideTopNavbar && "lg:h-[32rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[26rem]",
                (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "lg:h-[28rem]",
                className,
            )}
            {...{
                initial: { opacity: 0 },
                animate: { opacity: 1 },
                transition: { duration: 1.2 },
            }}
        >

            <BannerImage
                episode={currentEpisode}
                isTransitioning={isTransitioning}
                shouldBlurBanner={shouldBlurBanner}
            />

            <AnimatePresence>
                {currentEpisode && !isTransitioning && (
                    <>
                        <MediaMetadata
                            episode={currentEpisode}
                            episodes={episodes}
                            onHoverChange={setHoveringHeader}
                        />
                        <EpisodeCardSidebar
                            episode={currentEpisode}
                            isTransitioning={isTransitioning}
                        />
                    </>
                )}
            </AnimatePresence>
        </motion.div>
    )
}
