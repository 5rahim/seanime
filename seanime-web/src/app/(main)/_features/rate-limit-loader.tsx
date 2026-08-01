import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { ProgressBar } from "@/components/ui/progress-bar"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import React from "react"
import { FiClock } from "react-icons/fi"

const log = logger("RateLimitLoader")

export function RateLimitLoader() {
    const [totalSeconds, setTotalSeconds] = React.useState(0)
    const [secondsRemaining, setSecondsRemaining] = React.useState(0)

    useWebsocketMessageListener<number>({
        type: WSEvents.ANILIST_RATE_LIMIT,
        onMessage: (waitSeconds) => {
            if (typeof waitSeconds === "number" && waitSeconds > 0) {
                log.info(`Received AniList rate limit event: retrying in ${waitSeconds} seconds`)
                setTotalSeconds(waitSeconds)
                setSecondsRemaining(waitSeconds)
            }
        },
    })

    React.useEffect(() => {
        if (secondsRemaining <= 0) return

        const interval = setInterval(() => {
            setSecondsRemaining(prev => {
                if (prev <= 1) {
                    clearInterval(interval)
                    return 0
                }
                return prev - 1
            })
        }, 1000)

        return () => clearInterval(interval)
    }, [secondsRemaining])

    if (secondsRemaining <= 0) return null

    const progressValue = (secondsRemaining / totalSeconds) * 100

    return (
        <div className="fixed top-0 left-0 w-full z-[100] pointer-events-none flex flex-col items-center">
            <ProgressBar
                value={progressValue}
                size="xs"
                className="rounded-none bg-orange-950/20"
                indicatorClass="bg-orange-500 rounded-none transition-all duration-1000 ease-linear"
            />
            <div className="mt-2 bg-orange-950/90 border border-orange-500/30 text-orange-200 px-3 py-1.5 rounded-full text-xs font-medium shadow-lg backdrop-blur-sm animate-fade-in flex items-center gap-1.5">
                <FiClock className="animate-spin text-orange-400 size-3.5" style={{ animationDuration: "3s" }} />
                <span>AniList rate limit: retrying in {secondsRemaining}s</span>
            </div>
        </div>
    )
}
