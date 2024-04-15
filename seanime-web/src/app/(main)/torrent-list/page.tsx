"use client"
import { TorrentClientTorrent, TorrentClientTorrentActionProps } from "@/app/(main)/torrent-list/_lib/torrent-client.types"
import { serverStatusAtom } from "@/atoms/server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Tooltip } from "@/components/ui/tooltip"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import React, { useCallback } from "react"
import { BiDownArrow, BiFolder, BiLinkExternal, BiPause, BiPlay, BiStop, BiTime, BiTrash, BiUpArrow } from "react-icons/bi"
import * as upath from "upath"

export const dynamic = "force-static"

export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)

    return (
        <PageWrapper
            className="space-y-4 p-4 sm:p-8"
        >
            <div className="flex items-center w-full justify-between">
                <div>
                    <h2>Active torrents</h2>
                    <p className="text-[--muted]">
                        See torrents currently being downloaded
                    </p>
                </div>
                <div>
                    {/*Show embedded client button only for qBittorrent*/}
                    {serverStatus?.settings?.torrent?.defaultTorrentClient === "qbittorrent" && <Link href={`/qbittorrent`}>
                        <Button intent="white" rightIcon={<BiLinkExternal />}>Embedded client</Button>
                    </Link>}
                </div>
            </div>

            <div className="pb-10">
                <Content />
            </div>
        </PageWrapper>
    )
}

function Content() {
    const [enabled, setEnabled] = React.useState(true)

    const { data, isLoading, status, refetch } = useSeaQuery<TorrentClientTorrent[]>({
        endpoint: SeaEndpoints.TORRENT_CLIENT_LIST,
        queryKey: ["torrents"],
        refetchInterval: 2500,
        gcTime: 0,
        retry: false,
        refetchOnWindowFocus: false,
        enabled: enabled,
    })

    const { mutate, isPending } = useSeaMutation<boolean, TorrentClientTorrentActionProps>({
        endpoint: SeaEndpoints.TORRENT_CLIENT_ACTION,
        mutationKey: ["torrent-action"],
        onSuccess: () => {
            refetch()
        },
    })

    React.useEffect(() => {
        if (status === "error") {
            setEnabled(false)
        }
    }, [status])

    const handleTorrentAction = useCallback((props: TorrentClientTorrentActionProps) => {
        mutate(props)
    }, [mutate])

    if (!enabled) return <LuffyError title="Failed to connect">
        <div className="flex flex-col gap-4 items-center">
            <p className="max-w-md">Failed to connect to the torrent client, verify your settings and make sure it is running.</p>
            <Button
                intent="primary-subtle" onClick={() => {
                setEnabled(true)
            }}
            >Retry</Button>
        </div>
    </LuffyError>

    if (isLoading) return <LoadingSpinner />

    return (
        <AppLayoutStack className={""}>
            {data?.filter(Boolean)?.map(torrent => {
                return <TorrentItem
                    key={torrent.hash}
                    torrent={torrent}
                    onTorrentAction={handleTorrentAction}
                    isPending={isPending}
                />
            })}
            {(!isLoading && !data?.length) && <LuffyError title="Nothing to see">No active torrents</LuffyError>}
        </AppLayoutStack>
    )

}


type TorrentItemProps = {
    torrent: TorrentClientTorrent
    onTorrentAction: (props: TorrentClientTorrentActionProps) => void
    isPending?: boolean
}

const TorrentItem = React.memo(function TorrentItem({ torrent, onTorrentAction, isPending }: TorrentItemProps) {

    const progress = `${(torrent.progress * 100).toFixed(1)}%`

    const confirmDeleteTorrentProps = useConfirmationDialog({
        title: "Remove torrent",
        description: "This action cannot be undone.",
        onConfirm: () => {
            onTorrentAction({
                hash: torrent.hash,
                action: "remove",
                dir: torrent.contentPath,
            })
        },
    })

    return (
        <div className="p-4 border rounded-md  overflow-hidden relative flex gap-2">
            <div className="absolute top-0 w-full h-1 z-[1] bg-gray-700 left-0">
                <div
                    className={cn(
                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                        {
                            "bg-green-300": torrent.status === "downloading",
                            "bg-gray-500": torrent.status === "paused",
                            "bg-blue-500": torrent.status === "seeding",
                        },
                    )}
                    style={{ width: `${String(Math.floor(torrent.progress * 100))}%` }}
                ></div>
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
            <div className="flex-none flex gap-2 items-center">
                {torrent.status !== "seeding" ? (
                    <>
                        <Tooltip
                            trigger={<IconButton
                                icon={<BiPause />}
                                size="sm"
                                intent="gray-subtle"
                                className="flex-none"
                                onClick={async () => {
                                    onTorrentAction({
                                        hash: torrent.hash,
                                        action: "pause",
                                        dir: torrent.contentPath,
                                    })
                                }}
                                disabled={isPending}
                            />}
                        >Pause</Tooltip>
                        {torrent.status !== "downloading" && <Tooltip
                            trigger={<IconButton
                                icon={<BiPlay />}
                                size="sm"
                                intent="gray-subtle"
                                className="flex-none"
                                onClick={async () => {
                                    onTorrentAction({
                                        hash: torrent.hash,
                                        action: "resume",
                                        dir: torrent.contentPath,
                                    })
                                }}
                                disabled={isPending}
                            />}
                        >
                            Resume
                        </Tooltip>}
                    </>
                ) : <Tooltip
                    trigger={<IconButton
                        icon={<BiStop />}
                        size="sm"
                        intent="primary"
                        className="flex-none"
                        onClick={async () => {
                            onTorrentAction({
                                hash: torrent.hash,
                                action: "pause",
                                dir: torrent.contentPath,
                            })
                        }}
                        disabled={isPending}
                    />}
                >End</Tooltip>}

                <div className="flex-none flex gap-2 items-center">
                    <IconButton
                        icon={<BiFolder />}
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
                        disabled={isPending}
                    />
                    <IconButton
                        icon={<BiTrash />}
                        size="sm"
                        intent="alert-subtle"
                        className="flex-none"
                        onClick={async () => {
                            confirmDeleteTorrentProps.open()
                        }}
                        disabled={isPending}
                    />
                </div>
            </div>
            <ConfirmationDialog {...confirmDeleteTorrentProps} />
        </div>
    )
})
