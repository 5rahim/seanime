"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { ListSyncAnimeDiffKind, ListSyncAnimeMetadataDiffKind, ListSyncDiff } from "@/lib/server/types"
import { BiListPlus } from "@react-icons/all-files/bi/BiListPlus"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"
import { FiMinusCircle, FiPlusCircle } from "react-icons/fi"
import { LuUploadCloud } from "react-icons/lu"
import { SiAnilist, SiMyanimelist } from "react-icons/si"

type ListSyncDiffsProps = {
    children?: React.ReactNode
    diffs: ListSyncDiff[]
}

export function ListSyncDiffs(props: ListSyncDiffsProps) {

    const {
        children,
        diffs,
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)

    return (
        <div className="space-y-4">

            <p className="text-[--muted]">
                Source of truth: <span className="text-white font-semibold">{getSourceName(serverStatus?.settings?.listSync?.origin || "")}</span>
            </p>

            <div className="flex items-center justify-between w-full">
                <h4>Items to sync: <Badge size="lg" className="ml-1 px-1.5">{diffs.length}</Badge></h4>
                <div className="flex items-center gap-2">
                    <Button size="sm" intent="white-outline">Sync details</Button>
                    <Button size="sm" intent="success-outline">Sync additions</Button>
                    <Button size="sm" intent="alert-outline">Sync removals</Button>
                </div>
            </div>

            <div className="space-y-4">
                {diffs?.map((diff, idx) => {
                    return (
                        <div key={idx + diff.kind}>
                            <DiffItem item={diff} />
                        </div>
                    )
                })}
            </div>
        </div>
    )
}

type DiffItemProps = {
    item: ListSyncDiff
}

export function DiffItem(props: DiffItemProps) {

    const {
        item,
        ...rest
    } = props

    const entry = React.useMemo(() => {
        return item.targetEntry ?? item.originEntry
    }, [])

    if (!entry) return null

    return (
        <div
            className={cn("border border-[--border] rounded-[--radius] p-2 flex w-full relative")}
        >
            <div
                className="bg-cover bg-center w-16 h-16 rounded-[--radius] mr-4"
                style={{ backgroundImage: "url(" + entry.image + ")" }}
            />

            <div className="space-y-1">
                <p className="text-sm">{entry.displayTitle}</p>
                <div className="flex gap-2 items-center">
                    <div>
                        {item.kind === ListSyncAnimeDiffKind.MISSING_IN_TARGET && <FiPlusCircle className="text-green-200" />}
                        {item.kind === ListSyncAnimeDiffKind.MISSING_IN_ORIGIN && <FiMinusCircle className="text-red-300" />}
                        {item.kind === ListSyncAnimeDiffKind.METADATA && <LuUploadCloud className="text-blue-200" />}
                    </div>
                    <p className="text-sm">{getDiffKindDescription(item.kind).replace("{{source}}", getSourceName(item.targetSource))}</p>
                </div>
                {item.kind === ListSyncAnimeDiffKind.METADATA &&
                    <ul className="space-y-1 text-sm ml-4 mt-1 [&>li]:flex [&>li]:gap-1 [&>li]:items-center">
                        {item.metadataDiffKinds.map((mDiff, idx) => {
                            if (mDiff === ListSyncAnimeMetadataDiffKind.SCORE) return (
                                <li key={idx}>
                                    Score: <Badge tag intent="alert">{item.targetEntry?.score}</Badge> {`->`} <Badge
                                    tag
                                    intent="success"
                                >{item.originEntry?.score}</Badge>
                                </li>
                            )
                            if (mDiff === ListSyncAnimeMetadataDiffKind.PROGRESS) return (
                                <li key={idx}>
                                    Progress: <Badge tag intent="alert">{item.targetEntry?.progress}</Badge> {`->`} <Badge
                                    tag
                                    intent="success"
                                >{item.originEntry?.progress}</Badge>
                                </li>
                            )
                            if (mDiff === ListSyncAnimeMetadataDiffKind.STATUS) return (
                                <li key={idx}>
                                    Status: <Badge tag intent="alert">{capitalize(item.targetEntry?.status)}</Badge> {`->`} <Badge
                                    tag
                                    intent="success"
                                >{capitalize(item.originEntry?.status)}</Badge>
                                </li>
                            )
                        })}
                    </ul>}
            </div>

            <div className="absolute text-3xl right-4 bottom-2">
                {getSourceIcon(item.targetSource)}
            </div>

        </div>
    )
}

function getDiffKindDescription(kind: ListSyncAnimeDiffKind) {
    switch (kind) {
        case ListSyncAnimeDiffKind.METADATA:
            return "Details are mismatched"
        case ListSyncAnimeDiffKind.MISSING_IN_ORIGIN:
            return "Anime will be deleted from {{source}}"
        case ListSyncAnimeDiffKind.MISSING_IN_TARGET:
            return "Anime missing from {{source}}"
        default:
            return "N/A"
    }
}

function getSourceName(source: string) {
    switch (source) {
        case "anilist":
            return "AniList"
        case "mal":
            return "MyAnimeList"
        default:
            return source
    }
}


function getSourceIcon(source: string) {
    switch (source) {
        case "anilist":
            return <SiAnilist />
        case "mal":
            return <SiMyanimelist />
        default:
            return <BiListPlus />
    }
}
