"use client"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Tooltip } from "@/components/ui/tooltip"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { SeaTorrent, SeaTorrentActionProps } from "@/lib/server/types"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import React, { useCallback } from "react"
import { BiDownArrow, BiFolder, BiLinkExternal, BiPause, BiPlay, BiStop, BiTime, BiUpArrow } from "react-icons/bi"
import * as upath from "upath"

export default function Page() {

    return (
        <AppLayoutStack className="p-4 sm:p-8">
            <div className="flex items-center w-full justify-between">
                <div>
                    <h2>Active torrents</h2>
                    <p className="text-[--muted]">
                        See torrents currently being downloaded
                    </p>
                </div>
                <div className={""}>
                    <Link href={`/qbittorrent`}>
                        <Button intent="white" rightIcon={<BiLinkExternal />}>Embedded client</Button>
                    </Link>
                </div>
            </div>

            <div className="pb-10">
                <Content/>
            </div>
        </AppLayoutStack>
    )
}

function Content() {
    const { data, isLoading, refetch } = useSeaQuery<SeaTorrent[]>({
        endpoint: SeaEndpoints.TORRENT_CLIENT_LIST,
        queryKey: ["torrents"],
        refetchInterval: 2500,
        gcTime: 0,
        retry: false,
        refetchOnWindowFocus: false,
    })

    const { mutate, isPending } = useSeaMutation<boolean, SeaTorrentActionProps>({
        endpoint: SeaEndpoints.TORRENT_CLIENT_ACTION,
        mutationKey: ["torrent-action"],
    })

    const handleTorrentAction = useCallback((props: SeaTorrentActionProps) => {
        mutate(props)
    }, [mutate])

    if (isLoading) return <LoadingSpinner/>

    return (
        <AppLayoutStack className={""}>
            {data?.filter(Boolean)?.map(torrent => {
                return <TorrentItem
                    key={torrent.hash}
                    torrent={torrent}
                    refetch={refetch}
                    onTorrentAction={handleTorrentAction}
                />
            })}
            {(!isLoading && !data?.length) && <LuffyError title="Nothing to see">No active torrents</LuffyError>}
        </AppLayoutStack>
    )

}


type TorrentItemProps = {
    torrent: SeaTorrent
    refetch: () => void
    onTorrentAction: (props: SeaTorrentActionProps) => void
}

function TorrentItem({ torrent, refetch, onTorrentAction }: TorrentItemProps) {

    const progress = `${(torrent.progress * 100).toFixed(1)}%`

    return (
        <div className="p-4 border rounded-md  overflow-hidden relative flex gap-2">
            <div className="absolute top-0 w-full h-1 z-[1] bg-gray-700 left-0">
                <div className={cn(
                    "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                    {
                        "bg-green-300": torrent.status === "downloading",
                        "bg-gray-500": torrent.status === "paused",
                        "bg-blue-500": torrent.status === "seeding",
                    },
                )}
                     style={{ width: `${String(Math.floor(torrent.progress * 100))}%` }}></div>
            </div>
            <div className="w-full">
                <div
                    className={cn({
                        "opacity-50": torrent.status === "paused",
                    })}
                >{torrent.name}</div>
                <div className="text-[--muted]">
                    <span className={cn({ "text-green-300": torrent.status === "downloading" })}>{progress}</span>
                    {` `}
                    <BiDownArrow className="inline-block mx-2" />
                    {torrent.downSpeed}
                    {` `}
                    <BiUpArrow className="inline-block mx-2 mb-1" />
                    {torrent.upSpeed}
                    {` `}
                    <BiTime className="inline-block mx-2 mb-0.5" />
                    {torrent.eta}
                    {` - `}
                    <span>{torrent.seeds} {torrent.seeds !== 1 ? "seeds" : "seed"}</span>
                    {/*{` - `}*/}
                    {/*<span>{torrent.peers} {torrent.peers !== 1 ? "peers" : "peer"}</span>*/}
                    {` - `}
                    <strong
                        className={cn({
                            "text-blue-300": torrent.status === "seeding",
                        })}
                    >{capitalize(torrent.status)}</strong>
                </div>
            </div>
            <div className="flex gap-2 items-center">
                <div className="flex-none">
                    <IconButton
                        icon={<BiFolder/>}
                        size="sm"
                        intent="gray-subtle"
                        className="flex-none"
                        onClick={async () => {
                            onTorrentAction({
                                hash: torrent.hash,
                                action: "open",
                                dir: upath.dirname(torrent.contentPath),
                            })
                        }}
                    />
                </div>
                {torrent.status !== "seeding" ? (
                    <>
                        <Tooltip trigger={<IconButton
                            icon={<BiPause/>}
                            size="sm"
                            intent="gray-subtle"
                            className="flex-none"
                            onClick={async () => {
                                onTorrentAction({
                                    hash: torrent.hash,
                                    action: "pause",
                                    dir: torrent.contentPath,
                                })
                                refetch()
                            }}
                        />}>Pause</Tooltip>
                        <Tooltip trigger={<div>{torrent.status !== "downloading" && <IconButton
                            icon={<BiPlay/>}
                            size="sm"
                            intent="gray-subtle"
                            className="flex-none"
                            onClick={async () => {
                                onTorrentAction({
                                    hash: torrent.hash,
                                    action: "resume",
                                    dir: torrent.contentPath,
                                })
                                refetch()
                            }}
                        />}</div>}>
                            Resume
                        </Tooltip>
                    </>
                ) : <Tooltip trigger={<IconButton
                    icon={<BiStop/>}
                    size="sm"
                    intent="primary"
                    className="flex-none"
                    onClick={async () => {
                        onTorrentAction({
                            hash: torrent.hash,
                            action: "pause",
                            dir: torrent.contentPath,
                        })
                        refetch()
                    }}
                />}>End</Tooltip>}
            </div>
        </div>
    )
}
