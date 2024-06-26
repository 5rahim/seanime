import { AL_BaseMedia } from "@/api/generated/types"
import { __discover_headerIsTransitioningAtom, __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-trending"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { AnimatePresence, motion } from "framer-motion"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React from "react"
import { RiSignalTowerLine } from "react-icons/ri"

export const __discover_hoveringHeaderAtom = atom(false)

const MotionImage = motion(Image)

export function DiscoverPageHeader() {

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

    return (
        <motion.div
            className="__header lg:h-[26rem]"
            {...{
                initial: { opacity: 0 },
                animate: { opacity: 1 },
                transition: {
                    duration: 1.2,
                },
            }}
        >
            <div
                className="CUSTOM_LIB_BANNER_FADE_BG w-full absolute z-[1] top-0 h-[48rem] opacity-100 bg-gradient-to-b from-[--background] via-[--background] via-75% to-transparent via"
            />
            <div

                className="lg:h-[35rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[10rem] opacity-50 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div
                    className={cn(
                        "opacity-0 duration-1000 bg-[var(--background)] w-full h-full absolute z-[2]",
                        isTransitioning && "opacity-70",
                    )}
                />
                <AnimatePresence>
                    {(!!randomTrending?.bannerImage || !!randomTrending?.coverImage?.extraLarge) && (
                        <MotionImage
                            src={randomTrending.bannerImage || randomTrending.coverImage?.extraLarge!}
                            alt="banner image"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            {...{
                                initial: { opacity: 1 },
                                animate: { opacity: 1 },
                                exit: { opacity: 0 },
                                transition: {
                                    duration: 1.2,
                                },
                            }}
                            className={cn(
                                "object-cover object-center z-[1] transition-opacity duration-1000",
                            )}
                        />
                    )}
                </AnimatePresence>

                <Image
                    src={"/mask.png"}
                    alt="mask"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-right z-[2] transition-opacity duration-1000 opacity-60 hidden lg:block",
                    )}
                />
                {!randomTrending?.bannerImage && <Skeleton className="z-0 h-full absolute w-full" />}
                <AnimatePresence>
                    {(!!randomTrending && !isTransitioning) && (
                        <motion.div
                            {...{
                                initial: { opacity: 0, y: -40 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, y: -40 },
                                transition: {
                                    type: "spring",
                                    damping: 20,
                                    stiffness: 100,
                                },
                            }}
                            className="absolute bottom-[8rem] right-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block"
                        >
                            <div
                                className="flex flex-row-reverse items-center relative gap-6 p-6 pr-3 w-fit overflow-hidden"
                                onMouseEnter={() => setHoveringHeader(true)}
                                onMouseLeave={() => setHoveringHeader(false)}
                            >
                                <div className="flex-none">
                                    {randomTrending.coverImage?.large && <div
                                        className="w-[180px] h-[240px] relative rounded-md overflow-hidden bg-[--background] shadow-md"
                                    >
                                        <Image
                                            src={randomTrending.coverImage.large}
                                            alt="cover image"
                                            fill
                                            priority
                                            className={cn(
                                                "object-cover object-center transition-opacity duration-1000",
                                                isTransitioning && "opacity-30",
                                                !isTransitioning && "opacity-100",
                                            )}
                                        />
                                    </div>}
                                </div>
                                <div className="flex-auto space-y-1 z-[1] text-center">
                                    <h1 className="text-3xl text-gray-200 leading-8 line-clamp-2 font-bold max-w-md">{randomTrending.title?.userPreferred}</h1>
                                    {!!(randomTrending as AL_BaseMedia)?.nextAiringEpisode &&
                                        <div className="flex items-center justify-center max-w-md gap-4">
                                            {!!(randomTrending as AL_BaseMedia)?.nextAiringEpisode?.airingAt &&
                                                <p className="text-lg text-brand-200 inline-flex items-center gap-1.5">
                                                    <RiSignalTowerLine /> Releasing now
                                                </p>}
                                            {(!!(randomTrending as AL_BaseMedia)?.nextAiringEpisode || !!(randomTrending as AL_BaseMedia).episodes) && (
                                                <p className="text-lg font-semibold">
                                                    {!!(randomTrending as AL_BaseMedia).nextAiringEpisode?.episode ?
                                                        <span>{(randomTrending as AL_BaseMedia).nextAiringEpisode?.episode} episodes</span> :
                                                        <span>{(randomTrending as AL_BaseMedia).episodes} episodes</span>}
                                                </p>
                                            )}
                                        </div>}
                                    <div className="pt-2">
                                        <ScrollArea className="max-w-md leading-6 h-[72px] mb-4">{(randomTrending as any)?.description?.replace(
                                            /(<([^>]+)>)/ig,
                                            "")}</ScrollArea>
                                        <Link
                                            href={pageType === "anime"
                                                ? `/entry?id=${randomTrending.id}`
                                                : `/manga/entry?id=${randomTrending.id}`}
                                        >
                                            <Button
                                                intent="white-outline"
                                                size="md"
                                                className="text-md w-[14rem] border-opacity-50 text-sm"
                                            >
                                                {randomTrending.status === "NOT_YET_RELEASED" ? "Preview" :
                                                    pageType === "anime" ? "Watch now" : "Read now"}
                                            </Button>
                                        </Link>
                                    </div>
                                </div>
                            </div>
                        </motion.div>
                    )}
                </AnimatePresence>
                <div
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                />
                {/*<motion.div*/}
                {/*    {...{*/}
                {/*        initial: { opacity: 0, x: -40 },*/}
                {/*        animate: { opacity: 1, x: 0 },*/}
                {/*        exit: { opacity: 0, x: -40 },*/}
                {/*        transition: {*/}
                {/*            delay: 1,*/}
                {/*            type: "spring",*/}
                {/*            damping: 20,*/}
                {/*            stiffness: 100,*/}
                {/*        },*/}
                {/*    }}*/}
                {/*    className="absolute bottom-16 left-8 z-[3] cursor-pointer opacity-80 transition-opacity hover:opacity-100 ring-brand hover:ring-2 rounded-md"*/}
                {/*    onClick={() => router.push(`/search`)}*/}
                {/*>*/}
                {/*    <TextInput*/}
                {/*        leftIcon={<FiSearch />}*/}
                {/*        value={"Search by genres, seasons…"}*/}
                {/*        readonly*/}
                {/*        size="lg"*/}
                {/*        className="pointer-events-none w-60 md:w-96"*/}
                {/*        onChange={() => {*/}
                {/*        }}*/}
                {/*    />*/}
                {/*</motion.div>*/}
            </div>
        </motion.div>
    )

}
