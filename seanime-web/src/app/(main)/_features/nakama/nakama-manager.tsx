import { useGetNakamaStatus, useNakamaReconnectToHost, useNakamaRemoveStaleConnections } from "@/api/hooks/nakama.hooks"
import { Modal } from "@/components/ui/modal"
import { atom, useAtom, useAtomValue } from "jotai"
import React from "react"
import { useEffectOnce } from "react-use"
import { useWebsocketMessageListener } from "../../_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"
import { Nakama_NakamaStatus } from "@/api/generated/types"
import { SettingsCard } from "../../settings/_components/settings-card"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { MdOutlineConnectWithoutContact, MdRefresh, MdCleaningServices } from "react-icons/md"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"
import { BetaBadge } from "@/components/shared/beta-badge"

export const nakamaModalOpenAtom = atom(false)
export const nakamaStatusAtom = atom<Nakama_NakamaStatus | null>(null)

export function useNakamaStatus() {
    return useAtomValue(nakamaStatusAtom)
}

export function NakamaManager() {
    const [isModalOpen, setIsModalOpen] = useAtom(nakamaModalOpenAtom)
    const [nakamaStatus, setNakamaStatus] = useAtom(nakamaStatusAtom)

    const { data: status, refetch: refetchStatus, isLoading } = useGetNakamaStatus()
    const { mutate: reconnectToHost, isPending: isReconnecting } = useNakamaReconnectToHost()
    const { mutate: removeStaleConnections, isPending: isCleaningUp } = useNakamaRemoveStaleConnections()

    React.useEffect(() => {
        setNakamaStatus(status ?? null)
    }, [status])

    React.useEffect(() => {
        refetchStatus()
    }, [])

    React.useEffect(() => {
        if (isModalOpen) {
            refetchStatus()
        }
    }, [isModalOpen])

    const handleReconnect = React.useCallback(() => {
        reconnectToHost({}, {
            onSuccess: () => {
                toast.success("Reconnection initiated")
                refetchStatus()
            },
            onError: (error) => {
                toast.error(`Failed to reconnect: ${error.message}`)
            },
        })
    }, [reconnectToHost, refetchStatus])

    const handleCleanupStaleConnections = React.useCallback(() => {
        removeStaleConnections({}, {
            onSuccess: () => {
                toast.success("Stale connections cleaned up")
                refetchStatus()
            },
            onError: (error) => {
                toast.error(`Failed to cleanup: ${error.message}`)
            },
        })
    }, [removeStaleConnections, refetchStatus])

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_HOST_STARTED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_HOST_STOPPED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_PEER_CONNECTED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_PEER_DISCONNECTED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_HOST_CONNECTED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_HOST_DISCONNECTED,
        onMessage: () => {
            refetchStatus()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_ERROR,
        onMessage: () => {
            refetchStatus()
        },
    })

    return <>

        {/* Modal */}
        <Modal
            open={isModalOpen}
            onOpenChange={setIsModalOpen}
            title={<div className="flex items-center gap-2 w-full justify-center">
                <MdOutlineConnectWithoutContact className="size-6" />
                Nakama
                <BetaBadge />
            </div>}
            contentClass="max-w-3xl bg-gray-950/90 backdrop-blur-sm"
            overlayClass="bg-gray-950/70 backdrop-blur-sm"
            // allowOutsideInteraction
        >

            {isLoading && <LoadingSpinner />}

            {status?.isHost && !isLoading && (
                <>
                    <div className="flex items-center justify-between">
                        <h4>Currently hosting</h4>
                        <Button
                            onClick={handleCleanupStaleConnections}
                            disabled={isCleaningUp}
                            size="sm"
                            intent="gray-subtle"
                            leftIcon={<MdCleaningServices />}
                        >
                            {isCleaningUp ? "Cleaning up..." : "Remove stale connections"}
                        </Button>
                    </div>
                    <SettingsCard title="Connected peers">
                        {!status?.connectedPeers?.length && <p className="text-center text-sm text-[--muted]">No connected peers</p>}
                        {status?.connectedPeers?.map((peer, index) => (
                            <div key={index} className="flex items-center justify-between py-1">
                                <span className="font-medium">{peer}</span>
                            </div>
                        ))}
                    </SettingsCard>
                </>
            )}

            {status?.isConnectedToHost && !isLoading && (
                <>
                    <div className="flex items-center justify-between">
                        <h4>Connected to host</h4>
                        <Button
                            onClick={handleReconnect}
                            disabled={isReconnecting}
                            size="sm"
                            intent="primary-subtle"
                            leftIcon={<MdRefresh />}
                        >
                            {isReconnecting ? "Reconnecting..." : "Reconnect"}
                        </Button>
                    </div>
                    <SettingsCard title="Host connection">
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-[--muted]">Host:</span>
                                <span className="font-medium">
                                    {status?.hostConnectionStatus?.username || "Unknown"}
                                </span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-[--muted]">Status:</span>
                                <span
                                    className={`font-medium ${
                                        status?.hostConnectionStatus?.authenticated
                                            ? "text-green-500"
                                            : "text-red-500"
                                    }`}
                                >
                                    {status?.hostConnectionStatus?.authenticated ? "Connected" : "Disconnected"}
                                </span>
                            </div>
                            {status?.hostConnectionStatus?.url && (
                                <div className="flex items-center justify-between">
                                    <span className="text-sm text-[--muted]">URL:</span>
                                    <span className="font-mono text-xs">{status.hostConnectionStatus.url}</span>
                                </div>
                            )}
                        </div>
                    </SettingsCard>
                </>
            )}

            {!status?.isHost && !status?.isConnectedToHost && !isLoading && (
                <div className="text-center py-8">
                    <p className="text-[--muted]">Nakama is not active</p>
                    <p className="text-sm text-[--muted] mt-2">
                        Configure Nakama in settings to connect to a host or start hosting
                    </p>
                </div>
            )}
        </Modal>

    </>
}
