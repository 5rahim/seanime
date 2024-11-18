"use client"
import { Debrid_TorrentItem } from "@/api/generated/types"
import { useDebridCancelDownload, useDebridDeleteTorrent, useDebridDownloadTorrent, useDebridGetTorrents } from "@/api/hooks/debrid.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
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

                    {data?.filter(Boolean)?.map(torrent => {
                        return <TorrentItem
                            key={torrent.id}
                            torrent={torrent}
                        />
                    })}
                    {(!isLoading && !data?.length) && <LuffyError title="Nothing to see">No active torrents</LuffyError>}
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
        <div className="p-4 border rounded-md  overflow-hidden relative flex gap-2">
            <div className="absolute top-0 w-full h-1 z-[1] bg-gray-700 left-0">
                <div
                    className={cn(
                        "h-1 absolute z-[2] left-0 bg-gray-200 transition-all",
                        {
                            "bg-green-300": torrent.status === "downloading",
                            "bg-gray-500": torrent.status === "paused",
                            "bg-blue-500": torrent.status === "seeding",
                            "bg-gray-600": torrent.status === "completed",
                            "bg-orange-600": torrent.status === "other",
                        },
                    )}
                    style={{ width: `${String(Math.floor(torrent.completionPercentage))}%` }}
                ></div>
            </div>
            <div className="w-full">
                <div
                    className={cn({
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
                    <span className="text-[--foreground]">
                        {formatDate(torrent.added, "yyyy-MM-dd HH:mm")}
                    </span>
                    {` - `}
                    <strong
                        className={cn(
                            torrent.status === "seeding" && "text-blue-300",
                            torrent.status === "completed" && "text-green-300",
                        )}
                    >{(torrent.status === "other" || !torrent.isReady) ? "Processing" : capitalize(torrent.status)}</strong>
                </div>
            </div>
            <div className="flex-none flex gap-2 items-center">
                {(torrent.isReady && !progress) && <Button
                    leftIcon={<FiDownload />}
                    size="sm"
                    intent="white-subtle"
                    className="flex-none"
                    disabled={isDeleting || isCancelling}
                    onClick={() => {
                        setSelectedTorrentItem(torrent)
                    }}
                >
                    Download
                </Button>}
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

const downloadSchema = defineSchema(({ z }) => z.object({
    destination: z.string().min(2),
}))

function TorrentItemModal(props: TorrentItemModalProps) {

    const {
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const [selectedTorrentItem, setSelectedTorrentItem] = useAtom(selectedTorrentItemAtom)
    const { mutate: downloadTorrent, isPending: isDownloading } = useDebridDownloadTorrent()

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

            <Form
                schema={downloadSchema}
                onSubmit={data => {
                    downloadTorrent({
                        torrentItem: selectedTorrentItem!,
                        destination: data.destination,
                    }, {
                        onSuccess: () => {
                            setSelectedTorrentItem(null)
                        },
                    })
                }}
                defaultValues={{
                    destination: serverStatus?.settings?.library?.libraryPath ?? "",
                }}
            >
                <Field.DirectorySelector
                    name="destination"
                    label="Destination"
                    shouldExist={false}
                    help="Where to save the torrent"
                />
                <div className="flex justify-end">
                    <Field.Submit
                        intent="white"
                        leftIcon={<FiDownload className="text-xl" />}
                        loading={isDownloading}
                    >
                        Download
                    </Field.Submit>
                </div>
            </Form>
        </Modal>
    )
}
