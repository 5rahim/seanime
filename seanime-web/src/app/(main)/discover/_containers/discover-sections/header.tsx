import { __discover_headerIsTransitioningAtom, __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { motion } from "framer-motion"
import { atom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import Image from "next/image"
import Link from "next/link"
import { useRouter } from "next/navigation"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { RiSignalTowerLine } from "react-icons/ri"

export const __discover_hoveringHeaderAtom = atom(false)

export function DiscoverPageHeader() {

    const router = useRouter()

    const randomTrending = useAtomValue(__discover_randomTrendingAtom)
    const isTransitioning = useAtomValue(__discover_headerIsTransitioningAtom)

    const setHoveringHeader = useSetAtom(__discover_hoveringHeaderAtom)

    return (
        <motion.div
            {...{
                initial: { opacity: 0 },
                animate: { opacity: 1 },
                exit: { opacity: 0 },
                transition: { delay: 0.2, duration: 0.2 },
            }}
            className="__header lg:h-[26rem]"
        >
            <div
                className="lg:h-[35rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                {(!!randomTrending?.bannerImage || !!randomTrending?.coverImage?.extraLarge) && <Image
                    src={randomTrending.bannerImage || randomTrending.coverImage?.extraLarge!}
                    alt="banner image"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-center z-[1] transition-opacity duration-1000",
                        isTransitioning && "opacity-10",
                        !isTransitioning && "opacity-100",
                    )}
                />}

                <Image
                    src={"/mask.png"}
                    alt="mask"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-right z-[2] transition-opacity duration-1000 opacity-90 hidden lg:block",
                    )}
                />
                {!randomTrending?.bannerImage && <Skeleton className="z-0 h-full absolute w-full" />}
                {!!randomTrending && (
                    <motion.div
                        {...{
                            initial: { opacity: 0, y: -40 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, y: -40 },
                            transition: {
                                delay: 0.5,
                                type: "spring",
                                damping: 20,
                                stiffness: 100,
                            },
                        }}
                        className="absolute bottom-[8rem] right-2 w-fit h-[20rem] bg-gradient-to-t z-[3] hidden lg:block"
                    >
                        <div
                            className="flex flex-row-reverse relative items-start gap-6 p-6 pr-3 w-fit overflow-hidden"
                            onMouseEnter={() => setHoveringHeader(true)}
                            onMouseLeave={() => setHoveringHeader(false)}
                        >
                            <div className="flex-none">
                                {randomTrending.coverImage?.large && <div
                                    className="w-[180px] h-[240px] relative rounded-md overflow-hidden bg-[--background] shadow-md border"
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
                                <div className="flex items-center justify-center max-w-md gap-4">
                                    {!!randomTrending?.nextAiringEpisode?.airingAt &&
                                        <p className="text-lg text-brand-200 inline-flex items-center gap-1.5"><RiSignalTowerLine /> Airing now</p>}
                                    {(!!randomTrending?.nextAiringEpisode || !!randomTrending.episodes) && (
                                        <p className="text-lg font-semibold">
                                            {!!randomTrending.nextAiringEpisode?.episode ?
                                                <span>{randomTrending.nextAiringEpisode.episode} episodes</span> :
                                                <span>{randomTrending.episodes} episodes</span>}
                                        </p>
                                    )}
                                </div>
                                <div className="pt-2">
                                    <ScrollArea className="max-w-md h-[75px] mb-4">{(randomTrending as any)?.description?.replace(
                                        /(<([^>]+)>)/ig,
                                        "")}</ScrollArea>
                                    <Link
                                        href={`/entry?id=${randomTrending.id}`}
                                    >
                                        <Button
                                            intent="primary-outline"
                                            size="md"
                                            className="text-md w-[14rem] border-opacity-50 text-sm"
                                        >
                                            Watch now
                                        </Button>
                                    </Link>
                                </div>
                            </div>
                        </div>
                    </motion.div>
                )}
                <div
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                />
                <motion.div
                    {...{
                        initial: { opacity: 0, x: -40 },
                        animate: { opacity: 1, x: 0 },
                        exit: { opacity: 0, x: -40 },
                        transition: {
                            delay: 1,
                            type: "spring",
                            damping: 20,
                            stiffness: 100,
                        },
                    }}
                    className="absolute bottom-16 left-8 z-[3] cursor-pointer opacity-80 transition-opacity hover:opacity-100"
                    onClick={() => router.push(`/search`)}
                >
                    <TextInput
                        leftIcon={<FiSearch />}
                        value={"Search by genres, seasons…"}
                        readonly
                        size="lg"
                        className="pointer-events-none w-60 md:w-96"
                        onChange={() => {
                        }}
                    />
                </motion.div>
            </div>
        </motion.div>
    )

}
