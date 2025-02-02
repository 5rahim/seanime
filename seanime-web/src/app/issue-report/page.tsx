"use client"

import { Report_IssueReport } from "@/api/generated/types"
import { ScanLogViewer } from "@/app/scan-log-viewer/scan-log-viewer"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { TextInput } from "@/components/ui/text-input"
import { format, isSameSecond, parseISO } from "date-fns"
import { max, min } from "lodash"
import React, { useLayoutEffect, useRef, useState } from "react"
import { FiMousePointer } from "react-icons/fi"
import { HiServerStack } from "react-icons/hi2"
import { LuBrain, LuNetwork, LuTerminal } from "react-icons/lu"


export default function Page() {
    const [issueReport, setIssueReport] = useState<Report_IssueReport | null>(null)
    const [currentTime, setCurrentTime] = useState<Date | null>(null)
    const [startTime, setStartTime] = useState<Date | null>(null)
    const [endTime, setEndTime] = useState<Date | null>(null)
    const fileInputRef = useRef<HTMLInputElement>(null)
    const [errorTimestamps, setErrorTimestamps] = useState<Date[]>([])

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (file) {
            const reader = new FileReader()
            reader.onload = (e) => {
                const content = e.target?.result as string
                const parsedReport = JSON.parse(content) as Report_IssueReport
                setIssueReport(parsedReport)

                const allTimestamps = [
                    ...(parsedReport.clickLogs?.map(log => parseISO(log.timestamp!)) || []),
                    ...(parsedReport.networkLogs?.map(log => parseISO(log.timestamp!)) || []),
                    ...(parsedReport.reactQueryLogs?.map(log => parseISO(log.timestamp!)) || []),
                    ...(parsedReport.consoleLogs?.map(log => parseISO(log.timestamp!)) || []),
                    // ...parsedReport.serverLogs.split("\n").map(log => parseISO(log.substring(0, 19))),
                ].filter(date => !isNaN(date.getTime()))

                const earliestTimestamp = min(allTimestamps)
                const latestTimestamp = max(allTimestamps)
                setEndTime(latestTimestamp ?? null)
                setStartTime(earliestTimestamp ?? null)
                setCurrentTime(earliestTimestamp ?? null)

                let errorTimestamps: Date[] = []
                if (parsedReport.serverLogs) {
                    errorTimestamps = [
                        ...parsedReport.serverLogs?.split("\n")?.map(log => ({ timestamp: parseISO(log.substring(0, 19)), text: log })),
                    ].filter(log => log.text.includes("|ERR|")).map(log => log.timestamp)
                }
                errorTimestamps = [...errorTimestamps,
                    ...(parsedReport.networkLogs || []).filter(log => log.status >= 400).map(log => parseISO(log.timestamp!))]
                setErrorTimestamps(errorTimestamps)
            }
            reader.readAsText(file)
        }
    }

    const handleSliderChange = (value: number) => {
        if (issueReport && endTime && startTime) {
            const newTime = new Date(startTime.getTime() + (value * 1000))
            setCurrentTime(newTime)
        }
    }

    const filterLogsByTime = <T extends { timestamp?: string }>(logs: T[]): T[] => {
        if (!currentTime) return []
        return logs.filter((log) => parseISO(log.timestamp!) <= currentTime || isSameSecond(parseISO(log.timestamp!), currentTime))
    }

    const filterServerLogs = (logs: string): string[] => {
        if (!currentTime || !logs) return []
        return logs
            .split("\n")
            .filter((log) => {
                const logTime = parseISO(log.substring(0, 19))
                return logTime <= currentTime
            })
    }

    return (
        <div className="container mx-auto bg-gray-900 p-4 min-h-screen relative pb-40 space-y-4">
            <h1 className="text-3xl font-bold mb-6 text-brand-300 text-center">Issue Report Viewer</h1>
            <div className="container max-w-2xl">
                <TextInput
                    type="file"
                    ref={fileInputRef}
                    onChange={handleFileChange}
                    className="mb-6 p-1"
                />
            </div>
            {issueReport && currentTime && endTime && (
                <div className="relative">
                    <TimelineSlider
                        startTime={startTime}
                        endTime={endTime}
                        currentTime={currentTime}
                        errorTimestamps={errorTimestamps}
                        onChange={handleSliderChange}
                    />
                    <div className="space-y-6">
                        <LogConsole
                            title="Server Logs"
                            logs={filterServerLogs(issueReport.serverLogs || "")}
                            icon={<HiServerStack className="w-5 h-5" />}
                            type="server"
                            currentTime={currentTime}
                        />
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <LogConsole
                                title="Click Logs"
                                logs={filterLogsByTime(issueReport.clickLogs || [])}
                                icon={<FiMousePointer className="w-5 h-5" />}
                                type="click"
                                currentTime={currentTime}
                            />
                            <LogConsole
                                title="Network Logs"
                                logs={filterLogsByTime(issueReport.networkLogs || [])}
                                icon={<LuNetwork className="w-5 h-5" />}
                                type="network"
                                currentTime={currentTime}
                            />
                            <LogConsole
                                title="React Query Logs"
                                logs={filterLogsByTime(issueReport.reactQueryLogs || [])}
                                icon={<LuBrain className="w-5 h-5" />}
                                type="reactQuery"
                                currentTime={currentTime}
                            />
                            <LogConsole
                                title="Console Logs"
                                logs={filterLogsByTime(issueReport.consoleLogs || [])}
                                icon={<LuTerminal className="w-5 h-5" />}
                                type="console"
                                currentTime={currentTime}
                            />
                        </div>
                        {}
                    </div>
                </div>
            )}
            {issueReport && issueReport.scanLogs && (
                <div className="bg-gray-950 p-4 rounded-lg shadow-md mb-6">
                    <h2 className="text-lg font-semibold">Scan Logs</h2>
                    <div className="flex gap-2">
                        {issueReport.scanLogs.map((log, index) => (
                            <Drawer
                                title={`${index}`}
                                size="full"
                                trigger={<Button intent="gray-outline" className="mt-2">{index}</Button>}
                                key={index}
                            >
                                <ScanLogViewer content={log} />
                            </Drawer>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}

interface TimelineSliderProps {
    startTime: Date | null
    endTime: Date | null
    currentTime: Date | null
    errorTimestamps: Date[]
    onChange: (value: number) => void
}

function TimelineSlider({
    startTime,
    endTime,
    currentTime,
    errorTimestamps,
    onChange,
}: TimelineSliderProps) {
    const sliderRef = useRef<HTMLInputElement>(null)
    const [sliderWidth, setSliderWidth] = useState(0)

    useLayoutEffect(() => {
        if (sliderRef.current) {
            setSliderWidth(sliderRef.current.offsetWidth)
        }
    }, [sliderRef.current])

    if (!startTime || !endTime || !currentTime) {
        return <div className="text-red-500">Timeline data is not available</div>
    }

    const totalSeconds = Math.floor((endTime.getTime() - startTime.getTime()) / 1000)
    const currentSeconds = Math.floor((currentTime.getTime() - startTime.getTime()) / 1000)

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        onChange(parseInt(event.target.value, 10))
    }

    // const errorPositions = errorTimestamps.map(timestamp => {
    //     const errorSeconds = Math.floor((timestamp.getTime() - startTime.getTime()) / 1000)
    //     return (errorSeconds / totalSeconds) * sliderWidth
    // })

    const errorSeconds = errorTimestamps.map(timestamp => Math.ceil((timestamp.getTime() - startTime.getTime()) / 1000))

    return (
        <div className="px-4 bottom-0 left-0 fixed w-full">
            <div className="px-4 mb-8 bg-gray-950 p-4 rounded-lg shadow-md w-full ">
                <input
                    ref={sliderRef}
                    type="range"
                    min="0"
                    max={totalSeconds}
                    value={currentSeconds}
                    onChange={handleChange}
                    className="w-full h-2 bg-[--border] rounded-lg appearance-none cursor-pointer"
                />
                {/*{errorPositions.map((position, index) => (*/}
                {/*    <div*/}
                {/*        key={index}*/}
                {/*        className="absolute top-0 h-2 w-1 bg-red-500"*/}
                {/*        style={{ left: `${position}px` }}*/}
                {/*    />*/}
                {/*))}*/}
                <div
                    style={{
                        display: "grid",
                        gridTemplateColumns: `repeat(${totalSeconds}, 1fr)`,
                        gap: "1px",
                        height: "4px",
                        width: "100%",
                    }}
                >
                    {Array.from({ length: totalSeconds }).map((_, index) => (
                        <div
                            key={index}
                            className={cn(
                                "bg-gray-800",
                                index === currentSeconds && "bg-gray-600",
                                errorSeconds.includes(index) && "bg-red-500",
                            )}
                        />
                    ))}
                </div>
                {/*<div*/}
                {/*    style={{*/}
                {/*        background: `linear-gradient(to right, var(--brand) ${currentSeconds / totalSeconds * 100}%, var(--border) ${currentSeconds / totalSeconds * 100}%)`,*/}
                {/*        height: "2px",*/}
                {/*    }}*/}
                {/*/>*/}
                <div className="flex justify-between text-sm text-[--muted] mt-2">
                    <span>{format(startTime, "HH:mm:ss")}</span>
                    <span className="font-semibold text-[--brand]">{format(currentTime, "HH:mm:ss")}</span>
                    <span>{format(endTime, "HH:mm:ss")}</span>
                </div>
            </div>
        </div>
    )
}

interface LogConsoleProps {
    title: string
    logs: any[] | undefined
    icon: React.ReactNode
    type: "server" | "click" | "network" | "reactQuery" | "console"
    currentTime: Date | null
}

function LogConsole({ title, logs = [], icon, type, currentTime }: LogConsoleProps) {

    const listRef = useRef<HTMLDivElement>(null)

    React.useEffect(() => {
        if (listRef.current) {
            listRef.current.scrollTop = listRef.current.scrollHeight
        }
    }, [logs])

    const renderLog = (log: any, index: number, isSelected: boolean) => {
        switch (type) {
            case "server":
                return (
                    <div
                        key={index} className={cn(
                        "mb-2 p-2 rounded bg-gray-900",
                        isSelected && "bg-gray-800",
                        log.includes("|ERR|") && "bg-red-950",
                        log.includes("|WRN|") && "bg-yellow-950",
                        isSelected && log.includes("|ERR|") && "bg-red-900",
                        isSelected && log.includes("|WRN|") && "bg-yellow-900",
                    )}
                    >
                        <span className="text-xs font-mono text-white break-all">{log}</span>
                    </div>
                )
            case "click":
                return (
                    <div
                        key={index} className={cn(
                        "mb-2 p-2 rounded bg-gray-900",
                        isSelected && "bg-gray-800",
                    )}
                    >
                        <p className="font-semibold ">{log.element}</p>
                        {!!log.text?.length && <p className="text-sm italic text-[--brand]">"{log.text.slice(0, 100)}"</p>}
                        <p className="text-xs ">{log.pageUrl}</p>
                        <p className="text-xs ">{log.timestamp}</p>
                    </div>
                )
            case "network":
                return (
                    <div
                        key={index} className={cn(
                        "mb-2 p-2 rounded bg-gray-900",
                        isSelected && "bg-gray-800",
                        log.status >= 400 && "bg-red-950",
                        isSelected && log.status >= 400 && "bg-red-900",
                    )}
                    >
                        <Accordion type="single" collapsible>
                            <AccordionItem value="v">
                                <AccordionTrigger className="p-0">
                                    <div className="flex flex-col justify-start items-start">
                                        <p
                                            className={cn(
                                                "font-semibold text-md",
                                                log.method === "POST" && "text-[--green]",
                                            )}
                                        >{log.method} <span className="font-normal text-[--brand]">{log.url}</span></p>
                                        <p className="text-xs ">{log.pageUrl}</p>
                                        <p className="text-xs">Status: {log.status}, Time: {log.timestamp}</p>
                                    </div>
                                </AccordionTrigger>
                                <AccordionContent className="space-y-2">
                                    <p className="text-xs font-mono break-all">Body: {log.body}</p>
                                    <p className="text-xs font-mono break-all">Data: {log.dataPreview}</p>
                                </AccordionContent>
                            </AccordionItem>

                        </Accordion>
                    </div>
                )
            case "reactQuery":
                return (
                    <div
                        key={index} className={cn(
                        "mb-2 p-2 rounded bg-gray-900",
                        isSelected && "bg-gray-800",
                        log.status === "error" && "bg-red-950",
                        isSelected && log.status === "error" && "bg-red-900",
                    )}
                    >
                        <p className=" font-semibold ">{log.hash}</p>
                        <p className="text-xs ">{log.pageUrl}</p>
                        <p className="text-xs ">{log.type}, {log.timestamp}</p>
                    </div>
                )
            case "console":
                return (
                    <div
                        key={index} className={cn(
                        "mb-2 p-2 rounded bg-gray-900",
                        isSelected && "bg-gray-800",
                        log.type === "error" && "bg-red-950",
                        isSelected && log.type === "error" && "bg-red-900",
                        log.type === "warn" && "bg-yellow-950",
                        isSelected && log.type === "warn" && "bg-yellow-900",
                    )}
                    >
                        <p className="text-sm font-semibold ">{log.content}</p>
                        <p className="text-xs ">{log.timestamp}</p>
                    </div>
                )
            default:
                return <pre key={index} className="text-xs font-mono mb-2">{JSON.stringify(log, null, 2)}</pre>
        }
    }

    if (!currentTime) return null

    return (
        <div
            className={cn(
                "bg-gray-950 p-4 rounded-lg shadow-md",
                type === "server" && "col-span-full",
            )}
        >
            <div className="flex items-center mb-2">
                <span className={`p-2 rounded-full mr-2 bg-gray-900`}>
                    {icon}
                </span>
                <h2 className="text-lg font-semibold">{title}</h2>
            </div>
            <div ref={listRef} className={`text-[--foreground] p-3 rounded-[--radius-md] h-72 overflow-y-auto`}>
                {logs && logs.length > 0 ? (
                    logs.map((log, index) => {
                        if (type === "server") {
                            const timestamp = parseISO(log.substring(0, 19))
                            return renderLog(log, index, isSameSecond(timestamp, currentTime!))
                        }
                        return renderLog(log, index, isSameSecond(log.timestamp!, currentTime!))
                    })
                ) : (
                    <p className="text-[--muted] italic">No logs to display</p>
                )}
            </div>
        </div>
    )
}
