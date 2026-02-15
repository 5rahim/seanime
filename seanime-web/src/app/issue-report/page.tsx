import { Report_ClickLog, Report_ConsoleLog, Report_IssueReport, Report_NetworkLog } from "@/api/generated/types"

import { useDecompressIssueReport } from "@/api/hooks/report.hooks"
import { ScanLogViewer } from "@/app/scan-log-viewer/scan-log-viewer"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/shared/resizable"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { TextInput } from "@/components/ui/text-input"
import { format, parseISO } from "date-fns"
import React, { useEffect, useMemo, useRef, useState } from "react"
import {
    BiChevronDown,
    BiChevronRight,
    BiError,
    BiFile,
    BiImage,
    BiInfoCircle,
    BiNavigation,
    BiPlay,
    BiTrash,
    BiUpload,
    BiWifi,
} from "react-icons/bi"
import { FiMousePointer } from "react-icons/fi"
import { HiServerStack } from "react-icons/hi2"
import { LuBrain, LuNetwork, LuTerminal } from "react-icons/lu"
import { Virtuoso } from "react-virtuoso"
import { toast } from "sonner"

const DB_NAME = "seanime-issue-report-db"
const STORE_NAME = "reports"
const KEY = "latest_report"

const initDB = (): Promise<IDBDatabase> => {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, 1)
        request.onerror = () => reject(request.error)
        request.onsuccess = () => resolve(request.result)
        request.onupgradeneeded = (event) => {
            const db = (event.target as IDBOpenDBRequest).result
            if (!db.objectStoreNames.contains(STORE_NAME)) {
                db.createObjectStore(STORE_NAME)
            }
        }
    })
}

const saveReportToDB = async (report: ExtendedReport) => {
    try {
        const db = await initDB()
        return new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readwrite")
            const store = tx.objectStore(STORE_NAME)
            const request = store.put(report, KEY)
            request.onsuccess = () => resolve()
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to save report to DB:", error)
        toast.error("Failed to save report to browser storage")
    }
}

const getReportFromDB = async (): Promise<ExtendedReport | null> => {
    try {
        const db = await initDB()
        return new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readonly")
            const store = tx.objectStore(STORE_NAME)
            const request = store.get(KEY)
            request.onsuccess = () => resolve(request.result || null)
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to get report from DB:", error)
        return null
    }
}

const clearReportFromDB = async () => {
    try {
        const db = await initDB()
        return new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readwrite")
            const store = tx.objectStore(STORE_NAME)
            const request = store.delete(KEY)
            request.onsuccess = () => resolve()
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to clear report from DB:", error)
    }
}

// types
interface NavigationLog {
    from: string
    to: string
    timestamp: string
}

interface Screenshot {
    data: string
    caption?: string
    pageUrl: string
    timestamp: string
}

interface WebSocketLog {
    direction: "incoming" | "outgoing"
    eventType: string
    payload: any
    timestamp: string
}

interface ExtendedReport extends Report_IssueReport {
    description?: string
    navigationLogs?: NavigationLog[]
    screenshots?: Screenshot[]
    websocketLogs?: WebSocketLog[]
    rrwebEvents?: any[]
    viewportWidth?: number
    viewportHeight?: number
    recordingDurationMs?: number
}


// events
type EventType = "click" | "network" | "console" | "query" | "navigation" | "screenshot" | "server" | "websocket"

// unified event for timeline
interface UnifiedEvent {
    id: number
    type: EventType
    timestamp: Date
    pageUrl?: string
    summary: string
    level?: "error" | "warn" | "info" | "debug"
    raw: any
}

function buildUnifiedTimeline(report: ExtendedReport, includeServerLogs: boolean = true): UnifiedEvent[] {
    const events: UnifiedEvent[] = []
    let idx = 0

    for (const log of report.clickLogs || []) {
        events.push({
            id: idx++,
            type: "click",
            timestamp: parseISO(log.timestamp!),
            pageUrl: log.pageUrl,
            summary: `Click: <${log.element}>${log.text ? ` "${log.text.slice(0, 60)}"` : ""}`,
            level: "info",
            raw: log,
        })
    }

    for (const log of report.networkLogs || []) {
        // extract filename or endpoint from URL for better readability
        let urlPart = log.url
        try {
            const urlObj = new URL(log.url, "http://localhost")
            urlPart = urlObj.pathname.split("/").pop() || urlObj.pathname
            if (urlPart.length > 40) urlPart = urlPart.slice(0, 40) + "…"
        }
        catch {
            urlPart = log.url.slice(0, 50)
        }

        events.push({
            id: idx++,
            type: "network",
            timestamp: parseISO(log.timestamp!),
            pageUrl: log.pageUrl,
            summary: `${log.method} ${urlPart} → ${log.status} (${log.duration}ms)`,
            level: log.status >= 400 ? "error" : "info",
            raw: log,
        })
    }

    for (const log of report.consoleLogs || []) {
        events.push({
            id: idx++,
            type: "console",
            timestamp: parseISO(log.timestamp!),
            pageUrl: log.pageUrl,
            summary: log.content?.slice(0, 120) || "—",
            level: log.type === "error" ? "error" : log.type === "warn" ? "warn" : "debug",
            raw: log,
        })
    }

    for (const log of report.reactQueryLogs || []) {
        events.push({
            id: idx++,
            type: "query",
            timestamp: parseISO(log.timestamp!),
            pageUrl: log.pageUrl,
            summary: `[${log.type}] ${log.hash} → ${log.status}`,
            level: log.status === "error" ? "error" : "info",
            raw: log,
        })
    }

    for (const log of (report as ExtendedReport).navigationLogs || []) {
        events.push({
            id: idx++,
            type: "navigation",
            timestamp: parseISO(log.timestamp),
            summary: `${log.from} → ${log.to}`,
            level: "info",
            raw: log,
        })
    }

    for (const ss of (report as ExtendedReport).screenshots || []) {
        events.push({
            id: idx++,
            type: "screenshot",
            timestamp: parseISO(ss.timestamp),
            pageUrl: ss.pageUrl,
            summary: ss.caption || "Screenshot",
            level: "info",
            raw: ss,
        })
    }

    for (const ws of (report as ExtendedReport).websocketLogs || []) {
        events.push({
            id: idx++,
            type: "websocket",
            timestamp: parseISO(ws.timestamp),
            summary: `[${ws.direction}] ${ws.eventType}`,
            level: ws.eventType.includes("error") ? "error" : "info",
            raw: ws,
        })
    }

    // Parse and add server logs if enabled
    if (includeServerLogs && report.serverLogs) {
        const serverLines = report.serverLogs.split("\n").filter(Boolean)
        for (const line of serverLines) {
            // Try to extract timestamp from server log format (e.g., "2024-01-15T10:30:45..." or similar)
            const timestampMatch = line.match(/^(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2})/)
            let timestamp: Date
            if (timestampMatch) {
                try {
                    timestamp = parseISO(timestampMatch[1].replace(" ", "T"))
                }
                catch {
                    timestamp = new Date() // fallback
                }
            } else {
                timestamp = new Date() // fallback for lines without timestamp
            }

            const isError = line.includes("|ERR|")
            const isWarn = line.includes("|WRN|")

            events.push({
                id: idx++,
                type: "server",
                timestamp,
                summary: line.slice(0, 150),
                level: isError ? "error" : isWarn ? "warn" : "debug",
                raw: { line },
            })
        }
    }

    events.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime())
    return events
}

// stats
interface ReportStats {
    totalEvents: number
    clickCount: number
    networkCount: number
    consoleCount: number
    queryCount: number
    navigationCount: number
    screenshotCount: number
    websocketCount: number
    rrwebEventCount: number
    errorCount: number
    warningCount: number
    duration: string | null
    recordingDuration: string | null
    startTime: Date | null
    endTime: Date | null
    networkErrors: number
    hasReplay: boolean
}

function extractReportStats(report: ExtendedReport, events: UnifiedEvent[]): ReportStats {
    const stats: ReportStats = {
        totalEvents: events.length,
        clickCount: (report.clickLogs || []).length,
        networkCount: (report.networkLogs || []).length,
        consoleCount: (report.consoleLogs || []).length,
        queryCount: (report.reactQueryLogs || []).length,
        navigationCount: (report.navigationLogs || []).length,
        screenshotCount: (report.screenshots || []).length,
        websocketCount: (report.websocketLogs || []).length,
        rrwebEventCount: (report.rrwebEvents || []).length,
        errorCount: 0,
        warningCount: 0,
        duration: null,
        recordingDuration: null,
        startTime: null,
        endTime: null,
        networkErrors: 0,
        hasReplay: (report.rrwebEvents || []).length > 0,
    }

    for (const e of events) {
        if (e.level === "error") stats.errorCount++
        if (e.level === "warn") stats.warningCount++
    }

    stats.networkErrors = (report.networkLogs || []).filter(l => l.status >= 400).length

    if (report.recordingDurationMs) {
        const secs = Math.floor(report.recordingDurationMs / 1000)
        if (secs >= 60) {
            stats.recordingDuration = `${Math.floor(secs / 60)}m ${secs % 60}s`
        } else {
            stats.recordingDuration = `${secs}s`
        }
    }

    if (events.length > 0) {
        stats.startTime = events[0].timestamp
        stats.endTime = events[events.length - 1].timestamp
        const diffMs = stats.endTime.getTime() - stats.startTime.getTime()
        const secs = Math.floor(diffMs / 1000)
        if (secs >= 60) {
            stats.duration = `${Math.floor(secs / 60)}m ${secs % 60}s`
        } else {
            stats.duration = `${secs}s`
        }
    }

    return stats
}


type Phase = "overview" | "replay" | "timeline" | "network" | "console" | "clicks" | "websocket" | "server" | "screenshots" | "scanlogs"

export default function Page() {
    const [report, setReport] = useState<ExtendedReport | null>(null)
    const [isDragging, setIsDragging] = useState(false)
    const [isLoading, setIsLoading] = useState(true)
    const fileInputRef = useRef<HTMLInputElement>(null)
    const dragCounter = useRef(0)

    const { mutate: decompressReport, isPending: isDecompressing } = useDecompressIssueReport()

    // Load saved report on mount
    useEffect(() => {
        getReportFromDB().then((savedReport) => {
            if (savedReport) {
                setReport(savedReport)
                toast.success("Restored previous issue report")
            }
            setIsLoading(false)
        })
    }, [])

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (file) loadFile(file)
    }

    const loadFile = (file: File) => {
        if (file.type === "application/zip" || file.name.toLowerCase().endsWith(".zip")) {
            const formData = new FormData()
            formData.append("file", file)

            toast.info("Decompressing report...")

            decompressReport(formData, {
                onSuccess: (data) => {
                    setReport(data as ExtendedReport)
                    saveReportToDB(data as ExtendedReport)
                    toast.success("Report loaded")
                },
                onError: (error) => {
                    console.error(error)
                    toast.error("Failed to decompress report")
                },
            })
            return
        }

        const reader = new FileReader()
        reader.onload = (e) => {
            try {
                const content = e.target?.result as string
                const parsed = JSON.parse(content) as ExtendedReport
                setReport(parsed)
                saveReportToDB(parsed)
                toast.success("Report loaded")
            }
            catch {
                toast.error("Failed to parse report")
            }
        }
        reader.readAsText(file)
    }

    const handleDragEnter = (e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        dragCounter.current++
        setIsDragging(true)
    }
    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        dragCounter.current--
        if (dragCounter.current === 0) setIsDragging(false)
    }
    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
    }
    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        setIsDragging(false)
        dragCounter.current = 0
        const file = e.dataTransfer.files?.[0]
        if (file) loadFile(file)
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen text-[--muted]">
                <p>Loading saved report...</p>
            </div>
        )
    }

    return (
        <div
            className="container mx-auto p-4 min-h-screen relative"
            onDragEnter={handleDragEnter}
            onDragLeave={handleDragLeave}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
        >
            {isDragging && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/80 backdrop-blur-sm">
                    <div className="flex flex-col items-center gap-3 p-8 border-2 border-dashed border-indigo-500 rounded-xl bg-gray-900/50">
                        <BiUpload className="text-4xl text-indigo-400" />
                        <p className="text-lg font-medium text-indigo-300">Drop report file</p>
                    </div>
                </div>
            )}

            <div className="mb-4">
                <div className="flex items-center gap-4 justify-between">
                    <div className="flex items-center gap-4">
                        <h1 className="text-xl font-bold text-gray-200 tracking-tight">Issue Report</h1>
                        <label
                            className="flex items-center gap-2 px-3 py-1.5 bg-gray-800 border border-[--border] rounded-md cursor-pointer hover:bg-gray-700 transition-colors text-sm text-gray-300"
                        >
                            <BiUpload />
                            <span>{report ? "Load another file" : "Load report file"}</span>
                            <input
                                type="file"
                                ref={fileInputRef}
                                onChange={handleFileChange}
                                accept=".json,.zip"
                                className="hidden"
                            />
                        </label>
                    </div>

                    {report && (
                        <button
                            onClick={async () => {
                                await clearReportFromDB()
                                setReport(null)
                                toast.success("Cleared report")
                            }}
                            className="flex items-center gap-2 px-3 py-1.5 text-sm text-red-400 hover:text-red-300 hover:bg-red-950/30 rounded-md transition-colors"
                        >
                            <BiTrash />
                            Clear report
                        </button>
                    )}
                </div>
            </div>

            {report ? (
                <ReportViewer report={report} />
            ) : (
                <div className="flex items-center justify-center h-[40vh] text-[--muted]">
                    <p className="text-lg">Load an issue report JSON or ZIP file</p>
                </div>
            )}
        </div>
    )
}

function ReportViewer({ report }: { report: ExtendedReport }) {
    const [activePhase, setActivePhase] = useState<Phase>("overview")
    const [searchQuery, setSearchQuery] = useState("")
    const [includeServerLogs, setIncludeServerLogs] = useState(true)

    const events = useMemo(() => buildUnifiedTimeline(report, includeServerLogs), [report, includeServerLogs])
    const stats = useMemo(() => extractReportStats(report, events), [report, events])

    const hasScanLogs = !!report.scanLogs && report.scanLogs.length > 0
    const hasScreenshots = !!report.screenshots && report.screenshots.length > 0
    const hasReplay = stats.hasReplay
    const hasWsLogs = !!report.websocketLogs && report.websocketLogs.length > 0

    const tabs: { key: Phase; label: string; icon: React.ComponentType<any>; show?: boolean; accent?: boolean }[] = [
        { key: "overview", label: "Overview", icon: BiInfoCircle },
        { key: "replay", label: "Session Replay", icon: BiPlay, show: hasReplay, accent: true },
        { key: "timeline", label: "Timeline", icon: BiNavigation },
        { key: "network", label: `Network (${stats.networkCount})`, icon: LuNetwork as any },
        { key: "console", label: `Console (${stats.consoleCount})`, icon: LuTerminal as any },
        { key: "clicks", label: `Clicks (${stats.clickCount})`, icon: FiMousePointer as any },
        { key: "websocket", label: `WebSocket (${stats.websocketCount})`, icon: BiWifi, show: hasWsLogs },
        { key: "server", label: "Server Logs", icon: HiServerStack as any },
        { key: "screenshots", label: `Screenshots (${stats.screenshotCount})`, icon: BiImage, show: hasScreenshots },
        { key: "scanlogs", label: "Scan Logs", icon: BiFile, show: hasScanLogs },
    ]

    return (
        <div className="space-y-0">
            <div className="flex gap-1 bg-gray-950 border-b border-[--border] p-1 rounded-t-lg sticky top-0 z-20 overflow-x-auto">
                {tabs.filter(t => t.show !== false).map(({ key, label, icon: Icon, accent }) => (
                    <button
                        key={key}
                        onClick={() => setActivePhase(key)}
                        className={cn(
                            "flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-all font-medium whitespace-nowrap",
                            activePhase === key
                                ? accent ? "bg-indigo-600 text-white" : "bg-gray-800 text-white"
                                : accent
                                    ? "text-indigo-400 hover:text-indigo-300 hover:bg-indigo-950/30"
                                    : "text-gray-400 hover:text-gray-200 hover:bg-gray-900",
                        )}
                    >
                        <Icon className="text-base" />
                        {label}
                    </button>
                ))}
            </div>

            <div className="bg-gray-950 rounded-b-lg min-h-[80vh]">
                {activePhase === "overview" && (
                    <OverviewPanel report={report} stats={stats} events={events} />
                )}
                {activePhase === "replay" && (
                    <ReplayPanel
                        rrwebEvents={report.rrwebEvents || []}
                        unifiedEvents={events}
                        includeServerLogs={includeServerLogs}
                        setIncludeServerLogs={setIncludeServerLogs}
                    />
                )}
                {activePhase === "timeline" && (
                    <TimelinePanel
                        events={events}
                        searchQuery={searchQuery}
                        setSearchQuery={setSearchQuery}
                        includeServerLogs={includeServerLogs}
                        setIncludeServerLogs={setIncludeServerLogs}
                    />
                )}
                {activePhase === "network" && (
                    <NetworkPanel logs={report.networkLogs || []} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
                {activePhase === "console" && (
                    <ConsolePanel logs={report.consoleLogs || []} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
                {activePhase === "clicks" && (
                    <ClicksPanel logs={report.clickLogs || []} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
                {activePhase === "websocket" && (
                    <WebSocketPanel logs={report.websocketLogs || []} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
                {activePhase === "server" && (
                    <ServerLogsPanel serverLogs={report.serverLogs || ""} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
                {activePhase === "screenshots" && (
                    <ScreenshotsPanel screenshots={report.screenshots || []} />
                )}
                {activePhase === "scanlogs" && (
                    <ScanLogsPanel scanLogs={report.scanLogs || []} />
                )}
            </div>
        </div>
    )
}

function OverviewPanel({ report, stats, events }: { report: ExtendedReport; stats: ReportStats; events: UnifiedEvent[] }) {
    return (
        <div className="p-4 space-y-4">
            {/* report info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="bg-gray-900 border border-[--border] rounded-lg p-4 space-y-2">
                    <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Report Info</h3>
                    <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1">
                        <DetailLabel>Version</DetailLabel>
                        <DetailValue>{report.appVersion || "—"}</DetailValue>
                        <DetailLabel>OS</DetailLabel>
                        <DetailValue>{report.os || "—"} / {report.arch || "—"}</DetailValue>
                        <DetailLabel>User Agent</DetailLabel>
                        <DetailValue className="truncate max-w-[300px]">{report.userAgent || "—"}</DetailValue>
                        <DetailLabel>Created</DetailLabel>
                        <DetailValue>{report.createdAt ? format(parseISO(report.createdAt), "yyyy-MM-dd HH:mm:ss") : "—"}</DetailValue>
                        {report.viewportWidth ? (
                            <>
                                <DetailLabel>Viewport</DetailLabel>
                                <DetailValue>{report.viewportWidth}×{report.viewportHeight}</DetailValue>
                            </>
                        ) : null}
                    </div>
                </div>

                {/* description */}
                {report.description && (
                    <div className="bg-gray-900 border border-[--border] rounded-lg p-4 space-y-2">
                        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">User Description</h3>
                        <p className="text-sm text-gray-200 whitespace-pre-wrap">{report.description}</p>
                    </div>
                )}
            </div>

            {/* stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <StatCard label="Total Events" value={stats.totalEvents} icon={<BiInfoCircle />} color="text-blue-400" />
                <StatCard
                    label="Errors"
                    value={stats.errorCount}
                    icon={<BiError />}
                    color={stats.errorCount > 0 ? "text-red-400" : "text-gray-500"}
                />
                <StatCard
                    label="Net Errors"
                    value={stats.networkErrors}
                    icon={<LuNetwork />}
                    color={stats.networkErrors > 0 ? "text-red-400" : "text-gray-500"}
                />
                <StatCard
                    label="Recording"
                    value={0}
                    icon={<BiPlay />}
                    color="text-indigo-400"
                    sub={stats.recordingDuration || stats.duration || "—"}
                    hideValue
                />
            </div>

            {/* pipeline */}
            <div className="space-y-2">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Events</h3>
                <div className="flex items-center gap-2 flex-wrap">
                    <PipelineStep label="Clicks" detail={`${stats.clickCount}`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Network" detail={`${stats.networkCount} req`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Console" detail={`${stats.consoleCount}`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Queries" detail={`${stats.queryCount}`} />
                    {stats.navigationCount > 0 && (
                        <>
                            <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                            <PipelineStep label="Navigations" detail={`${stats.navigationCount}`} />
                        </>
                    )}
                    {stats.screenshotCount > 0 && (
                        <>
                            <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                            <PipelineStep label="Screenshots" detail={`${stats.screenshotCount}`} />
                        </>
                    )}
                </div>
            </div>

            {/* server status */}
            {report.status && (
                <div className="space-y-2">
                    <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Server status</h3>
                    <div className="bg-gray-900 border border-[--border] rounded-lg p-3 max-h-[40vh] overflow-auto">
                        <pre className="text-xs font-mono text-gray-300 whitespace-pre-wrap break-all">{report.status}</pre>
                    </div>
                </div>
            )}

            {/* unlocked local files */}
            {report.unlockedLocalFiles && report.unlockedLocalFiles.length > 0 && (
                <div className="space-y-2">
                    <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">
                        Unlocked local files ({report.unlockedLocalFiles.length})
                    </h3>
                    <div className="bg-gray-900 border border-[--border] rounded-lg max-h-[30vh] overflow-auto">
                        {report.unlockedLocalFiles.map((f, i) => (
                            <div key={i} className="px-3 py-1.5 border-b border-gray-800 last:border-0 text-sm">
                                <span className="text-gray-300 font-mono break-all">{f.path}</span>
                                <Badge size="sm" intent="gray" className="ml-2">{f.mediaId}</Badge>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}

const FILTER_OPTIONS: { key: EventType; label: string }[] = [
    { key: "click", label: "Clicks" },
    { key: "network", label: "Network" },
    { key: "console", label: "Console" },
    { key: "query", label: "Queries" },
    { key: "navigation", label: "Nav" },
    { key: "server", label: "Server" },
    { key: "websocket", label: "WebSocket" },
    { key: "screenshot", label: "Screenshots" },
]

function TimelinePanel({ events, searchQuery, setSearchQuery, includeServerLogs, setIncludeServerLogs }: {
    events: UnifiedEvent[]
    searchQuery: string
    setSearchQuery: (v: string) => void
    includeServerLogs: boolean
    setIncludeServerLogs: (v: boolean) => void
}) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
    // initialize with all filters selected
    const [typeFilters, setTypeFilters] = useState<Set<EventType>>(new Set(FILTER_OPTIONS.map(o => o.key)))
    const [showOnlyErrors, setShowOnlyErrors] = useState(false)

    const toggleFilter = (filter: EventType) => {
        setTypeFilters(prev => {
            const next = new Set(prev)
            if (next.has(filter)) next.delete(filter)
            else next.add(filter)
            return next
        })
    }

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        let result = events

        // Filter by error/warning if enabled
        if (showOnlyErrors) {
            result = result.filter(e => e.level === "error" || e.level === "warn")
        }

        // Filter by type
        result = result.filter(e => typeFilters.has(e.type))

        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            result = result.filter(e =>
                e.summary.toLowerCase().includes(q) ||
                (e.pageUrl || "").toLowerCase().includes(q) ||
                JSON.stringify(e.raw).toLowerCase().includes(q),
            )
        }

        return result
    }, [events, typeFilters, showOnlyErrors, searchQuery])

    const errorCount = events.filter(e => e.level === "error" || e.level === "warn").length

    return (
        <div className="flex h-[calc(100vh-140px)]">
            <div className="w-[260px] flex-none p-4 border-r border-[--border] space-y-4 bg-gray-950 overflow-y-auto">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search events..."
                    className="w-full"
                />

                <div className="space-y-1">
                    <div className="flex items-center justify-between">
                        <Checkbox
                            label="Errors only"
                            value={showOnlyErrors}
                            onValueChange={v => setShowOnlyErrors(v as boolean)}
                            size="sm"
                        />
                        <span className="text-xs text-red-400 font-mono tracking-wider tabular-nums bg-red-950/30 px-1.5 rounded">
                            {errorCount}
                        </span>
                    </div>
                </div>

                <div className="h-px bg-[--border]" />

                <div className="space-y-1">
                    {FILTER_OPTIONS.map(({ key, label }) => (
                        <div key={key} className="flex items-center justify-between group">
                            <Checkbox
                                label={label}
                                value={typeFilters.has(key)}
                                onValueChange={() => toggleFilter(key)}
                                size="sm"
                            />
                            <span className="text-xs text-gray-500 tabular-nums group-hover:text-gray-300 transition-colors">
                                {events.filter(e => e.type === key).length}
                            </span>
                        </div>
                    ))}
                </div>

                <div className="h-px bg-[--border]" />

                <Checkbox
                    label="Include server logs"
                    value={includeServerLogs}
                    onValueChange={v => setIncludeServerLogs(v as boolean)}
                    size="sm"
                />
            </div>

            <div className="flex-1 min-w-0 bg-gray-950">
                <div className="p-2 border-b border-[--border] flex items-center justify-between">
                    <span className="text-sm text-gray-400 font-medium">Events</span>
                    <span className="text-xs text-gray-500">{filtered.length} filtered events</span>
                </div>
                <Virtuoso
                    style={{ height: "calc(100% - 40px)" }}
                    totalCount={filtered.length}
                    itemContent={(index) => {
                        const event = filtered[index]
                        return (
                            <div className="pb-1 px-4 pt-1">
                                <TimelineEventRow
                                    event={event}
                                    isExpanded={expandedIds.has(event.id)}
                                    toggleExpanded={() => toggleExpanded(event.id)}
                                />
                            </div>
                        )
                    }}
                />
            </div>
        </div>
    )
}

function TimelineEventRow({ event, isExpanded, toggleExpanded }: {
    event: UnifiedEvent
    isExpanded: boolean
    toggleExpanded: () => void
}) {
    const typeIcons: Record<EventType, React.ReactNode> = {
        click: <FiMousePointer className="text-blue-400" />,
        network: <LuNetwork className={cn(event.level === "error" ? "text-red-400" : "text-green-400")} />,
        console: <LuTerminal
            className={cn(
                event.level === "error" ? "text-red-400" : event.level === "warn" ? "text-orange-400" : "text-gray-400",
            )}
        />,
        query: <LuBrain className={cn(event.level === "error" ? "text-red-400" : "text-purple-400")} />,
        navigation: <BiNavigation className="text-indigo-400" />,
        screenshot: <BiImage className="text-yellow-400" />,
        server: <HiServerStack className="text-gray-400" />,
        websocket: <BiWifi className="text-cyan-400" />,
    }

    const typeLabels: Record<EventType, string> = {
        click: "CLICK",
        network: "NET",
        console: "LOG",
        query: "QUERY",
        navigation: "NAV",
        screenshot: "IMG",
        server: "SRV",
        websocket: "WS",
    }

    const typeBadgeColors: Record<EventType, string> = {
        click: "text-blue-400 bg-blue-950/30",
        network: event.level === "error" ? "text-red-400 bg-red-950/30" : "text-green-400 bg-green-950/30",
        console: event.level === "error" ? "text-red-400 bg-red-950/30" : event.level === "warn"
            ? "text-orange-400 bg-orange-950/30"
            : "text-gray-400 bg-gray-800",
        query: event.level === "error" ? "text-red-400 bg-red-950/30" : "text-purple-400 bg-purple-950/30",
        navigation: "text-indigo-400 bg-indigo-950/30",
        screenshot: "text-yellow-400 bg-yellow-950/30",
        server: "text-gray-400 bg-gray-800",
        websocket: "text-cyan-400 bg-cyan-950/30",
    }

    return (
        <div
            className={cn(
                "bg-gray-900 border rounded-md overflow-hidden",
                event.level === "error" ? "border-red-800/50" : event.level === "warn" ? "border-orange-800/50" : "border-[--border]",
            )}
        >
            <button
                onClick={toggleExpanded}
                className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
            >
                {isExpanded ? <BiChevronDown className="text-gray-500 flex-shrink-0" /> :
                    <BiChevronRight className="text-gray-500 flex-shrink-0" />}

                <span className="text-xs text-gray-500 font-mono flex-shrink-0 w-[70px]">
                    {format(event.timestamp, "HH:mm:ss")}
                </span>

                <span className={cn("text-xs font-semibold px-1.5 py-0.5 rounded flex-shrink-0", typeBadgeColors[event.type])}>
                    {typeLabels[event.type]}
                </span>

                <span className="flex-shrink-0">{typeIcons[event.type]}</span>

                <span
                    className={cn(
                        "text-[0.8rem] leading-5 truncate flex-1 font-mono",
                        event.level === "error" ? "text-red-300" : event.level === "warn" ? "text-orange-300" : "text-gray-300",
                    )}
                >
                    {event.summary}
                </span>

                {event.pageUrl && (
                    <span className="text-xs text-gray-600 truncate max-w-[160px] flex-shrink-0">{event.pageUrl}</span>
                )}
            </button>
            {isExpanded && (
                <div className="border-t border-[--border] bg-gray-950 p-3">
                    {event.type === "screenshot" && event.raw.data ? (
                        <div className="space-y-2">
                            {event.raw.caption && <p className="text-sm text-gray-300 italic">"{event.raw.caption}"</p>}
                            <img src={event.raw.data} alt="Screenshot" className="max-w-full max-h-[50vh] rounded-lg border border-[--border]" />
                        </div>
                    ) : (
                        <DataGrid data={event.raw} />
                    )}
                </div>
            )}
        </div>
    )
}

function NetworkPanel({ logs, searchQuery, setSearchQuery }: {
    logs: Report_NetworkLog[]
    searchQuery: string
    setSearchQuery: (v: string) => void
}) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
    const [statusFilter, setStatusFilter] = useState<"all" | "errors">("all")

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        let result = [...logs]
        if (statusFilter === "errors") {
            result = result.filter(l => l.status >= 400)
        }
        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            result = result.filter(l =>
                l.url.toLowerCase().includes(q) ||
                l.method.toLowerCase().includes(q) ||
                String(l.status).includes(q),
            )
        }
        return result
    }, [logs, searchQuery, statusFilter])

    const errorCount = logs.filter(l => l.status >= 400).length

    return (
        <div className="p-4 space-y-3">
            <div className="flex flex-col sm:flex-row gap-2 sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search network requests..."
                    className="max-w-md"
                    fieldClass="w-fit"
                />
                <div className="flex gap-1">
                    <button
                        onClick={() => setStatusFilter("all")}
                        className={cn(
                            "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                            statusFilter === "all" ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                        )}
                    >
                        All ({logs.length})
                    </button>
                    <button
                        onClick={() => setStatusFilter("errors")}
                        className={cn(
                            "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                            statusFilter === "errors" ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                        )}
                    >
                        Errors ({errorCount})
                    </button>
                </div>
                <span className="text-sm text-gray-500 self-center flex-none">{filtered.length} requests</span>
            </div>

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const log = filtered[index]
                    const id = index
                    return (
                        <div className="pb-1">
                            <div
                                className={cn(
                                    "bg-gray-900 border rounded-md overflow-hidden",
                                    log.status >= 400 ? "border-red-800/50" : "border-[--border]",
                                )}
                            >
                                <button
                                    onClick={() => toggleExpanded(id)}
                                    className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
                                >
                                    {expandedIds.has(id)
                                        ? <BiChevronDown className="text-gray-500 flex-shrink-0" />
                                        : <BiChevronRight className="text-gray-500 flex-shrink-0" />}
                                    <span className="text-xs text-gray-500 font-mono flex-shrink-0 w-[70px]">
                                        {log.timestamp ? format(parseISO(log.timestamp), "HH:mm:ss") : "—"}
                                    </span>
                                    <Badge size="sm" intent={log.method === "POST" ? "success" : "gray"}>
                                        {log.method}
                                    </Badge>
                                    <span className="text-[0.8rem] text-blue-300 font-mono truncate flex-1">{log.url}</span>
                                    <Badge size="sm" intent={log.status >= 400 ? "alert" : "gray"}>
                                        {log.status}
                                    </Badge>
                                    <span className="text-xs text-gray-500 flex-shrink-0">{log.duration}ms</span>
                                </button>
                                {expandedIds.has(id) && (
                                    <div className="border-t border-[--border] bg-gray-950 p-3 space-y-2">
                                        <div className="text-xs text-gray-500 font-mono break-all">{log.pageUrl}</div>
                                        {log.body && log.body !== "null" && (
                                            <div className="space-y-1">
                                                <p className="text-sm font-semibold text-gray-400">Request Body</p>
                                                <pre className="text-xs font-mono text-gray-300 bg-gray-900 p-2 rounded break-all whitespace-pre-wrap max-h-[200px] overflow-auto">
                                                    {tryFormatJSON(log.body)}
                                                </pre>
                                            </div>
                                        )}
                                        {log.dataPreview && (
                                            <div className="space-y-1">
                                                <p className="text-sm font-semibold text-gray-400">Response</p>
                                                <pre className="text-xs font-mono text-gray-300 bg-gray-900 p-2 rounded break-all whitespace-pre-wrap max-h-[200px] overflow-auto">
                                                    {tryFormatJSON(log.dataPreview)}
                                                </pre>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ConsolePanel({ logs, searchQuery, setSearchQuery }: {
    logs: Report_ConsoleLog[]
    searchQuery: string
    setSearchQuery: (v: string) => void
}) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
    const [levelFilter, setLevelFilter] = useState<"all" | "error" | "warn" | "log">("all")

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        let result = [...logs]
        if (levelFilter !== "all") {
            result = result.filter(l => l.type === levelFilter)
        }
        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            result = result.filter(l => l.content.toLowerCase().includes(q))
        }
        return result
    }, [logs, searchQuery, levelFilter])

    const errorCount = logs.filter(l => l.type === "error").length
    const warnCount = logs.filter(l => l.type === "warn").length

    return (
        <div className="p-4 space-y-3">
            <div className="flex flex-col sm:flex-row gap-2 sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search console logs..."
                    className="max-w-md"
                    fieldClass="w-fit"
                />
                <div className="flex gap-1">
                    {([
                        { key: "all" as const, label: "All" },
                        { key: "error" as const, label: `Errors (${errorCount})` },
                        { key: "warn" as const, label: `Warnings (${warnCount})` },
                        { key: "log" as const, label: "Logs" },
                    ]).map(({ key, label }) => (
                        <button
                            key={key}
                            onClick={() => setLevelFilter(key)}
                            className={cn(
                                "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                                levelFilter === key ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                            )}
                        >
                            {label}
                        </button>
                    ))}
                </div>
                <span className="text-sm text-gray-500 self-center flex-none">{filtered.length} entries</span>
            </div>

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const log = filtered[index]
                    return (
                        <div className="pb-1">
                            <div
                                className={cn(
                                    "bg-gray-900 border rounded-md overflow-hidden",
                                    log.type === "error" ? "border-red-800/50" : log.type === "warn" ? "border-orange-800/50" : "border-[--border]",
                                )}
                            >
                                <button
                                    onClick={() => toggleExpanded(index)}
                                    className="flex items-start gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
                                >
                                    <LevelDot level={log.type === "error" ? "error" : log.type === "warn" ? "warn" : "debug"} />
                                    <span className="text-xs text-gray-500 font-mono flex-shrink-0 w-[70px] pt-0.5">
                                        {log.timestamp ? format(parseISO(log.timestamp), "HH:mm:ss") : "—"}
                                    </span>
                                    <LevelBadge level={log.type as any} />
                                    <span
                                        className={cn(
                                            "text-[0.8rem] leading-5 break-all flex-1",
                                            log.type === "error" ? "text-red-300" : log.type === "warn" ? "text-orange-300" : "text-gray-300",
                                        )}
                                    >
                                        {log.content?.slice(0, 200)}
                                    </span>
                                </button>
                                {expandedIds.has(index) && (
                                    <div className="border-t border-[--border] bg-gray-950 p-3 space-y-2">
                                        <pre className="text-xs font-mono text-gray-300 whitespace-pre-wrap break-all">{log.content}</pre>
                                        <div className="text-xs text-gray-600">{log.pageUrl}</div>
                                    </div>
                                )}
                            </div>
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ClicksPanel({ logs, searchQuery, setSearchQuery }: {
    logs: Report_ClickLog[]
    searchQuery: string
    setSearchQuery: (v: string) => void
}) {
    const filtered = useMemo(() => {
        if (!searchQuery) return logs
        const q = searchQuery.toLowerCase()
        return logs.filter(l =>
            (l.text || "").toLowerCase().includes(q) ||
            l.element.toLowerCase().includes(q) ||
            l.pageUrl.toLowerCase().includes(q),
        )
    }, [logs, searchQuery])

    return (
        <div className="p-4 space-y-3">
            <div className="flex gap-2 items-center sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search clicks..."
                    className="max-w-md"
                    fieldClass="w-fit"
                />
                <span className="text-sm text-gray-500">{filtered.length} clicks</span>
            </div>
            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const log = filtered[index]
                    return (
                        <div className="pb-1">
                            <div className="bg-gray-900 border border-[--border] rounded-md px-3 py-2 flex items-center gap-3">
                                <FiMousePointer className="text-blue-400 flex-shrink-0" />
                                <span className="text-xs text-gray-500 font-mono flex-shrink-0 w-[70px]">
                                    {log.timestamp ? format(parseISO(log.timestamp), "HH:mm:ss") : "—"}
                                </span>
                                <Badge size="sm" intent="gray">{log.element}</Badge>
                                <span className="text-[0.8rem] text-gray-200 truncate flex-1">
                                    {log.text ? `"${log.text.slice(0, 80)}"` : "—"}
                                </span>
                                <span className="text-xs text-gray-600 truncate max-w-[200px]">{log.pageUrl}</span>
                            </div>
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ServerLogsPanel({ serverLogs, searchQuery, setSearchQuery }: {
    serverLogs: string
    searchQuery: string
    setSearchQuery: (v: string) => void
}) {
    const lines = useMemo(() => {
        if (!serverLogs) return []
        return serverLogs.split("\n").filter(Boolean)
    }, [serverLogs])

    const filtered = useMemo(() => {
        if (!searchQuery) return lines
        const q = searchQuery.toLowerCase()
        return lines.filter(l => l.toLowerCase().includes(q))
    }, [lines, searchQuery])

    return (
        <div className="p-4 space-y-3">
            <div className="flex gap-2 items-center sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search server logs..."
                    className="max-w-md"
                    fieldClass="w-fit"
                />
                <span className="text-sm text-gray-500">{filtered.length} lines</span>
            </div>
            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const line = filtered[index]
                    const isError = line.includes("|ERR|")
                    const isWarn = line.includes("|WRN|")
                    return (
                        <div className="pb-0.5">
                            <div
                                className={cn(
                                    "px-3 py-1.5 rounded text-xs font-mono break-all",
                                    isError ? "bg-red-950/30 text-red-300" : isWarn ? "bg-orange-950/30 text-orange-300" : "text-gray-400",
                                )}
                            >
                                {line}
                            </div>
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ScreenshotsPanel({ screenshots }: { screenshots: Screenshot[] }) {
    const [selectedIndex, setSelectedIndex] = useState<number | null>(null)

    if (screenshots.length === 0) {
        return (
            <div className="flex items-center justify-center h-[40vh] text-gray-500">
                <p>No screenshots in this report</p>
            </div>
        )
    }

    return (
        <div className="p-4 space-y-4">
            <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">
                Screenshots ({screenshots.length})
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {screenshots.map((ss, i) => (
                    <button
                        key={i}
                        onClick={() => setSelectedIndex(selectedIndex === i ? null : i)}
                        className={cn(
                            "bg-gray-900 border rounded-lg overflow-hidden text-left hover:border-indigo-500 transition-colors",
                            selectedIndex === i ? "border-indigo-500" : "border-[--border]",
                        )}
                    >
                        <img src={ss.data} alt={ss.caption || `Screenshot ${i + 1}`} className="w-full h-[160px] object-cover" />
                        <div className="p-3 space-y-1">
                            <p className="text-sm text-gray-300 font-medium">{ss.caption || `Screenshot ${i + 1}`}</p>
                            <p className="text-xs text-gray-500">{ss.timestamp ? format(parseISO(ss.timestamp), "HH:mm:ss") : "—"}</p>
                            <p className="text-xs text-gray-600 truncate">{ss.pageUrl}</p>
                        </div>
                    </button>
                ))}
            </div>

            {/* fullscreen preview */}
            {selectedIndex !== null && (
                <div className="bg-gray-900 border border-[--border] rounded-lg p-4">
                    <img
                        src={screenshots[selectedIndex].data}
                        alt="Full screenshot"
                        className="max-w-full max-h-[60vh] rounded-lg mx-auto"
                    />
                    {screenshots[selectedIndex].caption && (
                        <p className="text-sm text-gray-300 text-center mt-3 italic">"{screenshots[selectedIndex].caption}"</p>
                    )}
                </div>
            )}
        </div>
    )
}

// session replay
function ReplayPanel({ rrwebEvents, unifiedEvents, includeServerLogs, setIncludeServerLogs }: {
    rrwebEvents: any[]
    unifiedEvents: UnifiedEvent[]
    includeServerLogs: boolean
    setIncludeServerLogs: (v: boolean) => void
}) {
    const containerRef = useRef<HTMLDivElement>(null)
    const playerRef = useRef<any>(null)
    const [currentTime, setCurrentTime] = useState(0) // relative time in ms
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())

    // Filters for replay
    const [typeFilters, setTypeFilters] = useState<Set<EventType>>(new Set(FILTER_OPTIONS.map(o => o.key)))

    const toggleFilter = (filter: EventType) => {
        setTypeFilters(prev => {
            const next = new Set(prev)
            if (next.has(filter)) next.delete(filter)
            else next.add(filter)
            return next
        })
    }

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    // derived start time from the first rrweb event
    const recordingStartTime = useMemo(() => {
        return rrwebEvents.length > 0 ? rrwebEvents[0].timestamp : 0
    }, [rrwebEvents])

    useEffect(() => {
        if (!containerRef.current || rrwebEvents.length === 0) return

        let mounted = true

        const initPlayer = async () => {
            try {
                const [rrwebPlayerModule] = await Promise.all([
                    import("rrweb-player"),
                    import("rrweb-player/dist/style.css"),
                ])

                if (!mounted || !containerRef.current) return

                containerRef.current.innerHTML = ""

                const RRWebPlayer = rrwebPlayerModule.default

                playerRef.current = new RRWebPlayer({
                    target: containerRef.current,
                    props: {
                        events: rrwebEvents,
                        width: containerRef.current.clientWidth,
                        height: containerRef.current.clientHeight,
                        autoPlay: false,
                        showController: true,
                        speedOption: [1, 2],
                        skipInactive: true,
                        insertStyleRules: [
                            ".UI-Modal__overlay { opacity: 1 !important; visibility: visible !important; }",
                            ".UI-Modal__content { opacity: 1 !important; visibility: visible !important; transform: none !important; }",
                            "[data-radix-portal] { opacity: 1 !important; visibility: visible !important; pointer-events: auto !important; }",
                            "body { pointer-events: auto !important; }",
                        ],
                    },
                })

                // syncing logic
                playerRef.current.addEventListener("ui-update-progress", (payload: any) => {
                    const time = playerRef.current?.getReplayer()?.getCurrentTime()
                    if (typeof time === "number") {
                        setCurrentTime(time)
                    }
                })
            }
            catch (err) {
                console.error("Failed to initialize rrweb player:", err)
                if (containerRef.current) {
                    containerRef.current.innerHTML = `
                        <div style="display:flex;align-items:center;justify-content:center;height:400px;color:#888;font-size:14px;">
                            Failed to load session replay player
                        </div>
                    `
                }
            }
        }

        // initialize player with a small delay to ensure layout is settled
        const timer = setTimeout(() => {
            initPlayer()
        }, 50)

        // resizing
        const resizeObserver = new ResizeObserver((entries) => {
            if (!playerRef.current) return
            for (const entry of entries) {
                const { width, height } = entry.contentRect
                try {
                    playerRef.current.$set({ width, height })
                    playerRef.current.triggerResize()
                }
                catch (e) {
                    console.warn("Failed to resize player", e)
                }
            }
        })

        if (containerRef.current) {
            resizeObserver.observe(containerRef.current)
        }

        return () => {
            mounted = false
            clearTimeout(timer)
            resizeObserver.disconnect()
            if (playerRef.current) {
                try {
                    playerRef.current.$destroy()
                }
                catch {
                }
                playerRef.current = null
            }
        }
    }, [rrwebEvents])

    // filter events based on current playback time
    const visibleEvents = useMemo(() => {
        if (!recordingStartTime) return []
        const currentAbsTime = recordingStartTime + currentTime

        // Filter by type
        const filtered = unifiedEvents.filter(e => typeFilters.has(e.type))

        // show events up to current time (with small buffer)
        return filtered.filter(e => e.timestamp.getTime() <= currentAbsTime + 100)
    }, [unifiedEvents, recordingStartTime, currentTime, typeFilters])

    const virtuosoRef = useRef<any>(null)

    // auto scroll as events are added
    useEffect(() => {
        if (visibleEvents.length > 0) {
            // small delay to ensure rendering
            requestAnimationFrame(() => {
                virtuosoRef.current?.scrollToIndex({ index: visibleEvents.length - 1, align: "end", behavior: "auto" })
            })
        }
    }, [visibleEvents.length])

    if (rrwebEvents.length === 0) {
        return (
            <div className="flex items-center justify-center h-[40vh] text-gray-500">
                <p>No session replay data in this report</p>
            </div>
        )
    }

    return (
        <div className="h-[calc(100vh-120px)] overflow-hidden flex flex-col">
            <div className="px-4 py-2 flex items-center gap-3 border-b border-[--border] bg-gray-950 flex-none">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Session Replay</h3>
                <span className="text-xs text-gray-500">{rrwebEvents.length} DOM events</span>
                <div className="ml-auto">
                    <Checkbox
                        label="Include server logs"
                        value={includeServerLogs}
                        onValueChange={v => setIncludeServerLogs(v as boolean)}
                        size="sm"
                    />
                </div>
            </div>

            <ResizablePanelGroup direction="horizontal" className="flex-1 overflow-hidden">
                <ResizablePanel defaultSize={70} minSize={30}>
                    <div className="h-full bg-gray-900 relative min-h-0 min-w-0 overflow-hidden">
                        <div
                            ref={containerRef}
                            className="absolute inset-0 w-full h-full [&_.rr-player]:!h-full [&_.rr-player]:!flex [&_.rr-player]:!flex-col [&_.replayer-wrapper]:!flex-1 [&_.rr-controller]:!flex-none"
                        />
                    </div>
                </ResizablePanel>

                <ResizableHandle withHandle />

                <ResizablePanel defaultSize={30} minSize={15}>
                    <div className="h-full bg-gray-950 flex flex-col flex-none">
                        <div className="p-2 border-b border-[--border] bg-gray-900/50 space-y-2">
                            <div className="flex items-center justify-between">
                                <p className="text-xs text-center font-medium text-gray-400">Timeline ({visibleEvents.length})</p>
                            </div>
                            <div className="flex gap-2 overflow-x-auto pb-1 no-scrollbar mask-fade-right">
                                {FILTER_OPTIONS.map(({ key, label }) => (
                                    <button
                                        key={key}
                                        onClick={() => toggleFilter(key)}
                                        className={cn(
                                            "flex-none px-2 py-0.5 text-[10px] rounded border transition-colors",
                                            typeFilters.has(key)
                                                ? "bg-gray-800 border-gray-600 text-gray-200"
                                                : "bg-transparent border-gray-800 text-gray-600 hover:border-gray-700",
                                        )}
                                    >
                                        {label}
                                    </button>
                                ))}
                            </div>
                        </div>
                        <div className="flex-1 overflow-hidden">
                            {visibleEvents.length === 0 ? (
                                <div className="h-full flex items-center justify-center text-gray-600 text-sm">
                                    Waiting for events...
                                </div>
                            ) : (
                                <Virtuoso
                                    ref={virtuosoRef}
                                    style={{ height: "100%" }}
                                    totalCount={visibleEvents.length}
                                    itemContent={(index) => {
                                        const event = visibleEvents[index]
                                        return (
                                            <div className="px-2 py-0.5">
                                                <TimelineEventRow
                                                    event={event}
                                                    isExpanded={expandedIds.has(event.id)}
                                                    toggleExpanded={() => toggleExpanded(event.id)}
                                                />
                                            </div>
                                        )
                                    }}
                                />
                            )}
                        </div>
                    </div>
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    )
}

// websocket panel
function WebSocketPanel({ logs, searchQuery, setSearchQuery }: {
    logs: WebSocketLog[]
    searchQuery: string
    setSearchQuery: (v: string) => void
}) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
    const [directionFilter, setDirectionFilter] = useState<"all" | "incoming" | "outgoing">("all")

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        let result = [...logs]
        if (directionFilter !== "all") {
            result = result.filter(l => l.direction === directionFilter)
        }
        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            result = result.filter(l =>
                l.eventType.toLowerCase().includes(q) ||
                JSON.stringify(l.payload).toLowerCase().includes(q),
            )
        }
        return result
    }, [logs, searchQuery, directionFilter])

    return (
        <div className="p-4 space-y-3">
            <div className="flex flex-col sm:flex-row gap-2 sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search WebSocket events..."
                    className="max-w-md"
                    fieldClass="w-fit"
                />
                <div className="flex gap-1">
                    {([
                        { key: "all" as const, label: "All" },
                        { key: "incoming" as const, label: "Incoming" },
                        { key: "outgoing" as const, label: "Outgoing" },
                    ]).map(({ key, label }) => (
                        <button
                            key={key}
                            onClick={() => setDirectionFilter(key)}
                            className={cn(
                                "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                                directionFilter === key ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                            )}
                        >
                            {label}
                        </button>
                    ))}
                </div>
                <span className="text-sm text-gray-500 self-center flex-none">{filtered.length} events</span>
            </div>

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const log = filtered[index]
                    return (
                        <div className="pb-1">
                            <div
                                className={cn(
                                    "bg-gray-900 border rounded-md overflow-hidden border-[--border]",
                                )}
                            >
                                <button
                                    onClick={() => toggleExpanded(index)}
                                    className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
                                >
                                    {expandedIds.has(index)
                                        ? <BiChevronDown className="text-gray-500 flex-shrink-0" />
                                        : <BiChevronRight className="text-gray-500 flex-shrink-0" />}
                                    <span className="text-xs text-gray-500 font-mono flex-shrink-0 w-[70px]">
                                        {log.timestamp ? format(parseISO(log.timestamp), "HH:mm:ss") : "—"}
                                    </span>
                                    <Badge size="sm" intent={log.direction === "incoming" ? "info" : "success"}>
                                        {log.direction === "incoming" ? "↓ IN" : "↑ OUT"}
                                    </Badge>
                                    <span className="text-[0.8rem] text-cyan-300 font-mono truncate flex-1">
                                        {log.eventType}
                                    </span>
                                </button>
                                {expandedIds.has(index) && (
                                    <div className="border-t border-[--border] bg-gray-950 p-3 space-y-2">
                                        <pre className="text-xs font-mono text-gray-300 bg-gray-900 p-2 rounded break-all whitespace-pre-wrap max-h-[300px] overflow-auto">
                                            {tryFormatJSON(JSON.stringify(log.payload))}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ScanLogsPanel({ scanLogs }: { scanLogs: string[] }) {
    return (
        <div className="p-4 space-y-3">
            <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">
                Scan Logs ({scanLogs.length})
            </h3>
            <p>
                First = earliest
            </p>
            <div className="flex gap-2 flex-wrap">
                {scanLogs.map((log, index) => (
                    <Drawer
                        title={`Scan Log ${index + 1}`}
                        size="full"
                        trigger={
                            <Button intent="gray-outline" className="mt-1" leftIcon={<BiFile />}>
                                Scan Log {index + 1}
                            </Button>
                        }
                        key={index}
                    >
                        <ScanLogViewer content={log} />
                    </Drawer>
                ))}
            </div>
        </div>
    )
}

function StatCard({ label, value, icon, color, sub, hideValue }: {
    label: string; value: number; icon: React.ReactNode; color: string; sub?: string; hideValue?: boolean
}) {
    return (
        <div className="bg-gray-900 border border-[--border] rounded-lg p-3 space-y-1">
            <div className="flex items-center gap-2">
                <span className={cn("text-lg", color)}>{icon}</span>
                <span className="text-sm text-gray-400 font-medium">{label}</span>
            </div>
            <div className="flex items-baseline gap-2">
                {!hideValue && <span className="text-2xl font-bold text-white">{value}</span>}
                {sub && <span className={cn("text-sm", hideValue ? "text-2xl font-bold text-white" : "text-gray-500")}>{sub}</span>}
            </div>
        </div>
    )
}

function PipelineStep({ label, detail }: { label: string; detail: string }) {
    return (
        <div className="bg-gray-900 border border-[--border] rounded-md px-3 py-2 text-center min-w-[100px]">
            <p className="text-sm font-semibold text-gray-300">{label}</p>
            <p className="text-sm text-gray-500 mt-0.5">{detail}</p>
        </div>
    )
}

function DetailLabel({ children }: { children: React.ReactNode }) {
    return <span className="text-sm text-gray-500 font-medium">{children}</span>
}

function DetailValue({ children, className }: { children: React.ReactNode; className?: string }) {
    return <span className={cn("text-sm text-gray-200", className)}>{children}</span>
}

type LogLevel = "error" | "warn" | "info" | "debug" | "log"

function LevelBadge({ level }: { level: LogLevel }) {
    const intents: Record<string, "alert" | "warning" | "success" | "info" | "gray"> = {
        error: "alert",
        warn: "warning",
        info: "info",
        debug: "gray",
        log: "gray",
    }
    return <Badge size="sm" intent={intents[level] || "gray"}>{level}</Badge>
}

function LevelDot({ level }: { level: string }) {
    return (
        <span
            className={cn(
                "inline-block w-1.5 h-1.5 rounded-full mt-1.5 flex-shrink-0",
                level === "error" && "bg-red-400",
                level === "warn" && "bg-orange-400",
                level === "info" && "bg-blue-400",
                level === "debug" && "bg-gray-600",
            )}
        />
    )
}

function DataGrid({ data }: { data: Record<string, any> }) {
    return (
        <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
            {Object.entries(data).map(([key, value]) => {
                // skip displaying base64 image data inline
                if (key === "data" && typeof value === "string" && value.startsWith("data:image")) {
                    return (
                        <React.Fragment key={key}>
                            <span className="text-gray-500 text-sm select-all">{key}</span>
                            <span className="text-sm text-gray-400">[image data]</span>
                        </React.Fragment>
                    )
                }

                // handle large text blocks or code blocks
                if ((key === "dataPreview" || key === "body" || key === "payload" || key === "stack") && (typeof value === "string" || (typeof value === "object" && value !== null))) {
                    const content = typeof value === "string" ? tryFormatJSON(value) : JSON.stringify(value, null, 2)
                    if (!content || content === "null" || content === "{}") return null

                    return (
                        <div key={key} className="col-span-2 mt-2">
                            <p className="text-gray-500 text-sm font-medium mb-1 font-mono">{key}</p>
                            <pre className="text-xs font-mono text-gray-300 bg-gray-900 p-2 rounded break-all whitespace-pre-wrap max-h-[300px] overflow-auto border border-gray-800">
                                {content}
                            </pre>
                        </div>
                    )
                }

                const isObj = typeof value === "object" && value !== null
                return (
                    <React.Fragment key={key}>
                        <span className="text-gray-500 text-sm select-all">{key}</span>
                        <span
                            className={cn(
                                "text-sm break-all select-all",
                                isObj ? "text-gray-400" : "text-gray-200",
                            )}
                        >
                            {isObj ? JSON.stringify(value, null, 1) : String(value)}
                        </span>
                    </React.Fragment>
                )
            })}
        </div>
    )
}

function tryFormatJSON(str: string): string {
    try {
        let parsed = JSON.parse(str)
        if (typeof parsed === "string") {
            try {
                const inner = JSON.parse(parsed)
                if (typeof inner === "object" && inner !== null) {
                    parsed = inner
                }
            }
            catch {
            }
        }
        return JSON.stringify(parsed, null, 2)
    }
    catch {
        return str
    }
}
