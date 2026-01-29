"use client"
import { Debrid_TorrentItem } from "@/api/generated/types"
import { useDebridCancelDownload, useDebridDeleteTorrent, useDebridDownloadTorrent, useDebridGetTorrents } from "@/api/hooks/debrid.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Tooltip } from "@/components/ui/tooltip"
import { WSEvents } from "@/lib/server/ws-events"
import { formatDate } from "date-fns"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"
import { BiDownArrow, BiLinkExternal, BiRefresh, BiTime, BiTrash, BiX } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { FiDownload } from "react-icons/fi"
import { HiFolderDownload } from "react-icons/hi"
import { toast } from "sonner"

export const dynamic = "force-static"

function getServiceName(provider: string) {
    switch (provider) {
        case "realdebrid":
            return "Real-Debrid"
        case "torbox":
            return "TorBox"
        case "alldebrid":
            return "AllDebrid"
        default:
            return provider
    }
}

function getDashboardLink(provider: string) {
    switch (provider) {
        case "torbox":
            return "https://torbox.app/dashboard"
        case "realdebrid":
            return "https://real-debrid.com/torrents"
        case "alldebrid":
            return "https://alldebrid.com/magnets/"
        default:
            return ""
    }
}

export default function Page() {
    const serverStatus = useServerStatus()

    if (!serverStatus) return <LoadingSpinner />

    if (!serverStatus?.debridSettings?.enabled || !serverStatus?.debridSettings?.provider) return <LuffyError
        title="Debrid not enabled"
    >
        Debrid service is not enabled or configured
    </LuffyError>

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper
                className="space-y-4 p-4 sm:p-8"
            >
                <Content />
            </PageWrapper>
            <TorrentItemModal />
        </>
    )
}

function Content() {
    const serverStatus = useServerStatus()
    const [enabled, setEnabled] = React.useState(true)
    const [refetchInterval, setRefetchInterval] = React.useState(30000)

    const { data, isLoading, status, refetch } = useDebridGetTorrents(enabled, refetchInterval)

    React.useEffect(() => {
        const hasDownloads = data?.filter(t => t.status === "downloading" || t.status === "paused")?.length ?? 0
        setRefetchInterval(hasDownloads ? 5000 : 30000)
    }, [data])

    React.useEffect(() => {
        if (status === "error") {
            setEnabled(false)
        }
    }, [status])

    if (!enabled) return <LuffyError title="Failed to connect">
        <div className="flex flex-col gap-4 items-center">
            <p className="max-w-md">Failed to connect to the Debrid service, verify your settings.</p>
            <Button
                intent="primary-subtle" onClick={() => {
                    setEnabled(true)
                }}
            >Retry</Button>
        </div>
    </LuffyError>

    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <div className="flex items-center w-full">
                <div>
                    <h2>{getServiceName(serverStatus?.debridSettings?.provider!)}</h2>
                    <p className="text-[--muted]">
                        See your debrid service torrents
                    </p>
                </div>
                <div className="flex flex-1"></div>
                <div className="flex gap-2 items-center">
                    <Button
                        intent="white-subtle"
                        leftIcon={<BiRefresh className="text-2xl" />}
                        onClick={() => {
                            refetch()
                            toast.info("Refreshed")
                        }}
                    >Refresh</Button>
                    {!!getDashboardLink(serverStatus?.debridSettings?.provider!) && (
                        <SeaLink href={getDashboardLink(serverStatus?.debridSettings?.provider!)} target="_blank">
                            <Button
                                intent="primary-subtle"
                                rightIcon={<BiLinkExternal className="text-xl" />}
                            >Dashboard</Button>
                        </SeaLink>
                    )}
                </div>
            </div>

            <div className="pb-10">
                <AppLayoutStack className={""}>

                    <div>
                        <ul className="text-[--muted] flex flex-wrap gap-4">
                            <li>Downloading: {data?.filter(t => t.status === "downloading" || t.status === "paused")?.length ?? 0}</li>
                            <li>Seeding: {data?.filter(t => t.status === "seeding")?.length ?? 0}</li>
                        </ul>
                    </div>

                    <Card className="p-0 overflow-hidden">
                        {data?.filter(Boolean)?.map(torrent => {
                            return <TorrentItem
                                key={torrent.id}
                                torrent={torrent}
                            />
                        })}
                        {(!isLoading && !data?.length) && <LuffyError title="Nothing to see">No active torrents</LuffyError>}
                    </Card>
                </AppLayoutStack>
            </div>
        </>
    )

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const selectedTorrentItemAtom = atom<Debrid_TorrentItem | null>(null)


type TorrentItemProps = {
    torrent: Debrid_TorrentItem
    isPending?: boolean
}

type DownloadProgress = {
    status: string
    itemID: string
    totalBytes: string
    totalSize: string
    speed: string
}

const TorrentItem = React.memo(function TorrentItem({ torrent, isPending }: TorrentItemProps) {

    const { mutate: deleteTorrent, isPending: isDeleting } = useDebridDeleteTorrent()

    const { mutate: cancelDownload, isPending: isCancelling } = useDebridCancelDownload()

    const [_, setSelectedTorrentItem] = useAtom(selectedTorrentItemAtom)

    const confirmDeleteTorrentProps = useConfirmationDialog({
        title: "Remove torrent",
        description: "This action cannot be undone.",
        onConfirm: () => {
            deleteTorrent({
                torrentItem: torrent,
            })
        },
    })

    const [progress, setProgress] = React.useState<DownloadProgress | null>(null)

    useWebsocketMessageListener<DownloadProgress>({
        type: WSEvents.DEBRID_DOWNLOAD_PROGRESS,
        onMessage: data => {
            if (data.itemID === torrent.id) {
                if (data.status === "downloading") {
                    setProgress(data)
                } else {
                    setProgress(null)
                }
            }
        },
    })

    function handleCancelDownload() {
        cancelDownload({
            itemID: torrent.id,
        })
    }

    return (
        <div
            data-torrent-item-container className={cn(
            "hover:bg-gray-900 hover:bg-opacity-70 px-4 py-3 relative flex gap-4 group/torrent-item",
            torrent.status === "paused" && "bg-gray-900 hover:bg-gray-900",
            torrent.status === "downloading" && "bg-green-900 bg-opacity-20 hover:hover:bg-opacity-30 hover:bg-green-900",
        )}
        >
            <div className="w-full">
                <div
                    className={cn("group-hover/torrent-item:text-white break-all", {
                        "opacity-50": torrent.status === "paused",
                    })}
                >{torrent.name}</div>
                <div className="text-[--muted]">
                    <span className={cn({ "text-green-300": torrent.status === "downloading" })}>{torrent.completionPercentage}%</span>
                    {` `}
                    <BiDownArrow className="inline-block mx-2" />
                    {torrent.speed}
                    {(torrent.eta && torrent.status === "downloading") && <>
                        {` `}
                        <BiTime className="inline-block mx-2 mb-0.5" />
                        {torrent.eta}
                    </>}
                    {` - `}
                    <span className="text-[--muted]">
                        {formatDate(torrent.added, "yyyy-MM-dd HH:mm")}
                    </span>
                    {` - `}
                    <strong
                        className={cn(
                            "text-sm",
                            torrent.status === "seeding" && "text-blue-300",
                            torrent.status === "completed" && "text-green-300",
                        )}
                    >{(torrent.status === "other" || !torrent.isReady) ? "" : capitalize(torrent.status)}</strong>
                </div>
                {torrent.status !== "seeding" && torrent.status !== "completed" &&
                    <div data-torrent-item-progress-bar className="w-full h-1 mr-4 mt-2 relative z-[1] bg-gray-700 left-0 overflow-hidden rounded-xl">
                        <div
                            className={cn(
                                "h-full absolute z-[2] left-0 bg-gray-200 transition-all",
                                {
                                    "bg-green-300": torrent.status === "downloading",
                                    "bg-gray-500": torrent.status === "paused",
                                    "bg-orange-800": torrent.status === "other",
                                },
                            )}
                            style={{ width: `${String(torrent.completionPercentage)}%` }}
                        ></div>
                    </div>}
            </div>
            <div className="flex-none flex gap-2 items-center">
                {(torrent.isReady && !progress) && <IconButton
                    icon={<FiDownload />}
                    size="sm"
                    intent="gray-subtle"
                    className="flex-none"
                    disabled={isDeleting || isCancelling}
                    onClick={() => {
                        setSelectedTorrentItem(torrent)
                    }}
                />}
                {(!!progress && progress.itemID === torrent.id) && <div className="flex gap-2 items-center">
                    <Tooltip
                        trigger={<p>
                            <HiFolderDownload className="text-2xl animate-pulse text-[--blue]" />
                        </p>}
                    >
                        Downloading locally
                    </Tooltip>
                    <p>
                        {progress?.totalBytes}<span className="text-[--muted]"> / {progress?.totalSize}</span>
                    </p>
                    <Tooltip
                        trigger={<p>
                            <IconButton
                                icon={<BiX className="text-xl" />}
                                intent="gray-subtle"
                                rounded
                                size="sm"
                                onClick={handleCancelDownload}
                                loading={isCancelling}
                            />
                        </p>}
                    >
                        Cancel download
                    </Tooltip>
                </div>}
                <IconButton
                    icon={<BiTrash />}
                    size="sm"
                    intent="alert-subtle"
                    className="flex-none"
                    onClick={async () => {
                        confirmDeleteTorrentProps.open()
                    }}
                    disabled={isCancelling}
                    loading={isDeleting}
                />
            </div>
            <ConfirmationDialog {...confirmDeleteTorrentProps} />
        </div>
    )
})

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TorrentItemModalProps = {}

function TorrentItemModal(props: TorrentItemModalProps) {
    const serverStatus = useServerStatus()

    const [selectedTorrentItem, setSelectedTorrentItem] = useAtom(selectedTorrentItemAtom)
    const { mutate: downloadTorrent, isPending: isDownloading } = useDebridDownloadTorrent()

    const [destination, setDestination] = React.useState("")

    const libraryPath = React.useMemo(() => serverStatus?.settings?.library?.libraryPath, [serverStatus])

    const libraryPathSelectionProps = useLibraryPathSelection({
        destination,
        setDestination,
    })

    React.useEffect(() => {
        if (selectedTorrentItem && libraryPath) {
            setDestination(libraryPath)
        }
    }, [selectedTorrentItem, libraryPath])

    const handleDownload = () => {
        if (!selectedTorrentItem || !destination) return
        downloadTorrent({
            torrentItem: selectedTorrentItem,
            destination: destination,
        }, {
            onSuccess: () => {
                setSelectedTorrentItem(null)
            },
        })
    }

    return (
        <Modal
            open={!!selectedTorrentItem}
            onOpenChange={() => {
                setSelectedTorrentItem(null)
            }}
            title="Download"
            contentClass="max-w-2xl"
        >
            <p className="text-center line-clamp-2 text-sm">
                {selectedTorrentItem?.name}
            </p>

            <div className="space-y-4 mt-4">
                <DirectorySelector
                    name="destination"
                    label="Destination"
                    leftIcon={<FcFolder />}
                    value={destination}
                    defaultValue={destination}
                    onSelect={setDestination}
                    shouldExist={false}
                    help="Where to save the torrent"
                    libraryPathSelectionProps={libraryPathSelectionProps}
                />

                <div className="flex justify-end">
                    <Button
                        intent="white"
                        leftIcon={<FiDownload className="text-xl" />}
                        loading={isDownloading}
                        disabled={!destination || destination.length < 2}
                        onClick={handleDownload}
                    >
                        Download
                    </Button>
                </div>
            </div>
        </Modal>
    )
}
