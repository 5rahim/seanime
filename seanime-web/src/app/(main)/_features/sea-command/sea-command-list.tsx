import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { useSeaCommandContext } from "./sea-command"

// renders when "/" is typed
export function SeaCommandList() {

    const { input, setInput, select, command: { isCommand, command, args }, scrollToTop } = useSeaCommandContext()

    const commands = [
        {
            command: "anime",
            description: "Find anime in your collection",
            show: true,
        },
        {
            command: "manga",
            description: "Find manga in your collection",
            show: true,
        },
        {
            command: "search",
            description: "Search for anime or manga",
            show: true,
        },
    ]

    return (
        <>
            <CommandGroup heading="Suggestions">
                {commands.filter(n => n.show).map(command => (
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
