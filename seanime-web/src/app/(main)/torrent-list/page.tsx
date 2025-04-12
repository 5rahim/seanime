"use client"
import { TorrentClientAction_Variables } from "@/api/generated/endpoint.types"
import { TorrentClient_Torrent } from "@/api/generated/types"
import { useGetActiveTorrentList, useTorrentClientAction } from "@/api/hooks/torrent_client.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Tooltip } from "@/components/ui/tooltip"
import capitalize from "lodash/capitalize"
import React from "react"
import { BiDownArrow, BiFolder, BiLinkExternal, BiPause, BiPlay, BiStop, BiTime, BiTrash, BiUpArrow } from "react-icons/bi"
import * as upath from "upath"

export const dynamic = "force-static"

export default function Page() {
    const serverStatus = useServerStatus()

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper
                data-torrent-list-page-container
                className="space-y-4 p-4 sm:p-8"
            >
                <div data-torrent-list-page-header className="flex items-center w-full justify-between">
                    <div data-torrent-list-page-header-title>
                        <h2>Active torrents</h2>
                        <p className="text-[--muted]">
                            See torrents currently being downloaded
                        </p>
                    </div>
                    <div data-torrent-list-page-header-actions>
                        {/*Show embedded client button only for qBittorrent*/}
                        {serverStatus?.settings?.torrent?.defaultTorrentClient === "qbittorrent" && <SeaLink href={`/qbittorrent`}>
                            <Button intent="white" rightIcon={<BiLinkExternal />}>Embedded client</Button>
                        </SeaLink>}
                    </div>
                </div>

                <div data-torrent-list-page-content className="pb-10">
                    <Content />
                </div>
            </PageWrapper>
        </>
    )
}

function Content() {
    const [enabled, setEnabled] = React.useState(true)

    const { data, isLoading, status, refetch } = useGetActiveTorrentList(enabled)

    const { mutate, isPending } = useTorrentClientAction(() => {
        refetch()
    })

    React.useEffect(() => {
        if (status === "error") {
            setEnabled(false)
        }
    }, [status])

    const handleTorrentAction = React.useCallback((props: TorrentClientAction_Variables) => {
        mutate(props)
    }, [mutate])


    const confirmStopAllSeedingProps = useConfirmationDialog({
        title: "Stop seeding all torrents",
        description: "This action will cause seeding to stop for all completed torrents.",
        actionIntent: "warning",
        onConfirm: () => {
            for (const torrent of data ?? []) {
                handleTorrentAction({
                    hash: torrent.hash,
                    action: "pause",
                    dir: torrent.contentPath,
                })
            }
        },
    })

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

            <div>
                <ul className="text-[--muted] flex flex-wrap gap-4">
                    <li>Downloading: {data?.filter(t => t.status === "downloading" || t.status === "paused")?.length ?? 0}</li>
                    <li>Seeding: {data?.filter(t => t.status === "seeding")?.length ?? 0}</li>
                    {!!data?.filter(t => t.status === "seeding")?.length && <li>
                        <Button
                            size="xs"
                            intent="primary-link"
                            onClick={() => confirmStopAllSeedingProps.open()}
                        >Stop seeding</Button>
                    </li>}
                </ul>
            </div>

            {data?.filter(Boolean)?.map(torrent => {
                return <TorrentItem
                    key={torrent.hash}
                    torrent={torrent}
                    onTorrentAction={handleTorrentAction}
                    isPending={isPending}
                />
            })}
            {(!isLoading && !data?.length) && <LuffyError title="Nothing to see">No active torrents</LuffyError>}

            <ConfirmationDialog {...confirmStopAllSeedingProps} />
        </AppLayoutStack>
    )

}


type TorrentItemProps = {
    torrent: TorrentClient_Torrent
    onTorrentAction: (props: TorrentClientAction_Variables) => void
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
        <div data-torrent-item-container className="p-4 border rounded-[--radius-md]  overflow-hidden relative flex gap-2">
            <div data-torrent-item-progress-bar className="absolute top-0 w-full h-1 z-[1] bg-gray-700 left-0">
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
            <div data-torrent-item-title-container className="w-full">
                <div
                    className={cn({
                        "opacity-50": torrent.status === "paused",
                    })}
                >{torrent.name}</div>
                <div data-torrent-item-info className="text-[--muted]">
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
            <div data-torrent-item-actions className="flex-none flex gap-2 items-center">
                {torrent.status !== "seeding" ? (
                    <>
                        {torrent.status !== "paused" && <Tooltip
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
                        >Pause</Tooltip>}
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

                <div data-torrent-item-actions-buttons className="flex-none flex gap-2 items-center">
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
