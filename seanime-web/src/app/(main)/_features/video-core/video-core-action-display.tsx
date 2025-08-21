// Flash notification system
import { vc_miniPlayer, vc_paused } from "@/app/(main)/_features/video-core/video-core"
import { cn } from "@/components/ui/core/styling"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/index"
import { motion } from "motion/react"
import React from "react"
import { PiPauseDuotone, PiPlayDuotone } from "react-icons/pi"

type VideoCoreFlashAction = {
    id: string
    message: string
    timestamp: number
    type: "message" | "time" | "icon"
}
export const vc_flashAction = atom<VideoCoreFlashAction | null>(null)
export const vc_flashActionTimeout = atom<ReturnType<typeof setTimeout> | null>(null)
export const vc_doFlashAction = atom(null, (get, set, payload: { message: string, type?: "message" | "time" | "icon", duration?: number }) => {
    const id = Date.now().toString()
    const timeout = get(vc_flashActionTimeout)
    const paused = get(vc_paused)
    set(vc_flashAction, { id, message: payload.message, timestamp: Date.now(), type: payload.type ?? "message" })
    if (timeout) {
        clearTimeout(timeout)
    }
    const t = setTimeout(() => {
        set(vc_flashAction, null)
        set(vc_flashActionTimeout, null)
    }, payload.duration ?? (payload.type === "icon" ? 200 : (paused ? 1000 : 500))) // stays longer when paused
    set(vc_flashActionTimeout, t)

})

export function VideoCoreActionDisplay() {
    const [notification] = useAtom(vc_flashAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)

    if (!notification) return null

    if (notification.type === "icon") {
        return (
            <motion.div
                initial={{ opacity: 0.2, scale: 1 }}
                animate={{ opacity: 0.5, scale: 1.6 }}
                exit={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.06, ease: "easeOut" }}
                className="absolute w-full h-full pointer-events-none flex z-[50] items-center justify-center"
            >
                {notification.message === "PLAY" &&
                    <PiPlayDuotone
                        className={cn("size-24 text-white", isMiniPlayer && "size-10")}
                        style={{ textShadow: "0 1px 10px rgba(0, 0, 0, 0.8)" }}
                    />}
                {notification.message === "PAUSE" &&
                    <PiPauseDuotone
                        className={cn("size-24 text-white", isMiniPlayer && "size-10")}
                        style={{ textShadow: "0 1px 10px rgba(0, 0, 0, 0.8)" }}
                    />}
            </motion.div>
        )
    }

    return (
        <div className="absolute top-16 left-1/2 transform -translate-x-1/2 z-50 pointer-events-none">
            <div
                className={cn(
                    "text-white px-2 py-1 !text-xl font-semibold rounded-lg bg-black/50 backdrop-blur-sm tracking-wide",
                    isMiniPlayer && "text-sm",
                )}
            >
                {notification.message}
            </div>
        </div>
    )
}

export function useVideoCoreFlashAction() {
    const [, flashAction] = useAtom(vc_doFlashAction)

    return { flashAction }
}
