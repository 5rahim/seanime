"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Spinner } from "@/components/ui/loading-spinner"
import { useBoolean } from "@/hooks/use-disclosure"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { ListSyncAnimeDiff, ListSyncAnimeDiffKind, ListSyncAnimeMetadataDiffKind } from "@/lib/server/types"
import { BiListPlus } from "@react-icons/all-files/bi/BiListPlus"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"
import toast from "react-hot-toast"
import { BiRefresh } from "react-icons/bi"
import { FiMinusCircle, FiPlusCircle } from "react-icons/fi"
import { LuUploadCloud } from "react-icons/lu"
import { SiAnilist, SiMyanimelist } from "react-icons/si"
import { useInterval } from "react-use"

type ListSyncDiffsProps = {
    children?: React.ReactNode
    diffs: ListSyncAnimeDiff[]
    onClearCache: () => void
    isDeletingCache: boolean
}

export function ListSyncDiffs(props: ListSyncDiffsProps) {

    const {
        children,
        diffs,
        onClearCache,
        isDeletingCache,
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()

    // const [data, setData] = React.useState(diffs)

    const { mutate: syncAnime, data: returnedData, isPending } = useSeaMutation<ListSyncAnimeDiff[], { kind: string }>({
        endpoint: SeaEndpoints.LIST_SYNC_ANIME,
        method: "post",
        onSuccess: async () => {
            toast.success("Item synced")
            await qc.refetchQueries({ queryKey: ["list-sync-anime-diffs"] })
        },
        retry: 3,
        retryDelay: 3000,
        onError: (err) => {
            if (err.response?.data?.error.includes("To many requests")) {
                toast.error("AniList: Too many requests, please wait a few seconds.")
            } else {
                toast.error("Oops, something went wrong. Try again.")
            }
            syncingMissingInOrigin.off()
            syncingMissingInTarget.off()
            syncingDetails.off()
        },
    })

    const missingInTargetCount = diffs.filter(diff => diff.kind === ListSyncAnimeDiffKind.MISSING_IN_TARGET).length
    const missingInOriginCount = diffs.filter(diff => diff.kind === ListSyncAnimeDiffKind.MISSING_IN_ORIGIN).length
    const detailsCount = diffs.filter(diff => diff.kind === ListSyncAnimeDiffKind.METADATA).length

    const syncingMissingInTarget = useBoolean(false)
    const syncingMissingInOrigin = useBoolean(false)
    const syncingDetails = useBoolean(false)

    useInterval(() => {
        if (syncingMissingInTarget.active) {
            if (missingInTargetCount === 0) {
                syncingMissingInTarget.off()
            } else {
                console.log("[ListSync] Syncing missing in target")
                syncAnime({ kind: ListSyncAnimeDiffKind.MISSING_IN_TARGET })
            }
        }
        if (syncingMissingInOrigin.active) {
            if (missingInOriginCount === 0) {
                syncingMissingInOrigin.off()
            } else {
                console.log("[ListSync] Syncing missing in origin")
                syncAnime({ kind: ListSyncAnimeDiffKind.MISSING_IN_ORIGIN })
            }
        }
        if (syncingDetails.active) {
            if (detailsCount === 0) {
                syncingDetails.off()
            } else {
                console.log("[ListSync] Syncing details")
                syncAnime({ kind: ListSyncAnimeDiffKind.METADATA })
            }
        }
    }, 2000)

    const disabledButton = isPending || isDeletingCache || syncingMissingInTarget.active || syncingMissingInOrigin.active || syncingDetails.active

    function handleSyncAdditions() {
        syncingMissingInTarget.on()
        syncingMissingInOrigin.off()
        syncingDetails.off()
    }

    function handleSyncRemovals() {
        syncingMissingInTarget.off()
        syncingMissingInOrigin.on()
        syncingDetails.off()
    }

    function handleSyncDetails() {
        syncingMissingInTarget.off()
        syncingMissingInOrigin.off()
        syncingDetails.on()
    }

    function handleStopSyncing() {
        syncingMissingInTarget.off()
        syncingMissingInOrigin.off()
        syncingDetails.off()
    }

    return (
        <div className="space-y-4">

            <p className="text-[--muted]">
                Source of truth: <span className="text-white font-semibold">{getSourceName(serverStatus?.settings?.listSync?.origin || "")}</span>
            </p>

            <ul className="text-sm text-[--muted] list-disc pl-4">
                <li><em className="font-semibold">MALSync</em> is recommended for a more complete solution.</li>
                <li>Refresh AniList (top right) to see changes reflected</li>
                <li>Some items may not be synced due to mapping limitations. You will notice this if they re-appear in the list below.</li>
            </ul>

            <div>
                <Button
                    size="sm"
                    intent="primary-outline"
                    onClick={onClearCache}
                    leftIcon={<BiRefresh />}
                    isLoading={isDeletingCache}
                    isDisabled={disabledButton}
                >
                    Refresh data
                </Button>
            </div>

            <div className="flex items-center justify-between w-full">
                <h4>Items to sync: <Badge size="lg" className="ml-1 px-1.5">{diffs.length}</Badge></h4>
                <div className="flex items-center gap-2">
                    <Button size="sm" intent="white-outline" onClick={handleSyncDetails} isDisabled={disabledButton || !diffs.length}>Sync
                                                                                                                                      details</Button>
                    <Button size="sm" intent="success-outline" onClick={handleSyncAdditions} isDisabled={disabledButton || !diffs.length}>
                        Sync additions</Button>
                    <Button size="sm" intent="alert-outline" onClick={handleSyncRemovals} isDisabled={disabledButton || !diffs.length}>Sync
                                                                                                                                       removals</Button>
                </div>
            </div>

            {(syncingMissingInTarget.active || syncingMissingInOrigin.active || syncingDetails.active) &&
                <div className="text-sm flex gap-2 items-center">
                    <Spinner className="w-4 h-4" />
                    {syncingMissingInTarget.active && <p>Syncing additions...</p>}
                    {syncingMissingInOrigin.active && <p>Syncing removals...</p>}
                    {syncingDetails.active && <p>Syncing details...</p>}
                    <Button size="sm" intent="alert" onClick={handleStopSyncing}>Stop</Button>
                </div>}


            {!diffs?.length && <div className="p-4 text-[--muted] text-center">No items to sync</div>}

            <div className="space-y-4">
                {diffs.map((diff, idx) => {
                    return (
                        <DiffItem key={diff.id} item={diff} />
                    )
                })}
            </div>
        </div>
    )
}

type DiffItemProps = {
    item: ListSyncAnimeDiff
}

function DiffItem(props: DiffItemProps) {

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
                <p className="text-sm max-w-sm line-clamp-1">{entry.displayTitle}</p>
                <div className="flex gap-2 items-center">
                    <div>
                        {item.kind === ListSyncAnimeDiffKind.MISSING_IN_TARGET && <FiPlusCircle className="text-green-200" />}
                        {item.kind === ListSyncAnimeDiffKind.MISSING_IN_ORIGIN && <FiMinusCircle className="text-red-300" />}
                        {item.kind === ListSyncAnimeDiffKind.METADATA && <LuUploadCloud className="text-blue-200" />}
                    </div>
                    <p className="text-sm text-gray-400">{getDiffKindDescription(item.kind)
                        .replace("{{source}}", getSourceName(item.targetSource))}</p>
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
