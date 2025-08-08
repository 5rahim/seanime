// Flash notification system
import { vc_paused } from "@/app/(main)/_features/video-core/video-core"
import { atom } from "jotai"
import { useAtom } from "jotai/index"
import React from "react"

type VideoCoreFlashAction = {
    id: string
    message: string
    timestamp: number
    type: "message" | "time"
}
export const vc_flashAction = atom<VideoCoreFlashAction | null>(null)
export const vc_flashActionTimeout = atom<ReturnType<typeof setTimeout> | null>(null)
export const vc_doFlashAction = atom(null, (get, set, payload: { message: string, type?: "message" | "time" }) => {
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
    }, paused ? 1000 : 500) // stays longer when paused
    set(vc_flashActionTimeout, t)

})

export function VideoCoreActionDisplay() {
    const [notification] = useAtom(vc_flashAction)

    if (!notification) return null

    return (
        <div className="absolute top-16 left-1/2 transform -translate-x-1/2 z-50 pointer-events-none">
            <div className="text-white px-2 py-1 !text-xl font-semibold rounded-lg bg-black/50 backdrop-blur-sm tracking-wide">
                {notification.message}
            </div>
        </div>
    )
    // return (
    //     <div className="absolute top-16 left-1/2 transform -translate-x-1/2 z-50 pointer-events-none">
    //         <div className="text-white px-4 py-2 !text-xl font-bold" style={{ textShadow: "0 1px 10px rgba(0, 0, 0, 0.8)" }}>
    //             {notification.message}
    //         </div>
    //     </div>
    // )
}

export function useVideoCoreFlashAction() {
    const [, flashAction] = useAtom(vc_doFlashAction)

    return { flashAction }
}
