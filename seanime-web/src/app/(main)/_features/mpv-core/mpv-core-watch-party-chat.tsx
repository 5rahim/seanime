import { MediaCoreControlButtonIcon } from "@/app/(main)/_features/media-core/media-core-control-bar"
import { MediaCoreMenu } from "@/app/(main)/_features/media-core/media-core-menu"
import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import {
    NakamaWatchPartyChat,
    watchPartyChat_chatMinimizedAtom,
    watchPartyChat_isPlayerAtom,
    watchPartyChat_unreadCountAtom,
} from "@/app/(main)/_features/nakama/nakama-watch-party-chat"
import { useAtom, useSetAtom } from "jotai"
import React from "react"
import { LuMessagesSquare } from "react-icons/lu"

export function MpvCoreWatchPartyChat(props: {
    isMiniPlayer: boolean
    isFullscreen: boolean
    containerElement: HTMLDivElement | null
    openMenu: string | null
    setOpenMenu: React.Dispatch<React.SetStateAction<string | null>>
    showMessage: (message: string, type?: "message" | "time" | "icon", durationMs?: number) => void
}) {
    const { isMiniPlayer, isFullscreen, containerElement, openMenu, setOpenMenu, showMessage } = props

    const { watchPartySession, isParticipant } = useNakamaWatchParty()
    const [unreadCount, setUnreadCount] = useAtom(watchPartyChat_unreadCountAtom)
    const [minimized, setMinimized] = useAtom(watchPartyChat_chatMinimizedAtom)

    const shouldHide = isMiniPlayer || !watchPartySession || !isParticipant

    const setIsPlayer = useSetAtom(watchPartyChat_isPlayerAtom)
    React.useEffect(() => {
        setIsPlayer(!shouldHide)
        return () => {
            setIsPlayer(false)
        }
    }, [shouldHide, setIsPlayer])

    React.useEffect(() => {
        setMinimized(openMenu !== "watch-party-chat")
    }, [openMenu, setMinimized])

    const previousCountRef = React.useRef<number | null>(null)
    React.useEffect(() => {
        if (previousCountRef.current === null) {
            previousCountRef.current = unreadCount
            return
        }
        previousCountRef.current = unreadCount
        if (!!unreadCount) {
            showMessage(`New chat message (${unreadCount})`, "message", 1000)
        }
    }, [unreadCount, showMessage])

    if (shouldHide) return null

    return (
        <MediaCoreMenu
            name="watch-party-chat"
            openMenu={openMenu}
            onOpenMenuChange={setOpenMenu}
            isFullscreen={isFullscreen}
            containerElement={containerElement}
            trigger={
                <MediaCoreControlButtonIcon
                    icons={[["default", LuMessagesSquare]]}
                    state="default"
                    className="text-2xl"
                    onClick={() => {
                        setUnreadCount(0)
                    }}
                    isMobile={false}
                    isMiniPlayer={isMiniPlayer}
                >
                    {unreadCount > 0 && (
                        <div className="absolute -top-1 -right-1 min-w-[1.2rem] h-5 px-1 rounded-full bg-red-600 text-white text-xs flex items-center justify-center">
                            {unreadCount}
                        </div>
                    )}
                </MediaCoreControlButtonIcon>
            }
        >
            <NakamaWatchPartyChat layout="videocore" />
        </MediaCoreMenu>
    )
}
