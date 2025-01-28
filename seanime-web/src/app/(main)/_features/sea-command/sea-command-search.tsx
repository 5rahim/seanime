import { useAnilistListAnime } from "@/api/hooks/anilist.hooks"
import { useAnilistListManga } from "@/api/hooks/manga.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CommandGroup, CommandItem } from "@/components/ui/command"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useDebounce } from "@/hooks/use-debounce"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React from "react"
import { CommandHelperText, CommandItemMedia } from "./_components/command-utils"
import { useSeaCommandContext } from "./sea-command"

export function SeaCommandSearch() {

    const serverStatus = useServerStatus()

    const { input, select, scrollToTop, commandListRef, command: { isCommand, command, args } } = useSeaCommandContext()

    const router = useRouter()

    const animeSearchInput = args.join(" ")
    const mangaSearchInput = args.slice(1).join(" ")
    const type = args[0] !== "manga" ? "anime" : "manga"

    const debouncedQuery = useDebounce(type === "anime" ? animeSearchInput : mangaSearchInput, 500)

    const { data: animeData, isLoading: animeIsLoading, isFetching: animeIsFetching } = useAnilistListAnime({
        search: debouncedQuery,
        page: 1,
        perPage: 10,
        status: ["FINISHED", "CANCELLED", "NOT_YET_RELEASED", "RELEASING"],
        sort: ["SEARCH_MATCH"],
    }, debouncedQuery.length > 0 && type === "anime")

    const { data: mangaData, isLoading: mangaIsLoading, isFetching: mangaIsFetching } = useAnilistListManga({
        search: debouncedQuery,
        page: 1,
        perPage: 10,
        status: ["FINISHED", "CANCELLED", "NOT_YET_RELEASED", "RELEASING"],
        sort: ["SEARCH_MATCH"],
    }, debouncedQuery.length > 0 && type === "manga")

    const isLoading = type === "anime" ? animeIsLoading : mangaIsLoading
    const isFetching = type === "anime" ? animeIsFetching : mangaIsFetching

    const media = React.useMemo(() => type === "anime" ? animeData?.Page?.media?.filter(Boolean) : mangaData?.Page?.media?.filter(Boolean),
        [animeData, mangaData, type])

    React.useEffect(() => {
        const cl = scrollToTop()
        return () => cl()
    }, [input, isLoading, isFetching])


    return (
        <>
            {(animeSearchInput === "" && mangaSearchInput === "") ? (
                <>
                    <CommandHelperText
                        command="/search [title]"
                        description="Search anime"
                        show={true}
                    />
                    <CommandHelperText
                        command="/search manga [title]"
                        description="Search manga"
                        show={true}
                    />
                </>
            ) : (

                <CommandGroup heading={`${type === "anime" ? "Anime" : "Manga"} results`}>
                    {(debouncedQuery !== "" && (!media || media.length === 0) && (isLoading || isFetching)) && (
                        <LoadingSpinner />
                    )}
                    {debouncedQuery !== "" && !isLoading && !isFetching && (!media || media.length === 0) && (
                        <div className="py-14 px-6 text-center text-sm sm:px-14">
                            {<div
                                className="h-[10rem] w-[10rem] mx-auto flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                            >
                                <Image
                                    src="/luffy-01.png"
                                    alt={""}
                                    fill
                                    quality={100}
                                    priority
                                    sizes="10rem"
                                    className="object-contain object-top"
                                />
                            </div>}
                            <h5 className="mt-4 font-semibold text-[--foreground]">Nothing
                                                                                   found</h5>
                            <p className="mt-2 text-[--muted]">
                                We couldn't find anything with that name. Please try again.
                            </p>
                        </div>
                    )}
                    {media?.map(item => (
                        <CommandItem
                            key={item?.id || ""}
                            onSelect={() => {
                                select(() => {
                                    if (type === "anime") {
                                        router.push(`/entry?id=${item.id}`)
                                    } else {
                                        router.push(`/manga/entry?id=${item.id}`)
                                    }
                                })
                            }}
                        >
                            <CommandItemMedia media={item} />
                        </CommandItem>
                    ))}
                </CommandGroup>
            )}
        </>
    )
}
