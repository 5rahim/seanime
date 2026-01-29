import { PluginProvider, registry, RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import { useIsMainTabRef } from "@/app/websocket-provider"
import { SeaImage } from "@/components/shared/sea-image"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { PopoverAnatomy } from "@/components/ui/popover"
import { Tooltip } from "@/components/ui/tooltip"
import { Vaul, VaulContent } from "@/components/vaul"
import { getPixelsFromLength } from "@/lib/helpers/css"
import { useIsMobile } from "@/lib/theme/hooks"
import * as PopoverPrimitive from "@radix-ui/react-popover"
import { useAtom, useAtomValue } from "jotai"
import React from "react"
import { BiX } from "react-icons/bi"
import { LuCircleDashed } from "react-icons/lu"
import {
    Plugin_Server_TrayIconEventPayload,
    usePluginListenTrayBadgeUpdatedEvent,
    usePluginListenTrayCloseEvent,
    usePluginListenTrayOpenEvent,
    usePluginListenTrayUpdatedEvent,
    usePluginSendRenderTrayEvent,
    usePluginSendTrayClickedEvent,
    usePluginSendTrayClosedEvent,
    usePluginSendTrayOpenedEvent,
} from "../generated/plugin-events"
import { __plugin_hasNavigatedAtom, __plugin_openedTrayPlugin, __plugin_unpinnedTrayIconClickedAtom } from "./plugin-sidebar-tray"

/**
 * TrayIcon
 */
export type TrayIcon = Plugin_Server_TrayIconEventPayload

type TrayPluginProps = {
    trayIcon: TrayIcon
    place: "sidebar" | "top"
    width: number | null
    isPinned: boolean
}

export const PluginTrayContext = React.createContext<TrayPluginProps>({
    place: "sidebar",
    width: null,
    isPinned: false,
    trayIcon: {
        extensionId: "",
        extensionName: "",
        iconUrl: "",
        withContent: false,
        isDrawer: false,
        tooltipText: "",
        badgeNumber: 0,
        badgeIntent: "info",
        width: "30rem",
        minHeight: "auto",
    },
})

function PluginTrayProvider(props: { children: React.ReactNode, props: TrayPluginProps }) {
    return <PluginTrayContext.Provider value={props.props}>
        {props.children}
    </PluginTrayContext.Provider>
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function usePluginTray() {
    const context = React.useContext(PluginTrayContext)
    if (!context) {
        throw new Error("usePluginTray must be used within a PluginTrayProvider")
    }
    return context
}

export function PluginTray(props: TrayPluginProps) {

    const [open, setOpen] = React.useState(false)
    const [badgeNumber, setBadgeNumber] = React.useState(0)
    const [badgeIntent, setBadgeIntent] = React.useState("info")

    const { sendTrayOpenedEvent } = usePluginSendTrayOpenedEvent()
    const { sendTrayClosedEvent } = usePluginSendTrayClosedEvent()
    const { sendTrayClickedEvent } = usePluginSendTrayClickedEvent()

    const isMainTabRef = useIsMainTabRef()

    const hasNavigated = useAtomValue(__plugin_hasNavigatedAtom)

    const [unpinnedTrayIconClicked, setUnpinnedTrayIconClicked] = useAtom(__plugin_unpinnedTrayIconClickedAtom)
    const [openedTrayPlugin, setOpenedTrayPlugin] = useAtom(__plugin_openedTrayPlugin)

    React.useEffect(() => {
        setBadgeNumber(props.trayIcon.badgeNumber)
        setBadgeIntent(props.trayIcon.badgeIntent)
    }, [])

    const firstRender = React.useRef(true)
    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (open) {
            sendTrayOpenedEvent({}, props.trayIcon.extensionId)
        } else {
            sendTrayClosedEvent({}, props.trayIcon.extensionId)
        }
    }, [open])

    // Handle unpinned tray icon click to open the tray
    const unpinnedTrayIconClickedOpenedRef = React.useRef(false)
    React.useEffect(() => {
        if (unpinnedTrayIconClicked?.extensionId === props.trayIcon.extensionId) {
            sendTrayClickedEvent({}, props.trayIcon.extensionId)
            setOpenedTrayPlugin(props.trayIcon.extensionId)
            if (!unpinnedTrayIconClickedOpenedRef.current) {
                const timeout = setTimeout(() => {
                    setOpen(true)
                    unpinnedTrayIconClickedOpenedRef.current = true
                }, 100)
                return () => clearTimeout(timeout)
            }
        }
    }, [unpinnedTrayIconClicked])

    // Reset unpinned tray icon click state when closing the tray
    React.useEffect(() => {
        if (unpinnedTrayIconClicked?.extensionId === props.trayIcon.extensionId && !open && unpinnedTrayIconClickedOpenedRef.current) {
            setUnpinnedTrayIconClicked(null)
            unpinnedTrayIconClickedOpenedRef.current = false
        }
    }, [open])

    usePluginListenTrayBadgeUpdatedEvent((data) => {
        setBadgeNumber(data.badgeNumber)
        setBadgeIntent(data.badgeIntent)
    }, props.trayIcon.extensionId)

    usePluginListenTrayOpenEvent((data) => {
        if (!isMainTabRef.current) return
        setOpen(true)
    }, props.trayIcon.extensionId)

    usePluginListenTrayCloseEvent((data) => {
        if (!isMainTabRef.current) return
        setOpen(false)
    }, props.trayIcon.extensionId)

    function handleClick() {
        sendTrayClickedEvent({}, props.trayIcon.extensionId)
    }

    const tooltipText = props.trayIcon.extensionName

    const TrayIcon = () => {
        return (
            <div
                data-plugin-tray-icon
                className="w-8 h-8 rounded-full flex items-center justify-center hover:bg-gray-800 cursor-pointer transition-all relative select-none"
                onClick={handleClick}
            >
                <div className="w-8 h-8 rounded-full flex items-center justify-center overflow-hidden relative" data-plugin-tray-icon-inner-container>
                    {props.trayIcon.iconUrl ? <SeaImage
                        src={props.trayIcon.iconUrl}
                        alt="logo"
                        fill
                        className="p-1 w-full h-full object-contain"
                        data-plugin-tray-icon-image
                        isExternal
                    /> : <div className="w-8 h-8 rounded-full flex items-center justify-center" data-plugin-tray-icon-image-fallback>
                        <LuCircleDashed className="text-2xl" />
                    </div>}
                </div>
                {!!badgeNumber && <Badge
                    intent={`${badgeIntent}-solid` as any}
                    size="sm"
                    className="absolute -top-2 -right-2 z-10 select-none pointer-events-none"
                    data-plugin-tray-icon-badge
                >
                    {badgeNumber}
                </Badge>}
            </div>
        )
    }

    const designatedWidthPx = getPixelsFromLength(props.trayIcon.width || "30rem")
    const popoverWidth = (props.width && props.width < 1024)
        ? (designatedWidthPx >= props.width ? `calc(100vw - 30px)` : designatedWidthPx)
        : props.trayIcon.width || "30rem"

    if (!props.trayIcon.withContent) {
        return <div className="cursor-pointer">
            {!!tooltipText ? <Tooltip
                side="right"
                trigger={<div data-plugin-tray-icon-tooltip-trigger>
                    <TrayIcon />
                </div>}
                data-plugin-tray-icon-tooltip
            >
                {tooltipText}
            </Tooltip> : <TrayIcon />}
        </div>
    }

    React.useEffect(() => {
        if (open) {
            setOpenedTrayPlugin(props.trayIcon.extensionId)
            setTimeout(() => {
                document.body.style.pointerEvents = "auto"
            }, 500)
        }
    }, [props.trayIcon.isDrawer, open])

    React.useLayoutEffect(() => {
        if (openedTrayPlugin !== props.trayIcon.extensionId) {
            setOpen(false)
        }
    }, [openedTrayPlugin])

    const { isMobile } = useIsMobile()


    // console.log("popoverWidth", popoverWidth)
    // console.log("designatedWidthPx", designatedWidthPx)
    // console.log("props.width", props.width)

    if (props.trayIcon.isDrawer) {
        return (
            <>
                <div
                    data-plugin-tray-icon-trigger={props.trayIcon.extensionId}
                    onClick={() => {
                        setOpenedTrayPlugin(props.trayIcon.extensionId)
                        setOpen(true)
                    }}
                    className="cursor-pointer"
                    data-plugin-tray-icon-trigger-drawer
                >
                    {!!tooltipText ? <Tooltip
                        side={props.place === "sidebar" ? "right" : "bottom"}
                        trigger={<div data-plugin-tray-icon-tooltip-trigger>
                            <TrayIcon />
                        </div>}
                        data-plugin-tray-icon-tooltip
                    >
                        {tooltipText}
                    </Tooltip> : <TrayIcon />}
                </div>
                <Vaul
                    open={open && openedTrayPlugin === props.trayIcon.extensionId}
                    onOpenChange={setOpen}
                    modal={false}
                >
                    <VaulContent
                        className={cn("bg-gray-950 p-0 rounded-t-xl mx-auto rounded-b-none border-b-0 !min-h-[120px]")}
                        onOpenAutoFocus={(e) => e.preventDefault()}
                        style={{
                            width: isMobile ? "100vw" : popoverWidth,
                            minHeight: props.trayIcon.minHeight || "auto",
                        }}
                        data-plugin-tray-popover-content={props.trayIcon.extensionId}
                    >
                        <div className="absolute w-full top-[-2.5rem]">
                            <div className="flex items-center justify-between">
                                <p
                                    className="text-sm border font-medium text-gray-300 px-1.5 py-0.5 rounded-lg bg-black/60"
                                    data-plugin-tray-vaul-title
                                >
                                    {props.trayIcon.extensionName}
                                </p>
                                <IconButton
                                    icon={<BiX />}
                                    data-plugin-tray-vaul-close-button
                                    intent="gray-glass"
                                    size="sm"
                                    className="rounded-full"
                                    onClick={() => setOpen(false)}
                                />
                            </div>
                        </div>
                        <PluginTrayProvider props={props}>
                            <PluginTrayContent
                                open={open}
                                setOpen={setOpen}
                                {...props}
                            />
                        </PluginTrayProvider>
                    </VaulContent>
                </Vaul>
            </>
        )
    }

    return (
        <>
            <PopoverPrimitive.Root
                open={open}
                onOpenChange={setOpen}
                modal={false}
            >
                <PopoverPrimitive.Trigger
                    asChild
                >
                    <div data-plugin-tray-icon-trigger={props.trayIcon.extensionId}>
                        {!!tooltipText ? <Tooltip
                            side={props.place === "sidebar" ? "right" : "bottom"}
                            trigger={<div data-plugin-tray-icon-tooltip-trigger>
                                <TrayIcon />
                            </div>}
                            data-plugin-tray-icon-tooltip
                        >
                            {tooltipText}
                        </Tooltip> : <TrayIcon />}
                    </div>
                </PopoverPrimitive.Trigger>
                <PopoverPrimitive.Portal>
                    <PopoverPrimitive.Content
                        sideOffset={10}
                        side={props.place === "sidebar" ? "right" : "bottom"}
                        className={cn(PopoverAnatomy.root(), "bg-gray-950 p-0 shadow-xl rounded-xl mb-4")}
                        onOpenAutoFocus={(e) => e.preventDefault()}
                        style={{
                            width: popoverWidth,
                            minHeight: props.trayIcon.minHeight || "auto",
                            marginLeft: props.width && props.width < 1024 ? `10px` : undefined,
                        }}
                        data-plugin-tray-popover-content={props.trayIcon.extensionId}
                    >
                        <PluginTrayProvider props={props}>
                            <PluginTrayContent
                                open={open}
                                setOpen={setOpen}
                                {...props}
                            />
                        </PluginTrayProvider>
                    </PopoverPrimitive.Content>
                </PopoverPrimitive.Portal>
            </PopoverPrimitive.Root>
        </>
    )
}

type PluginTrayContentProps = {
    open: boolean
    setOpen: (open: boolean) => void
} & TrayPluginProps

function PluginTrayContent(props: PluginTrayContentProps) {
    const {
        trayIcon,
        open,
        setOpen,
    } = props

    const { sendRenderTrayEvent } = usePluginSendRenderTrayEvent()

    React.useEffect(() => {
        if (open) {
            sendRenderTrayEvent({}, trayIcon.extensionId)
        }
    }, [open])

    const [data, setData] = React.useState<any>(null)

    usePluginListenTrayUpdatedEvent((data) => {
        // console.log("tray:updated", extensionID, data)
        setData(data)
    }, trayIcon.extensionId)

    return (
        <div>
            <div
                className={cn(
                    "max-h-[35rem] overflow-y-auto p-3",
                )}
            >

                <PluginProvider registry={registry}>
                    {!!data && data.components ? <RenderPluginComponents data={data.components} /> : <LoadingSpinner />}
                </PluginProvider>

            </div>
        </div>
    )
}
