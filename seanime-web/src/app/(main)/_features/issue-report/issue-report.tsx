import { getServerBaseUrl } from "@/api/client/server-url"
import { Report_ClickLog, Report_ConsoleLog, Report_NetworkLog, Report_ReactQueryLog } from "@/api/generated/types"
import { useSaveIssueReport } from "@/api/hooks/report.hooks"
import { useServerHMACAuth } from "@/app/(main)/_hooks/use-server-status"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { BiX } from "react-icons/bi"
import { PiRecordFill, PiStopCircleFill } from "react-icons/pi"
import { VscDebugAlt } from "react-icons/vsc"
import { toast } from "sonner"

export const __issueReport_overlayOpenAtom = atom<boolean>(false)
export const __issueReport_recordingAtom = atom<boolean>(false)
export const __issueReport_streamAtom = atom<any>()

export const __issueReport_clickLogsAtom = atom<Report_ClickLog[]>([])

export const __issueReport_consoleAtom = atom<Report_ConsoleLog[]>([])

export const __issueReport_networkAtom = atom<Report_NetworkLog[]>([])

export const __issueReport_reactQueryAtom = atom<Report_ReactQueryLog[]>([])

export function IssueReport() {
    const router = useRouter()
    const pathname = usePathname()
    const queryClient = useQueryClient()

    const [open, setOpen] = useAtom(__issueReport_overlayOpenAtom)
    const [isRecording, setRecording] = useAtom(__issueReport_recordingAtom)
    const [consoleLogs, setConsoleLogs] = useAtom(__issueReport_consoleAtom)
    const [clickLogs, setClickLogs] = useAtom(__issueReport_clickLogsAtom)
    const [networkLogs, setNetworkLogs] = useAtom(__issueReport_networkAtom)
    const [reactQueryLogs, setReactQueryLogs] = useAtom(__issueReport_reactQueryAtom)
    const [recordLocalFiles, setRecordLocalFiles] = React.useState(false)

    const { mutate, isPending } = useSaveIssueReport()

    React.useEffect(() => {
        if (!open) {
            setRecording(false)
        }
    }, [open])

    React.useEffect(() => {
        if (!isRecording) {
            setConsoleLogs([])
            setClickLogs([])
            setNetworkLogs([])
            setReactQueryLogs([])
        }
    }, [isRecording])

    React.useEffect(() => {
        if (!isRecording) return

        const captureClick = (e: MouseEvent) => {
            const element = e.target as HTMLElement
            setClickLogs(prev => [...prev, {
                pageUrl: window.location.href.replace(window.location.host, "{client}"),
                timestamp: new Date().toISOString(),
                element: element.tagName,
                className: JSON.stringify(element.className?.length && element.className.length > 50
                    ? element.className.slice(0, 100) + "..."
                    : element.className),
                text: element.innerText,
            }])
        }

        window.addEventListener("click", captureClick)

        return () => {
            window.removeEventListener("click", captureClick)
        }
    }, [isRecording])

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
                setConsoleLogs(prev => [...prev, {
                    type,
                    pageUrl: window.location.href.replace(window.location.host, "{client}"),
                    content: args.map(arg =>
                        typeof arg === "object" ? JSON.stringify(arg) : String(arg),
                    ).join(" "),
                    timestamp: new Date().toISOString(),
                }])
            }
            catch (e) {
                // console.error("Error capturing console logs", e)
            }
        }

        console.log = logInterceptor("log")
        console.error = logInterceptor("error")
        console.warn = logInterceptor("warn")

        return () => {
            Object.assign(console, originalConsole)
        }
    }, [isRecording])

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

            // Don't log the save issue report mutation to prevent feedback loop
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

    React.useEffect(() => {
        if (!isRecording) return

        const originalXhrOpen = XMLHttpRequest.prototype.open
        const originalXhrSend = XMLHttpRequest.prototype.send

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

            const _url = new URL(url)
            // remove host and port
            _url.host = "{server}"
            _url.port = ""

            this.addEventListener("load", () => {
                const duration = Date.now() - startTime
                setNetworkLogs(prev => [...prev, {
                    type: "xhr",
                    method,
                    url: _url.href,
                    pageUrl: window.location.href.replace(window.location.host, "{client}"),
                    status: this.status,
                    duration,
                    dataPreview: this.responseText.slice(0, 200),
                    timestamp: new Date().toISOString(),
                    body: JSON.stringify(body),
                }])
            })

            // @ts-ignore
            originalXhrSend.apply(this, arguments)
        }

        return () => {
            XMLHttpRequest.prototype.open = originalXhrOpen
            XMLHttpRequest.prototype.send = originalXhrSend
        }
    }, [isRecording])

    function handleStartRecording() {
        setRecording(true)
    }

    const { getHMACTokenQueryParam } = useServerHMACAuth()

    async function handleStopRecording() {
        const logsToSave = {
            clickLogs,
            consoleLogs,
            networkLogs,
            reactQueryLogs,
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
                        "p-4 bg-gray-900 border text-white rounded-xl",
                        "transition-colors duration-300",
                        isRecording && "p-0 border-transparent bg-transparent",
                    )}
                >
                    {!isRecording ? <div className="space-y-2">
                        <div className="flex items-center justify-center gap-4">
                            <VscDebugAlt className="text-2xl text-[--brand]" />
                            <div className="">
                                <p>
                                    Issue recorder
                                </p>
                                <p className="text-[--muted] text-sm text-center">
                                    Record your issue and generate a report
                                </p>
                                <div className="pt-2">
                                    <Checkbox
                                        label="Anime library issue"
                                        value={recordLocalFiles}
                                        onValueChange={v => typeof v === "boolean" && setRecordLocalFiles(v)}
                                        size="sm"
                                    />
                                </div>
                            </div>
                            <div className="flex items-center gap-0">
                                <Tooltip
                                    trigger={<IconButton
                                        intent="gray-basic"
                                        icon={<PiRecordFill className="text-red-500" />}
                                        onClick={handleStartRecording}
                                    />}
                                >
                                    Start recording
                                </Tooltip>
                                <Tooltip
                                    trigger={<IconButton
                                        intent="gray-basic"
                                        icon={<BiX />}
                                        onClick={() => setOpen(false)}
                                    />}
                                >
                                    Close
                                </Tooltip>
                            </div>
                        </div>
                    </div> : <div className="flex items-center justify-center gap-0">
                        <Tooltip
                            trigger={<IconButton
                                intent="alert"
                                icon={<PiStopCircleFill className="text-white animate-pulse" />}
                                onClick={handleStopRecording}
                            />}
                        >
                            Stop recording
                        </Tooltip>
                        <Tooltip
                            trigger={<IconButton
                                intent="white"
                                size="xs"
                                icon={<BiX />}
                                onClick={() => setRecording(false)}
                            />}
                        >
                            Cancel
                        </Tooltip>
                    </div>}
                </div>
            </div>}
        </>
    )
}
