import { useServerQuery } from "@/api/client/requests"
import {
    useDeleteLogs,
    useDownloadCPUProfile,
    useDownloadGoRoutineProfile,
    useDownloadMemoryProfile,
    useForceGC,
    useGetLogFilenames,
    useGetMemoryStats,
} from "@/api/hooks/status.hooks"
import { useHandleCopyLatestLogs } from "@/app/(main)/_hooks/logs"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { RowSelectionState } from "@tanstack/react-table"
import React from "react"
import { BiRefresh } from "react-icons/bi"
import { FaCopy, FaMemory, FaMicrochip } from "react-icons/fa"
import { FiDownload, FiTrash2 } from "react-icons/fi"
import { toast } from "sonner"
import { SettingsCard } from "../_components/settings-card"

type LogsSettingsProps = {}

export function LogsSettings(props: LogsSettingsProps) {

    const {} = props

    const [selectedFilenames, setSelectedFilenames] = React.useState<{ name: string }[]>([])
    const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})
    const onSelectChange = React.useCallback((event: DataGridRowSelectedEvent<{ name: string }>) => {
        setSelectedFilenames(event.data)
    }, [])
    const [globalFilter, setGlobalFilter] = React.useState<string>("")

    const { data: filenames, isLoading } = useGetLogFilenames()

    const { mutate: deleteLogs, isPending: isDeleting } = useDeleteLogs()

    const filenamesObj = React.useMemo(() => {
        return filenames?.map(f => ({ name: f })) ?? []
    }, [filenames])


    const columns = React.useMemo(() => defineDataGridColumns<{ name: string }>(() => [
        {
            accessorKey: "name",
            header: "Name",
            cell: info => (
                <LogModal filename={info.getValue<string>()} />
            ),
        },
    ]), [filenamesObj])

    const { handleCopyLatestLogs } = useHandleCopyLatestLogs()

    return (
        <>
            <SettingsCard>

                <div className="pb-3">
                    <Button
                        intent="white-subtle"
                        onClick={handleCopyLatestLogs}
                    >
                        Copy current server logs
                    </Button>
                </div>

                <Select
                    value={globalFilter === "seanime-" ? "seanime-" : globalFilter === "-scan" ? "-scan" : "-"}
                    onValueChange={value => {
                        setGlobalFilter(value === "-" ? "" : value)
                    }}
                    options={[
                        { value: "-", label: "All" },
                        { value: "seanime-", label: "Server" },
                        { value: "-scan", label: "Scanner" },
                    ]}
                />

                {selectedFilenames.length > 0 && (
                    <div className="flex items-center space-x-2">
                        <Button
                            onClick={() => deleteLogs({ filenames: selectedFilenames.map(f => f.name) }, {
                                onSuccess: () => {
                                    setSelectedFilenames([])
                                    setRowSelection({})
                                },
                            })}
                            intent="alert"
                            loading={isDeleting}
                            size="sm"
                        >
                            Delete selected
                        </Button>
                    </div>
                )}

                <DataGrid
                    data={filenamesObj}
                    columns={columns}
                    rowCount={filenamesObj.length}
                    isLoading={isLoading}
                    isDataMutating={isDeleting}
                    rowSelectionPrimaryKey="name"
                    enableRowSelection
                    initialState={{
                        pagination: {
                            pageIndex: 0,
                            pageSize: 5,
                        },
                    }}
                    state={{
                        rowSelection,
                        globalFilter,
                    }}
                    hideGlobalSearchInput
                    hideColumns={[
                        // {
                        //     below: 1000,
                        //     hide: ["number", "scanlator", "language"],
                        // },
                    ]}
                    onRowSelect={onSelectChange}
                    onRowSelectionChange={setRowSelection}
                    onGlobalFilterChange={setGlobalFilter}
                    className=""
                />
            </SettingsCard>

            <MemoryProfilingSettings />
        </>
    )
}

function LogModal(props: { filename: string }) {
    const { filename } = props
    const [open, setOpen] = React.useState(false)

    const { data, isPending } = useServerQuery<string>({
        endpoint: `/api/v1/log/${props.filename}`,
        method: "GET",
        queryKey: ["STATUS-get-log-content", props.filename],
        enabled: open,
    })

    function copyToClipboard() {
        if (!data) {
            return
        }
        navigator.clipboard.writeText(data)
        toast.success("Copied to clipboard")
    }

    return (
        <>
            <p
                onClick={() => setOpen(true)}
                className="cursor-pointer hover:text-[--muted]"
            >{filename}</p>
            <Modal
                open={open}
                title={filename}
                onOpenChange={v => setOpen(v)}
                contentClass="max-w-5xl"
            >

                <Button
                    onClick={copyToClipboard}
                    intent="gray-outline"
                    leftIcon={<FaCopy />}
                    className="w-fit"
                >
                    Copy to clipboard
                </Button>

                {isPending ? <LoadingSpinner /> :
                    <div className="bg-gray-900 rounded-[--radius-md] border max-w-full overflow-x-auto">
                        <pre className="text-md max-h-[40rem] p-2 min-h-12 whitespace-pre-wrap break-all">
                            {data?.split("\n").map((line, i) => (
                                <p
                                    key={i}
                                    className={cn(
                                        "w-full",
                                        i % 2 === 0 ? "bg-gray-800" : "bg-gray-900",
                                        line.includes("|ERR|") && "text-white bg-red-800",
                                        line.includes("|WRN|") && "text-orange-500",
                                        line.includes("|INF|") && "text-blue-200",
                                        line.includes("|TRC|") && "text-[--muted]",
                                    )}
                                >{line}</p>
                            ))}
                        </pre>
                    </div>}
            </Modal>
        </>
    )
}

function MemoryProfilingSettings() {
    const [cpuDuration, setCpuDuration] = React.useState(30)

    const { data: memoryStats, refetch: refetchMemoryStats, isLoading: isLoadingMemoryStats } = useGetMemoryStats()
    const { mutate: forceGC, isPending: isForceGCPending } = useForceGC()
    const { mutate: downloadHeapProfile, isPending: isDownloadingHeap } = useDownloadMemoryProfile()
    const { mutate: downloadAllocsProfile, isPending: isDownloadingAllocs } = useDownloadMemoryProfile()
    const { mutate: downloadGoRoutineProfile, isPending: isDownloadingGoroutine } = useDownloadGoRoutineProfile()
    const { mutate: downloadCPUProfile, isPending: isDownloadingCPU } = useDownloadCPUProfile()

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return "0 B"
        const k = 1024
        const sizes = ["B", "KB", "MB", "GB", "TB"]
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
    }

    const handleRefreshStats = () => {
        refetchMemoryStats()
    }

    const handleForceGC = () => {
        forceGC()
    }

    const handleDownloadHeapProfile = () => {
        downloadHeapProfile({ profileType: "heap" })
    }

    const handleDownloadAllocsProfile = () => {
        downloadAllocsProfile({ profileType: "allocs" })
    }

    const handleDownloadGoRoutineProfile = () => {
        downloadGoRoutineProfile()
    }

    const handleDownloadCPUProfile = () => {
        downloadCPUProfile({ duration: cpuDuration })
    }

    return (
        <SettingsCard title="Profiling">
            <div className="space-y-6">
                <div>
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-medium">Memory Statistics</h3>
                        <div className="flex gap-2">
                            <Button
                                intent="white-subtle"
                                size="sm"
                                leftIcon={<BiRefresh className="text-xl" />}
                                onClick={handleRefreshStats}
                                loading={isLoadingMemoryStats}
                            >
                                Refresh
                            </Button>
                            <Button
                                intent="gray-outline"
                                size="sm"
                                leftIcon={<FiTrash2 />}
                                onClick={handleForceGC}
                                loading={isForceGCPending}
                            >
                                Force GC
                            </Button>
                        </div>
                    </div>

                    {memoryStats && (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">Heap Allocated</div>
                                <div className="text-xl font-medium">{formatBytes(memoryStats.heapAlloc)}</div>
                            </div>
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">Heap In Use</div>
                                <div className="text-xl font-medium">{formatBytes(memoryStats.heapInuse)}</div>
                            </div>
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">Heap System</div>
                                <div className="text-xl font-medium">{formatBytes(memoryStats.heapSys)}</div>
                            </div>
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">Total Allocated</div>
                                <div className="text-xl font-medium">{formatBytes(memoryStats.totalAlloc)}</div>
                            </div>
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">Goroutines</div>
                                <div className="text-xl font-medium">{memoryStats.numGoroutine}</div>
                            </div>
                            <div className="bg-gray-800 p-4 rounded-md">
                                <div className="text-sm text-[--muted]">GC Cycles</div>
                                <div className="text-xl font-medium">{memoryStats.numGC}</div>
                            </div>
                        </div>
                    )}

                    {!memoryStats && !isLoadingMemoryStats && (
                        <div className="text-center py-4 text-[--muted]">
                            Click "Refresh" to load memory statistics
                        </div>
                    )}

                    {isLoadingMemoryStats && (
                        <div className="flex justify-center py-4">
                            <LoadingSpinner />
                        </div>
                    )}
                </div>

                <Separator />

                <div>
                    <div className="space-y-4">
                        <div>
                            <h4 className="text-md font-medium mb-2 flex items-center gap-2">
                                <FaMemory className="text-blue-400" />
                                Memory
                            </h4>
                            <div className="flex flex-wrap gap-2">
                                <Button
                                    intent="gray-subtle"
                                    size="sm"
                                    leftIcon={<FiDownload />}
                                    onClick={handleDownloadHeapProfile}
                                    loading={isDownloadingHeap}
                                >
                                    Heap Profile
                                </Button>
                                <Button
                                    intent="gray-subtle"
                                    size="sm"
                                    leftIcon={<FiDownload />}
                                    onClick={handleDownloadAllocsProfile}
                                    loading={isDownloadingAllocs}
                                >
                                    Allocations Profile
                                </Button>
                                <Button
                                    intent="gray-subtle"
                                    size="sm"
                                    leftIcon={<FiDownload />}
                                    onClick={handleDownloadGoRoutineProfile}
                                    loading={isDownloadingGoroutine}
                                >
                                    Goroutine Profile
                                </Button>
                            </div>
                        </div>

                        <Separator />

                        <div>
                            <h4 className="text-md font-medium mb-2 flex items-center gap-2">
                                <FaMicrochip className="text-green-400" />
                                CPU
                            </h4>
                            <div className="space-y-2">
                                <NumberInput
                                    label="Duration (seconds)"
                                    value={cpuDuration}
                                    onValueChange={(value) => setCpuDuration(value || 30)}
                                    min={1}
                                    max={300}
                                    className="w-32"
                                    size="sm"
                                />
                                <Button
                                    intent="gray-outline"
                                    size="sm"
                                    leftIcon={<FiDownload />}
                                    onClick={handleDownloadCPUProfile}
                                    loading={isDownloadingCPU}
                                >
                                    Download CPU Profile
                                </Button>
                            </div>
                            <p className="text-xs text-[--muted] mt-1">
                                CPU profiling will run for the specified duration (1-300 seconds)
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </SettingsCard>
    )
}
