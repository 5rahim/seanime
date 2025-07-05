import { useServerQuery } from "@/api/client/requests"
import { useDeleteLogs, useGetLogFilenames } from "@/api/hooks/status.hooks"
import { useHandleCopyLatestLogs } from "@/app/(main)/_hooks/logs"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Select } from "@/components/ui/select"
import { RowSelectionState } from "@tanstack/react-table"
import React from "react"
import { FaCopy } from "react-icons/fa"
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
