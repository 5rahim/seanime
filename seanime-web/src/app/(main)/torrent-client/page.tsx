import { TorrentClientAction_Variables } from "@/api/generated/endpoint.types"
import { TorrentClient_Torrent, TorrentClient_TorrentStatus } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useGetActiveTorrentList, useGetBuiltInTorrentDetails, useTorrentClientAction } from "@/api/hooks/torrent_client.hooks"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuSeparator, ContextMenuTrigger } from "@/components/ui/context-menu"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Pagination, PaginationEllipsis, PaginationItem, PaginationTrigger } from "@/components/ui/pagination"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { StaticTabs, Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { BiDownArrow, BiFolder, BiFolderOpen, BiPause, BiPlay, BiPlus, BiRefresh, BiRename, BiSearch, BiTrash, BiUpArrow } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { FiChevronDown, FiChevronUp } from "react-icons/fi"
import {
    LuCircleStop,
    LuDownload,
    LuFileCheck2,
    LuFolder,
    LuGauge,
    LuMagnet,
    LuNetwork,
    LuPause,
    LuPlay,
    LuRadioTower,
    LuSettings2,
    LuUpload,
    LuZap,
} from "react-icons/lu"

type StatusFilter = "all" | "downloading" | "seeding" | "paused" | "active" | "inactive"

const getBadgeClass = (value: StatusFilter, isSelected: boolean, count: number) => {
    if (isSelected) {
        return "bg-gray-700 text-white"
    }
    if (count > 0) {
        switch (value) {
            case "downloading":
                return "bg-green-500/15 text-green-400"
            case "seeding":
                return "bg-blue-500/15 text-blue-400"
            case "paused":
                return "bg-gray-500/20 text-gray-300"
            case "active":
                return "bg-[--brand]/15 text-[--brand]"
            case "inactive":
                return "bg-gray-500/20 text-gray-300"
            default:
                return "bg-gray-800 text-[--muted]"
        }
    }
    return "bg-gray-800 text-[--muted] group-hover/filter:bg-gray-700/80 group-hover/filter:text-[--foreground]"
}

function parseSpeed(speedStr: string): number {
    if (!speedStr) return 0
    const match = speedStr.match(/^([\d.]+)\s*([A-Za-z/]+)/)
    if (!match) return 0
    const value = parseFloat(match[1])
    const unit = match[2].toLowerCase()
    if (unit.startsWith("t")) return value * 1024 * 1024 * 1024 * 1024
    if (unit.startsWith("g")) return value * 1024 * 1024 * 1024
    if (unit.startsWith("m")) return value * 1024 * 1024
    if (unit.startsWith("k")) return value * 1024
    return value
}

function formatSpeed(value: number) {
    if (!value) return "0 KB/s"
    const units = ["B/s", "KB/s", "MB/s", "GB/s", "TB/s"]
    const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
    if (index < 0) return "0 B/s"
    return `${(value / Math.pow(1024, index)).toFixed(index > 1 ? 1 : 0)} ${units[index]}`
}

function truncatePath(path: string, maxLength = 35) {
    if (!path) return ""
    let cleanPath = path
    if ((cleanPath.endsWith("/") && cleanPath.length > 1) || (cleanPath.endsWith("\\") && cleanPath.length > 1)) {
        cleanPath = cleanPath.slice(0, -1)
    }
    if (cleanPath.length <= maxLength) return path

    const separator = cleanPath.includes("\\") ? "\\" : "/"
    const parts = cleanPath.split(separator)

    if (parts.length <= 3) {
        const half = Math.floor(maxLength / 2) - 2
        return cleanPath.substring(0, half) + "..." + cleanPath.substring(cleanPath.length - half)
    }

    const firstFolder = parts[0] === "" ? parts[1] : parts[0]
    const prefix = (parts[0] === "" ? separator : "") + firstFolder
    const last = parts[parts.length - 1]

    let rightSide = last
    let idx = parts.length - 2
    const startIdx = parts[0] === "" ? 2 : 1

    while (idx >= startIdx) {
        const nextRight = parts[idx] + separator + rightSide
        const combined = prefix + separator + "..." + separator + nextRight
        if (combined.length <= maxLength) {
            rightSide = nextRight
            idx--
        } else {
            break
        }
    }

    return prefix + separator + "..." + separator + rightSide
}

function getVisiblePages(currentPage: number, totalPages: number) {
    const pages: (number | "ellipsis")[] = []
    const maxVisiblePages = 5

    if (totalPages <= maxVisiblePages) {
        for (let i = 1; i <= totalPages; i++) {
            pages.push(i)
        }
    } else {
        pages.push(1)

        if (currentPage <= 3) {
            for (let i = 2; i <= 4; i++) {
                pages.push(i)
            }
            if (totalPages > 4) {
                pages.push("ellipsis")
                pages.push(totalPages)
            }
        } else if (currentPage >= totalPages - 2) {
            pages.push("ellipsis")
            for (let i = totalPages - 3; i <= totalPages; i++) {
                if (i > 1) pages.push(i)
            }
        } else {
            pages.push("ellipsis")
            pages.push(currentPage - 1)
            pages.push(currentPage)
            pages.push(currentPage + 1)
            pages.push("ellipsis")
            pages.push(totalPages)
        }
    }

    return pages
}

const filters: Array<{ value: StatusFilter, label: string, iconType: React.ElementType }> = [
    { value: "all", label: "All", iconType: LuFolder },
    { value: "downloading", label: "Downloading", iconType: LuDownload },
    { value: "seeding", label: "Seeding", iconType: LuUpload },
    { value: "paused", label: "Paused", iconType: LuPause },
    { value: "active", label: "Active", iconType: LuPlay },
    { value: "inactive", label: "Inactive", iconType: LuCircleStop },
]

export default function Page() {
    const serverStatus = useServerStatus()

    if (serverStatus?.settings?.torrent?.defaultTorrentClient !== "seanime") {
        return <PageWrapper className="p-4 sm:p-8">
            <LuffyError title="Seanime torrent client is not active">
                <p className="max-w-md">Select Seanime as the default torrent client to use this dashboard.</p>
                <SeaLink href="/settings"><Button intent="white">Open settings</Button></SeaLink>
            </LuffyError>
        </PageWrapper>
    }

    return <Dashboard />
}

function Dashboard() {
    const serverStatus = useServerStatus()
    const [filter, setFilter] = React.useState<StatusFilter>("all")
    const [search, setSearch] = React.useState("")
    const [selected, setSelected] = React.useState<Set<string>>(new Set())
    const [focusedHash, setFocusedHash] = React.useState<string>()
    const [addOpen, setAddOpen] = React.useState(false)
    const [moveOpen, setMoveOpen] = React.useState(false)
    const [renameOpen, setRenameOpen] = React.useState(false)
    const [limitsOpen, setLimitsOpen] = React.useState(false)
    const [magnet, setMagnet] = React.useState("")
    const [addDestination, setAddDestination] = React.useState(serverStatus?.settings?.library?.libraryPath ?? "")
    const [moveDestination, setMoveDestination] = React.useState("")
    const [newName, setNewName] = React.useState("")
    const [downloadLimit, setDownloadLimit] = React.useState(String(serverStatus?.settings?.torrent?.seanimeDownloadLimit ?? 0))
    const [uploadLimit, setUploadLimit] = React.useState(String(serverStatus?.settings?.torrent?.seanimeUploadLimit ?? 0))

    React.useEffect(() => {
        if (serverStatus?.settings?.torrent) {
            setDownloadLimit(String(serverStatus.settings.torrent.seanimeDownloadLimit ?? 0))
            setUploadLimit(String(serverStatus.settings.torrent.seanimeUploadLimit ?? 0))
        }
    }, [serverStatus?.settings?.torrent])

    const addLibraryPathSelectionProps = useLibraryPathSelection({
        destination: addDestination,
        setDestination: setAddDestination,
    })

    const moveLibraryPathSelectionProps = useLibraryPathSelection({
        destination: moveDestination,
        setDestination: setMoveDestination,
    })

    const list = useGetActiveTorrentList(true, "", "queue")
    const torrents = list.data ?? []
    const totalDownSpeed = React.useMemo(() => {
        return torrents.reduce((sum, t) => sum + parseSpeed(t.downSpeed), 0)
    }, [torrents])

    const totalUpSpeed = React.useMemo(() => {
        return torrents.reduce((sum, t) => sum + parseSpeed(t.upSpeed), 0)
    }, [torrents])
    const visible = React.useMemo(() => torrents.filter(torrent => {
        const matchesSearch = !search || torrent.name.toLowerCase().includes(search.toLowerCase()) || torrent.hash.includes(search.toLowerCase())
        if (!matchesSearch) return false
        switch (filter) {
            case "all":
                return true
            case "active":
                return torrent.status === "downloading" || torrent.status === "seeding"
            case "inactive":
                return torrent.status === "paused" || torrent.status === "queued"
            default:
                return torrent.status === filter
        }
    }), [torrents, filter, search])

    const [currentPage, setCurrentPage] = React.useState(1)
    const [pageSize, setPageSize] = React.useState(50)

    React.useEffect(() => {
        setCurrentPage(1)
    }, [search, filter])

    const totalPages = Math.ceil(visible.length / pageSize)

    React.useEffect(() => {
        if (currentPage > totalPages && totalPages > 0) {
            setCurrentPage(totalPages)
        }
    }, [currentPage, totalPages])

    const paginatedVisible = React.useMemo(() => {
        return visible.slice((currentPage - 1) * pageSize, currentPage * pageSize)
    }, [visible, currentPage, pageSize])

    React.useEffect(() => {
        setSelected(prev => {
            const next = new Set([...prev].filter(hash => torrents.some(torrent => torrent.hash === hash)))
            if (prev.size === next.size) return prev
            return next
        })
        if (focusedHash && !torrents.some(torrent => torrent.hash === focusedHash)) setFocusedHash(undefined)
    }, [torrents, focusedHash])

    const selectedTorrents = torrents.filter(torrent => selected.has(torrent.hash))
    const single = selectedTorrents.length === 1 ? selectedTorrents[0] : undefined
    const inspectorHash = single?.hash ?? focusedHash
    const details = useGetBuiltInTorrentDetails(inspectorHash)
    const { mutate: openInExplorer } = useOpenInExplorer()
    const action = useTorrentClientAction((variables) => {
        list.refetch()
        if (inspectorHash && variables?.action !== "remove") {
            details.refetch()
        }
        if (variables?.action === "set-limits") {
            setLimitsOpen(false)
        }
    })

    const perform = React.useCallback((payload: TorrentClientAction_Variables) => action.mutate(payload), [action])
    const performSelected = React.useCallback((actionName: string, value?: boolean) => {
        selectedTorrents.forEach(torrent => perform({ hash: torrent.hash, action: actionName, dir: torrent.contentPath, value }))
    }, [perform, selectedTorrents])

    const removeDialog = useConfirmationDialog({
        title: selected.size > 1 ? `Remove ${selected.size} torrents` : "Remove torrent",
        description: "Downloaded files will also be removed. This action cannot be undone.",
        actionIntent: "alert",
        onConfirm: () => performSelected("remove"),
    })

    if (list.isLoading) return <LoadingSpinner />
    if (list.isError) return <PageWrapper className="p-4 sm:p-8"><LuffyError title="Could not load torrents" /></PageWrapper>

    return <PageWrapper className="p-3 sm:p-6 lg:p-8 space-y-4">
        <header className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
            <div>
                <h2>Torrent client</h2>
                <p className="text-[--muted]">Manage downloads running directly in Seanime.</p>
            </div>
            <div className="flex flex-wrap gap-2">
                <Button leftIcon={<LuMagnet />} intent="white" onClick={() => setAddOpen(true)}>Add torrent</Button>
                <Button
                    intent="gray-outline"
                    className="border-gray-600 text-gray-200 hover:border-gray-500 hover:text-white"
                    disabled={torrents.length === 0 || torrents.every(t => t.status === "paused" || t.status === "stopped") || action.isPending}
                    onClick={() => perform({ action: "pause-all" })}
                >Pause all</Button>
                <Button
                    intent="gray-outline"
                    className="border-gray-600 text-gray-200 hover:border-gray-500 hover:text-white"
                    disabled={torrents.length === 0 || !torrents.some(t => t.status === "paused" || t.status === "stopped") || action.isPending}
                    onClick={() => perform({ action: "resume-all" })}
                >Resume all</Button>
            </div>
        </header>
        <div className="grid xl:min-h-[68vh] grid-cols-1 gap-4 xl:grid-cols-[12rem_minmax(0,1fr)]">
            <div className="flex flex-col gap-1 min-w-0">
                <StaticTabs
                    className="flex-wrap w-full xl:flex-col xl:gap-1"
                    triggerClass="rounded-lg text-sm xl:justify-start xl:px-3 xl:py-2 xl:h-auto xl:w-full"
                    pillClass="rounded-lg border-transparent"
                    items={filters.map(item => {
                        const count = torrents.filter(torrent => matchesFilter(torrent.status, item.value)).length
                        return {
                            name: item.label,
                            isCurrent: filter === item.value,
                            onClick: () => setFilter(item.value),
                            iconType: item.iconType,
                            addon: (
                                <span
                                    className={cn(
                                        "ml-2 rounded-full px-1.5 py-[0.5px] text-[10px] font-bold tabular-nums transition-colors",
                                        getBadgeClass(item.value, filter === item.value, count),
                                    )}
                                >{count}</span>
                            ),
                        }
                    })}
                />
                <div className="hidden xl:block mt-auto text-xs text-[--muted] px-3 py-2">
                    {torrents.length} torrents
                </div>
            </div>

            <main className="min-w-0 space-y-4">
                <Card className="p-0 overflow-hidden flex flex-col border border-[--border]">
                    <div className="p-2 sm:p-3 border-b border-[--border] bg-gray-950/20 w-full">
                        <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between w-full">
                            <div className="flex flex-wrap items-center gap-2 flex-shrink-0">
                                <div className="flex items-center gap-1">
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<BiPlay />}
                                            intent="gray-subtle"
                                            disabled={!selected.size || !selectedTorrents.some(t => t.status === "paused" || t.status === "stopped") || action.isPending}
                                            onClick={() => performSelected("resume")}
                                        />}
                                    >
                                        Resume
                                    </Tooltip>
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<BiPause />}
                                            intent="gray-subtle"
                                            disabled={!selected.size || selectedTorrents.every(t => t.status === "paused" || t.status === "stopped") || action.isPending}
                                            onClick={() => performSelected("pause")}
                                        />}
                                    >
                                        Pause
                                    </Tooltip>
                                </div>

                                <span className="mx-1 h-5 w-px bg-[--border]" />

                                <div className="flex items-center gap-1">
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<FiChevronUp />}
                                            intent="gray-subtle"
                                            disabled={!single || torrents.length <= 1 || single.queueIndex === 0 || action.isPending}
                                            onClick={() => single && perform({ hash: single.hash, action: "queue-up" })}
                                        />}
                                    >
                                        Move up
                                    </Tooltip>
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<FiChevronDown />}
                                            intent="gray-subtle"
                                            disabled={!single || torrents.length <= 1 || single.queueIndex === torrents.length - 1 || action.isPending}
                                            onClick={() => single && perform({ hash: single.hash, action: "queue-down" })}
                                        />}
                                    >
                                        Move down
                                    </Tooltip>
                                </div>

                                <span className="mx-1 h-5 w-px bg-[--border]" />

                                <div className="flex items-center gap-1">
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<LuZap />}
                                            intent={selectedTorrents.some(t => t.forceStart) ? "primary-subtle" : "gray-subtle"}
                                            disabled={!selected.size || selectedTorrents.every(t => t.progress === 1) || action.isPending}
                                            onClick={() => performSelected("force-start",
                                                !selectedTorrents.every(t => t.forceStart))}
                                        />}
                                    >
                                        Force start
                                    </Tooltip>
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<BiFolder />}
                                            intent="gray-subtle"
                                            disabled={!single || single.size === "0 B" || action.isPending}
                                            onClick={() => {
                                                if (!single) return
                                                setMoveDestination(single.contentPath)
                                                setMoveOpen(true)
                                            }}
                                        />}
                                    >
                                        Change save path
                                    </Tooltip>
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<BiRefresh />}
                                            intent="gray-subtle"
                                            disabled={!selected.size || selectedTorrents.every(t => t.size === "0 B") || action.isPending}
                                            onClick={() => performSelected("recheck")}
                                        />}
                                    >
                                        Recheck
                                    </Tooltip>
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<LuRadioTower />}
                                            intent="gray-subtle"
                                            disabled={!selected.size || action.isPending}
                                            onClick={() => performSelected("reannounce")}
                                        />}
                                    >
                                        Reannounce
                                    </Tooltip>
                                    <Popover
                                        open={limitsOpen}
                                        onOpenChange={setLimitsOpen}
                                        trigger={
                                            <div>
                                                <Tooltip trigger={<IconButton icon={<LuGauge />} intent="gray-subtle" />}>
                                                    Speed limits
                                                </Tooltip>
                                            </div>
                                        }
                                        className="w-72 space-y-3 p-3"
                                    >
                                        <p>Global speed limits</p>
                                        <TextInput
                                            label="Download (KB/s)"
                                            value={downloadLimit}
                                            onValueChange={setDownloadLimit}
                                            inputMode="numeric"
                                        />
                                        <TextInput label="Upload (KB/s)" value={uploadLimit} onValueChange={setUploadLimit} inputMode="numeric" />
                                        <Button
                                            size="sm" intent="white" className="w-full" disabled={action.isPending} onClick={() => perform({
                                            action: "set-limits",
                                            downloadLimit: Number(downloadLimit) || 0,
                                            uploadLimit: Number(uploadLimit) || 0,
                                        })}
                                        >Apply limits</Button>
                                    </Popover>
                                </div>

                                <span className="mx-1 h-5 w-px bg-[--border]" />

                                <div className="flex items-center gap-1">
                                    <Tooltip
                                        trigger={<IconButton
                                            icon={<BiTrash />}
                                            intent="alert-subtle"
                                            disabled={!selected.size || action.isPending}
                                            onClick={() => removeDialog.open()}
                                        />}
                                    >
                                        Remove
                                    </Tooltip>
                                </div>
                            </div>

                            <div className="flex-1 flex"></div>

                            <div className="hidden lg:flex items-center gap-4 text-xs font-semibold tabular-nums text-[--muted] bg-gray-950/40 px-3 py-1.5 h-10 rounded-xl border border-[--border] whitespace-nowrap flex-shrink-0">
                                <span className="flex items-center gap-1.5 whitespace-nowrap" title="Global download speed">
                                    <BiDownArrow className="text-green-500 flex-shrink-0" />
                                    <span>DL: {formatSpeed(totalDownSpeed)}</span>
                                </span>
                                <span className="mx-0.5 h-3 w-px bg-[--border] flex-shrink-0" />
                                <span className="flex items-center gap-1.5 whitespace-nowrap" title="Global upload speed">
                                    <BiUpArrow className="text-blue-400 flex-shrink-0" />
                                    <span>UL: {formatSpeed(totalUpSpeed)}</span>
                                </span>
                            </div>

                            <TextInput
                                value={search}
                                onValueChange={setSearch}
                                leftIcon={<BiSearch />}
                                placeholder="Filter torrents"
                                className="lg:w-64 flex-shrink-0"
                                fieldClass="w-fit"
                            />
                        </div>
                    </div>

                    <Table className="min-w-[1400px]">
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-10"><Checkbox
                                    value={visible.length > 0 && visible.every(torrent => selected.has(torrent.hash))}
                                    disabled={visible.length === 0}
                                    onValueChange={value => {
                                        setSelected(value === true ? new Set(visible.map(torrent => torrent.hash)) : new Set())
                                    }}
                                /></TableHead>
                                <TableHead className="w-[30%] min-w-80 whitespace-nowrap">Name</TableHead>
                                <TableHead className="w-[8%] min-w-24 whitespace-nowrap">Status</TableHead>
                                <TableHead className="w-[8%] min-w-24 text-right whitespace-nowrap">Size</TableHead>
                                <TableHead className="w-[10%] min-w-28 text-right whitespace-nowrap">Seeds / Peers</TableHead>
                                <TableHead className="text-right w-[10%] min-w-28 whitespace-nowrap"><BiDownArrow className="inline mr-1" />Speed</TableHead>
                                <TableHead className="text-right w-[10%] min-w-28 whitespace-nowrap"><BiUpArrow className="inline mr-1" />Speed</TableHead>
                                <TableHead className="w-[8%] min-w-20 text-right whitespace-nowrap">ETA</TableHead>
                                <TableHead className="w-[6%] min-w-16 text-right whitespace-nowrap">Ratio</TableHead>
                                <TableHead className="w-[12%] min-w-40 whitespace-nowrap">Date added</TableHead>
                                <TableHead className="w-[15%] min-w-60 whitespace-nowrap">Save path</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {paginatedVisible.map(torrent => (
                                <SeaContextMenu
                                    key={torrent.hash}
                                    content={
                                        <ContextMenuGroup>
                                            <ContextMenuLabel className="text-[--muted] line-clamp-1 py-0 my-1 font-semibold text-xs">
                                                {torrent.name}
                                            </ContextMenuLabel>
                                            <ContextMenuSeparator />
                                            {(torrent.status === "paused" || torrent.status === "stopped") && <ContextMenuItem
                                                disabled={torrent.status !== "paused" && torrent.status !== "stopped" || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "resume" })}
                                            >
                                                <BiPlay /> Resume
                                            </ContextMenuItem>}
                                            {!(torrent.status === "paused" || torrent.status === "stopped") && <ContextMenuItem
                                                disabled={action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "pause" })}
                                            >
                                                <BiPause /> Pause
                                            </ContextMenuItem>}
                                            <ContextMenuItem
                                                disabled={torrent.progress === 1 || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "force-start", value: !torrent.forceStart })}
                                            >
                                                <LuZap /> Force start
                                            </ContextMenuItem>
                                            <ContextMenuSeparator />
                                            <ContextMenuItem
                                                disabled={torrent.queueIndex === 0 || torrents.length <= 1 || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "queue-up" })}
                                            >
                                                <FiChevronUp /> Move up
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                disabled={torrent.queueIndex === torrents.length - 1 || torrents.length <= 1 || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "queue-down" })}
                                            >
                                                <FiChevronDown /> Move down
                                            </ContextMenuItem>
                                            <ContextMenuSeparator />
                                            <ContextMenuItem
                                                disabled={action.isPending}
                                                onClick={() => {
                                                    setFocusedHash(torrent.hash)
                                                    setNewName(torrent.name)
                                                    setRenameOpen(true)
                                                }}
                                            >
                                                <BiRename /> Rename
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                disabled={torrent.size === "0 B" || action.isPending}
                                                onClick={() => {
                                                    setFocusedHash(torrent.hash)
                                                    setMoveDestination(torrent.contentPath)
                                                    setMoveOpen(true)
                                                }}
                                            >
                                                <BiFolder /> Change save path
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                disabled={torrent.size === "0 B"}
                                                onClick={() => openInExplorer({ path: torrent.contentPath })}
                                            >
                                                <BiFolderOpen /> Open folder
                                            </ContextMenuItem>
                                            <ContextMenuSeparator />
                                            <ContextMenuItem
                                                disabled={torrent.size === "0 B" || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "recheck" })}
                                            >
                                                <BiRefresh /> Recheck
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                disabled={action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "reannounce" })}
                                            >
                                                <LuRadioTower /> Reannounce
                                            </ContextMenuItem>
                                            <ContextMenuItem
                                                disabled={torrent.progress === 1 || torrent.status === "seeding" || action.isPending}
                                                onClick={() => perform({ hash: torrent.hash, action: "set-sequential", value: !torrent.sequential })}
                                            >
                                                <LuSettings2 /> {torrent.sequential ? "Disable sequential mode" : "Enable sequential mode"}
                                            </ContextMenuItem>
                                            <ContextMenuSeparator />
                                            <ContextMenuItem
                                                className="text-red-500 hover:text-red-600 focus:bg-red-500/10 focus:text-red-500"
                                                disabled={action.isPending}
                                                onClick={() => {
                                                    setSelected(new Set([torrent.hash]))
                                                    setFocusedHash(torrent.hash)
                                                    setTimeout(() => removeDialog.open(), 0)
                                                }}
                                            >
                                                <BiTrash /> Remove
                                            </ContextMenuItem>
                                        </ContextMenuGroup>
                                    }
                                >
                                    <ContextMenuTrigger asChild>
                                        <TorrentRow
                                            torrent={torrent}
                                            selected={selected.has(torrent.hash)}
                                            focused={focusedHash === torrent.hash}
                                            onSelect={value => {
                                                setSelected(prev => toggleSet(prev, torrent.hash, value))
                                                if (value) {
                                                    setFocusedHash(torrent.hash)
                                                }
                                            }}
                                            onFocus={() => setFocusedHash(prev => prev === torrent.hash ? undefined : torrent.hash)}
                                            onRightClick={() => {
                                                setFocusedHash(torrent.hash)
                                                if (!selected.has(torrent.hash)) {
                                                    setSelected(new Set([torrent.hash]))
                                                }
                                            }}
                                        />
                                    </ContextMenuTrigger>
                                </SeaContextMenu>
                            ))}
                        </TableBody>
                    </Table>
                    {!visible.length && <div className="py-16 text-center text-[--muted]">No torrents match this view.</div>}

                    {totalPages > 1 && (
                        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 border-t border-[--border] bg-gray-950/10">
                            <div className="flex items-center gap-2">
                                <Select
                                    value={String(pageSize)}
                                    onValueChange={v => {
                                        setPageSize(Number(v))
                                        setCurrentPage(1)
                                    }}
                                    options={[
                                        { value: "20", label: "20 per page" },
                                        { value: "50", label: "50 per page" },
                                        { value: "100", label: "100 per page" },
                                        { value: "200", label: "200 per page" },
                                    ]}
                                    size="sm"
                                    fieldClass="w-36"
                                    className="w-36"
                                />
                                <span className="text-xs text-[--muted]">
                                    Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, visible.length)} of {visible.length}
                                </span>
                            </div>

                            <Pagination>
                                <PaginationTrigger
                                    direction="previous"
                                    onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                    isDisabled={currentPage <= 1}
                                />
                                {getVisiblePages(currentPage, totalPages).map((page, index) => (
                                    page === "ellipsis" ? (
                                        <PaginationEllipsis key={`ellipsis-${index}`} />
                                    ) : (
                                        <PaginationItem
                                            key={page}
                                            value={page}
                                            onClick={() => setCurrentPage(page as number)}
                                            data-selected={page === currentPage}
                                        />
                                    )
                                ))}
                                <PaginationTrigger
                                    direction="next"
                                    onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                    isDisabled={currentPage >= totalPages}
                                />
                            </Pagination>
                        </div>
                    )}
                </Card>

                {inspectorHash && <Inspector
                    torrent={single ?? torrents.find(torrent => torrent.hash === focusedHash)}
                    details={details.data}
                    isLoading={details.isLoading}
                    perform={perform}
                    isPending={action.isPending}
                    openInExplorer={openInExplorer}
                    onRename={() => {
                        const torrent = single ?? torrents.find(item => item.hash === focusedHash)
                        if (!torrent) return
                        setNewName(torrent.name)
                        setRenameOpen(true)
                    }}
                />}
            </main>
        </div>

        <Modal
            open={addOpen}
            onOpenChange={setAddOpen}
            title="Add torrent"
            description="Add a magnet link to the Seanime torrent client."
            contentClass="max-w-xl"
            footer={<Button
                intent="white" disabled={!magnet || !addDestination || action.isPending} onClick={() => {
                perform({ action: "add-magnet", magnet, dir: addDestination })
                setAddOpen(false)
                setMagnet("")
            }}
            >Start download</Button>}
        >
            <div className="space-y-4">
                <TextInput label="Magnet link" value={magnet} onValueChange={setMagnet} />
                <DirectorySelector
                    name="destination"
                    label="Save path"
                    leftIcon={<FcFolder />}
                    value={addDestination}
                    defaultValue={addDestination}
                    onSelect={setAddDestination}
                    shouldExist={false}
                    libraryPathSelectionProps={addLibraryPathSelectionProps}
                />
            </div>
        </Modal>

        <Modal
            open={moveOpen}
            onOpenChange={setMoveOpen}
            title="Change save path"
            description="Seanime will pause the torrent, move its files, and verify the data."
            contentClass="max-w-xl"
            footer={<Button
                intent="white" disabled={!single || !moveDestination || action.isPending} onClick={() => {
                if (single) perform({ hash: single.hash, action: "move-storage", dir: moveDestination })
                setMoveOpen(false)
            }}
            >Move files</Button>}
        >
            <DirectorySelector
                name="destination"
                label="New save path"
                leftIcon={<FcFolder />}
                value={moveDestination}
                defaultValue={moveDestination}
                onSelect={setMoveDestination}
                shouldExist={false}
                libraryPathSelectionProps={moveLibraryPathSelectionProps}
            />
        </Modal>

        <Modal
            open={renameOpen}
            onOpenChange={setRenameOpen}
            title="Rename torrent"
            contentClass="max-w-lg"
            footer={<Button
                intent="white" disabled={!focusedHash || !newName || action.isPending} onClick={() => {
                if (focusedHash) perform({ hash: focusedHash, action: "rename", name: newName })
                setRenameOpen(false)
            }}
            >Rename</Button>}
        >
            <TextInput label="Display name" value={newName} onValueChange={setNewName} />
        </Modal>

        <ConfirmationDialog {...removeDialog} />
    </PageWrapper>
}

const TorrentRow = React.forwardRef<HTMLTableRowElement, {
    torrent: TorrentClient_Torrent
    selected: boolean
    focused: boolean
    onSelect: (value: boolean) => void
    onFocus: () => void
    onRightClick: (event: React.MouseEvent) => void
} & Omit<React.ComponentPropsWithoutRef<"tr">, "onSelect">>((props, ref) => {
    const { torrent, selected, focused, onSelect, onFocus, onRightClick, ...rest } = props
    return <TableRow
        ref={ref}
        data-state={selected ? "selected" : undefined}
        className={cn(
            "cursor-pointer transition-all border-l-2",
            focused ? "bg-gray-800/80 border-l-[--brand]" : "border-l-transparent",
        )}
        onClick={onFocus}
        onContextMenu={(e) => {
            onRightClick(e)
            rest.onContextMenu?.(e)
        }}
        {...rest}
    >
        <TableCell onClick={event => event.stopPropagation()}><Checkbox
            value={selected}
            onValueChange={value => onSelect(value === true)}
        /></TableCell>
        <TableCell>
            <div className="max-w-md">
                <div className="truncate font-medium" title={torrent.name}>{torrent.name}</div>
                <div className="mt-2 h-1 overflow-hidden rounded-full bg-gray-700/70">
                    <div
                        className={cn("h-full bg-[--brand]",
                            torrent.status === "seeding" && "bg-blue-400",
                            torrent.status === "paused" && "bg-gray-500",
                            torrent.status === "error" && "bg-red-500")} style={{ width: `${Math.max(0, Math.min(100, torrent.progress * 100))}%` }}
                    />
                </div>
                <div className="mt-1 text-xs text-[--muted] flex items-center gap-1">
                    <span>{(torrent.progress * 100).toFixed(1)}%</span>
                    <span>·</span>
                    {torrent.error ? (
                        <span className="text-red-400 font-medium" title={torrent.error}>
                            {torrent.error}
                        </span>
                    ) : (
                        <span title={torrent.contentPath} className="cursor-help hover:text-[--foreground] transition-colors">
                            {truncatePath(torrent.contentPath, 35)}
                        </span>
                    )}
                </div>
            </div>
        </TableCell>
        <TableCell><Status status={torrent.status} forceStart={torrent.forceStart} /></TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.size}</TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.seeds} / {torrent.peers}</TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.downSpeed}</TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.upSpeed}</TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.eta}</TableCell>
        <TableCell className="text-right whitespace-nowrap tabular-nums">{torrent.ratio.toFixed(2)}</TableCell>
        <TableCell className="whitespace-nowrap">{torrent.addedAt ? new Date(torrent.addedAt).toLocaleString() : "Unknown"}</TableCell>
        <TableCell className="max-w-80 truncate cursor-help hover:text-[--foreground] transition-colors" title={torrent.contentPath}>
            {truncatePath(torrent.contentPath, 45)}
        </TableCell>
    </TableRow>
})
TorrentRow.displayName = "TorrentRow"

function Inspector(props: {
    torrent?: TorrentClient_Torrent
    details?: ReturnType<typeof useGetBuiltInTorrentDetails>["data"]
    isLoading: boolean
    perform: (payload: TorrentClientAction_Variables) => void
    onRename: () => void
    isPending: boolean
    openInExplorer: (payload: { path: string }) => void
}) {
    const { torrent, details, isLoading, perform, onRename, isPending, openInExplorer } = props
    const [tracker, setTracker] = React.useState("")

    if (!torrent) return <Card className="py-10 text-center text-[--muted]">Select a torrent to inspect files, trackers, and peers.</Card>
    if (isLoading && !details) return <Card className="py-10"><LoadingSpinner /></Card>

    return <Card className="p-0 overflow-hidden">
        <div className="flex flex-col gap-3 border-b border-[--border] px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
                <h5 className="truncate text-sm tracking-wise font-semibold">{torrent.name}</h5>
                <p className="truncate text-xs text-[--muted]">{torrent.hash}</p>
            </div>
            <div className="flex flex-wrap gap-2">
                <Button
                    size="xs"
                    intent="gray-outline"
                    leftIcon={<BiFolderOpen />}
                    disabled={isPending}
                    onClick={() => openInExplorer({ path: torrent.contentPath })}
                >Open folder</Button>
                <Button size="xs" intent="gray-outline" leftIcon={<BiRename />} disabled={isPending} onClick={onRename}>Rename</Button>
                <Button
                    size="xs"
                    intent={torrent.sequential ? "primary" : "gray-outline"}
                    disabled={torrent.progress === 1 || torrent.status === "seeding" || isPending}
                    onClick={() => perform({ hash: torrent.hash, action: "set-sequential", value: !torrent.sequential })}
                >Sequential</Button>
            </div>
        </div>
        <Tabs defaultValue="general">
            <TabsList className="justify-start overflow-x-auto overflow-y-hidden border-b border-[--border] px-2">
                <TabsTrigger value="general"><LuSettings2 className="mr-2" />General</TabsTrigger>
                <TabsTrigger value="files"><LuFileCheck2 className="mr-2" />Files</TabsTrigger>
                <TabsTrigger value="trackers"><LuRadioTower className="mr-2" />Trackers</TabsTrigger>
                <TabsTrigger value="peers"><LuNetwork className="mr-2" />Peers</TabsTrigger>
            </TabsList>
            <TabsContent value="general" className="p-4">
                <dl className="grid grid-cols-2 gap-x-6 gap-y-4 text-sm md:grid-cols-4">
                    <Metric label="Progress" value={`${(torrent.progress * 100).toFixed(1)}%`} />
                    <Metric label="Status" value={torrent.status} />
                    <Metric label="Downloaded" value={formatBytes(details?.torrent?.downloaded ?? 0)} />
                    <Metric label="Uploaded" value={formatBytes(details?.torrent?.uploaded ?? 0)} />
                    <Metric label="Save path" value={torrent.contentPath} wide disableCapitalize />
                    <Metric label="Date added" value={torrent.addedAt ? new Date(torrent.addedAt).toLocaleString() : "Unknown"} disableCapitalize />
                    <Metric label="Queue position" value={String(torrent.queueIndex + 1)} disableCapitalize />
                </dl>
            </TabsContent>
            <TabsContent value="files" className="p-0">
                <Table>
                    <TableHeader><TableRow><TableHead>File</TableHead><TableHead className="text-right whitespace-nowrap">Progress</TableHead><TableHead
                        className="text-right whitespace-nowrap"
                    >Size</TableHead><TableHead>Priority</TableHead></TableRow></TableHeader>
                    <TableBody>{details?.files?.map(file => <TableRow key={file.index}>
                        <TableCell className="max-w-xl break-all">{file.path}</TableCell>
                        <TableCell className="text-right whitespace-nowrap tabular-nums">{(file.progress * 100).toFixed(1)}%</TableCell>
                        <TableCell className="text-right whitespace-nowrap tabular-nums">{formatBytes(file.length)}</TableCell>
                        <TableCell>
                            <div className="flex gap-1">
                                {[0, 1, 2].map(priority => (
                                    <Button
                                        key={priority}
                                        size="xs"
                                        intent={file.priority === priority ? "primary" : "gray-outline"}
                                        disabled={file.progress === 1 || isPending}
                                        onClick={() => perform({ hash: torrent.hash, action: "set-file-priority", index: file.index, priority })}
                                    >
                                        {["Skip", "Normal", "High"][priority]}
                                    </Button>
                                ))}
                            </div>
                        </TableCell>
                    </TableRow>)}</TableBody>
                </Table>
                {!details?.files?.length && <div className="p-8 text-center text-[--muted]">Waiting for torrent metadata.</div>}
            </TabsContent>
            <TabsContent value="trackers" className="p-4 space-y-3">
                <div className="flex gap-2"><TextInput
                    value={tracker}
                    onValueChange={setTracker}
                    placeholder="https://tracker.example/announce"
                /><Button
                    intent="white" leftIcon={<BiPlus />} disabled={!tracker || isPending} onClick={() => {
                    perform({ hash: torrent.hash, action: "add-tracker", tracker })
                    setTracker("")
                }}
                >Add</Button></div>
                <div className="divide-y divide-[--border]">{details?.trackers?.map(item => <div
                    key={item}
                    className="flex items-center justify-between gap-3 py-2 text-sm"
                ><span className="break-all">{item}</span><Tooltip
                    trigger={<IconButton
                        icon={<BiTrash />}
                        size="xs"
                        intent="alert-subtle"
                        disabled={isPending}
                        onClick={() => perform({
                            hash: torrent.hash,
                            action: "remove-tracker",
                            tracker: item,
                        })}
                    />}
                >Remove tracker</Tooltip></div>)}</div>
                {!details?.trackers?.length && <div className="py-6 text-center text-[--muted]">No trackers are listed.</div>}
            </TabsContent>
            <TabsContent value="peers" className="p-0">
                <Table><TableHeader><TableRow><TableHead>Address</TableHead><TableHead>Client</TableHead></TableRow></TableHeader><TableBody>{details?.peers?.map(
                    (peer, index) =>
                        <TableRow key={`${peer.address}-${index}`}><TableCell>{peer.address || "Unknown"}</TableCell><TableCell>{peer.client || "Unknown"}</TableCell></TableRow>)}</TableBody></Table>
                {!details?.peers?.length && <div className="p-8 text-center text-[--muted]">No connected peers.</div>}
            </TabsContent>
        </Tabs>
    </Card>
}

function Metric({ label, value, wide, disableCapitalize }: { label: string, value: string, wide?: boolean, disableCapitalize?: boolean }) {
    return <div className={cn(wide && "col-span-2")}>
        <dt className="text-xs text-[--muted]">{label}</dt>
        <dd className={cn("mt-1 break-all", !disableCapitalize && "capitalize")}>{value}</dd>
    </div>
}

function Status({ status, forceStart }: { status: TorrentClient_TorrentStatus, forceStart: boolean }) {
    return <span
        className={cn(
            "inline-flex items-center rounded-full px-2 py-1 text-xs capitalize",
            status === "downloading" && "bg-green-500/15 text-green-300",
            status === "seeding" && "bg-blue-500/15 text-blue-300",
            status === "paused" && "bg-gray-500/20 text-gray-300",
            status === "queued" && "bg-orange-500/15 text-orange-200",
            status === "error" && "bg-red-500/15 text-red-300",
        )}
    >{forceStart ? "Forced" : status}</span>
}

function matchesFilter(status: TorrentClient_TorrentStatus, filter: StatusFilter) {
    if (filter === "all") return true
    if (filter === "active") return status === "downloading" || status === "seeding"
    if (filter === "inactive") return status === "paused" || status === "queued" || status === "error"
    return status === filter
}

function toggleSet(source: Set<string>, value: string, enabled: boolean) {
    const next = new Set(source)
    enabled ? next.add(value) : next.delete(value)
    return next
}

function formatBytes(value: number) {
    if (!value) return "0 B"
    const units = ["B", "KB", "MB", "GB", "TB"]
    const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
    return `${(value / Math.pow(1024, index)).toFixed(index > 1 ? 1 : 0)} ${units[index]}`
}
