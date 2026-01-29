import { useGetAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { useRouter } from "next/navigation"
import React from "react"
import { BiArrowBack } from "react-icons/bi"
import { CommandHelperText, CommandItemMedia } from "./_components/command-utils"
import { useSeaCommandContext } from "./sea-command"
import { seaCommand_compareMediaTitles } from "./utils"

// only rendered when typing "/anime", "/library" or "/manga"
export function SeaCommandUserMediaNavigation() {

    const { input, select, command: { isCommand, command, args }, scrollToTop } = useSeaCommandContext()
    const { data: animeCollection, isLoading: isAnimeLoading } = useGetAnimeCollection() // should be available instantly
    const { data: mangaCollection, isLoading: isMangaLoading } = useGetMangaCollection()

    const anime = animeCollection?.MediaListCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []
    const manga = mangaCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []

    const router = useRouter()

    const query = args.join(" ")
    const filteredAnime = (command === "anime" && query.length > 0) ? anime.filter(n => seaCommand_compareMediaTitles(n.title, query)) : []
    const filteredManga = (command === "manga" && query.length > 0) ? manga.filter(n => seaCommand_compareMediaTitles(n.title, query)) : []

    return (
        <>
            {query.length === 0 && (
                <>
                    <CommandHelperText
                        command="/anime [title]"
                        description="Find anime in your collection"
                        show={command === "anime"}
                    />
                    <CommandHelperText
                        command="/manga [title]"
                        description="Find manga in your collection"
                        show={command === "manga"}
                    />
                    <CommandHelperText
                        command="/library [title]"
                        description="Find anime in your library"
                        show={command === "library"}
                    />
                </>
            )}

            {command === "anime" && filteredAnime.length > 0 && (
                <CommandGroup heading="My anime">
                    {filteredAnime.map(n => (
                        <CommandItem
                            key={n.id}
                            onSelect={() => {
                                select(() => {
                                    router.push(`/entry?id=${n.id}`)
                                })
                            }}
                        >
                            <CommandItemMedia media={n} type="anime" />
                        </CommandItem>
                    ))}
                </CommandGroup>
            )}
            {command === "manga" && filteredManga.length > 0 && (
                <CommandGroup heading="My manga">
                    {filteredManga.map(n => (
                        <CommandItem
                            key={n.id}
                            onSelect={() => {
                                select(() => {
                                    router.push(`/manga/entry?id=${n.id}`)
                                })
                            }}
                        >
                            <CommandItemMedia media={n} type="manga" />
                        </CommandItem>
                    ))}
                </CommandGroup>
            )}
        </>
    )
}

export function SeaCommandNavigation() {

    const serverStatus = useServerStatus()

    const { input, select, command: { isCommand, command, args } } = useSeaCommandContext()

    const router = useRouter()

    const pages = [
        {
            name: "Home",
            href: "/",
            flag: "home",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Schedule",
            href: "/schedule",
            flag: "schedule",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Settings",
            href: "/settings",
            flag: "settings",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Manga",
            href: "/manga",
            flag: "manga",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Discover",
            href: "/discover",
            flag: "discover",
            show: !serverStatus?.isOffline,
        },
        {
            name: "My lists",
            href: "/lists",
            flag: "lists",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Auto Downloader",
            href: "/auto-downloader",
            flag: "auto-downloader",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Torrent list",
            href: "/torrent-list",
            flag: "torrent-list",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Scan summaries",
            href: "/scan-summaries",
            flag: "scan-summaries",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Extensions",
            href: "/extensions",
            flag: "extensions",
            show: !serverStatus?.isOffline,
        },
        {
            name: "Advanced search",
            href: "/search",
            flag: "search",
            show: !serverStatus?.isOffline,
        },
    ]

    // If no args, show all pages
    // If args, show pages that match the args
    const filteredPages = pages.filter(page => page.flag.startsWith(command))


    // if (!input.startsWith("/")) return null


    return (
        <>
            {command.startsWith("ba") && (
                <CommandGroup heading="Navigation">
                    <CommandItem
                        onSelect={() => {
                            select(() => {
                                router.back()
                            })
                        }}
                    >
                        <BiArrowBack className="mr-2 h-4 w-4" />
                        <span>Go back</span>
                    </CommandItem>
                </CommandGroup>
            )}
            {command.startsWith("fo") && (
                <CommandGroup heading="Navigation">
                    <CommandItem
                        onSelect={() => {
                            select(() => {
                                router.forward()
                            })
                        }}
                    >
                        <BiArrowBack className="mr-2 h-4 w-4 rotate-180" />
                        <span>Go forward</span>
                    </CommandItem>
                </CommandGroup>
            )}

            {/*Typing `/library`, `/schedule`, etc. without args*/}
            {isCommand && filteredPages.length > 0 && args.length === 0 && (
                <CommandGroup heading="Screens">
                    <>
                        {filteredPages.filter(page => page.show).map(page => (
                            <CommandItem
                                key={page.flag}
                                onSelect={() => {
                                    select(() => {
                                        router.push(page.href)
                                    })
                                }}
                            >
                                <span className="text-sm tracking-wide font-bold text-[--muted]">Go to:&nbsp;</span>{" "}{page.name}
                                {command === page.flag ? <CommandShortcut>Enter</CommandShortcut> : <CommandShortcut>/{page.flag}</CommandShortcut>}
                            </CommandItem>
                        ))}
                    </>
                </CommandGroup>
            )}
            {(command !== "back" && command !== "forward") && (
                <CommandGroup heading="Navigation">
                    {/* {command === "" && ( */}
                    <>
                        <CommandItem
                            onSelect={() => {
                                select(() => {
                                    router.back()
                                })
                            }}
                        >
                            <BiArrowBack className="mr-2 h-4 w-4" />
                            <span>Go back</span>
                        </CommandItem>
                        <CommandItem
                            onSelect={() => {
                                select(() => {
                                    router.forward()
                                })
                            }}
                        >
                            <BiArrowBack className="mr-2 h-4 w-4 rotate-180" />
                            <span>Go forward</span>
                        </CommandItem>
                    </>
                    {/* )} */}
                </CommandGroup>
            )}
        </>
    )
}
