import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import {
    NakamaWatchPartyChat,
    watchPartyChat_chatMinimizedAtom,
    watchPartyChat_isPlayerAtom,
    watchPartyChat_unreadCountAtom,
} from "@/app/(main)/_features/nakama/nakama-watch-party-chat"
import { nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"
import { vc_isFullscreen, vc_miniPlayer, vc_videoElement } from "@/app/(main)/_features/video-core/video-core"
import { vc_doFlashAction } from "@/app/(main)/_features/video-core/video-core-action-display"
import { VideoCoreControlButtonIcon } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { vc_menuOpen, VideoCoreMenu } from "@/app/(main)/_features/video-core/video-core-menu"
import { useAtom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import React from "react"
import { LuMessagesSquare } from "react-icons/lu"

export function VideoCoreWatchPartyChat() {
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const state = useAtomValue(nativePlayer_stateAtom)
    const videoElement = useAtomValue(vc_videoElement)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const flashAction = useSetAtom(vc_doFlashAction)

    const [open, setOpen] = useAtom(vc_menuOpen)

    const { watchPartySession, isParticipant, currentUserPeerId } = useNakamaWatchParty()
    const [unreadCount, setUnreadCount] = useAtom(watchPartyChat_unreadCountAtom)
    const [minimized, setMinimized] = useAtom(watchPartyChat_chatMinimizedAtom)

    const shouldHide = isMiniPlayer || !watchPartySession || !isParticipant

    const setIsPlayer = useSetAtom(watchPartyChat_isPlayerAtom)
    React.useEffect(() => {
        setIsPlayer(!shouldHide)
        return () => {
            setIsPlayer(false)
        }
    }, [shouldHide])

    React.useEffect(() => {
        setMinimized(open !== "watch-party-chat")
    }, [open])

    const previousCountRef = React.useRef<number | null>(null)
    React.useEffect(() => {
        if (previousCountRef.current === null) {
            previousCountRef.current = unreadCount
            return
        }
        previousCountRef.current = unreadCount
        if (!!unreadCount) {
            flashAction({ message: `New chat message (${unreadCount})`, duration: 600 })
        }
    }, [unreadCount])

    if (shouldHide) return null

    return (
        <VideoCoreMenu
            isDrawer
            name="watch-party-chat"
            sideOffset={8}
            className="bg-black/85 rounded-xl p-2 backdrop-blur-sm w-[30rem] !top-auto !h-fit max-h-[70dvh] z-[100] lg:top-10 lg:bottom-24"
            trigger={<VideoCoreControlButtonIcon
                icons={[
                    ["default", LuMessagesSquare],
                ]}
                state="default"
                className="text-2xl"
                onClick={() => {
                    setUnreadCount(0)
                }}
            >
                {unreadCount > 0 && (
                    <div className="absolute -top-1 -right-1 min-w-[1.2rem] h-5 px-1 rounded-full bg-red-600 text-white text-xs flex items-center justify-center">
                        {unreadCount}
                    </div>
                )}
            </VideoCoreControlButtonIcon>}
        >
            <NakamaWatchPartyChat layout="videocore" />
        </VideoCoreMenu>
    )
}
