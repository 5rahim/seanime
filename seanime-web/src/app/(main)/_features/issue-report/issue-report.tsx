import { getServerBaseUrl } from "@/api/client/server-url"
import { Report_ClickLog, Report_ConsoleLog, Report_NetworkLog, Report_ReactQueryLog } from "@/api/generated/types"
import { useSaveIssueReport } from "@/api/hooks/report.hooks"
import { WebSocketContext } from "@/app/(main)/_atoms/websocket.atoms"
import { useServerHMACAuth } from "@/app/(main)/_hooks/use-server-status"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { usePathname, useRouter } from "@/lib/navigation"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useCallback, useContext, useRef } from "react"
import { BiCamera, BiNote, BiX } from "react-icons/bi"
import { PiRecordFill, PiStopCircleFill } from "react-icons/pi"
import { VscDebugAlt } from "react-icons/vsc"
import { toast } from "sonner"

export const __issueReport_overlayOpenAtom = atom<boolean>(false)
export const __issueReport_recordingAtom = atom<boolean>(false)

export const __issueReport_clickLogsAtom = atom<Report_ClickLog[]>([])

export const __issueReport_consoleAtom = atom<Report_ConsoleLog[]>([])

export const __issueReport_networkAtom = atom<Report_NetworkLog[]>([])

export const __issueReport_reactQueryAtom = atom<Report_ReactQueryLog[]>([])

type NavigationLog = {
    from: string
    to: string
    timestamp: string
}

type ScreenshotEntry = {
    data: string // base64
    caption: string
    pageUrl: string
    timestamp: string
}

type WebSocketLogEntry = {
    direction: "incoming" | "outgoing"
    eventType: string
    payload: any
    timestamp: string
}

const __issueReport_navigationLogsAtom = atom<NavigationLog[]>([])
const __issueReport_screenshotsAtom = atom<ScreenshotEntry[]>([])

export function IssueReport() {
    const router = useRouter()
    const pathname = usePathname()
    const queryClient = useQueryClient()
    const socket = useContext(WebSocketContext)

    const [open, setOpen] = useAtom(__issueReport_overlayOpenAtom)
    const [isRecording, setRecording] = useAtom(__issueReport_recordingAtom)
    const [consoleLogs, setConsoleLogs] = useAtom(__issueReport_consoleAtom)
    const [clickLogs, setClickLogs] = useAtom(__issueReport_clickLogsAtom)
    const [networkLogs, setNetworkLogs] = useAtom(__issueReport_networkAtom)
    const [reactQueryLogs, setReactQueryLogs] = useAtom(__issueReport_reactQueryAtom)
    const [navigationLogs, setNavigationLogs] = useAtom(__issueReport_navigationLogsAtom)
    const [screenshots, setScreenshots] = useAtom(__issueReport_screenshotsAtom)
    const [recordLocalFiles, setRecordLocalFiles] = React.useState(false)
    const [description, setDescription] = React.useState("")
    const [showDescriptionInput, setShowDescriptionInput] = React.useState(false)
    const [recordingElapsed, setRecordingElapsed] = React.useState(0)

    // rrweb recording state
    const rrwebEventsRef = useRef<any[]>([])
    const rrwebStopFnRef = useRef<(() => void) | null>(null)
    const recordingStartTimeRef = useRef<number>(0)
    const recordingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

    // WebSocket event capture
    const wsEventsRef = useRef<WebSocketLogEntry[]>([])

    const { mutate, isPending } = useSaveIssueReport()

    const eventCount = consoleLogs.length + clickLogs.length + networkLogs.length + reactQueryLogs.length + navigationLogs.length
    const rrwebEventCount = rrwebEventsRef.current.length

    React.useEffect(() => {
        if (!open) {
            setRecording(false)
        }
    }, [open])

    // Reset all state when recording stops
    React.useEffect(() => {
        if (!isRecording) {
            setConsoleLogs([])
            setClickLogs([])
            setNetworkLogs([])
            setReactQueryLogs([])
            setNavigationLogs([])
            setScreenshots([])
            setDescription("")
            setShowDescriptionInput(false)
            setRecordingElapsed(0)

            // Stop rrweb recording
            if (rrwebStopFnRef.current) {
                rrwebStopFnRef.current()
                rrwebStopFnRef.current = null
            }

            // Clear recording timer
            if (recordingTimerRef.current) {
                clearInterval(recordingTimerRef.current)
                recordingTimerRef.current = null
            }

            rrwebEventsRef.current = []
            wsEventsRef.current = []
        }
    }, [isRecording])

    //
    // rrweb
    //
    const startRRWebRecording = useCallback(async () => {
        try {

            rrwebEventsRef.current = []
            const rrweb = await import("rrweb")

            const stopFn = rrweb.record({
                emit(event: any) {
                    rrwebEventsRef.current.push(event)
                },
                // take a full DOM snapshot every 200 events to allow
                // "checkout" for long sessions (only need last 200-400 events)
                checkoutEveryNth: 200,
                // mask sensitive inputs (passwords etc)
                maskAllInputs: false,
                maskInputOptions: { password: true },
                // block the issue reporter UI itself from being recorded
                blockSelector: ".issue-reporter-ui",
                inlineStylesheet: true,
                collectFonts: true,
                // reduce payload size
                slimDOMOptions: {
                    script: true,
                    comment: true,
                    headFavicon: true,
                    headWhitespace: true,
                    headMetaDescKeywords: true,
                    headMetaSocial: true,
                    headMetaRobots: true,
                    headMetaHttpEquiv: true,
                    headMetaAuthorship: true,
                    headMetaVerification: true,
                },
                // sampling config to keep event sizes reasonable
                sampling: {
                    mousemove: true,
                    mouseInteraction: true,
                    scroll: 250, // sample scroll events every 250ms
                    input: "last", // only capture last input event
                },
            })

            rrwebStopFnRef.current = stopFn ?? null
        }
        catch (err) {
            console.error("Failed to start rrweb recording:", err)
            toast.error("Failed to start DOM recording")
        }
    }, [])

    // websocket event capture
    React.useEffect(() => {
        if (!isRecording || !socket) return

        const handler = (event: MessageEvent) => {
            try {
                const parsed = JSON.parse(event.data) as { type?: string; payload?: any }
                // skip plugin events (too many of them) and keep server-side events
                if (parsed.type === "plugin") return

                const logEntry: WebSocketLogEntry = {
                    direction: "incoming",
                    eventType: parsed.type || "unknown",
                    payload: parsed.payload,
                    timestamp: new Date().toISOString(),
                }

                wsEventsRef.current.push(logEntry)

                // Add to rrweb timeline
                rrwebEventsRef.current.push({
                    type: 5, // Custom event
                    data: {
                        tag: "websocket",
                        payload: logEntry,
                    },
                    timestamp: Date.now(),
                })
            }
            catch {
                // non-json messages, skip
            }
        }

        socket.addEventListener("message", handler)

        return () => {
            socket.removeEventListener("message", handler)
        }
    }, [isRecording, socket])

    // Click logger
    React.useEffect(() => {
        if (!isRecording) return

        const captureClick = (e: MouseEvent) => {
            const element = e.target as HTMLElement
            // skip clicks within the issue reporter UI itself
            if (element.closest(".issue-reporter-ui")) return
            setClickLogs(prev => [...prev, {
                pageUrl: window.location.href.replace(window.location.host, "{client}"),
                timestamp: new Date().toISOString(),
                element: element.tagName,
                className: JSON.stringify(element.className?.length && element.className.length > 50
                    ? element.className.slice(0, 100) + "..."
                    : element.className),
                text: element.innerText?.slice(0, 150),
            }])
        }

        window.addEventListener("click", captureClick)

        return () => {
            window.removeEventListener("click", captureClick)
        }
    }, [isRecording])

    // Console logger
    React.useEffect(() => {
        if (!isRecording) return

        const originalConsole = {
            log: console.log,
            error: console.error,
            warn: console.warn,
        }

        const logInterceptor = (type: Report_ConsoleLog["type"]) => (...args: any[]) => {
            (originalConsole as any)[type](...args)
            try {
                const entry: Report_ConsoleLog = {
                    type,
                    pageUrl: window.location.href.replace(window.location.host, "{client}"),
                    content: args.map(arg =>
                        typeof arg === "object" ? JSON.stringify(arg) : String(arg),
                    ).join(" "),
                    timestamp: new Date().toISOString(),
                }
                // defer state update to avoid triggering setState during render
                queueMicrotask(() => {
                    setConsoleLogs(prev => [...prev, entry])
                })
            }
            catch (e) {
                // ignore capture errors
            }
        }

        console.log = logInterceptor("log")
        console.error = logInterceptor("error")
        console.warn = logInterceptor("warn")

        return () => {
            Object.assign(console, originalConsole)
        }
    }, [isRecording])

    // react query logger
    React.useEffect(() => {
        if (!isRecording) return

        const queryUnsubscribe = queryClient.getQueryCache().subscribe(listener => {
            if (listener.query.state.status === "pending") return
            setReactQueryLogs(prev => [...prev, {
                type: "query",
                pageUrl: window.location.href.replace(window.location.host, "{client}"),
                status: listener.query.state.status,
                hash: listener.query.queryHash,
                error: listener.query.state.error,
                timestamp: new Date().toISOString(),
                dataPreview: typeof listener.query.state.data === "object"
                    ? JSON.stringify(listener.query.state.data).slice(0, 200)
                    : "",
                dataType: typeof listener.query.state.data,
            }])
        })

        const mutationUnsubscribe = queryClient.getMutationCache().subscribe(listener => {
            if (!listener.mutation) return
            if (listener.mutation.state.status === "pending" || listener.mutation.state.status === "idle") return

            // don't log the save issue report mutation to prevent feedback loop
            const mutationKey = listener.mutation.options.mutationKey
            if (Array.isArray(mutationKey) && mutationKey.includes("REPORT-save-issue-report")) return

            setReactQueryLogs(prev => [...prev, {
                type: "mutation",
                pageUrl: window.location.href.replace(window.location.host, "{client}"),
                status: listener.mutation!.state.status,
                hash: JSON.stringify(listener.mutation!.options.mutationKey),
                error: listener.mutation!.state.error,
                timestamp: new Date().toISOString(),
                dataPreview: typeof listener.mutation!.state.data === "object" ? JSON.stringify(listener.mutation!.state.data)
                    .slice(0, 200) : "",
                dataType: typeof listener.mutation!.state.data,
            }])
        })

        return () => {
            queryUnsubscribe()
            mutationUnsubscribe()
        }
    }, [isRecording])

    // network logger
    React.useEffect(() => {
        if (!isRecording) return

        const originalXhrOpen = XMLHttpRequest.prototype.open
        const originalXhrSend = XMLHttpRequest.prototype.send
        const originalFetch = window.fetch

        const MAX_RESPONSE_SIZE = 500000 // 500KB

        const processNetworkLog = (
            method: string,
            url: string,
            status: number,
            duration: number,
            responseBody: string,
            requestBody: string,
            type: "xhr" | "fetch",
        ) => {
            let sanitizedUrl: string
            try {
                const urlStr = url
                // skip relative urls if they are not api calls
                if (urlStr.startsWith("/") && !urlStr.startsWith("//") && !urlStr.includes("/api/")) {
                    return
                }

                const _url = new URL(urlStr, window.location.origin)
                // normalize host
                if (_url.origin === window.location.origin) {
                    sanitizedUrl = _url.href.replace(window.location.host, "{client}")
                } else {
                    _url.host = "{server}"
                    _url.port = ""
                    sanitizedUrl = _url.href
                }
            }
            catch {
                return
            }

            const entry: Report_NetworkLog = {
                type,
                method,
                url: sanitizedUrl,
                pageUrl: window.location.href.replace(window.location.host, "{client}"),
                status,
                duration,
                dataPreview: responseBody,
                timestamp: new Date().toISOString(),
                body: requestBody,
            }

            setNetworkLogs(prev => [...prev, entry])

            // add to rrweb timeline
            rrwebEventsRef.current.push({
                type: 5,
                data: {
                    tag: "network",
                    payload: entry,
                },
                timestamp: Date.now(),
            })
        }

        //
        // XHR Interception
        //

        XMLHttpRequest.prototype.open = function (method, url) {
            // @ts-ignore
            this._url = url
            // @ts-ignore
            this._method = method
            // @ts-ignore
            originalXhrOpen.apply(this, arguments)
        }

        XMLHttpRequest.prototype.send = function (body) {
            const startTime = Date.now()
            // @ts-ignore
            const url = this._url
            // @ts-ignore
            const method = this._method

            this.addEventListener("load", () => {
                const duration = Date.now() - startTime
                let responseBody = ""
                try {
                    const contentType = this.getResponseHeader("content-type")
                    if (contentType && (contentType.includes("application/json") || contentType.includes("text/"))) {
                        if (this.responseText && this.responseText.length < MAX_RESPONSE_SIZE) {
                            responseBody = this.responseText
                        } else {
                            responseBody = "<response too large>"
                        }
                    }
                }
                catch {
                }

                processNetworkLog(
                    method,
                    typeof url === "string" ? url : String(url),
                    this.status,
                    duration,
                    responseBody,
                    body ? JSON.stringify(body) : "",
                    "xhr",
                )
            })

            // @ts-ignore
            originalXhrSend.apply(this, arguments)
        }

        //
        // fetch interception
        //

        window.fetch = async (...args) => {
            const startTime = Date.now()
            const [resource, config] = args
            const url = typeof resource === "string" ? resource : resource instanceof URL ? resource.href : resource.url
            const method = config?.method || "GET"

            try {
                const response = await originalFetch(...args)
                const clone = response.clone()

                // process response asynchronously
                clone.text().then(text => {
                    const duration = Date.now() - startTime
                    let responseBody = ""
                    const contentType = clone.headers.get("content-type")

                    if (contentType && (contentType.includes("application/json") || contentType.includes("text/"))) {
                        if (text.length < MAX_RESPONSE_SIZE) {
                            responseBody = text
                        } else {
                            responseBody = "<response too large>"
                        }
                    } else if (!contentType) {
                        if (text.length < MAX_RESPONSE_SIZE) {
                            responseBody = text
                        }
                    }

                    processNetworkLog(
                        method,
                        url,
                        response.status,
                        duration,
                        responseBody,
                        config?.body ? String(config.body) : "",
                        "fetch",
                    )
                }).catch(() => { })

                return response
            }
            catch (error) {
                // could log failed requests
                throw error
            }
        }

        return () => {
            XMLHttpRequest.prototype.open = originalXhrOpen
            XMLHttpRequest.prototype.send = originalXhrSend
            window.fetch = originalFetch
        }
    }, [isRecording])

    // page navigation tracker
    const prevPathnameRef = React.useRef(pathname)
    React.useEffect(() => {
        if (!isRecording) return
        if (prevPathnameRef.current !== pathname) {
            const entry: NavigationLog = {
                from: prevPathnameRef.current.replace(window.location.host, "{client}"),
                to: pathname.replace(window.location.host, "{client}"),
                timestamp: new Date().toISOString(),
            }
            // defer to avoid setState during render
            queueMicrotask(() => {
                setNavigationLogs(prev => [...prev, entry])
            })
            prevPathnameRef.current = pathname
        }
    }, [pathname, isRecording])

    function handleStartRecording() {
        recordingStartTimeRef.current = Date.now()
        setRecordingElapsed(0)

        // start elapsed timer
        recordingTimerRef.current = setInterval(() => {
            setRecordingElapsed(Math.floor((Date.now() - recordingStartTimeRef.current) / 1000))
        }, 1000)

        setRecording(true)

        // start rrweb recording
        startRRWebRecording()
    }

    function handleTakeScreenshot() {
        // use html2canvas-like approach via canvas capture
        const canvas = document.createElement("canvas")
        const ctx = canvas.getContext("2d")
        if (!ctx) {
            toast.error("Unable to capture screenshot")
            return
        }

        // capture via DOM serialization
        const input = document.createElement("input")
        input.type = "file"
        input.accept = "image/*"
        input.onchange = (e) => {
            const file = (e.target as HTMLInputElement).files?.[0]
            if (!file) return
            const reader = new FileReader()
            reader.onload = (ev) => {
                const base64 = ev.target?.result as string
                // limit size to ~500KB for the report
                if (base64.length > 700000) {
                    // compress by drawing to canvas
                    const img = new Image()
                    img.onload = () => {
                        const c = document.createElement("canvas")
                        const scale = Math.min(1, 1200 / Math.max(img.width, img.height))
                        c.width = img.width * scale
                        c.height = img.height * scale
                        const cx = c.getContext("2d")!
                        cx.drawImage(img, 0, 0, c.width, c.height)
                        const compressed = c.toDataURL("image/jpeg", 0.7)
                        addScreenshot(compressed)
                    }
                    img.src = base64
                } else {
                    addScreenshot(base64)
                }
            }
            reader.readAsDataURL(file)
        }
        input.click()
    }

    function addScreenshot(data: string) {
        const caption = prompt("Add a caption for this screenshot (optional):") || ""
        setScreenshots(prev => [...prev, {
            data,
            caption,
            pageUrl: window.location.href.replace(window.location.host, "{client}"),
            timestamp: new Date().toISOString(),
        }])
        toast.success("Screenshot added to report")
    }

    const { getHMACTokenQueryParam } = useServerHMACAuth()

    async function handleStopRecording() {
        // stop rrweb and capture final events
        if (rrwebStopFnRef.current) {
            rrwebStopFnRef.current()
            rrwebStopFnRef.current = null
        }

        const recordingDurationMs = Date.now() - recordingStartTimeRef.current

        const logsToSave = {
            description,
            clickLogs,
            consoleLogs,
            networkLogs,
            reactQueryLogs,
            navigationLogs,
            screenshots,
            websocketLogs: wsEventsRef.current,
            rrwebEvents: rrwebEventsRef.current,
            viewportWidth: window.innerWidth,
            viewportHeight: window.innerHeight,
            recordingDurationMs,
        }

        setRecording(false)

        mutate({
            ...logsToSave,
            isAnimeLibraryIssue: recordLocalFiles,
        }, {
            onSuccess: async () => {
                toast.success("Issue report saved successfully")

                setTimeout(async () => {
                    try {
                        const endpoint = "/api/v1/report/issue/download"
                        const tokenQuery = await getHMACTokenQueryParam(endpoint)
                        openTab(`${getServerBaseUrl()}${endpoint}${tokenQuery}`)
                    }
                    catch (error) {
                        toast.error("Failed to generate download token")
                    }
                }, 1000)
            },
        })
    }

    // format elapsed time as mm:ss
    const formatElapsed = (seconds: number) => {
        const m = Math.floor(seconds / 60).toString().padStart(2, "0")
        const s = (seconds % 60).toString().padStart(2, "0")
        return `${m}:${s}`
    }


    return (
        <>
            {open && <div
                className={cn(
                    "issue-reporter-ui",
                    "fixed z-[100] bottom-8 w-fit left-20 h-fit flex",
                    "transition-all duration-300 select-none",
                    !isRecording && "hover:translate-y-[-2px]",
                    isRecording && "justify-end",
                )}
            >
                <div
                    className={cn(
                        "rounded-xl border shadow-2xl shadow-black/50 backdrop-blur-sm",
                        "transition-colors duration-300",
                        isRecording
                            ? "p-3 bg-gray-950/95 border-red-900/50"
                            : "p-4 bg-gray-900/95 border-[--border] text-white",
                    )}
                >
                    {!isRecording ? <div className="space-y-3 min-w-[280px]">
                        <div className="flex items-center gap-3">
                            <div className="p-2 rounded-lg bg-brand-900/30">
                                <VscDebugAlt className="text-xl text-[--brand]" />
                            </div>
                            <div>
                                <p className="font-semibold text-sm text-gray-100">Issue Recorder</p>
                            </div>
                            <div className="ml-auto">
                                <IconButton
                                    intent="gray-basic"
                                    size="xs"
                                    icon={<BiX />}
                                    onClick={() => setOpen(false)}
                                />
                            </div>
                        </div>
                        <div className="border-t border-[--border] pt-2 space-y-2">
                            <Checkbox
                                label="Include library scanner logs"
                                value={recordLocalFiles}
                                onValueChange={v => typeof v === "boolean" && setRecordLocalFiles(v)}
                                size="md"
                            />
                        </div>
                        <div className="flex justify-end">
                            <button
                                onClick={handleStartRecording}
                                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 transition-colors text-sm font-medium text-white"
                            >
                                <PiRecordFill className="text-white animate-pulse" />
                                Start Recording
                            </button>
                        </div>
                    </div> : <div className="space-y-3 min-w-[320px]">

                        <div className="flex items-center gap-3">
                            <div className="relative">
                                <div className="w-3 h-3 rounded-full bg-red-500 animate-pulse" />
                                <div className="absolute inset-0 w-3 h-3 rounded-full bg-red-500 animate-ping opacity-50" />
                            </div>
                            <span className="text-sm font-semibold text-red-400">Recording</span>
                            <span className="text-xs text-gray-400 tabular-nums font-mono bg-gray-800 px-1.5 py-0.5 rounded">
                                {formatElapsed(recordingElapsed)}
                            </span>
                            <div className="ml-auto flex items-center gap-2">
                                <span className="text-xs text-gray-500 tabular-nums">{eventCount} events</span>
                                {screenshots.length > 0 && (
                                    <span className="text-xs text-gray-500">{screenshots.length} imgs</span>
                                )}
                            </div>
                        </div>

                        <div className="flex gap-1.5">
                            <button
                                onClick={handleTakeScreenshot}
                                className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium
                                        bg-gray-800 hover:bg-gray-700 text-gray-300 transition-colors border border-gray-700"
                            >
                                <BiCamera className="text-sm" />
                                Attach screenshot
                            </button>
                            <Tooltip
                                trigger={
                                    <button
                                        onClick={() => setShowDescriptionInput(!showDescriptionInput)}
                                        className={cn(
                                            "flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors border",
                                            description
                                                ? "bg-brand-900/30 border-brand-700/50 text-brand-300"
                                                : "bg-gray-800 hover:bg-gray-700 text-gray-300 border-gray-700",
                                        )}
                                    >
                                        <BiNote className="text-sm" />
                                        {description ? "Edit note" : "Add note"}
                                    </button>
                                }
                                className="z-[101]"
                            >
                                Add a description of what you're experiencing
                            </Tooltip>
                        </div>

                        {showDescriptionInput && (
                            <textarea
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Describe the issue you're experiencing..."
                                className="w-full px-3 py-2 rounded-lg bg-gray-900 border border-gray-700 text-sm text-gray-200
                                    placeholder-gray-500 resize-none focus:outline-none focus:border-brand-500 transition-colors"
                                rows={3}
                            />
                        )}

                        <div className="flex items-center gap-2 pt-1 border-t border-gray-800">
                            <button
                                onClick={handleStopRecording}
                                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500
                                    transition-colors text-sm font-medium text-white flex-1 justify-center"
                            >
                                <PiStopCircleFill className="text-lg" />
                                Stop & Save
                            </button>
                            <Tooltip
                                trigger={<IconButton
                                    intent="gray-basic"
                                    size="sm"
                                    icon={<BiX />}
                                    onClick={() => setRecording(false)}
                                />}
                                className="z-[101]"
                            >
                                Cancel recording
                            </Tooltip>
                        </div>
                    </div>}
                </div>
            </div>}
        </>
    )
}
