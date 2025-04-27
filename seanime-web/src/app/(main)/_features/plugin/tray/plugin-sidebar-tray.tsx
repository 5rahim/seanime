import { Extension_Extension, ExtensionRepo_StoredPluginSettingsData } from "@/api/generated/types"
import {
    useGetPluginSettings,
    useListDevelopmentModeExtensions,
    useReloadExternalExtension,
    useSetPluginSettingsPinnedTrays,
} from "@/api/hooks/extensions.hooks"
import { PluginTray, TrayIcon } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { Tooltip } from "@/components/ui/tooltip"
import { WSEvents } from "@/lib/server/ws-events"
import { useWindowSize } from "@uidotdev/usehooks"
import { useAtom } from "jotai/react"
import { atom } from "jotai/vanilla"
import Image from "next/image"
import { usePathname } from "next/navigation"
import React from "react"
import { LuBlocks, LuBug, LuCircleDashed, LuRefreshCw, LuShapes } from "react-icons/lu"
import { TbPinned, TbPinnedFilled } from "react-icons/tb"
import { usePluginListenTrayIconEvent, usePluginSendListTrayIconsEvent } from "../generated/plugin-events"

export const __plugin_trayIconsAtom = atom<TrayIcon[]>([])

export const __plugin_hasNavigatedAtom = atom<boolean>(false)

export const __plugin_unpinnedTrayIconClickedAtom = atom<TrayIcon | null>(null)

const ExtensionList = ({
    place,
    developmentModeExtensions,
    isReloadingExtension,
    reloadExternalExtension,
    trayIcons,
    settings,
    width,
}: {
    place: "sidebar" | "top";
    developmentModeExtensions: Extension_Extension[];
    isReloadingExtension: boolean;
    reloadExternalExtension: (params: { id: string }) => void;
    trayIcons: TrayIcon[];
    settings: ExtensionRepo_StoredPluginSettingsData | undefined;
    width: number | null;
}) => {

    const { mutate: setPluginSettingsPinnedTrays, isPending: isSettingPluginSettingsPinnedTrays } = useSetPluginSettingsPinnedTrays()

    const pinnedTrayPluginIds = settings?.pinnedTrayPluginIds || []

    const isPinned = (extensionId: string) => pinnedTrayPluginIds.includes(extensionId)

    const [unpinnedTrayIconClicked, setUnpinnedTrayIconClicked] = useAtom(__plugin_unpinnedTrayIconClickedAtom)

    const [trayIconListOpen, setTrayIconListOpen] = React.useState(false)

    const pinnedTrayIcons = trayIcons.filter(trayIcon => isPinned(trayIcon.extensionId) || trayIcon.extensionId === unpinnedTrayIconClicked?.extensionId)

    return (
        <>
            <div
                data-plugin-sidebar-tray
                className={cn(
                    "w-10 mx-auto p-1 my-2",
                    "flex flex-col gap-1 items-center border border-transparent justify-center rounded-full transition-all duration-300 select-none",
                    place === "top" && "flex-row w-auto my-0 justify-start px-2 py-2 border-none",
                    pinnedTrayIcons.length > 0 && "border-[--border]",
                )}
            >

                <Popover
                    open={trayIconListOpen}
                    onOpenChange={setTrayIconListOpen}
                    side={place === "top" ? "bottom" : "right"}
                    trigger={<div>
                        <Tooltip
                            side="right"
                            trigger={<IconButton
                                intent="gray-basic"
                                size="sm"
                                icon={<LuShapes className="size-5 text-[--muted]" />}
                                className="rounded-full hover:rotate-360 transition-all duration-300"
                            />}
                        >Tray plugins</Tooltip>
                    </div>}
                    className="p-2 w-[300px]"
                    data-plugin-sidebar-debug-popover
                    modal={false}
                >
                    <div className="space-y-1 max-h-[310px] overflow-y-auto" data-plugin-sidebar-debug-popover-content>
                        {/* <div className="text-sm">
                            <p className="font-bold">
                                Plugins
                            </p>
                        </div> */}
                        {trayIcons?.map(trayIcon => (
                            <div key={trayIcon.extensionId} className="flex items-center gap-2 justify-between bg-[--subtle] rounded-md px-2 py-1 max-w-full">
                                <div
                                    className="flex items-center gap-2 cursor-pointer max-w-full"
                                    onClick={() => {
                                        setUnpinnedTrayIconClicked(trayIcon)
                                        setTrayIconListOpen(false)
                                    }}
                                >
                                    <div className="w-8 h-8 rounded-full flex items-center justify-center overflow-hidden relative flex-none">
                                        {trayIcon.iconUrl ? <Image
                                            src={trayIcon.iconUrl}
                                            alt="logo"
                                            fill
                                            className="p-1 w-full h-full object-contain"
                                            data-plugin-tray-icon-image
                                        /> : <div
                                            className="w-8 h-8 rounded-full flex items-center justify-center flex-none"
                                            data-plugin-tray-icon-image-fallback
                                        >
                                            <LuCircleDashed className="text-2xl" />
                                        </div>}
                                    </div>
                                    <p className="text-sm font-medium line-clamp-1 tracking-wide">{trayIcon.extensionName}</p>
                                </div>
                                <div className="flex items-center gap-1">
                                    {/* <IconButton
                                        intent="gray-basic"
                                        size="sm"
                                        icon={<LuRefreshCw className="size-4" />}
                                        className="rounded-full"
                                        onClick={() => reloadExternalExtension({ id: trayIcon.extensionId })}
                                        loading={isReloadingExtension}
                                     /> */}
                                    <Tooltip
                                        trigger={<div>
                                            {isPinned(trayIcon.extensionId) ? <IconButton
                                                intent="primary-basic"
                                                size="sm"
                                                icon={<TbPinnedFilled className="size-5" />}
                                                className="rounded-full"
                                                onClick={() => {
                                                    setPluginSettingsPinnedTrays({
                                                        pinnedTrayPluginIds: pinnedTrayPluginIds.filter(id => id !== trayIcon.extensionId),
                                                    })
                                                }}
                                                disabled={isSettingPluginSettingsPinnedTrays}
                                            /> : <IconButton
                                                intent="gray-basic"
                                                size="sm"
                                                icon={<TbPinned className="size-5 text-[--muted]" />}
                                                className="rounded-full"
                                                onClick={() => {
                                                    setPluginSettingsPinnedTrays({
                                                        pinnedTrayPluginIds: [...pinnedTrayPluginIds, trayIcon.extensionId],
                                                    })
                                                }}
                                                disabled={isSettingPluginSettingsPinnedTrays}
                                            />}
                                        </div>}
                                    >
                                        {isPinned(trayIcon.extensionId) ? "Unpin" : "Pin"}
                                    </Tooltip>
                                </div>
                            </div>
                        ))}
                        {!trayIcons.length && <p className="text-sm text-[--muted] py-1 text-center w-full">
                            No tray plugins
                        </p>}

                        {/* {developmentModeExtensions?.map(extension => (
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
                        ))} */}
                    </div>
                </Popover>

                {!!developmentModeExtensions.length && <Popover
                    side={place === "top" ? "bottom" : "right"}
                    trigger={<div>
                        <IconButton
                            intent="warning-basic"
                            size="sm"
                            icon={<LuBug className="size-4" />}
                            className="rounded-full"
                        />
                    </div>}
                    className="p-2"
                    data-plugin-sidebar-debug-popover
                    modal={false}
                >
                    <div className="space-y-2" data-plugin-sidebar-debug-popover-content>
                        <div className="text-sm space-y-1">
                            <p className="font-bold">
                                Debug
                            </p>
                            <p className="text-xs text-[--muted]">
                                These extensions are loaded in development mode.
                            </p>
                        </div>
                        {developmentModeExtensions?.sort((a, b) => a.id.localeCompare(b.id, undefined, { numeric: true })).map(extension => (
                            <div key={extension.id} className="flex items-center gap-2 justify-between bg-[--subtle] rounded-md px-2 py-1">
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
                </Popover>}

                {pinnedTrayIcons.map((trayIcon, index) => (
                    <PluginTray
                        trayIcon={trayIcon}
                            isPinned={isPinned(trayIcon.extensionId)}
                            key={index}
                            place={place}
                            width={width}
                        />
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

    const { data: pluginSettings } = useGetPluginSettings()

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
                settings={pluginSettings}
                width={width}
            />}
            {isMobile && place === "top" && <div className="">
                <Popover
                    side="bottom"
                    trigger={<div>
                        <IconButton
                            intent="gray-basic"
                            size="sm"
                            icon={<LuBlocks className="size-5 text-[--muted]" />}
                        />
                    </div>}
                    className="rounded-full p-0 overflow-y-auto bg-[--background] mx-4"
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
                        settings={pluginSettings}
                        width={width}
                    />
                </Popover>
            </div>}
        </>
    )
}
