import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { useSeaCommandContext } from "./sea-command"

// renders when "/" is typed
export function SeaCommandList() {

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
            command: "search",
            description: "Search on AniList",
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
