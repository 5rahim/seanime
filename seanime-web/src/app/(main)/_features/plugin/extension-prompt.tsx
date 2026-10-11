import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useIsMainTab, websocketConnectedAtom } from "@/app/websocket-provider"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtomValue } from "jotai"
import * as React from "react"
import { LuDownload, LuKeyRound, LuPackage, LuSettings, LuShieldCheck, LuTerminal } from "react-icons/lu"

type ExtensionPromptRequest = {
    id: string
    kind: string
    extension: {
        id: string
        name: string
        icon?: string
    }
    action: string
    resource: string
    message: string
    details?: string[]
    allowLabel?: string
    denyLabel?: string
    expired?: boolean
}

export function ExtensionPrompt() {
    const isMainTab = useIsMainTab()
    const isConnected = useAtomValue(websocketConnectedAtom)
    const { sendMessage } = useWebsocketSender()
    const [queue, setQueue] = React.useState<ExtensionPromptRequest[]>([])
    const prompt = queue[0]
    const frameRef = React.useRef<HTMLDivElement>(null)
    const shakeRef = React.useRef<Animation | null>(null)

    const syncPrompt = React.useEffectEvent(() => {
        sendMessage({
            type: WSEvents.EXTENSION_PROMPT_SYNC,
            payload: {},
        })
    })

    function handleShake() {
        const modal = frameRef.current
        if (modal) {
            shakeRef.current?.cancel()
            shakeRef.current = modal.animate([
                { transform: "translateX(0) scale(1)" },
                { transform: "translateX(-16px) scale(0.995)" },
                { transform: "translateX(12px) scale(0.998)" },
                { transform: "translateX(-8px) scale(1)" },
                { transform: "translateX(5px) scale(1)" },
                { transform: "translateX(0) scale(1)" },
            ], {
                duration: 600,
                easing: "cubic-bezier(0.22, 1, 0.36, 1)",
            })
        }
    }

    useWebsocketMessageListener<ExtensionPromptRequest>({
        type: WSEvents.EXTENSION_PROMPT,
        onMessage: payload => {
            if (!payload?.id) return
            if (payload.expired) {
                setQueue(prev => prev.filter(item => item.id !== payload.id))
                return
            }
            if (!isMainTab) return
            setQueue(prev => prev.some(item => item.id === payload.id) ? prev : [...prev, payload])
        },
        deps: [isMainTab],
    })

    React.useEffect(() => {
        if (isConnected && isMainTab) {
            syncPrompt()
        }
    }, [isConnected, isMainTab])

    const respond = React.useCallback((allowed: boolean) => {
        if (!prompt?.id) return
        sendMessage({
            type: WSEvents.EXTENSION_PROMPT_RESPONSE,
            payload: {
                id: prompt.id,
                allowed,
            },
        })
        setQueue(prev => prev.slice(1))
    }, [prompt?.id, sendMessage])

    if (!isMainTab || !prompt) return null

    const Icon = getIcon(prompt.kind)
    const title = prompt.message || `Allow the extension "${prompt.extension?.name || "Extension"}" to ${prompt.action}?`

    return (
        <Modal
            open
            onOpenChange={open => {
                if (!open) respond(false)
            }}
            hideCloseButton
            onEscapeKeyDown={event => event.preventDefault()}
            onInteractOutside={event => {
                event.preventDefault()
                handleShake()
            }}
            overlayClass="z-[100] bg-gray-950/70"
            contentClass={cn(
                "extension-prompt-content max-w-[30rem] gap-0 overflow-hidden rounded-2xl p-0 shadow-2xl select-none",
                "sm:rounded-[2rem]",
            )}
        >
            <div ref={frameRef} className="grid gap-0 min-w-0">
                <div className="px-8 pb-6 pt-8 min-w-0">
                    <div className="mb-6 flex items-center gap-3 relative w-fit">
                        <div className="grid size-16 place-items-center rounded-xl border bg-gradient-to-br text-white shadow-md">
                            <Icon className="size-8" />
                        </div>
                        {!!prompt.extension?.icon && (
                            <img
                                src={prompt.extension.icon}
                                alt=""
                                className="size-7 rounded-xl bg-transparent object-cover shadow-md absolute -bottom-2 -right-2"
                            />
                        )}
                    </div>

                    <p className="text-sm text-[--muted] mb-2 text-pretty break-all">
                        {prompt.extension?.name || "An extension"} would like to perform the following action{!!prompt.resource
                        ? ` on "${prompt.resource}"`
                        : ""}:
                    </p>

                    <p
                        className={cn(
                            "text-pretty text-2xl font-semibold leading-[1.15] tracking-normal break-words",
                            title.length > 60 ? "text-lg leading-snug" : "text-2xl",
                        )}
                    >
                        {title}
                    </p>

                    {/* {!!prompt.resource && (
                     <p className="mt-3 text-base leading-snug">
                     {prompt.resource}
                     </p>
                     )} */}

                    {!!prompt.details?.length && (
                        <div className="mt-5 max-h-32 overflow-y-auto rounded-xl border bg-[--paper] p-3 text-sm max-w-full space-y-1.5 select-text">
                            {prompt.details.map(detail => (
                                <div key={detail} className="line-clamp-10 break-all">
                                    {detail}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                <div className="grid gap-3 px-6 pb-6 min-w-0">
                    <Button
                        intent="gray-subtle"
                        size="lg"
                        className="h-12 rounded-full border-0 bg-gray-200 text-base text-gray-950 shadow-none hover:bg-gray-300"
                        onClick={() => respond(false)}
                    >
                        {prompt.denyLabel || "Don't Allow"}
                    </Button>
                    <Button
                        intent="warning-subtle"
                        size="lg"
                        className="h-12 rounded-full border-0 text-base shadow-none"
                        onClick={() => respond(true)}
                    >
                        {prompt.allowLabel || "Allow"}
                    </Button>
                </div>
            </div>
        </Modal>
    )
}

function getIcon(kind: string) {
    switch (kind) {
        case "auth":
            return LuKeyRound
        case "settings":
            return LuSettings
        case "extensions":
            return LuPackage
        case "system":
            return LuTerminal
        case "download":
            return LuDownload
        default:
            return LuShieldCheck
    }
}
