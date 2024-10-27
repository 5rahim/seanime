"use client"
import { Anime_LocalFile, Summary_ScanSummaryFile, Summary_ScanSummaryLog } from "@/api/generated/types"
import { useGetScanSummaries } from "@/api/hooks/scan_summary.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { formatDateAndTimeShort } from "@/lib/server/utils"
import Image from "next/image"
import React from "react"
import { AiFillWarning } from "react-icons/ai"
import { BiCheckCircle, BiInfoCircle, BiXCircle } from "react-icons/bi"
import { BsFileEarmarkExcelFill, BsFileEarmarkPlayFill } from "react-icons/bs"
import { LuFileSearch, LuTextSelect } from "react-icons/lu"
import { TbListSearch } from "react-icons/tb"

export const dynamic = "force-static"

export default function Page() {

    const [selectedSummaryId, setSelectedSummaryId] = React.useState<string | undefined | null>(undefined)

    const { data, isLoading } = useGetScanSummaries()

    React.useEffect(() => {
        if (!!data?.length) {
            setSelectedSummaryId(data[data.length - 1]?.scanSummary?.id)
        }
    }, [data])

    const selectedSummary = React.useMemo(() => {
        const summary = data?.find(summary => summary.scanSummary?.id === selectedSummaryId)
        if (!summary || !summary?.createdAt || !summary?.scanSummary?.id) return undefined
        return {
            createdAt: summary?.createdAt,
            ...summary.scanSummary,
        }
    }, [selectedSummaryId, data])

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper
                className="p-4 sm:p-8 space-y-4"
            >
                <div className="flex justify-between items-center w-full relative">
                    <div>
                        <h2>Scan summaries</h2>
                        <p className="text-[--muted]">
                            View the logs and details of your latest scans
                        </p>
                    </div>
                </div>

                <div className="">
                    {isLoading && <LoadingSpinner />}
                    {(!isLoading && !data?.length) && <div className="p-4 text-[--muted] text-center">No scan summaries available</div>}
                    {!!data?.length && (
                        <div>
                            <Select
                                value={selectedSummaryId || "-"}
                                options={data?.filter(n => !!n.scanSummary)
                                    .map((summary) => ({ label: formatDateAndTimeShort(summary.createdAt!), value: summary.scanSummary!.id || "-" }))
                                    .toReversed()}
                                onValueChange={v => setSelectedSummaryId(v)}
                            />
                            {!!selectedSummary && (
                                <div className="mt-4 space-y-4 rounded-[--radius] ">
                                    <div>
                                        <p className="text-[--muted]">Seanime successfully scanned {selectedSummary.groups?.length} media</p>
                                        {!!selectedSummary?.unmatchedFiles?.length && (
                                            <p className="text-orange-300">{selectedSummary?.unmatchedFiles?.length} file{selectedSummary?.unmatchedFiles?.length > 1
                                                ? "s were "
                                                : " was "}not matched</p>
                                        )}
                                    </div>

                                    {!!selectedSummary?.unmatchedFiles?.length && <div className="space-y-2">
                                        <h5>Unmatched files</h5>
                                        <Accordion type="single" collapsible>
                                            <div className="grid grid-cols-1 gap-4">
                                                {selectedSummary?.unmatchedFiles?.map(file => (
                                                    <ScanSummaryGroupItem file={file} key={file.id} />
                                                ))}
                                            </div>
                                        </Accordion>
                                    </div>}

                                    {!!selectedSummary?.groups?.length && <div>
                                        <h5>Media scanned</h5>

                                        <div className="space-y-4 divide-y">
                                            {selectedSummary?.groups?.map(group => !!group?.files?.length ? (
                                                <div className="space-y-4 pt-4" key={group.id}>
                                                    <div className="flex gap-2">

                                                        <div
                                                            className="w-[5rem] h-[5rem] rounded-[--radius] flex-none object-cover object-center overflow-hidden relative"
                                                        >
                                                            <Image
                                                                src={group.mediaImage}
                                                                alt="banner"
                                                                fill
                                                                quality={80}
                                                                priority
                                                                sizes="20rem"
                                                                className="object-cover object-center"
                                                            />
                                                        </div>

                                                        <div className="space-y-1">
                                                            <SeaLink
                                                                href={`/entry?id=${group.mediaId}`}
                                                                className="font-medium tracking-wide"
                                                            >{group.mediaTitle}</SeaLink>
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

                                                    {group.files.flatMap(n => n.logs).some(n => n?.level === "error") &&
                                                        <p className="text-sm flex gap-1 text-red-300 items-center text-[--muted]">
                                                            <span className="text-base"><BiXCircle className="" /></span> Errors found
                                                        </p>}
                                                    {group.files.flatMap(n => n.logs).some(n => n?.level === "warning") &&
                                                        <p className="text-sm flex gap-1 text-orange-300 items-center text-[--muted]">
                                                            <span className="text-base"><AiFillWarning className="" /></span> Warnings found
                                                        </p>}

                                                    <div>


                                                        <Accordion type="single" collapsible>
                                                            <AccordionItem value="i1">
                                                                <AccordionTrigger className="p-0 dark:hover:bg-transparent text-[--muted] dark:hover:text-white">
                                                                    <span className="inline-flex text-base items-center gap-2"><LuTextSelect /> View
                                                                                                                                                scanner
                                                                                                                                                logs</span>
                                                                </AccordionTrigger>
                                                                <AccordionContent className="p-0 bg-[--paper] border mt-4 rounded-[--radius] overflow-hidden relative">
                                                                    <Accordion type="single" collapsible>
                                                                        <div className="grid grid-cols-1">
                                                                            {group.files.map(file => (
                                                                                <ScanSummaryGroupItem file={file} key={file.id} />
                                                                            ))}
                                                                        </div>
                                                                    </Accordion>
                                                                </AccordionContent>
                                                            </AccordionItem>
                                                        </Accordion>
                                                    </div>
                                                </div>
                                            ) : null)}
                                        </div>
                                    </div>}
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </PageWrapper>
        </>
    )

}

type ScanSummaryFileItem = {
    file: Summary_ScanSummaryFile
}

function ScanSummaryGroupItem(props: ScanSummaryFileItem) {
    const { file } = props

    const hasErrors = file.logs?.some(log => log.level === "error")
    const hasWarnings = file.logs?.some(log => log.level === "warning")

    if (!file.localFile || !file.logs) return null

    return (
        <AccordionItem value={file.localFile.path} className="bg-gray-950 overflow-x-auto">
            <AccordionTrigger
                className="w-full max-w-full py-2.5"
            >
                <div className="space-y-1 line-clamp-1 max-w-full w-full">
                    <p
                        className={cn(
                            "text-left font-normal text-gray-200 text-sm line-clamp-1 w-full flex items-center gap-2",
                            hasErrors && "text-red-300",
                            hasWarnings && "text-orange-300",
                        )}
                    >
                        <span>
                            {hasErrors ? <BsFileEarmarkExcelFill /> :
                                hasWarnings ? <BsFileEarmarkPlayFill /> :
                                    <BsFileEarmarkPlayFill />}
                        </span>
                        {file.localFile.name}</p>
                </div>
            </AccordionTrigger>
            <AccordionContent className="space-y-2 overflow-x-auto">
                <p className="text-sm text-left text-[--muted] italic line-clamp-1 max-w-full">{file.localFile.path}</p>
                <ScanSummaryFileParsedData localFile={file.localFile} />
                {file.logs.map(log => (
                    <ScanSummaryLog key={log.id} log={log} />
                ))}
            </AccordionContent>
        </AccordionItem>
    )

}

function ScanSummaryFileParsedData(props: { localFile: Anime_LocalFile }) {
    const { localFile } = props

    const folderTitles = localFile.parsedFolderInfo?.map(i => i.title).filter(Boolean).map(n => `"${n}"`).join(", ")
    const folderSeasons = localFile.parsedFolderInfo?.map(i => i.season).filter(Boolean).map(n => `"${n}"`).join(", ")
    const folderParts = localFile.parsedFolderInfo?.map(i => i.part).filter(Boolean).map(n => `"${n}"`).join(", ")

    return (
        <div className="flex-none">
            <div className="flex justify-between gap-2 items-center">
                <div className="flex gap-1 items-center">
                    <ul className="text-sm space-y-1 [&>li]:flex-none [&>li]:gap-1 [&>li]:line-clamp-1 [&>li]:flex [&>li]:items-center [&>li>span]:text-[--muted] [&>li>span]:uppercase">
                        <li><TbListSearch className="text-indigo-200" />
                            <span>Title</span> "{localFile.parsedInfo?.title}"{!!folderTitles?.length && `, ${folderTitles}`}</li>
                        <li><TbListSearch className="text-indigo-200" /> <span>Episode</span> "{localFile.parsedInfo?.episode || ""}"</li>
                        <li><TbListSearch className="text-indigo-200" />
                            <span>Season</span> "{localFile.parsedInfo?.season || ""}"{!!folderSeasons?.length && `, ${folderSeasons}`}</li>
                        <li><TbListSearch className="text-indigo-200" />
                            <span>Part</span> "{localFile.parsedInfo?.part || ""}"{!!folderParts?.length && `, ${folderParts}`}</li>
                        <li><TbListSearch className="text-indigo-200" /> <span>Episode Title</span> "{localFile.parsedInfo?.episodeTitle || ""}"</li>
                    </ul>
                </div>
            </div>
        </div>
    )
}

function ScanSummaryLog(props: { log: Summary_ScanSummaryLog }) {
    const { log } = props

    return (
        <div className="">
            <div className="flex justify-between gap-2 items-center w-full">
                <div className="flex gap-1 items-center w-full">
                    <div>
                        {log.level === "info" && <BiInfoCircle className="text-blue-300" />}
                        {log.level === "error" && <BiXCircle className="text-red-300" />}
                        {log.level === "warning" && <BiInfoCircle className="text-orange-300" />}
                    </div>
                    <ScanSummaryLogMessage message={log.message} level={log.level} />
                </div>
            </div>
        </div>
    )
}

function ScanSummaryLogMessage(props: { message: string, level: string }) {
    const { message, level } = props

    if (!message.startsWith("PANIC")) {
        return <div
            className={cn(
                "text-[--muted] hover:text-white text-sm tracking-wide flex-none",
                level === "error" && "text-red-300",
                level === "warning" && "text-orange-300",
            )}
        >{message}</div>
    }

    return (
        <div className="w-full text-sm">
            <p className="text-red-300 text-sm font-bold">Please report this issue on the GitHub repository</p>
            <pre className="p-4">
                {message}
            </pre>
        </div>
    )
}
