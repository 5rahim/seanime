import { Nakama_NakamaStatus, Nakama_WatchPartySession, Nakama_WatchPartySessionSettings } from "@/api/generated/types"
import {
    useNakamaCreateWatchParty,
    useNakamaJoinWatchParty,
    useNakamaLeaveWatchParty,
    useNakamaReconnectToHost,
    useNakamaRemoveStaleConnections,
} from "@/api/hooks/nakama.hooks"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useNakamaOnlineStreamWatchParty } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import { AlphaBadge } from "@/components/shared/beta-badge"
import { GlowingEffect } from "@/components/shared/glowing-effect"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Tooltip } from "@/components/ui/tooltip"
import { WSEvents } from "@/lib/server/ws-events"
import { atom, useAtom, useAtomValue } from "jotai"
import React from "react"
import { BiCog } from "react-icons/bi"
import { FaBroadcastTower } from "react-icons/fa"
import { HiOutlinePlay } from "react-icons/hi2"
import { LuPopcorn } from "react-icons/lu"
import { MdAdd, MdCleaningServices, MdOutlineConnectWithoutContact, MdPlayArrow, MdRefresh } from "react-icons/md"
import { toast } from "sonner"

export const nakamaModalOpenAtom = atom(false)
export const nakamaStatusAtom = atom<Nakama_NakamaStatus | null | undefined>(undefined)

export const watchPartySessionAtom = atom<Nakama_WatchPartySession | null | undefined>(undefined)

export function useNakamaStatus() {
    return useAtomValue(nakamaStatusAtom)
}

export function useWatchPartySession() {
    return useAtomValue(watchPartySessionAtom)
}

export function NakamaManager() {
    const { sendMessage } = useWebsocketSender()
    const [isModalOpen, setIsModalOpen] = useAtom(nakamaModalOpenAtom)
    const [nakamaStatus, setNakamaStatus] = useAtom(nakamaStatusAtom)
    const [watchPartySession, setWatchPartySession] = useAtom(watchPartySessionAtom)

    // const { data: status, refetch: refetchStatus, isLoading } = useGetNakamaStatus()
    const { mutate: reconnectToHost, isPending: isReconnecting } = useNakamaReconnectToHost()
    const { mutate: removeStaleConnections, isPending: isCleaningUp } = useNakamaRemoveStaleConnections()
    const { mutate: createWatchParty, isPending: isCreatingWatchParty } = useNakamaCreateWatchParty()
    const { mutate: joinWatchParty, isPending: isJoiningWatchParty } = useNakamaJoinWatchParty()
    const { mutate: leaveWatchParty, isPending: isLeavingWatchParty } = useNakamaLeaveWatchParty()

    // Watch party settings for creating a new session
    const [watchPartySettings, setWatchPartySettings] = React.useState<Nakama_WatchPartySessionSettings>({
        syncThreshold: 3.0,
        maxBufferWaitTime: 10,
    })

    function refetchStatus() {
        sendMessage({
            type: WSEvents.NAKAMA_STATUS_REQUESTED,
            payload: null,
        })
    }

    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_STATUS,
        onMessage: (data: Nakama_NakamaStatus | null) => {
            setNakamaStatus(data ?? null)
        },
    })

    // NAKAMA_WATCH_PARTY_STATE tells the client to refetch the status
    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_WATCH_PARTY_STATE,
        onMessage: (data: any) => {
            refetchStatus()
        },
    })

    React.useEffect(() => {
        if (nakamaStatus?.currentWatchPartySession) {
            setWatchPartySession(nakamaStatus.currentWatchPartySession)
        } else {
            setWatchPartySession(null)
        }
    }, [nakamaStatus])

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

    const handleCreateWatchParty = React.useCallback(() => {
        createWatchParty({ settings: watchPartySettings }, {
            onSuccess: () => {
                toast.success("Watch party created")
                refetchStatus()
            },
            onError: (error) => {
                toast.error(`Failed to create watch party: ${error.message}`)
            },
        })
    }, [createWatchParty, watchPartySettings, refetchStatus])

    const handleJoinWatchParty = React.useCallback(() => {
        joinWatchParty(undefined, {
            onSuccess: () => {
                toast.info("Joining watch party")
                refetchStatus()
            },
        })
    }, [joinWatchParty, refetchStatus])

    const handleLeaveWatchParty = React.useCallback(() => {
        leaveWatchParty(undefined, {
            onSuccess: () => {
                toast.info("Leaving watch party")
                setWatchPartySession(null)
                refetchStatus()
            },
        })
    }, [leaveWatchParty, setWatchPartySession, refetchStatus])

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

    /////// Online stream

    const { startOnlineStream } = useNakamaOnlineStreamWatchParty()
    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_ONLINE_STREAM_EVENT,
        onMessage: (_data: { type: string, payload: { type: string, payload: any } }) => {
            console.log(_data)
            switch (_data.type) {
                case "online-stream-playback-status":
                    const data = _data.payload
                    switch (data.type) {
                        case "start":
                            startOnlineStream(data.payload)
                            break
                    }
            }
        },
    })

    return <>
        <Modal
            open={isModalOpen}
            onOpenChange={setIsModalOpen}
            title={<div className="flex items-center gap-2 w-full justify-center">
                <MdOutlineConnectWithoutContact className="size-8" />
                Nakama
                <AlphaBadge className="border-transparent" />
            </div>}
            contentClass="max-w-3xl bg-gray-950 bg-opacity-60 backdrop-blur-sm firefox:bg-opacity-100 firefox:backdrop-blur-none sm:rounded-3xl"
            overlayClass="bg-gray-950/70 backdrop-blur-sm"
            // allowOutsideInteraction
        >

            <GlowingEffect
                variant="classic"
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                className="opacity-50"
            />

            <div className="absolute top-4 right-14">
                <SeaLink href="/settings?tab=nakama" onClick={() => setIsModalOpen(false)}>
                    <IconButton intent="gray-basic" size="sm" icon={<BiCog />} />
                </SeaLink>
            </div>

            {nakamaStatus === undefined && <LoadingSpinner />}

            {!nakamaStatus?.isHost && (
                <div className="flex items-center justify-between">
                    <div></div>
                    <Button
                        onClick={handleReconnect}
                        disabled={isReconnecting}
                        size="sm"
                        intent="gray-basic"
                        leftIcon={<MdRefresh />}
                    >
                        {isReconnecting ? "Reconnecting..." : "Reconnect"}
                    </Button>
                </div>
            )}

            {nakamaStatus !== undefined && (nakamaStatus?.isHost || nakamaStatus?.isConnectedToHost) && (
                <>

                    {nakamaStatus?.isHost && (
                        <>
                            <div className="flex items-center justify-between">
                                <Badge intent="success-solid" className="px-0 text-indigo-300 bg-transparent">Currently hosting</Badge>
                                <Button
                                    onClick={handleCleanupStaleConnections}
                                    disabled={isCleaningUp}
                                    size="sm"
                                    intent="gray-basic"
                                    leftIcon={<MdCleaningServices />}
                                >
                                    {isCleaningUp ? "Cleaning up..." : "Remove stale connections"}
                                </Button>
                            </div>
                            <h4>Connected peers ({nakamaStatus?.connectedPeers?.length ?? 0})</h4>
                            <div className="p-4 border rounded-lg bg-gray-950">
                                {!nakamaStatus?.connectedPeers?.length &&
                                    <p className="text-center text-sm text-[--muted]">No connected peers</p>}
                                {nakamaStatus?.connectedPeers?.map((peer, index) => (
                                    <div key={index} className="flex items-center justify-between py-1">
                                        <span className="font-medium">{peer}</span>
                                    </div>
                                ))}
                            </div>
                        </>
                    )}

                    {nakamaStatus?.isConnectedToHost && (
                        <>

                            <h4>Host connection</h4>
                            <div className="p-4 border rounded-lg bg-gray-950">
                                <div className="space-y-2">
                                    <div className="flex items-center justify-between">
                                        <span className="text-sm text-[--muted]">Host</span>
                                        <span className="font-medium text-sm tracking-wide">
                                            {nakamaStatus?.hostConnectionStatus?.username || "Unknown"}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </>
                    )}

                    {/* Watch Party Content */}
                    {(() => {
                        const isHost = nakamaStatus?.isHost || false
                        const isConnectedToHost = nakamaStatus?.isConnectedToHost || false
                        const currentPeerID = nakamaStatus?.hostConnectionStatus?.peerId

                        // Check if user is in the participant list by comparing peer ID
                        const isUserInSession = watchPartySession && (
                            isHost ||
                            (currentPeerID && watchPartySession.participants && currentPeerID in watchPartySession.participants)
                        )

                        // Show session view if there's a session AND user is in it
                        if (watchPartySession && isUserInSession) {
                            return (
                                <WatchPartySessionView
                                    session={watchPartySession}
                                    isHost={isHost}
                                    onLeave={handleLeaveWatchParty}
                                    isLeaving={isLeavingWatchParty}
                                />
                            )
                        }

                        // Otherwise show creation/join options
                        return (
                            <WatchPartyCreation
                                isHost={isHost}
                                isConnectedToHost={isConnectedToHost}
                                hasActiveSession={!!watchPartySession}
                                settings={watchPartySettings}
                                onSettingsChange={setWatchPartySettings}
                                onCreateWatchParty={handleCreateWatchParty}
                                onJoinWatchParty={handleJoinWatchParty}
                                isCreating={isCreatingWatchParty}
                                isJoining={isJoiningWatchParty}
                            />
                        )
                    })()}
                </>
            )}

            {!nakamaStatus?.isHost && !nakamaStatus?.isConnectedToHost && nakamaStatus !== undefined && (
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

interface WatchPartyCreationProps {
    isHost: boolean
    isConnectedToHost: boolean
    hasActiveSession: boolean
    settings: Nakama_WatchPartySessionSettings
    onSettingsChange: (settings: Nakama_WatchPartySessionSettings) => void
    onCreateWatchParty: () => void
    onJoinWatchParty: () => void
    isCreating: boolean
    isJoining: boolean
}

function WatchPartyCreation({
    isHost,
    isConnectedToHost,
    hasActiveSession,
    settings,
    onSettingsChange,
    onCreateWatchParty,
    onJoinWatchParty,
    isCreating,
    isJoining,
}: WatchPartyCreationProps) {
    return (
        <div className="space-y-4">
            <h4 className="flex items-center gap-2"><LuPopcorn className="size-6" /> Watch Party</h4>
            {isHost && (
                <div className="p-4 border rounded-lg bg-gray-950">
                    <div className="space-y-4">
                        {/* <div className="space-y-3">
                         <div className="flex items-center justify-between">
                         <label className="text-sm font-medium">Allow participant control</label>
                         <Switch
                         value={settings.allowParticipantControl}
                         onValueChange={(checked: boolean) =>
                         onSettingsChange({ ...settings, allowParticipantControl: checked })
                         }
                         />
                         </div>

                         <div className="space-y-2">
                         <label className="text-sm font-medium">Sync threshold (seconds)</label>
                         <NumberInput
                         value={settings.syncThreshold}
                         onValueChange={(value) =>
                         onSettingsChange({ ...settings, syncThreshold: value || 3.0 })
                         }
                         min={1}
                         max={10}
                         step={0.5}
                         />
                         <p className="text-xs text-[--muted]">How far out of sync before forcing synchronization</p>
                         </div>

                         <div className="space-y-2">
                         <label className="text-sm font-medium">Max buffer wait time (seconds)</label>
                         <NumberInput
                         value={settings.maxBufferWaitTime}
                         onValueChange={(value) =>
                         onSettingsChange({ ...settings, maxBufferWaitTime: value || 10 })
                         }
                         min={5}
                         max={60}
                         />
                         <p className="text-xs text-[--muted]">Maximum time to wait for peers to buffer</p>
                         </div>
                         </div> */}

                        <Button
                            onClick={onCreateWatchParty}
                            disabled={isCreating}
                            className="w-full"
                            intent="primary"
                            leftIcon={<MdAdd />}
                        >
                            {isCreating ? "Creating..." : "Create Watch Party"}
                        </Button>
                    </div>
                </div>
            )}

            {isConnectedToHost && !isHost && hasActiveSession && (
                <div className="p-4 border rounded-lg bg-gray-950">
                    <div className="space-y-4">
                        <p className="text-sm text-[--muted]">
                            There's an active watch party! Join to watch content together in sync.
                        </p>
                        <Button
                            onClick={onJoinWatchParty}
                            disabled={isJoining}
                            className="w-full"
                            intent="primary"
                            leftIcon={<MdPlayArrow />}
                        >
                            {isJoining ? "Joining..." : "Join Watch Party"}
                        </Button>
                    </div>
                </div>
            )}

            {!isHost && !isConnectedToHost && (
                <div className="text-center py-8">
                    <p className="text-[--muted]">Connect to a host to join a watch party</p>
                </div>
            )}

            {!isHost && isConnectedToHost && !hasActiveSession && (
                <div className="text-center py-8">
                    <p className="text-[--muted]">No active watch party</p>
                </div>
            )}
        </div>
    )
}

interface WatchPartySessionViewProps {
    session: Nakama_WatchPartySession
    isHost: boolean
    onLeave: () => void
    isLeaving: boolean
}

function WatchPartySessionView({ session, isHost, onLeave, isLeaving }: WatchPartySessionViewProps) {
    const { sendMessage } = useWebsocketSender()
    const nakamaStatus = useNakamaStatus()
    const participants = Object.values(session.participants || {})
    const participantCount = participants.length
    const serverStatus = useServerStatus()

    const [enablingRelayMode, setEnablingRelayMode] = React.useState(false)

    // Identify current user - either "host" if hosting, or the peer ID if connected as peer
    const currentUserId = isHost ? "host" : nakamaStatus?.hostConnectionStatus?.peerId

    function handleEnableRelayMode(peerId: string) {
        sendMessage({
            type: WSEvents.NAKAMA_WATCH_PARTY_ENABLE_RELAY_MODE,
            payload: { peerId },
        })
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h4 className="flex items-center gap-2"><LuPopcorn className="size-6" /> Watch Party</h4>
                <div className="flex items-center gap-2">
                    {/*Enable relay mode*/}
                    {isHost && !session.isRelayMode && (
                        <Tooltip
                            trigger={<IconButton
                                size="sm"
                                intent={!enablingRelayMode ? "primary-subtle" : "primary"}
                                icon={<FaBroadcastTower />}
                                onClick={() => setEnablingRelayMode(p => !p)}
                                className={cn(enablingRelayMode && "animate-pulse")}
                            />}
                        >
                            Enable relay mode
                        </Tooltip>
                    )}
                    <Button
                        onClick={onLeave}
                        disabled={isLeaving}
                        size="sm"
                        intent="alert-basic"
                        // leftIcon={isHost ? <MdStop /> : <MdExitToApp />}
                    >
                        {isLeaving ? "Leaving..." : isHost ? "Stop" : "Leave"}
                    </Button>
                </div>
            </div>

            {/* <SettingsCard title="Session Details">
             <div className="space-y-3">
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Session ID:</span>
             <span className="font-mono text-xs">{session.id}</span>
             </div>

             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Created:</span>
             <span className="text-sm">{session.createdAt ? new Date(session.createdAt).toLocaleString() : "Unknown"}</span>
             </div>

             {session.currentMediaInfo && (
             <>
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Current Media:</span>
             <span className="text-sm">Episode {session.currentMediaInfo.episodeNumber}</span>
             </div>
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Stream Type:</span>
             <Badge className="">
             {session.currentMediaInfo.streamType}
             </Badge>
             </div>
             </>
             )}
             </div>
             </SettingsCard> */}

            <h5>Participants ({participantCount})</h5>
            <div className="p-4 border rounded-lg bg-gray-950">
                <div className="space-y-0">
                    {participants.map((participant) => {
                        const isCurrentUser = participant.id === currentUserId
                        return (
                            <div key={participant.id} className="flex items-center justify-between py-1">
                                <div className="flex items-center gap-2">
                                    <span className="font-medium text-sm tracking-wide">
                                        {participant.username}
                                        {isCurrentUser && <span className="text-[--muted] font-normal"> (me)</span>}
                                    </span>
                                    {session.isRelayMode && participant.isHost && (
                                        <Badge intent="unstyled" className="text-xs" leftIcon={<FaBroadcastTower />}>Relay</Badge>
                                    )}
                                    {participant.isHost && (
                                        <Badge className="text-xs">Host</Badge>
                                    )}
                                    {participant.isRelayOrigin && (
                                        <Badge intent="warning" className="text-xs">Origin</Badge>
                                    )}
                                    {enablingRelayMode && !participant.isHost && !participant.isRelayOrigin && !session.isRelayMode && (
                                        <Button
                                            size="sm" intent="white" leftIcon={<HiOutlinePlay />}
                                            onClick={() => handleEnableRelayMode(participant.id)}
                                        >Promote to origin</Button>
                                    )}
                                </div>
                                <div className="flex items-center gap-2 text-xs text-[--muted]">
                                    {!participant.isHost && participant.bufferHealth !== undefined && (
                                        <Tooltip
                                            trigger={<div className="flex items-center gap-1">
                                                <span className="text-xs">Buffer</span>
                                            <div className="w-8 h-1 bg-gray-300 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full bg-green-500 transition-all duration-300"
                                                    style={{ width: `${Math.max(0, Math.min(100, participant.bufferHealth * 100))}%` }}
                                                />
                                            </div>
                                            <span className="text-xs">{Math.round(participant.bufferHealth * 100)}%</span>
                                            </div>}
                                        >
                                            Synchronization buffer health
                                        </Tooltip>
                                    )}
                                    {participant.latency > 0 && (
                                        <span>{participant.latency}ms</span>
                                    )}
                                    {participant.isBuffering ? (
                                        <Badge intent="alert-solid" className="text-xs">
                                            Buffering
                                        </Badge>
                                    ) : null}
                                </div>
                            </div>
                        )
                    })}
                </div>
            </div>

            {/* <SettingsCard title="Settings">
             <div className="space-y-2">
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Participant Control:</span>
             <span className="text-sm">
             {session.settings?.allowParticipantControl ? "Enabled" : "Disabled"}
             </span>
             </div>
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Sync Threshold:</span>
             <span className="text-sm">{session.settings?.syncThreshold}s</span>
             </div>
             <div className="flex items-center justify-between">
             <span className="text-sm text-[--muted]">Max Buffer Wait:</span>
             <span className="text-sm">{session.settings?.maxBufferWaitTime}s</span>
             </div>
             </div>
             </SettingsCard> */}
        </div>
    )
}
