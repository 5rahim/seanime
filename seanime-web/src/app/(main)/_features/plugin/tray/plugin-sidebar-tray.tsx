import { PluginTray, TrayIcon } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { cn } from "@/components/ui/core/styling"
import { useAtom } from "jotai/react"
import { atom } from "jotai/vanilla"
import React from "react"
import { usePluginListenTrayIconEvent, usePluginSendListTrayIconsEvent } from "../generated/plugin-events"


export const __plugin_trayIconsAtom = atom<TrayIcon[]>([])
// FIXME
// TODO Remove
// export const __plugin_trayIconsAtom = atomWithStorage<TrayIcon[]>("TEST_ONLY-plugin-tray-icons", [], undefined, { getOnInit: true })

export function PluginSidebarTray() {
    const { sendMessage } = useWebsocketSender()

    const [trayIcons, setTrayIcons] = useAtom(__plugin_trayIconsAtom)

    /**
     * 1. Send a request to the server to list all tray icons
     * 2. Receive the tray icons from the server
     * 3. Set the tray icons in the state to display them
     */
    const { sendListTrayIconsEvent } = usePluginSendListTrayIconsEvent()

    React.useEffect(() => {
        // Send a request to all plugins to list their tray icons.
        // Only plugins with a registered tray icon will respond.
        sendListTrayIconsEvent({}, "")
    }, [])

    /**
     * TODO: Listen to other events from Extension Repository to refetch tray icons
     * - When an extension is loaded
     * - When an extension is unloaded
     * - When an extension is updated
     */

    usePluginListenTrayIconEvent((data, extensionId) => {
        setTrayIcons(prev => {
            const oldTrayIcons = prev.filter(icon => icon.extensionId !== extensionId)
            return [...oldTrayIcons, {
                extensionId,
                ...data,
            }].sort((a, b) => a.extensionId.localeCompare(b.extensionId, undefined, { numeric: true }))
        })
    }, "")

    if (!trayIcons) return null

    return (
        <>
            <div
                className={cn(
                    "w-10 mx-auto p-1",
                    "flex flex-col gap-1 items-center justify-center rounded-full border hover:border-[--border] transition-all duration-300",
                )}
            >

                {trayIcons.map((trayIcon, index) => (
                    <PluginTray trayIcon={trayIcon} key={index} />
                ))}

                {/* <IconButton
                    intent="gray-basic"
                    size="sm"
                    icon={<LuShapes />}
                    className="rounded-full"
                 /> */}
            </div>
        </>
    )
}
