import { ExtensionRepo_TrayPluginExtensionItem } from "@/api/generated/types"
import { TrayPlugin } from "@/app/(main)/_features/plugin/tray/tray-plugin"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"
import { LuShapes } from "react-icons/lu"

export const __plugin_trayItemsAtom = atom<ExtensionRepo_TrayPluginExtensionItem[]>([])

export function PluginSidebarTray() {
    const { sendMessage } = useWebsocketSender()

    const [trayItems, setTrayItems] = useAtom(__plugin_trayItemsAtom)

    React.useEffect(() => {
        sendMessage({
            type: "tray:list",
            payload: {},
        })
    }, [])

    useWebsocketMessageListener({
        type: "tray:list",
        onMessage: (data: ExtensionRepo_TrayPluginExtensionItem[]) => {
            setTrayItems(data)
        },
    })

    if (!trayItems) return null

    return (
        <>
            <div
                className={cn(
                    "w-10 mx-auto p-1",
                    "flex flex-col gap-1 items-center justify-center rounded-full border hover:border-[--border] transition-all duration-300",
                )}
            >

                {trayItems.filter(n => n.isPinned).map((trayItem, index) => (
                    <TrayPlugin extensionID={trayItem.id} key={index} />
                ))}

                <IconButton
                    intent="gray-basic"
                    size="sm"
                    icon={<LuShapes />}
                    className="rounded-full"
                />
            </div>
        </>
    )
}
