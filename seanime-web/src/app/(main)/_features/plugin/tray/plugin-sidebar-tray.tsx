import { Extension_Extension } from "@/api/generated/types"
import { useListDevelopmentModeExtensions, useReloadExternalExtension } from "@/api/hooks/extensions.hooks"
import { PluginTray, TrayIcon } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { WSEvents } from "@/lib/server/ws-events"
import { useWindowSize } from "@uidotdev/usehooks"
import { useAtom } from "jotai/react"
import { atom } from "jotai/vanilla"
import { usePathname } from "next/navigation"
import React from "react"
import { LuBlocks, LuBug, LuRefreshCw } from "react-icons/lu"
import { usePluginListenTrayIconEvent, usePluginSendListTrayIconsEvent } from "../generated/plugin-events"

export const __plugin_trayIconsAtom = atom<TrayIcon[]>([])

export const __plugin_hasNavigatedAtom = atom<boolean>(false)

const ExtensionList = ({
    place,
    developmentModeExtensions,
    isReloadingExtension,
    reloadExternalExtension,
    trayIcons,
    width,
}: {
    place: "sidebar" | "top";
    developmentModeExtensions: Extension_Extension[];
    isReloadingExtension: boolean;
    reloadExternalExtension: (params: { id: string }) => void;
    trayIcons: TrayIcon[];
    width: number | null;
}) => {
    return (
        <>
            <div
                data-plugin-sidebar-tray
                className={cn(
                    "w-10 mx-auto p-1 my-2",
                    "flex flex-col gap-1 items-center justify-center rounded-full border hover:border-[--border] transition-all duration-300",
                    place === "top" && "flex-row w-auto my-0 justify-start px-2 py-2 border-none",
                )}
            >

                <Popover
                    side={place === "top" ? "bottom" : "right"}
                    trigger={<div>
                        <IconButton
                            intent="gray-basic"
                            size="sm"
                            icon={<LuBug className="size-4 text-[--orange]" />}
                            className="rounded-full"
                        />
                    </div>}
                    className="p-2"
                    data-plugin-sidebar-debug-popover
                >
                    <div className="space-y-2" data-plugin-sidebar-debug-popover-content>
                        {developmentModeExtensions?.map(extension => (
                            <div key={extension.id} className="flex items-center gap-2 justify-between bg-[--subtle] rounded-md p-2">
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

                {trayIcons.map((trayIcon, index) => (
                    <PluginTray trayIcon={trayIcon} key={index} place={place} width={width} />
                ))}
            </div>
        </>
    )
}

export function PluginSidebarTray({ place }: { place: "sidebar" | "top" }) {
    const { width } = useWindowSize()
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

    const isMobile = width && width < 1024


    if (!trayIcons) return null

    return (
        <>
            {!isMobile && place === "sidebar" && <ExtensionList
                place={place}
                developmentModeExtensions={developmentModeExtensions || []}
                isReloadingExtension={isReloadingExtension}
                reloadExternalExtension={reloadExternalExtension}
                trayIcons={trayIcons}
                width={width}
            />}
            {isMobile && place === "top" && <div className="">
                <Popover
                    side="bottom"
                    trigger={<div>
                        <IconButton
                            intent="gray-basic"
                            size="sm"
                            icon={<LuBlocks />}
                        />
                    </div>}
                    className="rounded-full p-0 overflow-y-auto bg-black/80 mx-4"
                    style={{
                        width: width ? width - 50 : "100%",
                        // transform: "translateX(10px)",
                    }}
                >
                    <ExtensionList
                        place={place}
                        developmentModeExtensions={developmentModeExtensions || []}
                        isReloadingExtension={isReloadingExtension}
                        reloadExternalExtension={reloadExternalExtension}
                        trayIcons={trayIcons}
                        width={width}
                    />
                </Popover>
            </div>}
        </>
    )
}
