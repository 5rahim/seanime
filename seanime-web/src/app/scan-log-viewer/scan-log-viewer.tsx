import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { TextInput } from "@/components/ui/text-input"
import React, { useMemo, useState } from "react"
import { BiCheck, BiChevronDown, BiChevronRight, BiError, BiFile, BiInfoCircle, BiLinkAlt, BiSearch, BiX } from "react-icons/bi"
import { RiFileSettingsFill } from "react-icons/ri"
import { Virtuoso } from "react-virtuoso"

type LogLevel = "trace" | "debug" | "info" | "warn" | "error"
type LogContext = "Matcher" | "FileHydrator" | "MediaFetcher" | "MediaContainer" | "Scanner"

interface ParsedLogLine {
    idx: number
    level: LogLevel
    context?: LogContext
    filename?: string
    message?: string
    raw: any
}

type Phase = "overview" | "parsing" | "matcher" | "hydrator" | "issues"

interface FileGroup {
    filename: string
    parsingLog?: ParsedLogLine
    matcherLogs: ParsedLogLine[]
    hydratorLogs: ParsedLogLine[]
    matchResult?: { id: number; match: string; score: number } | null
    hydrationResult?: { episode: number; aniDBEpisode: string; type: string } | null
    hasError: boolean
    hasWarning: boolean
    isUnmatched: boolean
}

interface ScanStats {
    totalFiles: number
    matchedFiles: number
    unmatchedFiles: number
    errorCount: number
    warningCount: number
    scanDuration: string | null
    matcherDuration: string | null
    hydratorDuration: string | null
    mediaCount: number
    fetchedMediaCount: number
    unknownMediaCount: number
    tokenIndexSize: number
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function parseLogLines(content: string): ParsedLogLine[] {
    const lines = content.split("\n")
    const parsed: ParsedLogLine[] = []
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim()
        if (!line) continue
        try {
            const data: any = JSON.parse(line)
            parsed.push({
                idx: i,
                level: data.level || "debug",
                context: data.context || undefined,
                filename: data.filename || undefined,
                message: data.message || undefined,
                raw: data,
            })
        }
        catch {
            // non-JSON lines ignored
        }
    }
    return parsed
}

function extractStats(lines: ParsedLogLine[]): ScanStats {
    const stats: ScanStats = {
        totalFiles: 0,
        matchedFiles: 0,
        unmatchedFiles: 0,
        errorCount: 0,
        warningCount: 0,
        scanDuration: null,
        matcherDuration: null,
        hydratorDuration: null,
        mediaCount: 0,
        fetchedMediaCount: 0,
        unknownMediaCount: 0,
        tokenIndexSize: 0,
    }

    const matchedFilenames = new Set<string>()
    const allMatcherFilenames = new Set<string>()

    for (const line of lines) {
        if (line.level === "error") stats.errorCount++
        if (line.level === "warn") stats.warningCount++

        const d = line.raw

        // matcher stats
        if (line.context === "Matcher") {
            if (line.filename) allMatcherFilenames.add(line.filename)
            if (d.message === "Best match found" && line.filename) {
                matchedFilenames.add(line.filename)
            }
            if (d.message === "Finished matching process") {
                stats.matcherDuration = d.ms ? `${d.ms}ms` : null
                stats.totalFiles = d.files || 0
                stats.unmatchedFiles = d.unmatched || 0
            }
        }

        // media fetcher stats
        if (line.context === "MediaFetcher") {
            if (d.message === "Finished creating media fetcher") {
                stats.fetchedMediaCount = d.allMediaCount || 0
                stats.unknownMediaCount = d.unknownMediaCount || 0
            }
        }

        // media container stats
        if (line.context === "MediaContainer") {
            if (d.message === "Created media container") {
                stats.mediaCount = d.mediaCount || 0
                stats.tokenIndexSize = d.tokenIndexSize || 0
            }
        }

        // hydrator timing
        if (line.context === "FileHydrator") {
            if (d.message === "Finished metadata hydration") {
                stats.hydratorDuration = d.ms ? `${d.ms}ms` : null
            }
        }

        // Scan completion
        if (d.message === "Scan completed" && d.count) {
            stats.totalFiles = d.count
        }
    }

    stats.matchedFiles = matchedFilenames.size
    if (stats.totalFiles === 0) stats.totalFiles = allMatcherFilenames.size
    if (stats.unmatchedFiles === 0) stats.unmatchedFiles = stats.totalFiles - stats.matchedFiles

    return stats
}

function buildFileGroups(lines: ParsedLogLine[]): FileGroup[] {
    const fileMap = new Map<string, FileGroup>()

    const getOrCreate = (filename: string): FileGroup => {
        let group = fileMap.get(filename)
        if (!group) {
            group = {
                filename,
                parsingLog: undefined,
                matcherLogs: [],
                hydratorLogs: [],
                matchResult: null,
                hydrationResult: null,
                hasError: false,
                hasWarning: false,
                isUnmatched: false,
            }
            fileMap.set(filename, group)
        }
        return group
    }

    for (const line of lines) {
        if (!line.filename) continue

        // Parsed file lines (no context, has path)
        if (!line.context && line.raw.path) {
            const group = getOrCreate(line.filename)
            group.parsingLog = line
        }

        if (line.context === "Matcher") {
            const group = getOrCreate(line.filename)
            group.matcherLogs.push(line)
            if (line.level === "error") group.hasError = true
            if (line.level === "warn") group.hasWarning = true

            if (line.raw.message === "Best match found" || line.raw.message === "Hook overrode match" || line.raw.message === "Matched by rule") {
                group.matchResult = {
                    id: line.raw.id,
                    match: line.raw.match,
                    score: line.raw.score || 0,
                }
            }
            if (line.raw.message === "No match found") {
                group.isUnmatched = true
            }
        }

        if (line.context === "FileHydrator") {
            const group = getOrCreate(line.filename)
            group.hydratorLogs.push(line)
            if (line.level === "error") group.hasError = true
            if (line.level === "warn") group.hasWarning = true

            if (line.raw.hydrated) {
                group.hydrationResult = {
                    episode: line.raw.hydrated.episode || 0,
                    aniDBEpisode: line.raw.hydrated.aniDBEpisode || "",
                    type: (line.raw.message?.includes("File has been marked as")
                        ? line.raw.message.replace("File has been marked as ", "")
                        : group.hydrationResult?.type) || "unknown",
                }
            }
        }
    }

    return Array.from(fileMap.values())
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function ScanLogViewer({ content }: { content: string }) {
    const [activePhase, setActivePhase] = useState<Phase>("overview")
    const [searchQuery, setSearchQuery] = useState("")
    const [levelFilter, setLevelFilter] = useState<LogLevel | "all">("all")
    const [statusFilter, setStatusFilter] = useState<"all" | "matched" | "unmatched" | "errors">("all")
    const [selectedFile, setSelectedFile] = useState<string | null>(null)

    const lines = useMemo(() => parseLogLines(content), [content])
    const stats = useMemo(() => extractStats(lines), [lines])
    const fileGroups = useMemo(() => buildFileGroups(lines), [lines])

    const fileGroupMap = useMemo(() => {
        const map = new Map<string, FileGroup>()
        for (const g of fileGroups) map.set(g.filename, g)
        return map
    }, [fileGroups])

    const parsingLines = useMemo(
        () => lines.filter((l) => !l.context && l.raw.path && l.raw.filename),
        [lines],
    )

    const systemLines = useMemo(
        () => lines.filter((l) => l.context === "MediaFetcher" || l.context === "MediaContainer" || (!l.context && !l.raw.path)),
        [lines],
    )

    const issueLines = useMemo(
        () => lines.filter((l) => l.level === "error" || l.level === "warn"),
        [lines],
    )

    const onSelectFile = (filename: string) => {
        setSelectedFile(filename)
    }

    if (!content) {
        return (
            <div className="flex items-center justify-center h-[40vh] text-[--muted]">
                <p className="text-lg">Load a scan log file to begin analysis</p>
            </div>
        )
    }

    // file flow view (full journey of a file)
    if (selectedFile) {
        const group = fileGroupMap.get(selectedFile)
        return (
            <div className="space-y-0">
                <div className="flex items-center gap-2 bg-gray-950 border-b border-[--border] p-2 rounded-t-lg sticky top-0 z-20">
                    <Button intent="gray" size="sm" onClick={() => setSelectedFile(null)}>
                        ← Back
                    </Button>
                    <BiFile className="text-blue-400" />
                    <span className="text-sm font-medium text-gray-200 break-all">{selectedFile}</span>
                </div>
                <div className="bg-gray-950 rounded-b-lg min-h-[60vh] p-4">
                    {group ? (
                        <FileFlowPanel group={group} />
                    ) : (
                        <p className="text-gray-500 text-sm">No logs found for this file.</p>
                    )}
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-0">
            <div className="flex gap-1 bg-gray-950 border-b border-[--border] p-1 rounded-t-lg sticky top-0 z-20">
                {([
                    { key: "overview", label: "Overview", icon: BiInfoCircle },
                    { key: "parsing", label: "Parsed Files", icon: BiFile },
                    { key: "matcher", label: "Matcher", icon: BiSearch },
                    { key: "hydrator", label: "Hydrator", icon: RiFileSettingsFill },
                    { key: "issues", label: `Issues (${stats.errorCount + stats.warningCount})`, icon: BiError },
                ] as const).map(({ key, label, icon: Icon }) => (
                    <button
                        key={key}
                        onClick={() => setActivePhase(key as Phase)}
                        className={cn(
                            "flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-all font-medium",
                            activePhase === key
                                ? "bg-gray-800 text-white"
                                : "text-gray-400 hover:text-gray-200 hover:bg-gray-900",
                        )}
                    >
                        <Icon className="text-base" />
                        {label}
                    </button>
                ))}
            </div>

            <div className="bg-gray-950 rounded-b-lg min-h-[60vh]">
                {activePhase === "overview" && (
                    <OverviewPanel stats={stats} lines={systemLines} />
                )}
                {activePhase === "parsing" && (
                    <ParsingPanel lines={parsingLines} searchQuery={searchQuery} setSearchQuery={setSearchQuery} onSelectFile={onSelectFile} />
                )}
                {activePhase === "matcher" && (
                    <MatcherPanel
                        fileGroups={fileGroups}
                        searchQuery={searchQuery}
                        setSearchQuery={setSearchQuery}
                        statusFilter={statusFilter}
                        setStatusFilter={setStatusFilter}
                        onSelectFile={onSelectFile}
                    />
                )}
                {activePhase === "hydrator" && (
                    <HydratorPanel
                        fileGroups={fileGroups}
                        searchQuery={searchQuery}
                        setSearchQuery={setSearchQuery}
                        levelFilter={levelFilter}
                        setLevelFilter={setLevelFilter}
                        onSelectFile={onSelectFile}
                    />
                )}
                {activePhase === "issues" && (
                    <IssuesPanel lines={issueLines} searchQuery={searchQuery} setSearchQuery={setSearchQuery} />
                )}
            </div>
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function OverviewPanel({ stats, lines }: { stats: ScanStats; lines: ParsedLogLine[] }) {
    return (
        <div className="p-4 space-y-4">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <StatCard label="Total Files" value={stats.totalFiles} icon={<BiFile />} color="text-blue-400" />
                <StatCard
                    label="Matched"
                    value={stats.matchedFiles}
                    icon={<BiCheck />}
                    color="text-green-400"
                    sub={stats.totalFiles > 0 ? `${((stats.matchedFiles / stats.totalFiles) * 100).toFixed(0)}%` : undefined}
                />
                <StatCard
                    label="Unmatched"
                    value={stats.unmatchedFiles}
                    icon={<BiX />}
                    color={stats.unmatchedFiles > 0 ? "text-orange-400" : "text-gray-500"}
                />
                <StatCard
                    label="Issues"
                    value={stats.errorCount + stats.warningCount}
                    icon={<BiError />}
                    color={stats.errorCount > 0 ? "text-red-400" : "text-gray-500"}
                />
            </div>

            <div className="space-y-2">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Pipeline</h3>
                <div className="flex items-center gap-2 flex-wrap">
                    <PipelineStep label="File Discovery" detail={`${stats.totalFiles} files`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Media Fetch" detail={`${stats.fetchedMediaCount} media (${stats.unknownMediaCount} new)`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Token Index" detail={`${stats.tokenIndexSize} tokens`} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Matcher" detail={stats.matcherDuration || "—"} />
                    <BiChevronRight className="text-[--muted] text-lg flex-shrink-0" />
                    <PipelineStep label="Hydrator" detail={stats.hydratorDuration || "—"} />
                </div>
            </div>

            <div className="space-y-2">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Events</h3>
                <Virtuoso
                    style={{ height: "40vh" }}
                    totalCount={lines.length}
                    itemContent={(index) => (
                        <div className="pb-1">
                            <SystemLogLine line={lines[index]} />
                        </div>
                    )}
                />
            </div>
        </div>
    )
}

function StatCard({ label, value, icon, color, sub }: { label: string; value: number; icon: React.ReactNode; color: string; sub?: string }) {
    return (
        <div className="bg-gray-900 border border-[--border] rounded-lg p-3 space-y-1">
            <div className="flex items-center gap-2">
                <span className={cn("text-lg", color)}>{icon}</span>
                <span className="text-sm text-gray-400 font-medium">{label}</span>
            </div>
            <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold text-white">{value}</span>
                {sub && <span className="text-sm text-gray-500">{sub}</span>}
            </div>
        </div>
    )
}

function PipelineStep({ label, detail }: { label: string; detail: string }) {
    return (
        <div className="bg-gray-900 border border-[--border] rounded-md px-3 py-2 text-center min-w-[120px]">
            <p className="text-sm font-semibold text-gray-300">{label}</p>
            <p className="text-sm text-gray-500 mt-0.5">{detail}</p>
        </div>
    )
}

function SystemLogLine({ line }: { line: ParsedLogLine }) {
    const d = line.raw
    return (
        <div
            className={cn(
                "flex items-start gap-2 px-2 py-1 rounded text-sm font-mono",
                line.level === "error" && "bg-red-950/30 text-red-300",
                line.level === "warn" && "bg-orange-950/30 text-orange-300",
                (line.level === "info" || line.level === "debug") && "text-gray-400",
            )}
        >
            <LevelBadge level={line.level} />
            {d.context && <span className="text-indigo-400 flex-shrink-0">[{d.context}]</span>}
            <span className="break-all">{d.message}</span>
            {d.count !== undefined && <span className="text-gray-500 flex-shrink-0">count={d.count}</span>}
            {d.startTime !== undefined &&
                <span className="text-gray-500 flex-shrink-0">start=<span className="text-white">{d.startTime}</span></span>}
            {d.duration !== undefined &&
                <span className="text-gray-500 flex-shrink-0">duration=<span className="text-white">{d.duration}</span></span>}
            {d.ms !== undefined && <span className="text-white flex-shrink-0">{d.ms}ms</span>}
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function ParsingPanel({ lines, searchQuery, setSearchQuery, onSelectFile }: {
    lines: ParsedLogLine[];
    searchQuery: string;
    setSearchQuery: (v: string) => void;
    onSelectFile: (f: string) => void
}) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        if (!searchQuery) return lines
        const q = searchQuery.toLowerCase()
        return lines.filter((l) => l.raw.filename?.toLowerCase().includes(q) || l.raw.path?.toLowerCase().includes(q))
    }, [lines, searchQuery])

    return (
        <div className="p-4 space-y-3">
            <div className="flex gap-2 items-center sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search parsed files..."
                    className="max-w-md"
                />
                <span className="text-sm text-gray-500">{filtered.length} files</span>
            </div>
            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const line = filtered[index]
                    return (
                        <div className="pb-1">
                            <ParsedFileLine
                                line={line}
                                onSelectFile={onSelectFile}
                                isExpanded={expandedIds.has(line.idx)}
                                toggleExpanded={() => toggleExpanded(line.idx)}
                            />
                        </div>
                    )
                }}
            />
        </div>
    )
}

function ParsedFileLine({ line, onSelectFile, isExpanded, toggleExpanded }: {
    line: ParsedLogLine;
    onSelectFile?: (f: string) => void;
    isExpanded?: boolean;
    toggleExpanded?: () => void
}) {
    const [internalExpanded, setInternalExpanded] = useState(false)
    const expanded = isExpanded !== undefined ? isExpanded : internalExpanded
    const handleToggle = () => {
        if (toggleExpanded) toggleExpanded()
        else setInternalExpanded(!internalExpanded)
    }

    const d = line.raw // raw log data
    const pd = d.parsedData || {}
    const hasEpisode = !!pd.episode
    const hasSeason = !!pd.season
    const title = pd.title || ""

    return (
        <div className="bg-gray-900 border border-[--border] rounded-md overflow-hidden">
            <button
                onClick={handleToggle}
                className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
            >
                {expanded ? <BiChevronDown className="text-gray-500 flex-shrink-0" /> : <BiChevronRight className="text-gray-500 flex-shrink-0" />}
                <BiFile className="text-blue-400 flex-shrink-0" />
                <span className="text-sm text-blue-200 font-mono truncate flex-1">{d.filename}</span>
                <div className="flex gap-1 flex-shrink-0">
                    {title && <Badge size="sm" intent="unstyled">{title}</Badge>}
                    {hasSeason && <Badge size="sm" intent="blue">S{pd.season}</Badge>}
                    {hasEpisode && <Badge size="sm" intent="gray">{pd.episode}</Badge>}
                    {/* {pd.releaseGroup && <Badge size="sm" intent="gray">{pd.releaseGroup}</Badge>} */}
                </div>
            </button>
            {expanded && (
                <div className="px-3 py-2 border-t border-[--border] bg-gray-950 space-y-2">
                    {onSelectFile && (
                        <div className="flex justify-start">
                            <Button
                                intent="primary-subtle" size="xs" onClick={(e) => {
                                e.stopPropagation()
                                onSelectFile(d.filename)
                            }}
                            >
                                View full flow
                            </Button>
                        </div>
                    )}
                    <div className="text-sm text-gray-500 font-mono break-all">{d.path}</div>
                    {d.parsedData && (
                        <div className="space-y-1">
                            <p className="text-sm font-semibold text-gray-400">Parsed Data</p>
                            <DataGrid data={d.parsedData} />
                        </div>
                    )}
                    {d.parsedFolderData && d.parsedFolderData.length > 0 && (
                        <div className="space-y-1">
                            <p className="text-sm font-semibold text-gray-400">Folder Data
                                                                               ({d.parsedFolderData.length} level{d.parsedFolderData.length > 1
                                    ? "s"
                                    : ""})</p>
                            {d.parsedFolderData.map((fd: any, i: number) => (
                                <div key={i} className="pl-2 border-l-2 border-gray-700" style={{ marginLeft: `${i * 10}px` }}>
                                    <DataGrid data={fd} />
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function MatcherPanel({
    fileGroups,
    searchQuery,
    setSearchQuery,
    statusFilter,
    setStatusFilter,
    onSelectFile,
}: {
    fileGroups: FileGroup[]
    searchQuery: string
    setSearchQuery: (v: string) => void
    statusFilter: "all" | "matched" | "unmatched" | "errors"
    setStatusFilter: (v: "all" | "matched" | "unmatched" | "errors") => void
    onSelectFile: (f: string) => void
}) {
    const [expandedFiles, setExpandedFiles] = useState<Set<string>>(new Set())

    const toggleExpanded = (filename: string) => {
        setExpandedFiles(prev => {
            const next = new Set(prev)
            if (next.has(filename)) next.delete(filename)
            else next.add(filename)
            return next
        })
    }

    const filtered = useMemo(() => {
        let groups = fileGroups.filter((g) => g.matcherLogs.length > 0)

        // status filter
        if (statusFilter === "matched") groups = groups.filter((g) => !!g.matchResult)
        if (statusFilter === "unmatched") groups = groups.filter((g) => g.isUnmatched || !g.matchResult)
        if (statusFilter === "errors") groups = groups.filter((g) => g.hasError || g.hasWarning)

        // search
        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            groups = groups.filter(
                (g) =>
                    g.filename.toLowerCase().includes(q) ||
                    g.matchResult?.match?.toLowerCase().includes(q) ||
                    String(g.matchResult?.id || "").includes(q),
            )
        }

        return groups
    }, [fileGroups, searchQuery, statusFilter])

    const matchedCount = fileGroups.filter((g) => g.matcherLogs.length > 0 && g.matchResult).length
    const unmatchedCount = fileGroups.filter((g) => g.matcherLogs.length > 0 && (g.isUnmatched || !g.matchResult)).length

    return (
        <div className="p-4 space-y-3">
            <div className="flex flex-col sm:flex-row gap-2 sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search by filename or match..."
                    className="max-w-md"
                />
                <div className="flex gap-1">
                    {([
                        { key: "all" as const, label: "All" },
                        { key: "matched" as const, label: `Matched (${matchedCount})` },
                        { key: "unmatched" as const, label: `Unmatched (${unmatchedCount})` },
                        { key: "errors" as const, label: "Issues" },
                    ]).map(({ key, label }) => (
                        <button
                            key={key}
                            onClick={() => setStatusFilter(key)}
                            className={cn(
                                "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                                statusFilter === key ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                            )}
                        >
                            {label}
                        </button>
                    ))}
                </div>
                <span className="text-sm text-gray-500 self-center">{filtered.length} files</span>
            </div>

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const group = filtered[index]
                    return (
                        <div className="pb-1">
                            <MatcherFileGroup
                                group={group}
                                onSelectFile={onSelectFile}
                                isExpanded={expandedFiles.has(group.filename)}
                                toggleExpanded={() => toggleExpanded(group.filename)}
                            />
                        </div>
                    )
                }}
            />
        </div>
    )
}

function MatcherFileGroup({ group, onSelectFile, isExpanded, toggleExpanded }: {
    group: FileGroup;
    onSelectFile: (f: string) => void;
    isExpanded?: boolean;
    toggleExpanded?: () => void
}) {
    const [internalExpanded, setInternalExpanded] = useState(false)
    const expanded = isExpanded !== undefined ? isExpanded : internalExpanded
    const handleToggle = () => {
        if (toggleExpanded) toggleExpanded()
        else setInternalExpanded(!internalExpanded)
    }

    const mr = group.matchResult

    return (
        <div
            className={cn(
                "bg-gray-900 border rounded-md overflow-hidden",
                group.hasError ? "border-red-800/50" : group.isUnmatched ? "border-orange-800/50" : "border-[--border]",
            )}
        >
            <button
                onClick={handleToggle}
                className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
            >
                {expanded ? <BiChevronDown className="text-gray-500" /> : <BiChevronRight className="text-gray-500" />}

                {mr ? (
                    <BiCheck className="text-green-400 text-lg flex-shrink-0" />
                ) : (
                    <BiX className="text-orange-400 text-lg flex-shrink-0" />
                )}

                <span className="text-[0.8rem] leading-5 text-gray-200 truncate flex-1">{group.filename}</span>

                <div className="flex gap-1.5 items-center flex-shrink-0">
                    {mr && (
                        <>
                            <BiLinkAlt className="text-gray-500" />
                            <span className="text-sm text-indigo-300 font-medium max-w-[200px] truncate">{mr.match}</span>
                            <Badge size="sm" intent="primary">{mr.id}</Badge>
                            <Badge size="sm" intent={mr.score >= 15 ? "success" : mr.score >= 10 ? "warning" : "alert"}>
                                score: {mr.score}
                            </Badge>
                        </>
                    )}
                    {group.isUnmatched && <Badge size="sm" intent="warning">unmatched</Badge>}
                    {group.hasError && <Badge size="sm" intent="alert">error</Badge>}
                    <Badge size="sm" intent="gray">{group.matcherLogs.length} logs</Badge>
                </div>
            </button>

            {expanded && (
                <div className="border-t border-[--border] bg-gray-950">
                    <div className="flex justify-start px-2 pt-2">
                        <Button
                            intent="primary-subtle" size="xs" onClick={(e) => {
                            e.stopPropagation()
                            onSelectFile(group.filename)
                        }}
                        >
                            View full flow
                        </Button>
                    </div>
                    <div className="space-y-0.5 p-2">
                        {group.matcherLogs.map((log) => (
                            <MatcherLogLine key={log.idx} line={log} />
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}

function MatcherLogLine({ line }: { line: ParsedLogLine }) {
    const [showDetail, setShowDetail] = useState(false)
    const d = line.raw
    const msg = d.message || ""

    const isMatch = msg === "Best match found"
    const isNoMatch = msg === "No match found"
    const isComparison = msg === "Comparison"
    const isVariations = msg === "Found title variations"
    const isCandidates = msg === "Found candidates"
    const isMetadata = msg === "Extracted metadata"

    return (
        <div className="group">
            <button
                onClick={() => setShowDetail(!showDetail)}
                className={cn(
                    "flex items-start gap-2 w-full px-2 py-1 text-left rounded text-[0.8rem] leading-5 transition-colors",
                    isMatch && "text-green-300 bg-green-950/20",
                    isNoMatch && "text-orange-400 bg-orange-950/20",
                    isComparison && "text-gray-400",
                    isVariations && "text-blue-300/80",
                    isCandidates && "text-gray-400",
                    isMetadata && "text-gray-500",
                    line.level === "error" && "text-red-400 bg-red-950/20",
                    !isMatch && !isNoMatch && !isComparison && !isVariations && !isCandidates && !isMetadata && line.level !== "error" && "text-gray-400",
                    "hover:bg-gray-800/50",
                )}
            >
                <LevelDot level={line.level} />
                <span className="break-all flex-1">
                    {isMatch && (
                        <>✓ Matched → <span className="text-indigo-300">{d.match}</span> <span className="text-gray-500">[{d.id}]</span> <span
                            className="text-green-400"
                        >(score: {d.score})</span></>
                    )}
                    {isNoMatch && (
                        <>✗ No match found <span className="text-gray-500">(best score: {d.score})</span></>
                    )}
                    {isComparison && d.match && (
                        <>
                            <span className="text-gray-500 mr-1">vs</span>
                            <span className="text-indigo-300">{d.match}</span>
                            <span className="text-[--muted] mx-1">[{d.id}]</span>
                            <span className="text-gray-300">= {d.score}</span>
                            <span className="text-[--muted] ml-1">(title={d.titleScore} base={d.baseTitleScore} season={d.seasonPartScore} year={d.yearScore})</span>
                        </>
                    )}
                    {isVariations && d.titleVariations && (
                        <span className="flex flex-wrap gap-1 items-center">
                            <span className="text-gray-500 mr-0.5">Variations:</span>
                            {(d.titleVariations as string[]).map((v: string, i: number) => (
                                <span key={i} className="text-blue-200 px-1.5 py-0.5 rounded text-sm">{v}</span>
                            ))}
                        </span>
                    )}
                    {isCandidates && (
                        <><span className="text-white">{d.candidates}</span> candidates found</>
                    )}
                    {isMetadata && (
                        <>
                            <span className="text-gray-500">season=</span><span className="text-gray-300">{d.season}</span>{" "}
                            <span className="text-gray-500">part=</span><span className="text-gray-300">{d.part}</span>{" "}
                            <span className="text-gray-500">year=</span><span className="text-gray-300">{d.year}</span>
                        </>
                    )}
                    {!isMatch && !isNoMatch && !isComparison && !isVariations && !isCandidates && !isMetadata && (
                        <>{msg}</>
                    )}
                </span>
            </button>
            {showDetail && (
                <div className="ml-6 px-2 py-2 mb-1 bg-gray-900 rounded text-[0.8rem] leading-5 space-y-2">
                    {isComparison && d.titles && Array.isArray(d.titles) ? (
                        <div className="space-y-2">
                            <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
                                {Object.entries(d).map(([key, value]) => {
                                    if (key === "level" || key === "context" || key === "titles") return null
                                    const isObj = typeof value === "object" && value !== null
                                    return (
                                        <React.Fragment key={key}>
                                            <span className="text-gray-500 text-sm select-all">{key}</span>
                                            <span className={cn("text-sm break-all select-all", isObj ? "text-gray-400" : "text-gray-200")}>
                                                {isObj ? JSON.stringify(value) : String(value)}
                                            </span>
                                        </React.Fragment>
                                    )
                                })}
                            </div>
                            <p className="text-sm font-semibold text-gray-400">Titles ({d.titles.length})</p>
                            <div className="grid gap-1">
                                {(d.titles as any[]).map((t: any, i: number) => (
                                    <div
                                        key={i} className={cn(
                                        "flex flex-wrap items-center gap-x-2 gap-y-0.5 px-2 py-1 rounded text-sm",
                                        t.IsMain ? "bg-indigo-950/30 border border-indigo-800/30" : "bg-gray-800/50",
                                    )}
                                    >
                                        {/* {t.IsMain && <Badge size="sm" intent="primary">main</Badge>} */}
                                        <span className="text-gray-200 font-medium">{t.Original}</span>
                                        {t.Season > 0 && <Badge size="sm" intent="blue">S{t.Season}</Badge>}
                                        {t.Part > 0 && <Badge size="sm" intent="gray">P{t.Part}</Badge>}
                                        {t.Year > 0 && <Badge size="sm" intent="gray">{t.Year}</Badge>}
                                        <span className="text-[--muted]">→ {t.Normalized}</span>
                                        {t.Tokens && <span className="text-[--muted]">[{t.Tokens.join(", ")}]</span>}
                                    </div>
                                ))}
                            </div>
                        </div>
                    ) : (
                        <DataGrid data={d} />
                    )}
                </div>
            )}
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function HydratorPanel({
    fileGroups,
    searchQuery,
    setSearchQuery,
    levelFilter,
    setLevelFilter,
    onSelectFile,
}: {
    fileGroups: FileGroup[]
    searchQuery: string
    setSearchQuery: (v: string) => void
    levelFilter: LogLevel | "all"
    setLevelFilter: (v: LogLevel | "all") => void
    onSelectFile: (f: string) => void
}) {
    const [expandedFiles, setExpandedFiles] = useState<Set<string>>(new Set())

    const toggleExpanded = (filename: string) => {
        setExpandedFiles(prev => {
            const next = new Set(prev)
            if (next.has(filename)) next.delete(filename)
            else next.add(filename)
            return next
        })
    }

    const filtered = useMemo(() => {
        let groups = fileGroups.filter((g) => g.hydratorLogs.length > 0)

        if (levelFilter !== "all") {
            groups = groups.filter((g) => g.hydratorLogs.some((l) => l.level === levelFilter))
        }

        if (searchQuery) {
            const q = searchQuery.toLowerCase()
            groups = groups.filter(
                (g) =>
                    g.filename.toLowerCase().includes(q) ||
                    String(g.hydratorLogs[0]?.raw?.mediaId || "").includes(q),
            )
        }

        return groups
    }, [fileGroups, searchQuery, levelFilter])

    return (
        <div className="p-4 space-y-3">
            <div className="flex flex-col sm:flex-row gap-2 sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search by filename or media ID..."
                    className="max-w-md"
                />
                <div className="flex gap-1">
                    {(["all", "error", "warn", "debug"] as const).map((lvl) => (
                        <button
                            key={lvl}
                            onClick={() => setLevelFilter(lvl)}
                            className={cn(
                                "px-2.5 py-1 text-sm rounded-md font-medium transition-colors",
                                levelFilter === lvl ? "bg-gray-700 text-white" : "text-gray-500 hover:text-gray-300 hover:bg-gray-800",
                            )}
                        >
                            {lvl}
                        </button>
                    ))}
                </div>
                <span className="text-sm text-gray-500 self-center flex-none">{filtered.length} files</span>
            </div>

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const group = filtered[index]
                    return (
                        <div className="pb-1">
                            <HydratorFileGroup
                                group={group}
                                onSelectFile={onSelectFile}
                                isExpanded={expandedFiles.has(group.filename)}
                                toggleExpanded={() => toggleExpanded(group.filename)}
                            />
                        </div>
                    )
                }}
            />
        </div>
    )
}

function HydratorFileGroup({ group, onSelectFile, isExpanded, toggleExpanded }: {
    group: FileGroup;
    onSelectFile: (f: string) => void;
    isExpanded?: boolean;
    toggleExpanded?: () => void
}) {
    const [internalExpanded, setInternalExpanded] = useState(false)
    const expanded = isExpanded !== undefined ? isExpanded : internalExpanded
    const handleToggle = () => {
        if (toggleExpanded) toggleExpanded()
        else setInternalExpanded(!internalExpanded)
    }

    const hr = group.hydrationResult
    const lastLog = group.hydratorLogs[group.hydratorLogs.length - 1]
    const mediaId = lastLog?.raw?.mediaId

    return (
        <div
            className={cn(
                "bg-gray-900 border rounded-md overflow-hidden",
                group.hasError ? "border-red-800/50" : group.hasWarning ? "border-orange-800/50" : "border-[--border]",
            )}
        >
            <button
                onClick={handleToggle}
                className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
            >
                {expanded ? <BiChevronDown className="text-gray-500" /> : <BiChevronRight className="text-gray-500" />}
                <RiFileSettingsFill className="text-cyan-400 flex-shrink-0" />
                <span className="text-[0.8rem] leading-5 text-gray-200 truncate flex-1">{group.filename}</span>

                <div className="flex gap-1.5 items-center flex-shrink-0">
                    {mediaId && <Badge size="sm" intent="primary">{mediaId}</Badge>}
                    {hr && (
                        <>
                            <Badge size="sm" intent={hr.type === "main" ? "success" : hr.type === "special" ? "info" : "info"}>
                                {hr.type || "unknown"}
                            </Badge>
                            <span className="text-sm text-gray-500">{`ep${hr.episode}`}{hr.aniDBEpisode && ` (${hr.aniDBEpisode})`}</span>
                        </>
                    )}
                    {group.hasError && <Badge size="sm" intent="alert">error</Badge>}
                    {group.hasWarning && <Badge size="sm" intent="warning">warning</Badge>}
                </div>
            </button>

            {expanded && (
                <div className="border-t border-[--border] bg-gray-950 p-2 space-y-1">
                    <div className="flex justify-start">
                        <Button
                            intent="primary-subtle" size="xs" onClick={(e) => {
                            e.stopPropagation()
                            onSelectFile(group.filename)
                        }}
                        >
                            View full flow
                        </Button>
                    </div>
                    {group.hydratorLogs.map((log) => (
                        <HydratorLogLine key={log.idx} line={log} />
                    ))}
                </div>
            )}
        </div>
    )
}

function HydratorLogLine({ line }: { line: ParsedLogLine }) {
    const [showDetail, setShowDetail] = useState(false)
    const d = line.raw

    return (
        <div>
            <button
                onClick={() => setShowDetail(!showDetail)}
                className={cn(
                    "flex items-start gap-2 w-full px-2 py-1 text-left rounded text-[0.8rem] leading-5 transition-colors hover:bg-gray-800/50",
                    line.level === "error" && "text-red-400",
                    line.level === "warn" && "text-orange-300",
                    line.level === "debug" && "text-gray-400",
                    line.level === "info" && "text-cyan-300",
                )}
            >
                <LevelDot level={line.level} />
                <span className="break-all flex-1">
                    {d.message}
                    {d.parsed && <> | Parsed: <span className="text-white">"{d.parsed.parsedEpisode || "?"}"</span></>}
                    {d.hydrated && <> → <span className="text-white">Ep: {d.hydrated.episode}</span>
                        <span className="text-white">(AniDB: {d.hydrated.aniDBEpisode || "—"})</span></>}
                    {d.mediaTreeAnalysis?.normalized !== undefined && (
                        d.mediaTreeAnalysis.normalized ? <span className="text-green-400"> [normalized]</span> :
                            <span className="text-orange-400"> [not normalized]</span>
                    )}
                    {d.mediaTreeAnalysis?.error && <span className="text-red-400"> ⚠ {d.mediaTreeAnalysis.error}</span>}
                </span>
            </button>
            {showDetail && (
                <div className="ml-6 px-2 py-2 mb-1 bg-gray-900 rounded text-[0.8rem] leading-5">
                    <DataGrid data={d} />
                </div>
            )}
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function IssuesPanel({ lines, searchQuery, setSearchQuery }: { lines: ParsedLogLine[]; searchQuery: string; setSearchQuery: (v: string) => void }) {
    const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())

    const toggleExpanded = (id: number) => {
        setExpandedIds(prev => {
            const next = new Set(prev)
            if (next.has(id)) next.delete(id)
            else next.add(id)
            return next
        })
    }

    const filtered = useMemo(() => {
        if (!searchQuery) return lines
        const q = searchQuery.toLowerCase()
        return lines.filter((l) =>
            l.raw.filename?.toLowerCase().includes(q) ||
            l.raw.message?.toLowerCase().includes(q) ||
            JSON.stringify(l.raw).toLowerCase().includes(q),
        )
    }, [lines, searchQuery])

    return (
        <div className="p-4 space-y-3">
            <div className="flex gap-2 items-center sticky top-12 z-10 bg-gray-950 py-2">
                <TextInput
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search issues..."
                    className="max-w-md"
                />
                <span className="text-sm text-gray-500">{filtered.length} issues</span>
            </div>

            {filtered.length === 0 && (
                <div className="flex items-center justify-center h-[20vh] text-green-400">
                    <p className="flex items-center gap-2"><BiCheck className="text-xl" /> No issues found</p>
                </div>
            )}

            <Virtuoso
                style={{ height: "calc(100vh - 200px)" }}
                totalCount={filtered.length}
                itemContent={(index) => {
                    const line = filtered[index]
                    return (
                        <div className="pb-1">
                            <IssueLine
                                line={line}
                                isExpanded={expandedIds.has(line.idx)}
                                toggleExpanded={() => toggleExpanded(line.idx)}
                            />
                        </div>
                    )
                }}
            />
        </div>
    )
}

function IssueLine({ line, isExpanded, toggleExpanded }: { line: ParsedLogLine; isExpanded?: boolean; toggleExpanded?: () => void }) {
    const [internalExpanded, setInternalExpanded] = useState(false)
    const expanded = isExpanded !== undefined ? isExpanded : internalExpanded
    const handleToggle = () => {
        if (toggleExpanded) toggleExpanded()
        else setInternalExpanded(!internalExpanded)
    }

    const d = line.raw

    return (
        <div
            className={cn(
                "bg-gray-900 border rounded-md overflow-hidden",
                line.level === "error" ? "border-red-800/50" : "border-orange-800/50",
            )}
        >
            <button
                onClick={handleToggle}
                className="flex items-start gap-2 w-full px-3 py-2 text-left hover:bg-gray-800/50 transition-colors"
            >
                <LevelBadge level={line.level} />
                {d.context && <span className="text-sm text-indigo-400 flex-shrink-0">[{d.context}]</span>}
                {d.filename && <span className="text-sm text-blue-300 font-mono flex-shrink-0 max-w-[250px] truncate">{d.filename}</span>}
                <span
                    className={cn(
                        "text-sm break-all flex-1",
                        line.level === "error" ? "text-red-300" : "text-orange-300",
                    )}
                >
                    {d.message}
                </span>
            </button>
            {expanded && (
                <div className="border-t border-[--border] bg-gray-950 p-3">
                    <DataGrid data={d} />
                </div>
            )}
        </div>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function LevelBadge({ level }: { level: LogLevel }) {
    const intents: Record<string, "alert" | "warning" | "success" | "info" | "gray"> = {
        error: "alert",
        warn: "warning",
        info: "info",
        debug: "gray",
        trace: "gray",
    }
    return <Badge size="sm" intent={intents[level] || "gray"}>{level}</Badge>
}

function LevelDot({ level }: { level: LogLevel }) {
    return (
        <span
            className={cn(
                "inline-block w-1.5 h-1.5 rounded-full mt-1.5 flex-shrink-0",
                level === "error" && "bg-red-400",
                level === "warn" && "bg-orange-400",
                level === "info" && "bg-blue-400",
                level === "debug" && "bg-gray-600",
                level === "trace" && "bg-gray-700",
            )}
        />
    )
}

function DataGrid({ data }: { data: Record<string, any> }) {
    return (
        <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
            {Object.entries(data).map(([key, value]) => {
                if (key === "level" || key === "context") return null
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


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


function FileFlowPanel({ group }: { group: FileGroup }) {
    return (
        <div className="space-y-4">
            {/* Parsing */}
            {group.parsingLog && (
                <FlowSection title="Parsing" icon={<BiFile className="text-blue-400" />}>
                    <ParsedFileLine line={group.parsingLog} />
                </FlowSection>
            )}

            {/* Matcher */}
            {group.matcherLogs.length > 0 && (
                <FlowSection
                    title="Matcher"
                    icon={<BiSearch className="text-indigo-400" />}
                    badge={group.matchResult
                        ? <Badge size="sm" intent="unstyled" className="text-[--green]">→ {group.matchResult.match} [{group.matchResult.id}]
                                                                                        (score: {group.matchResult.score})</Badge>
                        : group.isUnmatched
                            ? <Badge size="sm" intent="warning">unmatched</Badge>
                            : undefined
                    }
                >
                    <div className="space-y-0.5">
                        {group.matcherLogs.map((log) => (
                            <MatcherLogLine key={log.idx} line={log} />
                        ))}
                    </div>
                </FlowSection>
            )}

            {/* Hydrator */}
            {group.hydratorLogs.length > 0 && (
                <FlowSection
                    title="Hydrator"
                    icon={<RiFileSettingsFill className="text-cyan-400" />}
                    badge={group.hydrationResult
                        ? <Badge size="sm" intent={group.hydrationResult.type === "main" ? "success" : "warning"}>{group.hydrationResult.type} →
                                                                                                                                               ep{group.hydrationResult.episode}</Badge>
                        : undefined
                    }
                >
                    <div className="space-y-0.5">
                        {group.hydratorLogs.map((log) => (
                            <HydratorLogLine key={log.idx} line={log} />
                        ))}
                    </div>
                </FlowSection>
            )}

            {group.matcherLogs.length === 0 && group.hydratorLogs.length === 0 && !group.parsingLog && (
                <p className="text-gray-500 text-sm">No logs found for this file across any phase.</p>
            )}
        </div>
    )
}

function FlowSection({ title, icon, badge, children }: { title: string; icon: React.ReactNode; badge?: React.ReactNode; children: React.ReactNode }) {
    return (
        <div className="border border-[--border] rounded-lg overflow-hidden">
            <div className="flex items-center gap-2 px-3 py-2 bg-gray-900">
                {icon}
                <span className="text-sm font-semibold text-gray-200">{title}</span>
                {badge}
            </div>
            <div className="p-2 bg-gray-950">
                {children}
            </div>
        </div>
    )
}
