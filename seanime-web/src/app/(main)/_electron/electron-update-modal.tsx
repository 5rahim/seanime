"use client"
import { useDownloadMacDenshiUpdate } from "@/api/hooks/download.hooks"
import { useGetLatestUpdate } from "@/api/hooks/releases.hooks"
import { UpdateChangelogBody } from "@/app/(main)/_features/update/update-helper"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"
import { BiLinkExternal } from "react-icons/bi"
import { FiArrowRight } from "react-icons/fi"
import { GrInstall } from "react-icons/gr"
import { toast } from "sonner"

type UpdateModalProps = {
    collapsed?: boolean
}

const updateModalOpenAtom = atom<boolean>(false)
export const isUpdateInstalledAtom = atom<boolean>(false)
export const isUpdatingAtom = atom<boolean>(false)

export function ElectronUpdateModal(props: UpdateModalProps) {
    const serverStatus = useServerStatus()
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)
    const [isUpdating, setIsUpdating] = useAtom(isUpdatingAtom)

    const { data: updateData, isLoading, refetch } = useGetLatestUpdate(!!serverStatus && !serverStatus?.settings?.library?.disableUpdateCheck)

    useWebsocketMessageListener({
        type: WSEvents.CHECK_FOR_UPDATES,
        onMessage: () => {
            refetch().then(() => {
                checkElectronUpdate()
            })
        },
    })

    const [updateLoading, setUpdateLoading] = React.useState(true)
    const [electronUpdate, setUpdate] = React.useState<boolean>(false)
    const [updateError, setUpdateError] = React.useState("")
    const [isInstalled, setIsInstalled] = useAtom(isUpdateInstalledAtom)
    const [isDownloading, setIsDownloading] = React.useState(false)
    const [isDownloaded, setIsDownloaded] = React.useState(false)
    const [downloadProgress, setDownloadProgress] = React.useState(0)

    const isMacOS = window.electron?.platform === "darwin"
    const { mutate: downloadMacUpdate, isPending: isMacUpdatePending } = useDownloadMacDenshiUpdate()

    const checkElectronUpdate = React.useCallback(() => {
        if (serverStatus?.settings?.library?.disableUpdateCheck) return
        try {
            if (window.electron) {
                // Check if the update is available
                setUpdateLoading(true)
                window.electron.checkForUpdates()
                    .then((updateAvailable: boolean) => {
                        setUpdate(updateAvailable)
                        setUpdateLoading(false)
                    })
                    .catch((error: any) => {
                        logger("ELECTRON").error("Failed to check for updates", error)
                        setUpdateError(JSON.stringify(error))
                        setUpdateLoading(false)
                    })
            }
        }
        catch (e) {
            logger("ELECTRON").error("Failed to check for updates", e)
            setIsUpdating(false)
        }
    }, [serverStatus])

    React.useEffect(() => {
        checkElectronUpdate()

        // Listen for update events from Electron
        if (window.electron) {
            // Register listeners for update events
            const removeUpdateDownloaded = window.electron.on("update-downloaded", () => {
                if (!isMacOS) {
                    toast.success("Update downloaded and ready to install")
                    setIsDownloading(false)
                    setIsDownloaded(true)
                    setDownloadProgress(100)
                }
            })

            const removeUpdateError = window.electron.on("update-error", (error: string) => {
                logger("ELECTRON").error("Update error", error)
                if (!isMacOS) {
                    toast.error(`Update error: ${error}`)
                    setIsUpdating(false)
                    setIsDownloading(false)
                }
            })

            const removeDownloadProgress = window.electron.on("download-progress", (progress: { percent: number }) => {
                if (!isMacOS) {
                    setDownloadProgress(Math.round(progress.percent))
                }
            })

            const removeUpdateAvailable = window.electron.on("update-available", () => {
                setIsDownloading(true)
                setIsDownloaded(false)
                setDownloadProgress(0)
                toast.info("Update found, downloading...")
            })

            return () => {
                // Clean up listeners
                removeUpdateDownloaded?.()
                removeUpdateError?.()
                removeDownloadProgress?.()
                removeUpdateAvailable?.()
            }
        }
    }, [])

    React.useEffect(() => {
        if (updateData && updateData.release) {
            setUpdateModalOpen(true)
        }
    }, [updateData])

    // Auto-install when download completes if user already clicked install
    React.useEffect(() => {
        if (isDownloaded && isUpdating && !isDownloading) {
            // Retry installation now that download is complete
            handleInstallUpdate()
        }
    }, [isDownloaded, isUpdating, isDownloading])

    async function handleInstallUpdate() {
        if (!electronUpdate || isUpdating) return

        try {
            setIsUpdating(true)

            if (window.electron) {
                // macOS: Use manual download and install flow
                if (isMacOS) {
                    if (!updateData?.release?.version) {
                        toast.error("Update version not found")
                        setIsUpdating(false)
                        return
                    }

                    // Find the macOS arm64 asset
                    const macAsset = updateData.release.assets?.find(asset =>
                        asset.name.includes("denshi")
                        && asset.name.includes("MacOS")
                        && asset.name.includes("arm64")
                        && asset.name.endsWith(".zip")
                    )

                    if (!macAsset) {
                        toast.error("macOS update asset not found")
                        setIsUpdating(false)
                        return
                    }

                    toast.info("Downloading and installing update...")
                    setIsDownloading(true)

                    downloadMacUpdate({
                        download_url: macAsset.browser_download_url,
                        version: updateData.release.version,
                    }, {
                        onSuccess: () => {
                            setIsInstalled(true)
                            toast.success("Update installed! Closing app...")
                            // Close the app after a short delay
                            setTimeout(() => {
                                window.electron?.send("quit-app")
                            }, 2000)
                        },
                        onError: (error) => {
                            logger("ELECTRON").error("Failed to install macOS update", error)
                            toast.error(`Failed to install update: ${error.message}`)
                            setIsUpdating(false)
                            setIsDownloading(false)
                        },
                    })
                    return
                }

                // Windows/Linux: Use electron-updater flow
                // If not downloaded yet, trigger download first
                if (!isDownloaded) {
                    toast.info("Downloading update...")
                    setIsDownloading(true)

                    // Trigger update check which will start download
                    await window.electron.checkForUpdates()

                    // Wait for download to complete
                    // The update-downloaded event will set isDownloaded to true
                    return
                }

                // Kill the currently running server before installing update
                try {
                    toast.info("Shutting down server...")
                    await window.electron.killServer()
                }
                catch (e) {
                    logger("ELECTRON").error("Failed to kill server", e)
                }

                // Install update
                toast.info("Installing update...")
                await window.electron.installUpdate()
                setIsInstalled(true)

                // Electron will automatically restart the app
                toast.info("Update installed. Restarting app...")
            }
        }
        catch (e) {
            logger("ELECTRON").error("Failed to install update", e)
            toast.error(`Failed to install update: ${JSON.stringify(e)}`)
            setIsUpdating(false)
            setIsDownloading(false)
        }
    }

    if (serverStatus?.settings?.library?.disableUpdateCheck) return null

    if (isLoading || updateLoading || !updateData || !updateData.release) return null

    if (isInstalled) return (
        <div className="fixed top-0 left-0 w-full h-full bg-[--background] flex items-center z-[9999]">
            <div className="container max-w-4xl py-10">
                <div className="mb-4 flex justify-center w-full">
                    <img src="/seanime-logo.png" alt="logo" className="w-14 h-auto" />
                </div>
                <p className="text-center text-lg">
                    Update installed. Restart the app.
                </p>
            </div>
        </div>
    )

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
                onOpenChange={v => !isUpdating && setUpdateModalOpen(v)}
                contentClass="max-w-3xl"
            >
                <div className="space-y-2">
                    <h3 className="text-center">A new update is available!</h3>
                    <h4 className="font-bold flex gap-2 text-center items-center justify-center">
                        <span className="text-[--muted]">{updateData.current_version}</span> <FiArrowRight />
                        <span className="text-indigo-200">{updateData.release.version}</span></h4>

                    {!electronUpdate && !isMacOS && (
                        <Alert intent="warning">
                            This update is not yet available for desktop clients.
                            Wait a few minutes or check the GitHub page for more information.
                        </Alert>
                    )}

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full !mt-4">
                        {electronUpdate && !isMacOS && <Button
                            leftIcon={<GrInstall className="text-2xl" />}
                            onClick={handleInstallUpdate}
                            loading={isUpdating || isDownloading || isMacUpdatePending}
                            disabled={isLoading}
                        >

                            {isDownloading ? `Downloading... ${downloadProgress}%` :
                                isDownloaded ? "Install now" : "Download & Install"}
                        </Button>}
                        {electronUpdate && isMacOS && <Button
                            leftIcon={<GrInstall className="text-2xl" />}
                            onClick={handleInstallUpdate}
                            loading={isUpdating || isMacUpdatePending}
                            disabled={isLoading}
                        >
                            {(isMacUpdatePending) ? "Installing..." : "Install now"}
                        </Button>}
                        <div className="flex flex-1" />
                        {!updateData?.release?.tag_name?.includes("v2.") && <SeaLink href={updateData?.release?.html_url || ""} target="_blank">
                            <Button intent="white-subtle" rightIcon={<BiLinkExternal />}>See on GitHub</Button>
                        </SeaLink>}
                    </div>
                </div>
            </Modal>
        </>
    )
}
