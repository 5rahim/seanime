import { ExtensionRepo_TrayPluginExtensionItem } from "@/api/generated/types"
import { PluginTray } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { cn } from "@/components/ui/core/styling"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"

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
                    <PluginTray extensionID={trayItem.id} key={index} />
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
