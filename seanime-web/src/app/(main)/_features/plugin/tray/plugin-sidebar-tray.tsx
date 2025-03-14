import { useListDevelopmentModeExtensions, useReloadExternalExtension } from "@/api/hooks/extensions.hooks"
import { PluginTray, TrayIcon } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom } from "jotai/react"
import { atom } from "jotai/vanilla"
import { usePathname } from "next/navigation"
import React from "react"
import { LuBug, LuRefreshCw } from "react-icons/lu"
import { usePluginListenTrayIconEvent, usePluginSendListTrayIconsEvent } from "../generated/plugin-events"

export const __plugin_trayIconsAtom = atom<TrayIcon[]>([])

export const __plugin_hasNavigatedAtom = atom<boolean>(false)

export function PluginSidebarTray() {
    const queryClient = useQueryClient()
    const [trayIcons, setTrayIcons] = useAtom(__plugin_trayIconsAtom)

    const [hasNavigated, setHasNavigated] = useAtom(__plugin_hasNavigatedAtom)
    const pathname = usePathname()

    const firstRender = React.useRef(true)
    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (!hasNavigated) {
            setHasNavigated(true)
        }
    }, [pathname, hasNavigated])

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

    const { data: developmentModeExtensions, refetch } = useListDevelopmentModeExtensions()
    const { mutate: reloadExternalExtension, isPending: isReloadingExtension } = useReloadExternalExtension()

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (extensionId) => {
            setTrayIcons(prev => prev.filter(icon => icon.extensionId !== extensionId))
            setTimeout(() => {
                refetch()
            }, 1000)
        }
    })


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

                <Popover
                    side="right"
                    trigger={<div>
                        <IconButton
                            intent="gray-basic"
                            size="sm"
                            icon={<LuBug />}
                            className="rounded-full"
                        />
                    </div>}>
                        <div className="space-y-2">
                            {developmentModeExtensions?.map(extension => (
                                <div key={extension.id} className="flex items-center gap-2 justify-between bg-[--subtle] rounded-md p-2 px-4">
                                    <p className="text-sm font-medium">{extension.id}</p>
                                    <div>
                                        <IconButton
                                            intent="warning-basic"
                                            size="sm"
                                            icon={<LuRefreshCw className="size-5" />}
                                            className="rounded-full"
                                            onClick={() => reloadExternalExtension({ id: extension.id })}
                                            loading={isReloadingExtension}
                                        />
                                    </div>
                                </div>
                            ))}
                        </div>
                </Popover>
            </div>
        </>
    )
}
