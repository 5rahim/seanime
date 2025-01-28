"use client"

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { TextInput } from "@/components/ui/text-input"
import json2toml from "json2toml"
import React, { useState } from "react"
import { BiCheck, BiFile, BiSearch, BiSolidStar } from "react-icons/bi"
import { RiFileSettingsFill } from "react-icons/ri"
import { Virtuoso } from "react-virtuoso"


export function ScanLogViewer({ content }: { content: string }) {
    const [lines, setLines] = useState<string[]>([])

    React.useEffect(() => {
        if (content) {
            setLines(content.split("\n"))
        }
    }, [content])

    const [selected, setSelected] = React.useState<string | null>(null)

    const linesDisplayed = React.useMemo(() => {
        if (!selected?.length) return lines
        return lines.filter((line) => line.includes(selected))
    }, [lines, selected])

    return (
        <div className="container mx-auto bg-gray-900 p-4 min-h-[60vh] relative space-y-2">
            <TextInput
                value={selected || ""}
                onChange={(e) => setSelected(e.target.value)}
                placeholder="Search..."
            />
            {selected && (
                <Button intent="white" size="sm" onClick={() => setSelected(null)}>
                    Clear selection
                </Button>
            )}
            <Accordion type={selected ? "multiple" : "single"} collapsible={selected ? undefined : true}>
                {linesDisplayed.length > 0 && (
                    <div className="h-[calc(100vh-150px)]">
                        <Virtuoso
                            data={linesDisplayed}
                            itemContent={(index, line) => (
                                <Line
                                    index={index}
                                    line={line}
                                    onFileSelect={(file) => setSelected(file)}
                                />
                            )}
                            className="h-full"
                        />
                    </div>
                )}
            </Accordion>
        </div>
    )
}

interface LineProps {
    line: string
    index: number
    onFileSelect: (file: string) => void
}

function Line({ line, index, onFileSelect }: LineProps) {

    const [data, setData] = React.useState<Record<string, any> | null>(null)

    React.useEffect(() => {
        try {
            setData(JSON.parse(line) as any)
        }
        catch (e) {
            console.log("Not a JSON", e)
        }
    }, [])

    const isParsedFileLine = data && data.path && data.filename
    const isMediaFetcher = data && data.context === "MediaFetcher"
    const isMatcher = data && data.context === "Matcher"
    const isFileHydrator = data && data.context === "FileHydrator"
    const isNotModule = !isParsedFileLine && !isMediaFetcher && !isMatcher && !isFileHydrator

    if (!data) return <div className="h-1"></div>

    return (
        <AccordionItem value={String(index)}>
            <div
                className={cn(
                    "bg-gray-950 rounded-[--radius-md]",
                )}
            >
                <span className="font-mono text-white break-all">
                    <div
                        className={cn(
                            "mb-2 rounded",
                            //isSelected && "bg-gray-800",
                            //log.status >= 400 && "bg-red-950",
                            //isSelected && log.status >= 400 && "bg-red-900",
                        )}
                    >
                        <AccordionTrigger className="p-2 ">
                            <div className="flex flex-row gap-2">
                                {isParsedFileLine && (
                                    <>
                                        <BiFile className="text-blue-500" />
                                        <p className="text-sm break-all tracking-wide text-blue-200 text-left">
                                            {data.path}
                                        </p>
                                    </>
                                )}
                                {isMediaFetcher && (
                                    <p className="text-sm tracking-wide text-green-100">
                                        {JSON.stringify(data)}
                                    </p>
                                )}
                                {isMatcher && (
                                    <div className="flex flex-col justify-start items-start gap-2">
                                        <div className="flex gap-2">
                                            <BiSearch className="text-indigo-500" />
                                            <p
                                                className={cn(
                                                    "text-sm tracking-wide",
                                                    data.filename && "text-blue-200",
                                                    data.fileRatings && "text-yellow-300",
                                                    data.titleVariations && "text-blue-200 opacity-50",
                                                    data.id && "text-blue-200",
                                                    data.rating && "text-green-200",
                                                    data.match && "text-orange-200 opacity-80",
                                                    (data.rating && data.threshold && data.rating < data.threshold) && "text-red-400",
                                                    data.message?.includes("un-matching") && "text-red-400",
                                                )}
                                            >
                                                {data.filename ? data.filename :
                                                    data.fileRatings ? `${data.fileRatings.length} ratings for ${data.mediaId}` :
                                                        data.message}
                                            </p>
                                        </div>

                                        {data.hasOwnProperty("rating") ? (
                                            <p className="text-sm tracking-wide flex gap-1 items-center ">
                                                <BiSolidStar className="text-[--yellow]" /> {data.rating?.toFixed(2)}{data.highestRating && "/"}{data.highestRating?.toFixed(
                                                2)}, {data.message}
                                            </p>
                                        ) : data.hasOwnProperty("id") ? (
                                            <p className="text-sm tracking-wide flex gap-1 items-center ">
                                                {data.message}
                                                <BiCheck className="text-[--green]" /> <span className="text-indigo-300">{data.title}</span>
                                                <span>[{data.id}]</span>
                                            </p>
                                        ) : data.titleVariations ? (
                                            <p className="text-xs tracking-wide break-words text-left">
                                                {data.titleVariations.length} title variations
                                            </p>
                                        ) : data.match && (
                                            <p className="text-sm tracking-wide flex gap-1 items-center ">
                                                {data.message} <BiCheck className="text-[--green]" />
                                                <span className="text-[--muted]">{data.match.Value}</span>
                                                <span className="flex items-center">
                                                    (<BiSolidStar className="text-[--yellow]" /> {data.match.Rating?.toFixed(2)}{data.match.Distance})
                                                </span>
                                            </p>
                                        )}
                                    </div>
                                )}
                                {isFileHydrator && (
                                    <div className="flex flex-col justify-start items-start gap-2">
                                        <div className="flex gap-2">
                                            <RiFileSettingsFill className="text-cyan-500" />
                                            <p
                                                className={cn(
                                                    "text-sm tracking-wide",
                                                    data.filename && "text-blue-200",
                                                    data.branches && "text-yellow-300",
                                                    data.level === "warn" && "text-orange-300",
                                                    data.level === "error" && "text-red-400",
                                                )}
                                            >
                                                {data.filename ? data.filename :
                                                    data.fileRatings ? `${data.fileRatings.length} ratings for ${data.mediaId}` :
                                                        data.branches ? `${data.branches.length} branches fetched for ${data.mediaId}` :
                                                            data.message}
                                            </p>

                                        </div>
                                        {data.metadata && (
                                            <p className="text-sm tracking-wide text-green-100">
                                                {data.message} {JSON.stringify(data.metadata)}
                                            </p>
                                        )}
                                    </div>
                                )}
                                {isNotModule && (
                                    <p className="text-sm tracking-wide text-yellow-100 opacity-70">
                                        {JSON.stringify(data)}
                                    </p>
                                )}
                            </div>
                        </AccordionTrigger>
                        <AccordionContent className="space-y-2">
                            {data.filename && (
                                <Button intent="white" size="xs" onClick={() => onFileSelect(data.filename)}>
                                    Select file
                                </Button>
                            )}
                            {data.titleVariations && (
                                <div className="text-xs tracking-wide break-words text-left">
                                    {data.titleVariations.map((title: string) => {
                                        return (
                                            <p key={title} className="">
                                                {title}
                                            </p>
                                        )
                                    })}
                                </div>
                            )}
                            {<pre className="text-sm">{json2toml(data, { newlineAfterSection: true, indent: 2 })}</pre>}
                        </AccordionContent>

                    </div>
                </span>
            </div>
        </AccordionItem>
    )
}
