import { Nakama_NakamaStatus, Nakama_WatchPartySession, Nakama_WatchPartySessionSettings } from "@/api/generated/types"
import {
    useGetNakamaStatus,
    useNakamaCreateWatchParty,
    useNakamaJoinWatchParty,
    useNakamaLeaveWatchParty,
    useNakamaReconnectToHost,
    useNakamaRemoveStaleConnections,
} from "@/api/hooks/nakama.hooks"
import { BetaBadge } from "@/components/shared/beta-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { WSEvents } from "@/lib/server/ws-events"
import { atom, useAtom, useAtomValue } from "jotai"
import React from "react"
import { MdAdd, MdCleaningServices, MdExitToApp, MdOutlineConnectWithoutContact, MdPeople, MdPlayArrow, MdRefresh, MdStop } from "react-icons/md"
import { toast } from "sonner"
import { useWebsocketMessageListener } from "../../_hooks/handle-websockets"
import { SettingsCard } from "../../settings/_components/settings-card"

export const nakamaModalOpenAtom = atom(false)
export const nakamaStatusAtom = atom<Nakama_NakamaStatus | null>(null)


type WatchPartySessionParticipant = {
    id: string
    username: string
    isHost: boolean
    canControl: boolean
    isReady: boolean
    lastSeen: string
    latency: number
    isBuffering: boolean
    bufferHealth: number
    playbackStatus?: any
}

type WatchPartySessionMediaInfo = {
    mediaId: number
    episodeNumber: number
    aniDBEpisode: string
    streamType: string
    streamPath: string
}

export const watchPartySessionAtom = atom<Nakama_WatchPartySession | null>(null)

export function useNakamaStatus() {
    return useAtomValue(nakamaStatusAtom)
}

export function useWatchPartySession() {
    return useAtomValue(watchPartySessionAtom)
}

export function NakamaManager() {
    const [isModalOpen, setIsModalOpen] = useAtom(nakamaModalOpenAtom)
    const [nakamaStatus, setNakamaStatus] = useAtom(nakamaStatusAtom)
    const [watchPartySession, setWatchPartySession] = useAtom(watchPartySessionAtom)

    const { data: status, refetch: refetchStatus, isLoading } = useGetNakamaStatus()
    const { mutate: reconnectToHost, isPending: isReconnecting } = useNakamaReconnectToHost()
    const { mutate: removeStaleConnections, isPending: isCleaningUp } = useNakamaRemoveStaleConnections()
    const { mutate: createWatchParty, isPending: isCreatingWatchParty } = useNakamaCreateWatchParty()
    const { mutate: joinWatchParty, isPending: isJoiningWatchParty } = useNakamaJoinWatchParty()
    const { mutate: leaveWatchParty, isPending: isLeavingWatchParty } = useNakamaLeaveWatchParty()

    // Watch party settings for creating a new session
    const [watchPartySettings, setWatchPartySettings] = React.useState<Nakama_WatchPartySessionSettings>({
        allowParticipantControl: false,
        syncThreshold: 3.0,
        maxBufferWaitTime: 10,
    })

    React.useEffect(() => {
        setNakamaStatus(status ?? null)
        if (status?.currentWatchPartySession) {
            setWatchPartySession(status.currentWatchPartySession)
        } else {
            setWatchPartySession(null)
        }
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

    const handleCreateWatchParty = React.useCallback(() => {
        createWatchParty({ settings: watchPartySettings }, {
            onSuccess: () => {
                toast.success("Watch party created successfully")
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
                toast.success("Joined watch party")
                refetchStatus()
            },
            onError: (error) => {
                toast.error(`Failed to join watch party: ${error.message}`)
            },
        })
    }, [joinWatchParty, refetchStatus])

    const handleLeaveWatchParty = React.useCallback(() => {
        leaveWatchParty(undefined, {
            onSuccess: () => {
                toast.success("Left watch party")
                setWatchPartySession(null)
                refetchStatus()
            },
            onError: (error) => {
                toast.error(`Failed to leave watch party: ${error.message}`)
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

    // Watch Party websocket listeners
    useWebsocketMessageListener({
        type: WSEvents.NAKAMA_WATCH_PARTY_STATE,
        onMessage: (data: any) => {
            // if (data === null || data === undefined) {
            //     // Watch party was stopped
            //     setWatchPartySession(null)
            // } else {
            //     // Session data received
            //     const session = data as Nakama_WatchPartySession
            //     setWatchPartySession(session)
            // }
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

            {!status?.isHost && (
                <div className="flex items-center justify-between">
                    <div></div>
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
            )}

            {!isLoading && (status?.isHost || status?.isConnectedToHost) && (
                <Tabs defaultValue="connection" className="w-full">
                    <TabsList className="grid w-full grid-cols-2 mb-4">
                        <TabsTrigger value="connection">Connection</TabsTrigger>
                        <TabsTrigger value="watch-party">Watch Party</TabsTrigger>
                    </TabsList>

                    <TabsContent value="connection" className="space-y-4">
                        {status?.isHost && (
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

                        {status?.isConnectedToHost && (
                            <>

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
                                                className={`font-medium ${status?.hostConnectionStatus?.authenticated
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
                    </TabsContent>

                    <TabsContent value="watch-party" className="space-y-4">
                        {/* Watch Party Content */}
                        {(() => {
                            const isHost = status?.isHost || false
                            const isConnectedToHost = status?.isConnectedToHost || false
                            const currentPeerID = status?.hostConnectionStatus?.peerId

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
                    </TabsContent>
                </Tabs>
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
            {isHost && (
                <SettingsCard title="Create Watch Party">
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
                </SettingsCard>
            )}

            {isConnectedToHost && !isHost && hasActiveSession && (
                <SettingsCard title="Join Watch Party">
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
                </SettingsCard>
            )}

            {!isHost && !isConnectedToHost && (
                <div className="text-center py-8">
                    <p className="text-[--muted]">Connect to a host to join a watch party</p>
                </div>
            )}

            {!isHost && isConnectedToHost && !hasActiveSession && (
                <div className="text-center py-8">
                    <p className="text-[--muted]">No active watch party</p>
                    <p className="text-sm text-[--muted] mt-2">
                        Waiting for the host to create a watch party...
                    </p>
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
    const participants = Object.values(session.participants || {})
    const participantCount = participants.length

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h4>Active Watch Party</h4>
                <div className="flex items-center gap-2">
                    <Badge className="flex items-center gap-1">
                        <MdPeople className="size-3" />
                        {participantCount} participant{participantCount !== 1 ? "s" : ""}
                    </Badge>
                    <Button
                        onClick={onLeave}
                        disabled={isLeaving}
                        size="sm"
                        intent="warning-subtle"
                        leftIcon={isHost ? <MdStop /> : <MdExitToApp />}
                    >
                        {isLeaving ? "Leaving..." : isHost ? "Stop Party" : "Leave Party"}
                    </Button>
                </div>
            </div>

            <SettingsCard title="Session Details">
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
            </SettingsCard>

            <SettingsCard title="Participants">
                <div className="space-y-2">
                    {participants.map((participant) => (
                        <div key={participant.id} className="flex items-center justify-between py-2">
                            <div className="flex items-center gap-2">
                                <span className="font-medium">{participant.username}</span>
                                {participant.isHost && (
                                    <Badge className="text-xs">Host</Badge>
                                )}
                                {participant.canControl && !participant.isHost && (
                                    <Badge className="text-xs">Controller</Badge>
                                )}
                            </div>
                            <div className="flex items-center gap-2 text-xs text-[--muted]">
                                {(participant as any).isBuffering ? (
                                    <Badge className="text-xs bg-red-500 text-white">
                                        Buffering
                                    </Badge>
                                ) : participant.isReady ? (
                                    <Badge className="text-xs bg-green-500 text-white">
                                        Ready
                                    </Badge>
                                ) : (
                                    <Badge className="text-xs bg-gray-500 text-white">
                                        Not Ready
                                    </Badge>
                                )}
                                {!participant.isHost && (participant as any).bufferHealth !== undefined && (
                                    <div className="flex items-center gap-1">
                                        <span className="text-xs">Buffer:</span>
                                        <div className="w-8 h-1 bg-gray-300 rounded-full overflow-hidden">
                                            <div
                                                className="h-full bg-green-500 transition-all duration-300"
                                                style={{ width: `${Math.max(0, Math.min(100, (participant as any).bufferHealth * 100))}%` }}
                                            />
                                        </div>
                                        <span className="text-xs">{Math.round((participant as any).bufferHealth * 100)}%</span>
                                    </div>
                                )}
                                {participant.latency > 0 && (
                                    <span>{participant.latency}ms</span>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            </SettingsCard>

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
