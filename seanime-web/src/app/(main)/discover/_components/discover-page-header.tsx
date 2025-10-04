import { AL_BaseAnime } from "@/api/generated/types"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import {
    __discover_animeRandomNumberAtom,
    __discover_animeTotalItemsAtom,
    __discover_headerIsTransitioningAtom,
    __discover_randomTrendingAtom,
    __discover_setAnimeRandomNumberAtom,
} from "@/app/(main)/discover/_containers/discover-trending"
import {
    __discover_mangaRandomNumberAtom,
    __discover_mangaTotalItemsAtom,
    __discover_setMangaRandomNumberAtom,
} from "@/app/(main)/discover/_containers/discover-trending-country"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { SeaLink } from "@/components/shared/sea-link"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { ThemeMediaPageBannerSize, ThemeMediaPageBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { __isDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { AnimatePresence, motion } from "motion/react"
import { usePathname } from "next/navigation"
import React from "react"
import { RiSignalTowerLine } from "react-icons/ri"

// Atoms for state management
export const __discover_hoveringHeaderAtom = atom(false)
export const __discover_clickedCarouselDotAtom = atom(0)

const MotionImage = motion.create(SeaImage)
const MotionIframe = motion.create("iframe")

interface HeaderCarouselDotsProps {
    className?: string
}

interface BannerImageProps {
    media: AL_BaseAnime | null
    isTransitioning: boolean
    shouldBlurBanner: boolean
    showTrailer: boolean
    trailerLoaded: boolean
    onTrailerLoad: () => void
    onTrailerError: () => void
}

interface MediaMetadataProps {
    media: AL_BaseAnime | null
    pageType: "anime" | "manga" | "schedule"
    isTransitioning: boolean
    onHoverChange: (hovering: boolean) => void
}

function HeaderCarouselDots({ className }: HeaderCarouselDotsProps) {
    const [pageType] = useAtom(__discord_pageTypeAtom)
    const setClickedCarouselDot = useSetAtom(__discover_clickedCarouselDotAtom)

    const animeRandomNumber = useAtomValue(__discover_animeRandomNumberAtom)
    const animeTotalItems = useAtomValue(__discover_animeTotalItemsAtom)
    const setAnimeRandomNumber = useSetAtom(__discover_setAnimeRandomNumberAtom)

    const mangaRandomNumber = useAtomValue(__discover_mangaRandomNumberAtom)
    const mangaTotalItems = useAtomValue(__discover_mangaTotalItemsAtom)
    const setMangaRandomNumber = useSetAtom(__discover_setMangaRandomNumberAtom)

    const currentIndex = pageType === "anime" ? animeRandomNumber : mangaRandomNumber
    const totalItems = pageType === "anime" ? animeTotalItems : mangaTotalItems
    const setCurrentIndex = pageType === "anime" ? setAnimeRandomNumber : setMangaRandomNumber

    // Don't render if there are no items or only one item
    if (totalItems <= 1) return null

    const maxDots = Math.min(totalItems, 12)

    return (
        <div
            className={cn(
                "hidden lg:flex items-center gap-2 z-[10] pl-8 max-w-[20rem] flex-wrap top-[4.5rem]",
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
                    onClick={() => {
                        setCurrentIndex(index)
                        setClickedCarouselDot(n => n + 1)
                    }}
                    aria-label={`Go to slide ${index + 1}`}
                />
            ))}
        </div>
    )
}

function BannerImage({ media, isTransitioning, shouldBlurBanner, showTrailer, trailerLoaded, onTrailerLoad, onTrailerError }: BannerImageProps) {
    const ts = useThemeSettings()
    const bannerImage = media?.bannerImage || media?.coverImage?.extraLarge
    const trailerId = (media as AL_BaseAnime)?.trailer?.id
    const trailerSite = (media as AL_BaseAnime)?.trailer?.site || "youtube"

    return (
        <div
            className={cn(
                "lg:h-[35rem] w-full flex-none object-cover object-center absolute top-0 bg-[--background] overflow-hidden",
                !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                __isDesktop__ && "top-[-2rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[30rem]",
            )}
        >
            {/* Gradients */}
            <div className="w-full z-[2] absolute bottom-[-10rem] h-[10rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent" />
            <div className="w-full absolute z-[2] top-0 h-[10rem] opacity-50 bg-gradient-to-b from-[--background] to-transparent" />
            <div
                className={cn(
                    "opacity-0 duration-1000 bg-[var(--background)] w-full h-full absolute z-[2]",
                    isTransitioning && "opacity-70",
                )}
            />

            {/* Banner Image */}
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
                                !media?.bannerImage && "opacity-35",
                                trailerLoaded && "opacity-0",
                            )}
                        />
                    )}
                </div>
            </AnimatePresence>

            {/* Trailer */}
            {(showTrailer && trailerId && trailerSite === "youtube") && (
                <div
                    className={cn(
                        "absolute w-full h-full overflow-hidden z-[1]",
                        !trailerLoaded && "opacity-0",
                    )}
                >
                    <MotionIframe
                        src={`https://www.youtube-nocookie.com/embed/${trailerId}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&vq=medium&playlist=${trailerId}&cc_lang_pref=ja&enablejsapi=true`}
                        className="w-full h-full absolute left-0 object-cover object-center lg:scale-[1.8] 2xl:scale-[2.5]"
                        allow="autoplay"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: trailerLoaded ? 1 : 0 }}
                        transition={{ duration: 0.5 }}
                        onLoad={onTrailerLoad}
                        onError={onTrailerError}
                    />
                </div>
            )}

            {shouldBlurBanner && (
                <div className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2]" />
            )}

            <div
                className={cn(
                    "hidden lg:block max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-5% via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
                    "opacity-70 duration-500",
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

function MediaMetadata({ media, pageType, isTransitioning, onHoverChange }: MediaMetadataProps) {
    const ts = useThemeSettings()
    const { setPreviewModalMediaId } = useMediaPreviewModal()

    if (!media) return null

    return (
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
                <SeaLink href={pageType === "manga" ? `/manga/entry?id=${media.id}` : `/entry?id=${media.id}`}>
                    {media.coverImage?.large && (
                        <div className="w-[180px] h-[280px] relative rounded-[--radius-md] overflow-hidden bg-[--background] shadow-md">
                            <SeaImage
                                src={media.coverImage.large}
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
                <SeaLink href={pageType === "manga" ? `/manga/entry?id=${media.id}` : `/entry?id=${media.id}`}>
                    <TextGenerateEffect
                        className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white leading-8 line-clamp-2 pb-1 max-w-md text-pretty text-3xl overflow-ellipsis"
                        words={media.title?.userPreferred || ""}
                    />
                </SeaLink>

                <div className="flex flex-wrap gap-2">
                    {media.genres?.slice(0, 3).map((genre) => (
                        <div key={genre} className="text-sm font-semibold px-1 text-gray-300">
                            {genre}
                        </div>
                    ))}
                </div>

                <div className="flex items-center max-w-lg gap-4">
                    {media.meanScore && (
                        <div className="rounded-full w-fit inline-block">
                            <MediaEntryAudienceScore meanScore={media.meanScore} />
                        </div>
                    )}

                    {(media as AL_BaseAnime)?.nextAiringEpisode?.airingAt && (
                        <p className="text-base text-brand-200 inline-flex items-center gap-1.5">
                            <RiSignalTowerLine /> Releasing now
                        </p>
                    )}

                    {((media as AL_BaseAnime)?.nextAiringEpisode || (media as AL_BaseAnime).episodes) && (media as AL_BaseAnime)?.format !== "MOVIE" && (
                        <p className="text-base font-medium">
                            {(media as AL_BaseAnime).nextAiringEpisode?.episode ? (
                                <span>
                                    {(media as AL_BaseAnime).nextAiringEpisode?.episode! - 1} episode{(media as AL_BaseAnime).nextAiringEpisode?.episode! - 1 === 1
                                    ? ""
                                    : "s"} released
                                </span>
                            ) : (
                                <span>
                                    {(media as AL_BaseAnime).episodes} total episode{(media as AL_BaseAnime).episodes === 1 ? "" : "s"}
                                </span>
                            )}
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
                        {(media as any)?.description?.replace(/(<([^>]+)>)/ig, "")}
                    </ScrollArea>

                    <Button
                        size="sm"
                        intent="gray-outline"
                        className="rounded-full"
                        onClick={() => setPreviewModalMediaId(media?.id, pageType === "manga" ? "manga" : "anime")}
                    >
                        Preview
                    </Button>
                </motion.div>
            </motion.div>
        </motion.div>

    )
}

export function DiscoverPageHeader({ playTrailer }: { playTrailer?: boolean }) {
    const ts = useThemeSettings()
    const pathname = usePathname()

    const [pageType, setPageType] = useAtom(__discord_pageTypeAtom)
    const randomTrending = useAtomValue(__discover_randomTrendingAtom)
    const isTransitioning = useAtomValue(__discover_headerIsTransitioningAtom)
    const [isHoveringHeader, setHoveringHeader] = useAtom(__discover_hoveringHeaderAtom)


    const [showTrailer, setShowTrailer] = React.useState(false)
    const [trailerLoaded, setTrailerLoaded] = React.useState(false)
    const [trailerPlaying, setTrailerPlaying] = React.useState(false)
    const hoverTimerRef = React.useRef<NodeJS.Timeout | null>(null)
    const trailerPlayTimerRef = React.useRef<NodeJS.Timeout | null>(null)
    const trailerStopTimerRef = React.useRef<NodeJS.Timeout | null>(null)

    const shouldBlurBanner = ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable &&
        !randomTrending?.bannerImage

    // Reset page type to anime when on home page
    React.useLayoutEffect(() => {
        if (pathname === "/") {
            setPageType("anime")
        }
    }, [pathname])

    React.useEffect(() => {
        if (isTransitioning) {
            setShowTrailer(false)
            setTrailerLoaded(false)
            setTrailerPlaying(false)
        }
        if ((isHoveringHeader || playTrailer) && randomTrending && !isTransitioning && (randomTrending as AL_BaseAnime)?.trailer?.id) {
            hoverTimerRef.current = setTimeout(() => {
                setShowTrailer(true)
            }, 1000)
        }

        if (!isHoveringHeader && !playTrailer) {
            setShowTrailer(false)
            setTrailerLoaded(false)
            setTrailerPlaying(false)
        }

        return () => {
            if (hoverTimerRef.current) {
                clearTimeout(hoverTimerRef.current)
                hoverTimerRef.current = null
            }
        }
    }, [isHoveringHeader, randomTrending, isTransitioning, playTrailer])

    React.useEffect(() => {
        if (trailerLoaded && !trailerPlaying && !isHoveringHeader && (randomTrending as AL_BaseAnime)?.trailer?.id) {
            trailerPlayTimerRef.current = setTimeout(() => {
                if (!isHoveringHeader) {
                    setTrailerPlaying(true)

                    trailerStopTimerRef.current = setTimeout(() => {
                        if (!isHoveringHeader) {
                            setShowTrailer(false)
                            setTrailerLoaded(false)
                            setTrailerPlaying(false)
                        }
                    }, 6000)
                }
            }, 1000)
        }

        return () => {
            if (trailerPlayTimerRef.current) {
                clearTimeout(trailerPlayTimerRef.current)
                trailerPlayTimerRef.current = null
            }
            if (trailerStopTimerRef.current) {
                clearTimeout(trailerStopTimerRef.current)
                trailerStopTimerRef.current = null
            }
        }
    }, [trailerLoaded, trailerPlaying, isHoveringHeader, randomTrending])

    const handleTrailerLoad = () => {
        setTimeout(() => {
            setTrailerLoaded(true)
        }, 1000)
    }
    const handleTrailerError = () => {
        setShowTrailer(false)
        setTrailerLoaded(false)
    }

    if (!randomTrending) return <div>
        <Skeleton
            className={cn(
                "__header lg:h-[28rem] overflow-hidden",
                ts.hideTopNavbar && "lg:h-[32rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[24rem]",
                (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "lg:h-[28rem]",
            )}
        />
    </div>

    return (
        <motion.div
            className={cn(
                "__header lg:h-[28rem] overflow-hidden",
                ts.hideTopNavbar && "lg:h-[32rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[24rem]",
                (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "lg:h-[28rem]",
            )}
            {...{
                initial: { opacity: 0 },
                animate: { opacity: 1 },
                transition: { duration: 1.2 },
            }}
        >
            <BannerImage
                media={randomTrending}
                isTransitioning={isTransitioning}
                shouldBlurBanner={shouldBlurBanner}
                showTrailer={showTrailer}
                trailerLoaded={trailerLoaded}
                onTrailerLoad={handleTrailerLoad}
                onTrailerError={handleTrailerError}
            />

            <div
                className={cn(
                    "absolute left-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block",
                    "top-[5rem]",
                    ts.hideTopNavbar && "top-[4rem]",
                    (__isDesktop__ && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "top-[0rem]",
                    (__isDesktop__ && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "top-[2rem]",
                )}
                data-media-id={randomTrending.id}
                data-media-mal-id={randomTrending.idMal}
            >

                <HeaderCarouselDots />
                <AnimatePresence>
                    {randomTrending && !isTransitioning && (
                        <MediaMetadata
                            media={randomTrending}
                            pageType={pageType}
                            isTransitioning={isTransitioning}
                            onHoverChange={setHoveringHeader}
                        />
                    )}
                </AnimatePresence>
            </div>
        </motion.div>
    )
}
