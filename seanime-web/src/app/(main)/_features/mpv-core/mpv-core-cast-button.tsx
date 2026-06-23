import { getServerBaseUrl } from "@/api/client/server-url"
import type { Player_PlaybackInfo } from "@/api/generated/types"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { __CAST_ENABLED__, __isElectronDesktop__ } from "@/types/constants"
import React from "react"
import { BiCast } from "react-icons/bi"
import { mc_resolveSource } from "./mpv-core"

export interface MpvCoreCastButtonProps {
    info: Player_PlaybackInfo | null
    paused: boolean
    onCastingStart: () => void
    onCastingEnd: () => void
}

export function MpvCoreCastButton(props: MpvCoreCastButtonProps) {
    const [devices, setDevices] = React.useState<CastDevice[]>([])
    const [discovering, setDiscovering] = React.useState(false)
    const [casting, setCasting] = React.useState(false)
    const [modalOpen, setModalOpen] = React.useState(false)

    React.useEffect(() => {
        const removeDevice = window.electron?.on?.("cast:deviceFound", (device: CastDevice) => {
            setDevices(current => current.some(item => item.id === device.id) ? current : [...current, device])
        })
        const removeSession = window.electron?.on?.("cast:sessionUpdate", (session: CastSessionState) => {
            setCasting(session.connected)
            if (!session.connected) props.onCastingEnd()
        })
        return () => {
            removeDevice?.()
            removeSession?.()
        }
    }, [props.onCastingEnd])

    React.useEffect(() => {
        if (modalOpen && !casting) discover().then()
    }, [modalOpen])

    if (!__CAST_ENABLED__ || !__isElectronDesktop__ || !window.electron?.cast) return null

    async function discover() {
        setDevices([])
        setDiscovering(true)
        await window.electron?.cast?.discover()
        window.setTimeout(async () => {
            await window.electron?.cast?.stopDiscovery()
            setDevices(await window.electron?.cast?.getDevices() ?? [])
            setDiscovering(false)
        }, 5000)
    }

    async function connect(device: CastDevice) {
        if (!props.info || !window.electron?.cast) return
        await window.electron.cast.connect(device.id)
        props.onCastingStart()
        const serverBaseUrl = getServerBaseUrl()
        const streamUrl = mc_resolveSource(props.info.streamUrl)
        await window.electron.cast.loadMedia({
            streamUrl,
            contentType: props.info.mimeType || "video/mp4",
            title: props.info.media?.title?.userPreferred || "Seanime",
            subtitle: props.info.episode?.displayTitle || "",
            imageUrl: props.info.media?.coverImage?.large || "",
            serverPort: Number(serverBaseUrl.split(":").pop()) || 43211,
        })
        setCasting(true)
        setModalOpen(false)
    }

    async function disconnect() {
        await window.electron?.cast?.disconnect()
        setCasting(false)
        props.onCastingEnd()
    }

    return (
        <>
            <IconButton
                intent={casting ? "primary" : "gray-basic"}
                size="sm"
                icon={<BiCast className={cn("text-lg", casting && "text-brand-300")} />}
                onClick={() => setModalOpen(true)}
                title={casting ? "Casting" : "Cast to device"}
            />
            <Modal open={modalOpen} onOpenChange={setModalOpen} title="Cast to Device" contentClass="max-w-md">
                <div className="space-y-4">
                    {casting && (
                        <div className="flex items-center justify-between p-3 bg-gray-900 rounded-md border border-brand-700">
                            <div>
                                <p className="text-sm font-medium text-brand-300">Connected</p>
                                <p className="text-base font-semibold">Chromecast</p>
                            </div>
                            <Button intent="alert-subtle" size="sm" onClick={() => void disconnect()}>
                                Disconnect
                            </Button>
                        </div>
                    )}

                    {discovering && (
                        <div className="flex items-center gap-2 text-sm text-[--muted]">
                            <LoadingSpinner />
                            <span>Searching for devices...</span>
                        </div>
                    )}

                    {!discovering && !devices.length && (
                        <div className="text-center py-6">
                            <p className="text-sm text-[--muted]">No devices found</p>
                            <Button intent="gray-subtle" size="sm" className="mt-2" onClick={() => void discover()}>
                                Scan again
                            </Button>
                        </div>
                    )}

                    {!!devices.length && (
                        <div className="space-y-2">
                            {devices.map(device => (
                                <button
                                    key={device.id}
                                    className="w-full flex items-center gap-3 p-3 rounded-md transition-colors hover:bg-gray-800 text-left"
                                    onClick={() => void connect(device)}
                                >
                                    <BiCast className="text-xl text-gray-400" />
                                    <p className="text-sm font-medium">{device.name}</p>
                                </button>
                            ))}
                        </div>
                    )}

                    {!discovering && !!devices.length && (
                        <Button intent="gray-subtle" size="sm" className="w-full" onClick={() => void discover()}>
                            Scan again
                        </Button>
                    )}
                </div>
            </Modal>
        </>
    )
}
