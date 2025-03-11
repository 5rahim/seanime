import { useAtom } from "jotai/react"
import { atom } from "jotai/vanilla"
import React from "react"
import { usePluginListenCommandPaletteInfoEvent, usePluginSendListCommandPalettesEvent } from "../generated/plugin-events"
import { PluginCommandPalette, PluginCommandPaletteInfo } from "./plugin-command-palette"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"


export const __plugin_commandPalettesAtom = atom<PluginCommandPaletteInfo[]>([])

export function PluginCommandPalettes() {
    const [commandPalettes, setCommandPalettes] = useAtom(__plugin_commandPalettesAtom)

    /**
     * 1. Send a request to the server to list all command palettes
     * 2. Receive the command palettes from the server
     * 3. Set the command palettes in the state to display them
     */
    const { sendListCommandPalettesEvent } = usePluginSendListCommandPalettesEvent()

    React.useEffect(() => {
        // Send a request to all plugins to list their command palettes.
        // Only plugins with a registered command palette will respond.
        sendListCommandPalettesEvent({}, "")
    }, [])

    /**
     * TODO: Listen to other events from Extension Repository to refetch command palettes
     * - When an extension is loaded
     * - When an extension is unloaded
     * - When an extension is updated
     */

    usePluginListenCommandPaletteInfoEvent((data, extensionId) => {
        if (data.keyboardShortcut === "q" || data.keyboardShortcut === "meta+j") return

        setCommandPalettes(prev => {
            const oldCommandPalettes = prev.filter(palette => palette.extensionId !== extensionId)
            return [...oldCommandPalettes, {
                extensionId,
                ...data,
            }].sort((a, b) => a.extensionId.localeCompare(b.extensionId, undefined, { numeric: true }))
        })
    }, "")

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (extensionId) => {
            setCommandPalettes(prev => prev.filter(palette => palette.extensionId !== extensionId))
        }
    })

    if (!commandPalettes) return null

    return (
        <>
            {commandPalettes.map((palette, index) => (
                <PluginCommandPalette extensionId={palette.extensionId} info={palette} key={index} />
            ))}
        </>
    )
}
