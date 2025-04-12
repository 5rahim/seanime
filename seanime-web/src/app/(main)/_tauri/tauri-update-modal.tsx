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
import { emit } from "@tauri-apps/api/event"
import { platform } from "@tauri-apps/plugin-os"
import { relaunch } from "@tauri-apps/plugin-process"
import { check, Update } from "@tauri-apps/plugin-updater"
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

export function TauriUpdateModal(props: UpdateModalProps) {
    const serverStatus = useServerStatus()
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)

    const [isUpdating, setIsUpdating] = useAtom(isUpdatingAtom)

    const { data: updateData, isLoading, refetch } = useGetLatestUpdate(!!serverStatus && !serverStatus?.settings?.library?.disableUpdateCheck)

    useWebsocketMessageListener({
        type: WSEvents.CHECK_FOR_UPDATES,
        onMessage: () => {
            refetch().then(() => checkTauriUpdate())
        },
    })

    const [updateLoading, setUpdateLoading] = React.useState(true)
    const [tauriUpdate, setUpdate] = React.useState<Update | null>(null)
    const [tauriError, setTauriError] = React.useState("")
    const [isInstalled, setIsInstalled] = useAtom(isUpdateInstalledAtom)

    const checkTauriUpdate = React.useCallback(() => {
        try {
            (async () => {
                try {
                    const update = await check()
                    setUpdate(update)
                    setUpdateLoading(false)
                }
                catch (error) {
                    logger("TAURI").error("Failed to check for updates", error)
                    setTauriError(JSON.stringify(error))
                    setUpdateLoading(false)
                }
            })()
        }
        catch (e) {
            logger("TAURI").error("Failed to check for updates", e)
            setIsUpdating(false)
        }
    }, [])

    React.useEffect(() => {
        checkTauriUpdate()
    }, [])

    const [currentPlatform, setCurrentPlatform] = React.useState("")

    React.useEffect(() => {
        (async () => {
            setCurrentPlatform(platform())
        })()
    }, [])


    async function handleInstallUpdate() {
        if (!tauriUpdate || isUpdating) return

        try {
            setIsUpdating(true)

            // Wait for the update be downloaded
            toast.info("Downloading update...")
            await tauriUpdate.download()
            // Kill the currently running server
            toast.info("Shutting down server...")
            await emit("kill-server")
            // Wait 1 second before installing the update
            toast.info("Installing update...")
            setTimeout(async () => {
                await tauriUpdate.install()
                setIsInstalled(true)
                // Relaunch the app once the update is installed
                // on macOS, the app will be closed and the user will have to reopen it
                if (currentPlatform === "macos") {
                    toast.info("Update installed. Please reopen the app.")
                } else {
                    toast.info("Relaunching app...")
                }

                await relaunch()
            }, 1000)
        }
        catch (e) {
            logger("TAURI").error("Failed to download update", e)
            toast.error(`Failed to download update: ${JSON.stringify(e)}`)
            setIsUpdating(false)
        }
    }

    React.useEffect(() => {
        if (updateData && updateData.release) {
            setUpdateModalOpen(true)
        }
    }, [updateData])

    if (serverStatus?.settings?.library?.disableUpdateCheck) return null

    if (isLoading || updateLoading || !updateData || !updateData.release) return null

    if (isInstalled) return (
        <div className="fixed top-0 left-0 w-full h-full bg-[--background] flex items-center z-[9999]">
            <div className="container max-w-4xl py-10">
                <div className="mb-4 flex justify-center w-full">
                    <img src="/logo_2.png" alt="logo" className="w-36 h-auto" />
                </div>
                <p className="text-center text-lg">
                    Update installed. Please reopen the app if it doesn't restart automatically.
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

                    {!tauriUpdate && (
                        <Alert intent="warning">
                            This update is not yet available for desktop clients.
                            Wait a few minutes or check the GitHub page for more information.
                        </Alert>
                    )}

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full !mt-4">
                        {!!tauriUpdate && <Button
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
