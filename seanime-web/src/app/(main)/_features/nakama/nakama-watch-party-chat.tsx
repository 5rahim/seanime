import { useNakamaSendChatMessage } from "@/api/hooks/nakama.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { TextInput } from "@/components/ui/text-input"
import { WSEvents } from "@/lib/server/ws-events"
import { atom, useAtom } from "jotai"
import React from "react"
import { BiChevronDown, BiChevronUp } from "react-icons/bi"
import { HiOutlineChatBubbleLeftRight } from "react-icons/hi2"
import { IoSend } from "react-icons/io5"
import { useNakamaStatus, useWatchPartySession } from "./nakama-manager"

type ChatMessage = {
    peerId: string
    username: string
    message: string
    timestamp: string
    messageId: string
}

const chatMessagesAtom = atom<ChatMessage[]>([])
const chatMinimizedAtom = atom<boolean>(true)
const unreadCountAtom = atom<number>(0)

export function NakamaWatchPartyChat() {
    const watchPartySession = useWatchPartySession()
    const nakamaStatus = useNakamaStatus()
    const [messages, setMessages] = useAtom(chatMessagesAtom)
    const [minimized, setMinimized] = useAtom(chatMinimizedAtom)
    const [unreadCount, setUnreadCount] = useAtom(unreadCountAtom)
    const [inputValue, setInputValue] = React.useState("")
    const messagesEndRef = React.useRef<HTMLDivElement>(null)
    const chatContainerRef = React.useRef<HTMLDivElement>(null)
    const inputRef = React.useRef<HTMLInputElement>(null)

    const { mutate: sendChatMessage, isPending: isSending } = useNakamaSendChatMessage()

    const currentUserPeerId = React.useMemo(() => {
        if (nakamaStatus?.isHost) {
            return "host"
        }
        return nakamaStatus?.hostConnectionStatus?.peerId || null
    }, [nakamaStatus])

    const isParticipant = React.useMemo(() => {
        if (!watchPartySession) return false
        const participants = watchPartySession.participants || {}
        if (nakamaStatus?.isHost) {
            return true
        }
        // If user is a peer, check if they're in the participants list
        return !!(currentUserPeerId && participants[currentUserPeerId])
    }, [watchPartySession, nakamaStatus, currentUserPeerId])

    // Listen for chat messages
    useWebsocketMessageListener<ChatMessage>({
        type: WSEvents.NAKAMA_WATCH_PARTY_CHAT_MESSAGE,
        onMessage: (data) => {
            setMessages(prev => [...prev, data])

            // Track own messages by storing the peerId when sending
            // If minimized and not own message, increment unread count
            if (minimized && data.peerId !== currentUserPeerId) {
                setUnreadCount(prev => prev + 1)
            }
        },
    })

    // Auto-scroll to bottom when new messages arrive
    React.useEffect(() => {
        if (!minimized && messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: "smooth" })
        }
    }, [messages, minimized])

    // Clear unread count when maximized
    React.useEffect(() => {
        if (!minimized) {
            setUnreadCount(0)
            inputRef.current?.focus()
        }
    }, [minimized])

    // Clear messages when session ends
    React.useEffect(() => {
        if (!watchPartySession) {
            setMessages([])
            setUnreadCount(0)
        }
    }, [watchPartySession, setMessages])

    const handleSendMessage = () => {
        if (!inputValue.trim() || isSending) return
        inputRef.current?.focus()

        sendChatMessage({ message: inputValue.trim() }, {
            onSuccess: () => {
                setInputValue("")
                setTimeout(() => {
                    inputRef.current?.focus()
                }, 200)
            },
            onError: (error) => {
                console.error("Failed to send message:", error)
            },
        })
    }

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault()
            handleSendMessage()
        }
    }

    function isHostMessage(msg: ChatMessage) {
        return msg.peerId === "host"
    }

    // Don't show chat if there's no session or user is not a participant
    if (!watchPartySession || !isParticipant) return null

    return (
        <div
            className={cn(
                "fixed bottom-4 right-4 z-[40] flex flex-col bg-gray-900 border rounded-lg shadow-2xl transition-all duration-300",
                minimized ? "w-64 h-14" : "w-[400px] h-[500px]",
            )}
        >
            <div
                className="flex items-center justify-between px-4 py-3 border-b cursor-pointer hover:bg-gray-800/50 transition-colors rounded-t-lg"
                onClick={() => setMinimized(!minimized)}
            >
                <div className="flex items-center gap-2">
                    <HiOutlineChatBubbleLeftRight className="text-xl text-white" />
                    <span className="font-semibold text-sm">Watch Party Chat</span>
                    {minimized && unreadCount > 0 && (
                        <span className="bg-red-500 text-white text-xs font-bold w-5 flex justify-center items-center rounded-full animate-bounce shadow-lg">
                            {unreadCount > 9 ? "9+" : unreadCount}
                        </span>
                    )}
                </div>
                <IconButton
                    icon={minimized ? <BiChevronUp /> : <BiChevronDown />}
                    intent="gray-basic"
                    size="xs"
                    onClick={(e) => {
                        e.stopPropagation()
                        setMinimized(!minimized)
                    }}
                />
            </div>

            {!minimized && (
                <>
                    <div
                        ref={chatContainerRef}
                        className="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent"
                    >
                        {messages.length === 0 ? (
                            <div className="flex items-center justify-center h-full text-gray-500 text-sm">
                                No messages yet
                            </div>
                        ) : (
                            messages.map((msg) => {
                                const isOwnMessage = msg.peerId === currentUserPeerId
                                return (
                                    <div
                                        key={msg.messageId}
                                        className={cn(
                                            "flex flex-col gap-1 p-2 rounded",
                                            isOwnMessage && "bg-gray-800",
                                        )}
                                    >
                                        <div className="flex items-baseline justify-between gap-2">
                                            <span
                                                className={cn(
                                                    "font-semibold text-sm tracking-wide",
                                                    "text-white",
                                                )}
                                            >
                                                {isOwnMessage ? "Me" : msg.username}{isHostMessage(msg) && " (Host)"}:{" "}
                                            </span>
                                            <span className="text-xs text-gray-500">
                                                {new Date(msg.timestamp).toLocaleTimeString([], {
                                                    hour: "2-digit",
                                                    minute: "2-digit",
                                                })}
                                            </span>
                                        </div>
                                        <p className="text-sm text-gray-200 break-words whitespace-pre-wrap">{msg.message}</p>
                                    </div>
                                )
                            })
                        )}
                        <div ref={messagesEndRef} />
                    </div>

                    <div className="p-2 border-t border-gray-800">
                        <div className="flex gap-2 items-center">
                            <TextInput
                                value={inputValue}
                                onValueChange={setInputValue}
                                onKeyDown={handleKeyPress}
                                placeholder="Type a message..."
                                disabled={isSending}
                                className="flex-1 h-10"
                                size="sm"
                                ref={inputRef}
                                autoFocus
                                autoComplete="off"
                            />
                            <IconButton
                                icon={<IoSend />}
                                onClick={handleSendMessage}
                                disabled={!inputValue.trim() || isSending}
                                intent="primary"
                                size="sm"
                            />
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}

