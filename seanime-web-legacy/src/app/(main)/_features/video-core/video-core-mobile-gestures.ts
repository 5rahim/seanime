import { vc_duration, vc_isFullscreen, vc_isMobile, vc_isSwiping, vc_swipeSeekTime } from "@/app/(main)/_features/video-core/video-core"
import { vc_timeRangeElement } from "@/app/(main)/_features/video-core/video-core-time-range"
import { logger } from "@/lib/helpers/debug"
import { useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { useEffect, useRef } from "react"

const log = logger("VIDEO CORE MOBILE GESTURES")

type UseVideoCoreMobileGestures = {
    videoElement: HTMLVideoElement | null
    containerElement: HTMLElement | null
    onSeek: (time: number) => void
}

export function useVideoCoreMobileGestures({
    videoElement,
    containerElement,
    onSeek,
}: UseVideoCoreMobileGestures) {
    const [isSwiping, setIsSwiping] = useAtom(vc_isSwiping)
    const setSwipeSeekTime = useSetAtom(vc_swipeSeekTime)
    const isMobile = useAtomValue(vc_isMobile)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const duration = useAtomValue(vc_duration)
    const timeRangeElement = useAtomValue(vc_timeRangeElement)
    const touchStartRef = useRef<{ x: number; y: number; time: number; currentTime: number } | null>(null)

    useEffect(() => {
        // Only enable on mobile and when not in fullscreen
        if (!isMobile || isFullscreen || !containerElement || !videoElement || duration <= 1) {
            return
        }

        const handleTouchStart = (e: TouchEvent) => {
            // Only handle single touch
            if (e.touches.length !== 1) return

            // Ignore touches that start on the time range element
            if (timeRangeElement && (e.target === timeRangeElement || timeRangeElement.contains(e.target as Node))) {
                setIsSwiping(false)
                return
            }

            const touch = e.touches[0]
            touchStartRef.current = {
                x: touch.clientX,
                y: touch.clientY,
                time: Date.now(),
                currentTime: videoElement.currentTime,
            }

            log.info("Touch start", touchStartRef.current)
        }

        const handleTouchMove = (e: TouchEvent) => {
            if (!touchStartRef.current || e.touches.length !== 1) return
            // e.preventDefault()

            const touch = e.touches[0]
            const deltaX = touch.clientX - touchStartRef.current.x
            const deltaY = touch.clientY - touchStartRef.current.y

            // Check if it's not a vertical swipe for scrolling
            const isHorizontal = Math.abs(deltaX) > Math.abs(deltaY) || isSwiping

            if (isHorizontal) {
                e.preventDefault()

                if (!isSwiping) {
                    setIsSwiping(true)
                    log.info("Started swiping")
                }

                const screenWidth = window.innerWidth
                const seekRatio = deltaX / screenWidth
                const seekDelta = seekRatio * duration
                const newTime = Math.max(0, Math.min(duration, touchStartRef.current.currentTime + seekDelta))

                setSwipeSeekTime(newTime)
            }
        }

        const handleTouchEnd = (e: TouchEvent) => {
            if (!touchStartRef.current) return

            if (isSwiping) {
                // Apply the seek
                const touch = e.changedTouches[0]
                const deltaX = touch.clientX - touchStartRef.current.x
                const deltaY = touch.clientY - touchStartRef.current.y

                const isHorizontal = Math.abs(deltaX) > Math.abs(deltaY) || isSwiping

                // Only seek if it was a horizontal swipe
                if (isHorizontal) {
                    const screenWidth = window.innerWidth
                    const seekRatio = deltaX / screenWidth
                    const seekDelta = seekRatio * duration
                    const newTime = Math.max(0, Math.min(duration, touchStartRef.current.currentTime + seekDelta))

                    log.info("Applying seek to", newTime)
                    onSeek(newTime)
                }

                // Clear swiping state immediately to hide time range
                setIsSwiping(false)
                setSwipeSeekTime(null)
                log.info("Ended swiping")
            }

            touchStartRef.current = null
        }

        const handleTouchCancel = () => {
            if (isSwiping) {
                setIsSwiping(false)
                setSwipeSeekTime(null)
                log.info("Swipe cancelled")
            }
            touchStartRef.current = null
        }

        // Add touch event listeners
        containerElement.addEventListener("touchstart", handleTouchStart, { passive: true })
        containerElement.addEventListener("touchmove", handleTouchMove, { passive: false })
        containerElement.addEventListener("touchend", handleTouchEnd, { passive: true })
        containerElement.addEventListener("touchcancel", handleTouchCancel, { passive: true })

        return () => {
            containerElement.removeEventListener("touchstart", handleTouchStart)
            containerElement.removeEventListener("touchmove", handleTouchMove)
            containerElement.removeEventListener("touchend", handleTouchEnd)
            containerElement.removeEventListener("touchcancel", handleTouchCancel)
        }
    }, [videoElement, containerElement, isMobile, isFullscreen, isSwiping, setIsSwiping, setSwipeSeekTime, duration, onSeek, timeRangeElement])
}

