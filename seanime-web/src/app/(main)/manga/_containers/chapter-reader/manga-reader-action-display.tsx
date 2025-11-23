import { cn } from "@/components/ui/core/styling"
import { atom, useAtom } from "jotai"
import React from "react"

type MangaReaderFlashAction = {
    id: string
    message: string
    timestamp: number
}

export const manga_flashAction = atom<MangaReaderFlashAction | null>(null)
export const manga_flashActionTimeout = atom<ReturnType<typeof setTimeout> | null>(null)

export const manga_doFlashAction = atom(null, (get, set, payload: { message: string, duration?: number }) => {
    const id = Date.now().toString()
    const timeout = get(manga_flashActionTimeout)
    set(manga_flashAction, { id, message: payload.message, timestamp: Date.now() })

    if (timeout) {
        clearTimeout(timeout)
    }

    const t = setTimeout(() => {
        set(manga_flashAction, null)
        set(manga_flashActionTimeout, null)
    }, payload.duration ?? 800)

    set(manga_flashActionTimeout, t)
})

export function MangaReaderActionDisplay() {
    const [notification] = useAtom(manga_flashAction)

    if (!notification) return null

    return (
        <div className="absolute top-16 left-1/2 transform -translate-x-1/2 z-50 pointer-events-none">
            <div
                className={cn(
                    "text-white px-3 py-2 !text-lg font-semibold rounded-lg bg-black/50 backdrop-blur-sm tracking-wide",
                )}
            >
                {notification.message}
            </div>
        </div>
    )
}

export function useMangaReaderFlashAction() {
    const [, flashAction] = useAtom(manga_doFlashAction)

    return { flashAction }
}

