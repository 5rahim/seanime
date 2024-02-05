"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { Tooltip } from "@/components/ui/tooltip"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { ScanSummary, ScanSummaryFile, ScanSummaryLog } from "@/lib/server/types"
import { formatDateAndTimeShort } from "@/lib/server/utils"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React from "react"
import { BiCheckCircle, BiChevronDown, BiChevronUp, BiInfoCircle, BiXCircle } from "react-icons/bi"
import { LuFileSearch } from "react-icons/lu"
import { PiClockCounterClockwiseFill } from "react-icons/pi"


export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()

    const [selectedSummaryId, setSelectedSummaryId] = React.useState<string | null>(null)

    const { data, isLoading } = useSeaQuery<ScanSummary[] | null>({
        queryKey: ["scan-summaries"],
        endpoint: SeaEndpoints.SCAN_SUMMARIES,
    })

    React.useEffect(() => {
        if (!!data?.length) {
            setSelectedSummaryId(data[data.length - 1].id)
        }
    }, [data])

    const selectSummary = React.useMemo(() => data?.find(summary => summary.id === selectedSummaryId), [selectedSummaryId, data])


    return (
        <div className="p-12 space-y-4">
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>Scan summaries</h2>
                    <p className="text-[--muted]">
                        View the summaries of your 5 latest scans.
                    </p>
                </div>
            </div>

            <div className="border border-[--border] rounded-[--radius] bg-[--paper] text-lg space-y-2 p-4">
                {isLoading && <LoadingSpinner />}
                {(!isLoading && !data?.length) && <div className="p-4 text-[--muted] text-center">No scan summaries available</div>}
                {!!data?.length && (
                    <div>
                        <Select
                            label="Summary"
                            leftIcon={<PiClockCounterClockwiseFill />}
                            value={selectedSummaryId || ""}
                            options={data.map((summary, i) => ({ label: formatDateAndTimeShort(summary.createdAt), value: summary.id })).toReversed()}
                            onChange={e => setSelectedSummaryId(e.target.value)}
                        />
                        {!!selectSummary && (
                            <div className="mt-4 space-y-4 rounded-[--radius] ">
                                <div>
                                    <h5>Media that were scanned</h5>
                                    <p className="text-[--muted]">Seanime successfully scanned {selectSummary.groups.length} anime</p>
                                </div>

                                <div className="space-y-4">
                                    {selectSummary.groups.map(group => (
                                        <div className="border border-[--border] rounded-[--radius] p-4 bg-gray-900 space-y-4">
                                            <div className="flex gap-2">

                                                <div
                                                    className="w-[5rem] h-[5rem] rounded-[--radius] flex-none object-cover object-center overflow-hidden relative"
                                                >
                                                    <Image
                                                        src={group.mediaImage}
                                                        alt={"banner"}
                                                        fill
                                                        quality={80}
                                                        priority
                                                        sizes="20rem"
                                                        className="object-cover object-center"
                                                    />
                                                </div>

                                                <div className="space-y-1">
                                                    <p className="font-medium tracking-wide">{group.mediaTitle}</p>
                                                    <p className="flex gap-1 items-center text-sm text-[--muted]">
                                                        <span className="text-lg">{group.mediaIsInCollection ?
                                                            <BiCheckCircle className="text-green-200" /> :
                                                            <BiXCircle className="text-red-300" />}</span> Anime {group.mediaIsInCollection
                                                        ? "is present"
                                                        : "is not present"} in your AniList collection</p>
                                                    <p className="text-sm flex gap-1 items-center text-[--muted]">
                                                        <span className="text-base"><LuFileSearch className="text-brand-200" /></span>{group.files.length} file{group.files.length > 1 && "s"} scanned
                                                    </p>
                                                </div>

                                            </div>

                                            <div>
                                                <div className="grid grid-cols-1 gap-4">
                                                    {group.files.map(file => (
                                                        <ScanSummaryGroupItem file={file} key={file.id} />
                                                    ))}
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    )

}

type ScanSummaryFileItem = {
    file: ScanSummaryFile
}

function ScanSummaryGroupItem(props: ScanSummaryFileItem) {
    const { file } = props

    const [open, setOpen] = React.useState(false)

    const hasErrors = file.logs.some(log => log.level === "error")
    const hasWarnings = file.logs.some(log => log.level === "warning")

    return (
        <div className="rounded-[--radius] p-3 bg-[--background-color]">
            <div className="flex justify-between gap-2 items-center cursor-pointer" onClick={() => setOpen(p => !p)}>

                <div className="space-y-1">
                    <p
                        className={cn(
                            "font-medium text-base tracking-wide line-clamp-1",
                            hasErrors && "text-red-300",
                            hasWarnings && "text-yellow-300",
                        )}
                    >{file.localFile.name}</p>
                    <Tooltip
                        trigger={
                            <p className="text-sm text-gray-500 italic line-clamp-1">{file.localFile.path}</p>}
                    >
                        {file.localFile.path}
                    </Tooltip>
                </div>

                <div>
                    <IconButton intent="white-basic" icon={!open ? <BiChevronDown /> : <BiChevronUp />} size="sm" />
                </div>
            </div>
            {open && (
                <div className="space-y-2 mt-2 border border-[--border] rounded-[--radius] p-3">
                    {file.logs.map(log => (
                        <ScanSummaryLog key={log.id} log={log} />
                    ))}
                </div>
            )}
        </div>
    )

}

function ScanSummaryLog(props: { log: ScanSummaryLog }) {
    const { log } = props

    return (
        <div className="">
            <div className="flex justify-between gap-2 items-center">
                <div className="flex gap-1 items-center">
                    <div>
                        {log.level === "info" && <BiInfoCircle className="text-blue-300" />}
                        {log.level === "error" && <BiXCircle className="text-red-300" />}
                        {log.level === "warning" && <BiInfoCircle className="text-yellow-300" />}
                    </div>
                    <p
                        className={cn(
                            "text-[--muted] hover:text-white text-sm tracking-wide line-clamp-1",
                            log.level === "error" && "text-red-300",
                            log.level === "warning" && "text-yellow-300",
                        )}
                    >{log.message}</p>
                </div>
            </div>
        </div>
    )
}
