"use client"
import { Updater_Release } from "@/api/generated/types"
import { useDownloadRelease } from "@/api/hooks/download.hooks"
import { useGetLatestUpdate, useInstallLatestUpdate } from "@/api/hooks/releases.hooks"
import { UpdateChangelogBody } from "@/app/(main)/_features/update/update-helper"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { RadioGroup } from "@/components/ui/radio-group"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"
import { BiDownload, BiLinkExternal } from "react-icons/bi"
import { FiArrowRight } from "react-icons/fi"
import { GrInstall } from "react-icons/gr"
import { toast } from "sonner"

type UpdateModalProps = {
    collapsed?: boolean
}


export const updateModalOpenAtom = atom<boolean>(false)
const downloaderOpenAtom = atom<boolean>(false)

export function UpdateModal(props: UpdateModalProps) {
    const serverStatus = useServerStatus()
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)
    const [downloaderOpen, setDownloaderOpen] = useAtom(downloaderOpenAtom)

    const { data: updateData, isLoading, refetch } = useGetLatestUpdate(!!serverStatus && !serverStatus?.settings?.library?.disableUpdateCheck)

    useWebsocketMessageListener({
        type: WSEvents.CHECK_FOR_UPDATES,
        onMessage: () => {
            refetch()
        },
    })

    // Install update
    const { mutate: installUpdate, isPending } = useInstallLatestUpdate()
    const [fallbackDestination, setFallbackDestination] = React.useState<string>("")

    React.useEffect(() => {
        if (updateData && updateData.release) {
            localStorage.setItem("latest-available-update", JSON.stringify(updateData.release.version))
            const latestVersionNotified = localStorage.getItem("notified-available-update")
            if (!latestVersionNotified || (latestVersionNotified !== JSON.stringify(updateData.release.version))) {
                setUpdateModalOpen(true)
            }
        }
    }, [updateData])

    const ignoreUpdate = () => {
        if (updateData && updateData.release) {
            localStorage.setItem("notified-available-update", JSON.stringify(updateData.release.version))
            setUpdateModalOpen(false)
        }
    }

    function handleInstallUpdate() {
        installUpdate({ fallback_destination: "" })
    }

    if (serverStatus?.settings?.library?.disableUpdateCheck) return null

    if (isLoading || !updateData || !updateData.release) return null

    return (
        <>
            <VerticalMenu
                collapsed={props.collapsed}
                items={[
                    {
                        iconType: AiFillExclamationCircle,
                        name: "Update available",
                        onClick: () => setUpdateModalOpen(true),
                    },
                ]}
                itemIconClass="text-brand-300"
            />
            <Modal
                open={updateModalOpen}
                onOpenChange={() => ignoreUpdate()}
                contentClass="max-w-3xl"
            >
                <Downloader release={updateData.release} />
                {/*<div*/}
                {/*    className="bg-[url(/pattern-2.svg)] z-[-1] w-full h-[4rem] absolute opacity-60 left-0 bg-no-repeat bg-right bg-cover"*/}
                {/*>*/}
                {/*    <div*/}
                {/*        className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"*/}
                {/*    />*/}
                {/*</div>*/}
                <div className="space-y-2">
                    <h3 className="text-center">A new update is available!</h3>
                    <h4 className="font-bold flex gap-2 text-center items-center justify-center">
                        <span className="text-[--muted]">{updateData.current_version}</span> <FiArrowRight />
                        <span className="text-indigo-200">{updateData.release.version}</span></h4>

                    {serverStatus?.isDesktopSidecar && <Alert
                        intent="info"
                        description="Update Seanime from the desktop application."
                    />}

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full items-center !mt-4">
                        {!serverStatus?.isDesktopSidecar && <Modal
                            trigger={<Button leftIcon={<GrInstall className="text-2xl" />}>
                                Update now
                            </Button>}
                            contentClass="max-w-xl"
                            title={<span>Update Seanime</span>}
                        >
                            <div className="space-y-4">
                                <p>
                                    Seanime will perform an update by downloading and replacing existing files.
                                    Refer to the documentation for more information.
                                </p>
                                <Button className="w-full" onClick={handleInstallUpdate} disabled={isPending}>
                                    Download and Install
                                </Button>
                            </div>
                        </Modal>}
                        <div className="flex flex-1" />
                        <SeaLink href={updateData?.release?.html_url || ""} target="_blank">
                            <Button intent="white-subtle" rightIcon={<BiLinkExternal />}>See on GitHub</Button>
                        </SeaLink>
                        {!serverStatus?.isDesktopSidecar &&
                            <Button intent="white" leftIcon={<BiDownload />} onClick={() => setDownloaderOpen(true)}>Download</Button>}
                    </div>
                </div>
            </Modal>
        </>
    )

}

type DownloaderProps = {
    children?: React.ReactNode
    release?: Updater_Release
}

export function Downloader(props: DownloaderProps) {

    const [downloaderOpen, setDownloaderOpen] = useAtom(downloaderOpenAtom)
    const [destination, setDestination] = React.useState<string>("")
    const [asset, setAsset] = React.useState<string>("")

    const {
        children,
        release,
        ...rest
    } = props

    const { mutate, isPending } = useDownloadRelease()

    function handleDownloadRelease() {
        if (!asset || !destination) {
            return toast.error("Missing options")
        }
        mutate({ destination, download_url: asset }, {
            onSuccess: () => {
                setDownloaderOpen(false)
            },
        })
    }

    if (!release) return null

    return (
        <Modal
            open={downloaderOpen}
            onOpenChange={() => setDownloaderOpen(false)}
            title="Download new release"
            contentClass="space-y-4 max-w-2xl overflow-hidden"
        >
            <div>
                <RadioGroup
                    itemClass={cn(
                        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
                    )}
                    itemIndicatorClass="hidden"
                    itemLabelClass="font-normal tracking-wide line-clamp-1 truncate flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"
                    itemContainerClass={cn(
                        "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-2 w-full",
                        "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                        "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                        "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                    )}
                    value={asset}
                    onValueChange={v => !!v ? setAsset(v) : {}}
                    options={release.assets?.filter(n => !n.name.endsWith(".txt")).map((asset) => ({
                        label: asset.name,
                        value: asset.browser_download_url,
                    })) || []}
                />
            </div>
            <DirectorySelector
                label="Select destination"
                onSelect={setDestination}
                value={destination}
                rightAddon={`/seanime-${release.version}`}
            />
            <div className="flex gap-2 justify-end mt-2">
                <Button intent="white" leftIcon={<BiDownload />} onClick={handleDownloadRelease} loading={isPending}>Download</Button>
            </div>
        </Modal>
    )
}
