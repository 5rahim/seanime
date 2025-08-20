"use client"
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
            refetch().then(() => checkElectronUpdate())
        },
    })

    const [updateLoading, setUpdateLoading] = React.useState(true)
    const [electronUpdate, setUpdate] = React.useState<boolean>(false)
    const [updateError, setUpdateError] = React.useState("")
    const [isInstalled, setIsInstalled] = useAtom(isUpdateInstalledAtom)

    const checkElectronUpdate = React.useCallback(() => {
        try {
            if (window.electron) {
                // Check if the update is available
                setUpdateLoading(true);
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
    }, [])

    React.useEffect(() => {
        checkElectronUpdate()

        // Listen for update events from Electron
        if (window.electron) {
            // Register listeners for update events
            const removeUpdateDownloaded = window.electron.on("update-downloaded", () => {
                toast.info("Update downloaded and ready to install")
            })

            const removeUpdateError = window.electron.on("update-error", (error: string) => {
                logger("ELECTRON").error("Update error", error)
                toast.error(`Update error: ${error}`)
                setIsUpdating(false)
            })

            return () => {
                // Clean up listeners
                removeUpdateDownloaded?.()
                removeUpdateError?.()
            }
        }
    }, [])

    React.useEffect(() => {
        if (updateData && updateData.release) {
            setUpdateModalOpen(true)
        }
    }, [updateData])

    async function handleInstallUpdate() {
        if (!electronUpdate || isUpdating) return

        try {
            setIsUpdating(true)

            // Tell Electron to download and install the update
            if (window.electron) {
                toast.info("Downloading update...")

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
            logger("ELECTRON").error("Failed to download update", e)
            toast.error(`Failed to download update: ${JSON.stringify(e)}`)
            setIsUpdating(false)
        }
    }

    if (serverStatus?.settings?.library?.disableUpdateCheck) return null

    if (isLoading || updateLoading || !updateData || !updateData.release) return null

    if (isInstalled) return (
        <div className="fixed top-0 left-0 w-full h-full bg-[--background] flex items-center z-[9999]">
            <div className="container max-w-4xl py-10">
                <div className="mb-4 flex justify-center w-full">
                    <img src="/logo_2.png" alt="logo" className="w-36 h-auto" />
                </div>
                <p className="text-center text-lg">
                    Update installed. The app will restart automatically.
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

                    {!electronUpdate && (
                        <Alert intent="warning">
                            This update is not yet available for desktop clients.
                            Wait a few minutes or check the GitHub page for more information.
                        </Alert>
                    )}

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full !mt-4">
                        {!!electronUpdate && <Button
                            leftIcon={<GrInstall className="text-2xl" />}
                            onClick={handleInstallUpdate}
                            loading={isUpdating}
                            disabled={isLoading}
                        >
                            Update now
                        </Button>}
                        <div className="flex flex-1" />
                        <SeaLink href={updateData?.release?.html_url || ""} target="_blank">
                            <Button intent="white-subtle" rightIcon={<BiLinkExternal />}>See on GitHub</Button>
                        </SeaLink>
                    </div>
                </div>
            </Modal>
        </>
    )
}
