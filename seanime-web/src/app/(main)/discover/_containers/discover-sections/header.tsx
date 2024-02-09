import { __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { RiSignalTowerLine } from "@react-icons/all-files/ri/RiSignalTowerLine"
import { useAtomValue } from "jotai"
import Image from "next/image"
import Link from "next/link"
import { useRouter } from "next/navigation"
import React from "react"

export function DiscoverPageHeader() {

    const router = useRouter()

    const randomTrending = useAtomValue(__discover_randomTrendingAtom)


    return (
        <div className={"__header h-[20rem]"}>
            <div
                className="h-[30rem] w-full md:w-[calc(100%-5rem)] flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className={"w-full absolute z-[2] top-0 h-[15rem] bg-gradient-to-b from-[--background-color] to-transparent via"}
                />
                {(!!randomTrending?.bannerImage || !!randomTrending?.coverImage?.extraLarge) && <Image
                    src={randomTrending.bannerImage || randomTrending.coverImage?.extraLarge!}
                    alt={"banner image"}
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className="object-cover object-center z-[1]"
                />}
                {!randomTrending?.bannerImage && <Skeleton className={"z-0 h-full absolute w-full"} />}
                {!!randomTrending && (
                    <div className={"absolute w-full flex justify-center bottom-16 z-[3] text-4xl font-bold h-fit flex-none leading-auto"}>
                        <p className="max-w-[30rem] line-clamp-2 text-center">
                            {randomTrending.title?.userPreferred}
                        </p>
                    </div>
                )}
                {!!randomTrending && (
                    <div
                        className={"absolute bottom-[6rem] right-2 w-fit h-[10rem] bg-gradient-to-t z-[3]"}
                    >
                        <div className={"flex flex-row-reverse relative items-start gap-6 p-6 w-fit overflow-hidden rounded-xl bg-[#121212] bg-opacity-80 shadow-2xl shadow-[#121212]"}>
                            <div className={"flex-none"}>
                                {randomTrending.coverImage?.large && <div
                                    className="w-[140px] h-[180px] relative rounded-md overflow-hidden bg-[--background-color] shadow-md border border-[--border]"
                                >
                                    <Image
                                        src={randomTrending.coverImage.large}
                                        alt={"cover image"}
                                        fill
                                        priority
                                        className="object-cover object-center"
                                    />
                                </div>}
                            </div>
                            <div className={"flex-auto space-y-1 z-[1]"}>
                                <h1 className={"text-lg text-gray-300 line-clamp-2 font-medium max-w-[16rem] leading-6"}>{randomTrending.title?.userPreferred}</h1>
                                {!!randomTrending?.nextAiringEpisode?.airingAt &&
                                    <p className="text-lg text-brand-200 flex items-center gap-1.5"><RiSignalTowerLine /> Airing now</p>}
                                {(!!randomTrending?.nextAiringEpisode || !!randomTrending.episodes) && (
                                    <p className="text-lg font-semibold">
                                        {!!randomTrending.nextAiringEpisode?.episode ?
                                            <span>{randomTrending.nextAiringEpisode.episode} episodes</span> :
                                            <span>{randomTrending.episodes} episodes</span>}
                                    </p>
                                )}
                                <div className="pt-2">
                                    <Link
                                        href={`/entry?id=${randomTrending.id}`}
                                    >
                                        <Button
                                            intent={"primary-outline"}
                                            size={"md"}
                                            className={"text-md w-[14rem] border-opacity-50 text-sm"}
                                        >
                                            Watch now
                                        </Button>
                                    </Link>
                                </div>
                                {/*<p className={"text-[--muted]"}>{randomTrending.}</p>*/}
                            </div>
                            <div
                                className="bg-[url(/pattern-1.svg)] z-[0] w-full h-full absolute opacity-50 top-0 left-0 bg-no-repeat bg-right bg-contain"
                            />
                        </div>
                    </div>
                )}
                <div
                    className={"w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background-color] via-[--background-color] via-opacity-50 via-10% to-transparent"}
                />
                <div
                    className={"absolute bottom-16 left-8 z-[3] cursor-pointer opacity-80 transition-opacity hover:opacity-100"}
                    onClick={() => router.push(`/search`)}
                >
                    <TextInput
                        leftIcon={<FiSearch />}
                        value={"Search by genres, seasons…"}
                        isReadOnly
                        size={"lg"}
                        className={"pointer-events-none w-96"}
                        onChange={() => {
                        }}
                    />
                </div>
            </div>
        </div>
    )

}
