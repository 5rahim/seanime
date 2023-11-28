import Image from "next/image"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import React from "react"
import { useAtomValue } from "jotai"
import { __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { useRouter } from "next/navigation"

export function DiscoverPageHeader() {

    const router = useRouter()

    const randomTrending = useAtomValue(__discover_randomTrendingAtom)


    return (
        <div className={"__header h-[20rem]"}>
            <div
                className="h-[30rem] w-full md:w-[calc(100%-5rem)] flex-none object-cover object-center absolute top-0 overflow-hidden">
                <div
                    className={"w-full absolute z-[2] top-0 h-[15rem] bg-gradient-to-b from-[--background-color] to-transparent via"}
                />
                {randomTrending?.bannerImage && <Image
                    src={randomTrending.bannerImage}
                    alt={"banner image"}
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className="object-cover object-center z-[1]"
                />}
                {!randomTrending?.bannerImage && <Skeleton className={"z-0 h-full absolute w-full"}/>}
                <div
                    className={"w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background-color] via-[--background-color] via-opacity-50 via-10% to-transparent"}
                />
                <div
                    className={"absolute bottom-16 right-8 z-[2] cursor-pointer opacity-80 transition-opacity hover:opacity-100"}
                    onClick={() => router.push(`/search`)}>
                    <TextInput
                        leftIcon={<FiSearch/>}
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