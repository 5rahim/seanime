import { PluginProvider, registry, RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import {
    Plugin_Server_ModalUpdatedEventPayload,
    usePluginListenModalUpdatedEvent,
    usePluginSendModalClosedEvent,
} from "@/app/(main)/_features/plugin/generated/plugin-events"
import { PluginTrayContext, usePluginTray } from "@/app/(main)/_features/plugin/tray/plugin-tray"
import { useIsMainTabRef } from "@/app/websocket-provider"
import { Modal } from "@/components/ui/modal"
import React from "react"

export function PluginActionModal({ extensionId }: { extensionId: string }) {
    const [modal, setModal] = React.useState<Plugin_Server_ModalUpdatedEventPayload | null>(null)
    const isMainTabRef = useIsMainTabRef()
    const tray = usePluginTray()
    const { sendModalClosedEvent } = usePluginSendModalClosedEvent()

    usePluginListenModalUpdatedEvent(event => {
        if (!isMainTabRef.current) return
        setModal(prev => {
            if (event.action === "open") return event
            if (event.action === "close") return null
            return prev ? { ...prev, ...event } : null
        })
    }, extensionId)

    if (!modal) return null

    return <PluginProvider registry={registry}>
        <PluginTrayContext.Provider value={{ ...tray, trayIcon: { ...tray.trayIcon, extensionId } }}>
            <Modal
                open
                title={modal.title}
                description={modal.description}
                contentClass={modal.className}
                onOpenChange={open => {
                    if (!open) {
                        setModal(null)
                        sendModalClosedEvent({}, extensionId)
                    }
                }}
            >
                {modal.components && <RenderPluginComponents data={modal.components} />}
            </Modal>
        </PluginTrayContext.Provider>
    </PluginProvider>
}
