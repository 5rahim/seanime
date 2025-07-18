import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { ProgressBar } from "@/components/ui/progress-bar"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import React from "react"

const log = logger("IndefiniteLoader")

export function TopIndefiniteLoader() {

    const [showStack, setShowStack] = React.useState<string[]>([])

    // Empty after 3 minutes
    React.useEffect(() => {
        const timeout = setTimeout(() => {
            setShowStack([])
        }, 3 * 60 * 1000)
        return () => clearTimeout(timeout)
    }, [showStack])

    useWebsocketMessageListener<string>({
        type: WSEvents.SHOW_INDEFINITE_LOADER,
        onMessage: data => {
            if (data) {
                log.info("Showing indefinite loader", data)
                setShowStack(prev => {
                    if (prev.includes(data)) {
                        return prev
                    }
                    return [...prev, data]
                })
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.HIDE_INDEFINITE_LOADER,
        onMessage: data => {
            if (data) {
                log.info("Hiding indefinite loader", data)
                setShowStack(prev => {
                    return prev.filter(item => item !== data)
                })
            }
        },
    })

    return (
        <>
            {showStack.length > 0 && <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                <ProgressBar size="xs" isIndeterminate />
            </div>}
        </>
    )
}
