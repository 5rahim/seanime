import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { usePathname, useSearchParams } from "@/lib/navigation"
import { useSeaCommandContext } from "./sea-command"

// renders when "/" is typed
export function SeaCommandList() {

    const pathname = usePathname()
    const searchParams = useSearchParams()
    const mediaId = Number(searchParams.get("id"))
    const isAnimePage = (pathname === "/entry" || pathname === "/offline/entry/anime") && Number.isFinite(mediaId) && mediaId > 0

    const { input, setInput, select, command: { isCommand, command, args }, scrollToTop } = useSeaCommandContext()

    const commands = [
        {
            command: "anime",
            description: "Find in your collection",
            show: true,
        },
        {
            command: "manga",
            description: "Find in your collection",
            show: true,
        },
        {
            command: "library",
            description: "Find in your anime library",
            show: true,
        },
        {
            command: "search",
            description: "Search on AniList",
            show: true,
        },
        {
            command: "magnet",
            description: "Stream or download via magnet link",
            show: true,
        },
        {
            command: "logs",
            description: "Copy the current logs",
            show: true,
        },
        {
            command: "issue",
            description: "Record an issue",
            show: true,
        },
        {
            command: "droptorrent",
            description: "Drop current torrentstream torrent",
            show: input.startsWith("/d"),
        },
        {
            command: "reload",
            description: "Reload the page",
            show: input.startsWith("/r"),
        },
        {
            command: "spoilers",
            description: "Toggle spoilers for this anime",
            show: isAnimePage,
        },
    ]

    const filtered = commands.filter(n => n.show && n.command.startsWith(command) && n.command != command)

    if (!filtered?.length) return null

    return (
        <>
            <CommandGroup heading="Autocomplete">
                {filtered.map(command => (
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
