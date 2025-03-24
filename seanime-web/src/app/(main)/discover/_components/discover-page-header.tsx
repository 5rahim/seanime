import { AL_BaseAnime } from "@/api/generated/types"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { __discover_headerIsTransitioningAtom, __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-trending"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaLink } from "@/components/shared/sea-link"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
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

export const __discover_hoveringHeaderAtom = atom(false)

const MotionImage = motion.create(Image)

export function DiscoverPageHeader() {
    const ts = useThemeSettings()
    const pathname = usePathname()

    const [pageType, setPageType] = useAtom(__discord_pageTypeAtom)

    const randomTrending = useAtomValue(__discover_randomTrendingAtom)
    const isTransitioning = useAtomValue(__discover_headerIsTransitioningAtom)

    const setHoveringHeader = useSetAtom(__discover_hoveringHeaderAtom)

    // Reset page type to anime when on home page
    React.useLayoutEffect(() => {
        if (pathname === "/") {
            setPageType("anime")
        }
    }, [pathname])

    const bannerImage = randomTrending?.bannerImage || randomTrending?.coverImage?.extraLarge

    const shouldBlurBanner = (ts.mediaPageBannerType === ThemeMediaPageBannerType.BlurWhenUnavailable && !randomTrending?.bannerImage)
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
                            )}
                        />
                    )}
                </AnimatePresence>
                {shouldBlurBanner && <div
                    data-discover-page-header-banner-image-blur
                    className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "
                ></div>}

                {/*RIGHT FADE*/}
                <div
                    data-discover-page-header-banner-image-right-gradient
                    className={cn(
                        "hidden lg:block max-w-[60rem] w-full z-[2] h-full absolute right-0 bg-gradient-to-l from-[--background] from-5% via-[--background] transition-opacity via-opacity-50 via-5% to-transparent",
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
                            "absolute right-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block",
                            "top-[5rem]",
                            ts.hideTopNavbar && "top-[4rem]",
                            ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "top-[4rem]",
                            (ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && ts.hideTopNavbar) && "top-[2rem]",
                            (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small) && "top-[0rem]",
                            (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && ts.mediaPageBannerSize !== ThemeMediaPageBannerSize.Small) && "top-[2rem]",
                        )}
                    >
                        <div
                            data-discover-page-header-metadata-inner-container
                            className="flex flex-row-reverse items-center relative gap-6 p-6 pr-3 w-fit overflow-hidden"
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
                                        className="w-[190px] h-[290px] relative rounded-[--radius-md] overflow-hidden bg-[--background] shadow-md"
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
                                className="flex-auto space-y-2 z-[1] text-center"
                                initial={{ opacity: 0, x: 10 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ duration: 0.5, delay: 0.6 }}
                                data-discover-page-header-metadata-container
                            >
                                <SeaLink
                                    href={pageType === "anime"
                                        ? `/entry?id=${randomTrending.id}`
                                        : `/manga/entry?id=${randomTrending.id}`}
                                    data-discover-page-header-metadata-media-title
                                >
                                    <TextGenerateEffect
                                        className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] text-white leading-8 line-clamp-2 pb-1 text-center max-w-md text-pretty text-3xl overflow-ellipsis"
                                        words={randomTrending.title?.userPreferred || ""}
                                    />
                                </SeaLink>
                                {/*<h1 className="text-3xl text-gray-200 leading-8 line-clamp-2 font-bold max-w-md">{randomTrending.title?.userPreferred}</h1>*/}
                                <div className="flex justify-center items-center max-w-md gap-4" data-discover-page-header-metadata-media-info>
                                    {!!(randomTrending as AL_BaseAnime)?.nextAiringEpisode?.airingAt &&
                                        <p
                                            className="text-lg text-brand-200 inline-flex items-center gap-1.5"
                                            data-discover-page-header-metadata-media-airing-now
                                        >
                                            <RiSignalTowerLine /> Releasing now
                                        </p>}
                                    {((!!(randomTrending as AL_BaseAnime)?.nextAiringEpisode || !!(randomTrending as AL_BaseAnime).episodes) && (randomTrending as AL_BaseAnime)?.format !== "MOVIE") && (
                                        <p className="text-lg font-semibold" data-discover-page-header-metadata-media-episodes>
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
                                    {randomTrending.meanScore &&
                                        <div className="rounded-full w-fit inline-block" data-discover-page-header-metadata-media-score>
                                        <MediaEntryAudienceScore
                                            meanScore={randomTrending.meanScore}
                                        />
                                    </div>}
                                </div>
                                <motion.div
                                    className="pt-2"
                                    initial={{ opacity: 0, x: 10 }}
                                    animate={{ opacity: 1, x: 10 }}
                                    transition={{ duration: 0.5, delay: 0.7 }}
                                    data-discover-page-header-metadata-media-description-container
                                >
                                    <ScrollArea
                                        data-discover-page-header-metadata-media-description-scroll-area
                                        className="max-w-md leading-6 h-[72px] mb-4"
                                    >{(randomTrending as any)?.description?.replace(
                                        /(<([^>]+)>)/ig,
                                        "")}</ScrollArea>
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
