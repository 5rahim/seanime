import { Plugin_Server_DebugLogEventPayload } from "@/app/(main)/_features/plugin/generated/plugin-events.ts"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling.ts"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LuClipboard } from "react-icons/lu"
import { LuSearch } from "react-icons/lu"
import { LuTrash2 } from "react-icons/lu"
import { LuX } from "react-icons/lu"

export type DebugLogEntry = Plugin_Server_DebugLogEventPayload & {
    id: number
    extensionId: string
}

export const maxDebugLogs = 500

const windowPadding = 12
const windowW = 680
const windowH = 520
const minWindowH = 280
const autoScrollThrshld = 48
const dbgLvls = ["all", "log", "info", "warn", "error", "debug"] as const

type DebugLevelFilter = (typeof dbgLvls)[number]

function clamp(value: number, min: number, max: number) {
    return Math.min(Math.max(value, min), max)
}

function getInitHeight() {
    if (typeof window === "undefined") return windowH

    return Math.min(windowH, Math.max(minWindowH, window.innerHeight - windowPadding * 2))
}

function formatDebugValue(value: any) {
    if (value === undefined) return "undefined"
    if (typeof value === "string") return value

    try {
        return JSON.stringify(value, null, 2)
    }
    catch {
        return String(value)
    }
}

function levelClass(level: string) {
    switch (level) {
        case "error":
            return "border-red-500/40 bg-red-500/10 text-[--red]"
        case "warn":
            return "border-yellow-500/40 bg-yellow-500/10 text-[--yellow]"
        case "info":
            return "border-sky-500/40 bg-sky-500/10 text-[--sky]"
        case "debug":
            return "border-violet-500/40 bg-violet-500/10 text-[--violet]"
        default:
            return "border-[--border] bg-[--subtle] text-[--foreground]"
    }
}

function isElNearBottom(element: HTMLDivElement) {
    return element.scrollHeight - element.scrollTop - element.clientHeight <= autoScrollThrshld
}

function getSearchText(log: DebugLogEntry) {
    const values = log.values?.length
        ? formatDebugValue(log.values.length === 1 ? log.values[0] : log.values)
        : ""

    return [log.level, log.message, log.extensionId, log.at, values].join("\n").toLowerCase()
}

export function PluginDebugWindow({
    extensionId,
    extensionName,
    logs,
    index,
    onClear,
    onClose,
}: {
    extensionId: string
    extensionName?: string
    logs: DebugLogEntry[]
    index: number
    onClear: () => void
    onClose: () => void
}) {
    const initialHeight = React.useMemo(() => getInitHeight(), [])
    const [height, setHeight] = React.useState(initialHeight)
    const [position, setPosition] = React.useState(() => ({
        x: Math.max(windowPadding, Math.min((typeof window !== "undefined" ? window.innerWidth : 900) - 700, 88 + index * 28)),
        y: typeof window === "undefined"
            ? 84 + index * 28
            : clamp(84 + index * 28, windowPadding, Math.max(windowPadding, window.innerHeight - initialHeight - windowPadding)),
    }))
    const [level, setLevel] = React.useState<DebugLevelFilter>("all")
    const [search, setSearch] = React.useState("")
    const deferredSearch = React.useDeferredValue(search)
    const drag = React.useRef<{ x: number, y: number, left: number, top: number } | null>(null)
    const resize = React.useRef<{ y: number, height: number } | null>(null)
    const panelRef = React.useRef<HTMLDivElement | null>(null)
    const listRef = React.useRef<HTMLDivElement | null>(null)
    const shouldAutoScrollRef = React.useRef(true)

    const clampHeight = React.useCallback((nextHeight: number, top: number) => {
        if (typeof window === "undefined") return nextHeight

        const maxHeight = Math.max(minWindowH, window.innerHeight - top - windowPadding)
        return clamp(nextHeight, minWindowH, maxHeight)
    }, [])

    const clampPosition = React.useCallback((nextX: number, nextY: number, nextHeight: number) => {
        if (typeof window === "undefined") {
            return { x: nextX, y: nextY }
        }

        const panelWidth = panelRef.current?.offsetWidth ?? Math.min(windowW, window.innerWidth - windowPadding * 2)
        const maxX = Math.max(windowPadding, window.innerWidth - panelWidth - windowPadding)
        const maxY = Math.max(windowPadding, window.innerHeight - nextHeight - windowPadding)

        return {
            x: clamp(nextX, windowPadding, maxX),
            y: clamp(nextY, windowPadding, maxY),
        }
    }, [])

    const visibleLogs = React.useMemo(() => {
        const query = deferredSearch.trim().toLowerCase()

        return logs.filter(log => {
            if (level !== "all" && log.level !== level) return false
            if (!query) return true

            return getSearchText(log).includes(query)
        })
    }, [deferredSearch, level, logs])

    React.useEffect(() => {
        const list = listRef.current
        if (!list || !shouldAutoScrollRef.current) return

        list.scrollTo({ top: list.scrollHeight })
    }, [visibleLogs.length])

    React.useEffect(() => {
        setHeight(prev => clampHeight(prev, position.y))
    }, [clampHeight, position.y])

    React.useEffect(() => {
        setPosition(prev => {
            const next = clampPosition(prev.x, prev.y, height)
            if (next.x === prev.x && next.y === prev.y) return prev
            return next
        })
    }, [clampPosition, height])

    React.useEffect(() => {
        if (typeof window === "undefined") return

        const handleResize = () => {
            const nextHeight = clampHeight(panelRef.current?.offsetHeight ?? height, position.y)

            setHeight(nextHeight)
            setPosition(prev => clampPosition(prev.x, prev.y, nextHeight))
        }

        window.addEventListener("resize", handleResize)

        return () => {
            window.removeEventListener("resize", handleResize)
        }
    }, [clampHeight, clampPosition, height, position.y])

    const onPointerDown = React.useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (event.button !== 0) return

        event.currentTarget.setPointerCapture(event.pointerId)
        drag.current = {
            x: event.clientX,
            y: event.clientY,
            left: position.x,
            top: position.y,
        }
    }, [position])

    const onPointerMove = React.useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (!drag.current) return

        const nextX = drag.current.left + event.clientX - drag.current.x
        const nextY = drag.current.top + event.clientY - drag.current.y

        setPosition(clampPosition(nextX, nextY, panelRef.current?.offsetHeight ?? height))
    }, [clampPosition, height])

    const onResizePointerDown = React.useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (event.button !== 0) return

        event.currentTarget.setPointerCapture(event.pointerId)
        resize.current = {
            y: event.clientY,
            height: panelRef.current?.offsetHeight ?? height,
        }
    }, [height])

    const onResizePointerMove = React.useCallback((event: React.PointerEvent<HTMLDivElement>) => {
        if (!resize.current) return

        const nextHeight = clampHeight(resize.current.height + event.clientY - resize.current.y, position.y)
        setHeight(nextHeight)
        setPosition(prev => clampPosition(prev.x, prev.y, nextHeight))
    }, [clampHeight, clampPosition, position.y])

    const onPointerUp = React.useCallback(() => {
        drag.current = null
        resize.current = null
    }, [])

    const onListScroll = React.useCallback((event: React.UIEvent<HTMLDivElement>) => {
        shouldAutoScrollRef.current = isElNearBottom(event.currentTarget)
    }, [])

    const copyLogs = React.useCallback(() => {
        const text = logs.map(log => {
            const time = new Date(log.at).toISOString()
            return `[${time}] [${log.level}] ${log.message}`
        }).join("\n")

        navigator.clipboard?.writeText(text)
    }, [logs])

    return (
        <div
            ref={panelRef}
            className="fixed z-[80] flex w-[min(680px,calc(100vw_-_24px))] min-h-[280px] flex-col overflow-hidden rounded-2xl border border-[--border] bg-[--background] shadow-2xl"
            style={{
                left: position.x,
                top: position.y,
                height,
                maxHeight: `calc(100vh - ${position.y + windowPadding}px)`,
            }}
        >
            <div
                className="flex cursor-move select-none items-center justify-between gap-3 border-b border-[--border] px-3 py-2"
                onPointerDown={onPointerDown}
                onPointerMove={onPointerMove}
                onPointerUp={onPointerUp}
                onPointerCancel={onPointerUp}
            >
                <div className="min-w-0">
                    <p className="truncate text-sm font-semibold">{extensionName || extensionId}</p>
                    <p className="truncate font-mono text-xs text-[--muted]">{extensionId}</p>
                </div>
                <div className="flex items-center gap-1" onPointerDown={event => event.stopPropagation()}>
                    <Tooltip
                        trigger={<div>
                            <IconButton
                                intent="gray-basic"
                                size="sm"
                                icon={<LuClipboard className="size-4" />}
                                className="rounded-full"
                                onClick={copyLogs}
                            />
                        </div>}
                    >Copy</Tooltip>
                    <Tooltip
                        trigger={<div>
                            <IconButton
                                intent="gray-basic"
                                size="sm"
                                icon={<LuTrash2 className="size-4" />}
                                className="rounded-full"
                                onClick={onClear}
                            />
                        </div>}
                    >Clear</Tooltip>
                    <Tooltip
                        trigger={<div>
                            <IconButton
                                intent="gray-basic"
                                size="sm"
                                icon={<LuX className="size-4" />}
                                className="rounded-full"
                                onClick={onClose}
                            />
                        </div>}
                    >Close</Tooltip>
                </div>
            </div>
            <div className="flex flex-col gap-2 border-b border-[--border] px-3 py-2 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex min-w-0 flex-1 items-center gap-2">
                    <TextInput
                        value={search}
                        onChange={event => setSearch(event.target.value)}
                        placeholder="Search logs..."
                        size="sm"
                        leftIcon={<LuSearch className="size-3.5" />}
                        fieldClass="w-full max-w-xs"
                        className="text-xs"
                    />
                </div>
                <div className="flex min-w-0 items-center gap-1 overflow-x-auto">
                    {dbgLvls.map(item => (
                        <button
                            key={item}
                            className={cn(
                                "rounded-md border px-2 py-1 text-xs capitalize transition-colors",
                                level === item
                                    ? "bg-white text-black"
                                    : "border-[--border] text-[--muted] hover:text-[--foreground]",
                            )}
                            onClick={() => setLevel(item)}
                        >
                            {item}
                        </button>
                    ))}
                </div>
                <p className="flex-none font-mono text-xs text-right text-[--muted] min-w-10">{visibleLogs.length}/{logs.length}</p>
            </div>
            <div ref={listRef} onScroll={onListScroll} className="min-h-0 flex-1 space-y-2 overflow-y-auto p-3">
                {!visibleLogs.length && <div className="flex h-full items-center justify-center text-sm text-[--muted]">
                    No logs
                </div>}
                {visibleLogs.map(log => (
                    <div key={log.id} className="rounded-md border border-[--border] bg-[--paper] p-2 hover:bg-[--subtle]">
                        <div className="mb-1 flex min-w-0 items-center gap-2">
                            <span className="font-mono text-xs text-[--muted]">{new Date(log.at).toLocaleTimeString()}</span>
                            <span
                                className={cn(
                                    "rounded border px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
                                    levelClass(log.level),
                                )}
                            >
                                {log.level}
                            </span>
                        </div>
                        <p className="whitespace-pre-wrap break-words font-mono text-xs leading-relaxed text-[--foreground]">{log.message}</p>
                        {!!log.values?.length &&
                            <pre className="mt-2 max-h-52 overflow-auto rounded-md bg-black/30 p-2 text-xs leading-relaxed text-[--muted]">
                                {formatDebugValue(log.values.length === 1 ? log.values[0] : log.values)}
                            </pre>}
                    </div>
                ))}
            </div>
            <div
                className="flex h-3 flex-none cursor-ns-resize items-center justify-center border-t border-[--border] bg-[--subtle] touch-none"
                onPointerDown={onResizePointerDown}
                onPointerMove={onResizePointerMove}
                onPointerUp={onPointerUp}
                onPointerCancel={onPointerUp}
            >
                <div className="h-1 w-12 rounded-full bg-[--border]" />
            </div>
        </div>
    )
}

