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
import { __discover_mangaTotalItemsAtom } from "@/app/(main)/discover/_containers/discover-trending-country"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaLink } from "@/components/shared/sea-link"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ThemeMediaPageBannerSize, ThemeMediaPageBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { AnimatePresence, motion } from "framer-motion"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import { usePathname } from "next/navigation"
import React from "react"
import { RiSignalTowerLine } from "react-icons/ri"
import { __discover_mangaRandomNumberAtom, __discover_setMangaRandomNumberAtom } from "../_containers/discover-trending-country"


export const __discover_hoveringHeaderAtom = atom(false)

const MotionImage = motion.create(Image)
const MotionIframe = motion.create("iframe")

type HeaderCarouselDotsProps = {
    className?: string
}

function HeaderCarouselDots({ className }: HeaderCarouselDotsProps) {
    const ts = useThemeSettings()
    const [pageType] = useAtom(__discord_pageTypeAtom)
    const pathname = usePathname()

    // Get the appropriate atoms based on the page type
    const animeRandomNumber = useAtomValue(__discover_animeRandomNumberAtom)
    const animeTotalItems = useAtomValue(__discover_animeTotalItemsAtom)
    const setAnimeRandomNumber = useSetAtom(__discover_setAnimeRandomNumberAtom)

    const mangaRandomNumber = useAtomValue(__discover_mangaRandomNumberAtom)
    const mangaTotalItems = useAtomValue(__discover_mangaTotalItemsAtom)
    const setMangaRandomNumber = useSetAtom(__discover_setMangaRandomNumberAtom)

    // Use the appropriate values based on the page type
    const currentIndex = pageType === "anime" ? animeRandomNumber : mangaRandomNumber
    const totalItems = pageType === "anime" ? animeTotalItems : mangaTotalItems
    const setCurrentIndex = pageType === "anime" ? setAnimeRandomNumber : setMangaRandomNumber

    // Don't render if there are no items or only one item
    if (totalItems <= 1) return null

    // Limit to a maximum of 10 dots
    const maxDots = Math.min(totalItems, 12)

    return (
        <div
            data-discover-page-header-carousel-dots
            className={cn(
                "absolute hidden lg:flex items-center justify-center gap-2 z-[10] pl-8",
                ts.hideTopNavbar && process.env.NEXT_PUBLIC_PLATFORM !== "desktop" && "top-[4rem]",
                process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "top-[2rem]",
                pathname === "/" && "hidden lg:hidden",
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
                    onClick={() => setCurrentIndex(index)}
                    aria-label={`Go to slide ${index + 1}`}
                />
            ))}
        </div>
    )
}

export function DiscoverPageHeader() {
    const ts = useThemeSettings()
    const pathname = usePathname()

    const [pageType, setPageType] = useAtom(__discord_pageTypeAtom)

    const randomTrending = useAtomValue(__discover_randomTrendingAtom)
    const isTransitioning = useAtomValue(__discover_headerIsTransitioningAtom)

    const [isHoveringHeader, setHoveringHeader] = useAtom(__discover_hoveringHeaderAtom)
    const [showTrailer, setShowTrailer] = React.useState(false)
    const [trailerLoaded, setTrailerLoaded] = React.useState(false)
    const hoverTimerRef = React.useRef<NodeJS.Timeout | null>(null)

    // Reset page type to anime when on home page
    React.useLayoutEffect(() => {
        if (pathname === "/") {
            setPageType("anime")
        }
    }, [pathname])

    // Handle hover with timer
    React.useEffect(() => {
        if (isHoveringHeader && !showTrailer && (randomTrending as AL_BaseAnime)?.trailer?.id) {
            hoverTimerRef.current = setTimeout(() => {
                setShowTrailer(true)
            }, 1000) // 1 second delay before showing trailer
        } else if (!isHoveringHeader) {
            if (hoverTimerRef.current) {
                clearTimeout(hoverTimerRef.current)
                hoverTimerRef.current = null
            }
            setShowTrailer(false)
            setTrailerLoaded(false)
        }

        return () => {
            if (hoverTimerRef.current) {
                clearTimeout(hoverTimerRef.current)
                hoverTimerRef.current = null
            }
        }
    }, [isHoveringHeader, showTrailer, randomTrending])

    const bannerImage = randomTrending?.bannerImage || randomTrending?.coverImage?.extraLarge
    const trailerId = (randomTrending as AL_BaseAnime)?.trailer?.id
    const trailerSite = (randomTrending as AL_BaseAnime)?.trailer?.site || "youtube"

    const shouldBlurBanner = (ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable && !randomTrending?.bannerImage)

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    return (
        <motion.div
            data-discover-page-header
            className={cn(
                "__header lg:h-[28rem]",
                ts.hideTopNavbar && "lg:h-[32rem]",
                ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[24rem]",
                (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "lg:h-[28rem]",
                // (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "lg:h-[32rem]",
                // (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "lg:h-[33rem]",
            )}
            {...{
                initial: { opacity: 0 },
                animate: { opacity: 1 },
                transition: {
                    duration: 1.2,
                },
            }}
        >
            {/* Carousel Rectangular Dots */}
            <HeaderCarouselDots />
            <div
                data-discover-page-header-banner-image-container
                className={cn(
                    "lg:h-[35rem] w-full overflow-hidden flex-none object-cover object-center absolute top-0 bg-[--background]",
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "top-[-2rem]",
                    ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "lg:h-[30rem]",
                )}
            >
                <div
                    data-discover-page-header-banner-image-top-gradient
                    className={cn(
                        "w-full z-[2] absolute bottom-[-10rem] h-[10rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                    )}
                />

                <div
                    data-discover-page-header-banner-image-top-gradient-2
                    className="w-full absolute z-[2] top-0 h-[10rem] opacity-50 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div
                    data-discover-page-header-banner-image-background
                    className={cn(
                        "opacity-0 duration-1000 bg-[var(--background)] w-full h-full absolute z-[2]",
                        isTransitioning && "opacity-70",
                    )}
                />
                <AnimatePresence>
                    {(!!bannerImage) && (
                        <MotionImage
                            data-discover-page-header-banner-image
                            src={bannerImage}
                            alt="banner image"
                            fill
                            quality={100}
                            priority
                            className={cn(
                                "object-cover object-center z-[1] transition-all duration-1000",
                                isTransitioning && "scale-[1.01] -translate-x-0.5",
                                !isTransitioning && "scale-100 translate-x-0",
                                !randomTrending?.bannerImage && "opacity-35",
                                trailerLoaded && "opacity-0",
                            )}
                        />
                    )}
                </AnimatePresence>

                {/* Trailer */}
                {(showTrailer && trailerId && trailerSite === "youtube") && (
                    <div
                        data-discover-page-header-trailer-container
                        className={cn(
                            "absolute w-full h-full overflow-hidden z-[1]",
                            !trailerLoaded && "opacity-0",
                        )}
                    >
                        <MotionIframe
                            data-discover-page-header-trailer
                            src={`https://www.youtube-nocookie.com/embed/${trailerId}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&vq=medium&playlist=${trailerId}&cc_lang_pref=ja`}
                            className="w-full h-full absolute left-0 object-cover object-center lg:scale-[1.8] 2xl:scale-[2.5]"
                            allow="autoplay"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: trailerLoaded ? 1 : 0 }}
                            transition={{ duration: 0.5 }}
                            onLoad={() => setTrailerLoaded(true)}
                            onError={() => {
                                setShowTrailer(false)
                                setTrailerLoaded(false)
                            }}
                        />
                    </div>
                )}
                {shouldBlurBanner && <div
                    data-discover-page-header-banner-image-blur
                    className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "
                ></div>}

                {/*LEFT FADE*/}
                <div
                    data-discover-page-header-banner-image-right-gradient
                    className={cn(
                        "hidden lg:block max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-5% via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
                        "opacity-100 duration-500",
                    )}
                />

                {/*LEFT FADE IF SIDEBAR IS TRANSPARENT*/}
                {!ts.disableSidebarTransparency && <div
                    data-discover-page-header-banner-image-left-gradient
                    className={cn(
                        "hidden lg:block max-w-[10rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
                        "opacity-70 duration-500",
                    )}
                />}
                <div
                    data-discover-page-header-banner-image-bottom-gradient
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                />
            </div>
            <AnimatePresence>
                {(!!randomTrending && !isTransitioning && (pageType === "anime" || pageType === "manga")) && (
                    <motion.div
                        data-discover-page-header-metadata-container
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
                        className={cn(
                            "absolute left-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block",
                            "top-[5rem]",
                            ts.hideTopNavbar && "top-[4rem]",
                            ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "top-[4rem]",
                            (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "top-[2rem]",
                            (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "top-[0rem]",
                            (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "top-[2rem]",
                        )}
                        data-media-id={randomTrending?.id}
                        data-media-mal-id={randomTrending?.idMal}
                    >
                        <div
                            data-discover-page-header-metadata-inner-container
                            className="flex items-center relative gap-6 p-6 pr-3 w-fit overflow-hidden"
                            onMouseEnter={() => setHoveringHeader(true)}
                            onMouseLeave={() => setHoveringHeader(false)}
                        >
                            <motion.div
                                className="flex-none"
                                initial={{ opacity: 0, scale: 0.7, skew: 5 }}
                                animate={{ opacity: 1, scale: 1, skew: 0 }}
                                exit={{ opacity: 1, scale: 1, skewY: 1 }}
                                transition={{ duration: 0.5 }}
                                data-discover-page-header-metadata-media-image-container
                            >
                                <SeaLink
                                    href={pageType === "anime"
                                        ? `/entry?id=${randomTrending.id}`
                                        : `/manga/entry?id=${randomTrending.id}`}
                                    data-discover-page-header-metadata-media-image-link
                                >
                                    {randomTrending.coverImage?.large && <div
                                        className="w-[180px] h-[280px] relative rounded-[--radius-md] overflow-hidden bg-[--background] shadow-md"
                                        data-discover-page-header-metadata-media-image-inner-container
                                    >
                                        <Image
                                            src={randomTrending.coverImage.large}
                                            alt="cover image"
                                            fill
                                            priority
                                            placeholder={imageShimmer(700, 475)}
                                            className={cn(
                                                "object-cover object-center transition-opacity duration-1000",
                                                isTransitioning && "opacity-30",
                                                !isTransitioning && "opacity-100",
                                            )}
                                            data-discover-page-header-metadata-media-image
                                        />
                                    </div>}
                                </SeaLink>
                            </motion.div>
                            <motion.div
                                className="flex-auto space-y-2 z-[1]"
                                initial={{ opacity: 0, x: 10 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ duration: 0.5, delay: 0.6 }}
                                data-discover-page-header-metadata-inner-container
                            >
                                <SeaLink
                                    href={pageType === "anime"
                                        ? `/entry?id=${randomTrending.id}`
                                        : `/manga/entry?id=${randomTrending.id}`}
                                    data-discover-page-header-metadata-media-title
                                >
                                    <TextGenerateEffect
                                        className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white leading-8 line-clamp-2 pb-1 max-w-md text-pretty text-3xl overflow-ellipsis"
                                        words={randomTrending.title?.userPreferred || ""}
                                    />
                                </SeaLink>
                                <div className="flex flex-wrap gap-2">
                                    {randomTrending.genres?.map((genre) => (
                                        <div key={genre} className="text-sm font-semibold px-1 text-gray-300">
                                            {genre}
                                        </div>
                                    ))}
                                </div>
                                {/*<h1 className="text-3xl text-gray-200 leading-8 line-clamp-2 font-bold max-w-md">{randomTrending.title?.userPreferred}</h1>*/}
                                <div className="flex items-center max-w-lg gap-4" data-discover-page-header-metadata-media-info>
                                    {randomTrending.meanScore &&
                                        <div className="rounded-full w-fit inline-block" data-discover-page-header-metadata-media-score>
                                            <MediaEntryAudienceScore
                                                meanScore={randomTrending.meanScore}
                                            />
                                        </div>}
                                    {!!(randomTrending as AL_BaseAnime)?.nextAiringEpisode?.airingAt &&
                                        <p
                                            className="text-base text-brand-200 inline-flex items-center gap-1.5"
                                            data-discover-page-header-metadata-media-airing-now
                                        >
                                            <RiSignalTowerLine /> Releasing now
                                        </p>}
                                    {((!!(randomTrending as AL_BaseAnime)?.nextAiringEpisode || !!(randomTrending as AL_BaseAnime).episodes) && (randomTrending as AL_BaseAnime)?.format !== "MOVIE") && (
                                        <p className="text-base font-medium" data-discover-page-header-metadata-media-episodes>
                                            {!!(randomTrending as AL_BaseAnime).nextAiringEpisode?.episode ?
                                                <span>{(randomTrending as AL_BaseAnime).nextAiringEpisode?.episode! - 1} episode{(randomTrending as AL_BaseAnime).nextAiringEpisode?.episode! - 1 === 1
                                                    ? ""
                                                    : "s"} released</span> :
                                                <span>{(randomTrending as AL_BaseAnime).episodes} total
                                                                                                  episode{(randomTrending as AL_BaseAnime).episodes === 1
                                                        ? ""
                                                        : "s"}</span>}
                                        </p>
                                    )}

                                </div>
                                <motion.div
                                    className="pt-0 left-0"
                                    initial={{ opacity: 0, x: 10 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    transition={{ duration: 0.5, delay: 0.7 }}
                                    data-discover-page-header-metadata-media-description-container
                                >
                                    <ScrollArea
                                        data-discover-page-header-metadata-media-description-scroll-area
                                        className="max-w-lg leading-3 h-[77px] mb-4 p-0 text-sm"
                                    >{(randomTrending as any)?.description?.replace(
                                        /(<([^>]+)>)/ig,
                                        "")}</ScrollArea>

                                    <Button
                                        size="sm" intent="gray-outline" className="rounded-full"
                                        // rightIcon={<ImEnlarge2 />}
                                        onClick={() => setPreviewModalMediaId(randomTrending?.id, pageType)}
                                    >
                                        Preview
                                    </Button>
                                    {/*<SeaLink*/}
                                    {/*    href={pageType === "anime"*/}
                                    {/*        ? `/entry?id=${randomTrending.id}`*/}
                                    {/*        : `/manga/entry?id=${randomTrending.id}`}*/}
                                    {/*>*/}
                                    {/*    <Button*/}
                                    {/*        intent="white-basic"*/}
                                    {/*        size="md"*/}
                                    {/*        className="text-md w-[14rem] border-opacity-50 text-sm"*/}
                                    {/*    >*/}
                                    {/*        {randomTrending.status === "NOT_YET_RELEASED" ? "Preview" :*/}
                                    {/*            pageType === "anime" ? "Watch now" : "Read now"}*/}
                                    {/*    </Button>*/}
                                    {/*</SeaLink>*/}
                                </motion.div>
                            </motion.div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </motion.div>
    )

}
