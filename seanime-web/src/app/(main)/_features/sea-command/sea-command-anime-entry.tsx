import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { CommandGroup, CommandItem } from "@/components/ui/command"
import React from "react"
import { useSeaCommandContext } from "./sea-command"

export function SeaCommandAnimeEntry() {
    const { params: { page, pageParams } } = useSeaCommandContext<"anime-entry">()

    if (!pageParams?.entry || page !== "anime-entry") return null

    return <Cmd />
}

function Cmd() {
    const { params: { page, pageParams }, input, setInput, close, command: { isCommand } } = useSeaCommandContext<"anime-entry">()
    const serverStatus = useServerStatus()
    const { currentView, setView } = useAnimeEntryPageView()

    const goToCommands = [
        {
            command: "library",
            description: "Downloaded episodes",
            show: currentView !== "library",
        },
        {
            command: "torrentstream",
            description: "Torrent streaming",
            show: currentView !== "torrentstream" && serverStatus?.torrentstreamSettings?.enabled,
        },
        {
            command: "debridstream",
            description: "Debrid streaming",
            show: currentView !== "debridstream" && serverStatus?.debridSettings?.enabled,
        },
        {
            command: "onlinestream",
            description: "Online streaming",
            show: currentView !== "onlinestream" && serverStatus?.settings?.library?.enableOnlinestream,
        },
    ]


    if (pageParams?.entry?.media?.status !== "RELEASING" && pageParams?.entry?.media?.status !== "FINISHED") return null

    return (
        <>
            <CommandGroup heading="Go to">
                {goToCommands.filter(command => command.show).map(command => (
                    <CommandItem
                        key={command.command}
                        onSelect={() => {
                            setView(command.command as any)
                            setInput("")
                        }}
                    >
                        {command.description}
                    </CommandItem>
                ))}
            </CommandGroup>
        </>
    )
}
