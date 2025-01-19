import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { useRouter } from "next/navigation"
import { BiArrowBack } from "react-icons/bi"
import { useSeaCommandContext } from "./sea-command"

// renders when "/" is typed
export function SeaCommandList() {

    const { params, input, setInput, select, command: { isCommand, command, args }, scrollToTop } = useSeaCommandContext<"other">()

    const commands = [
        {
            command: "anime",
            description: "Find anime in your collection",
        },
        {
            command: "manga",
            description: "Find manga in your collection",
        },
        {
            command: "search",
            description: "Search for anime or manga",
        },
    ]

    return (
        <>
            <CommandGroup heading="Suggestions">
                {commands.map(command => (
                    <CommandItem
                        key={command.command}
                        onSelect={() => {
                            setInput(`/${command.command}`)
                        }}
                    >
                        <span className="tracking-widest text-sm">/{command.command}</span>
                        <CommandShortcut className="text-[--muted]">{command.description}</CommandShortcut>
                    </CommandItem>
                ))}
            </CommandGroup>
        </>
    )
}

export function SeaCommandNavigation() {

    const serverStatus = useServerStatus()

    const { params, input, select, command: { isCommand, command, args } } = useSeaCommandContext<"other">()

    const router = useRouter()

    const pages = [
        {
            name: "My library",
            href: "/",
            flag: "library",
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
            name: "AniList",
            href: "/anilist",
            flag: "anilist",
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
            name: "Search",
            href: "/search",
            flag: "search",
            show: !serverStatus?.isOffline,
        },
    ]

    // If no args, show all pages
    // If args, show pages that match the args
    const filteredPages = pages.filter(page => page.flag.includes(command))


    // if (!input.startsWith("/")) return null


    return (
        <>
            {command === "back" && (
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
            )}
            {command === "forward" && (
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
                                {page.name}
                                <CommandShortcut>/{page.flag}</CommandShortcut>
                            </CommandItem>
                        ))}
                    </>
                </CommandGroup>
            )}
            {(command !== "back" && command !== "forward" && params.page === "other") && (
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
