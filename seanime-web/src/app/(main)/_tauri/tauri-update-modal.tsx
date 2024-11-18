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
import { relaunch } from "@tauri-apps/plugin-process"
import { check, Update } from "@tauri-apps/plugin-updater"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"
import { BiLinkExternal } from "react-icons/bi"
import { GrInstall } from "react-icons/gr"
import { toast } from "sonner"


type UpdateModalProps = {
    collapsed?: boolean
}


const updateModalOpenAtom = atom<boolean>(false)

export function TauriUpdateModal(props: UpdateModalProps) {
    const serverStatus = useServerStatus()
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)

    const [isUpdating, setIsUpdating] = React.useState(false)

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


    async function handleInstallUpdate() {
        if (!tauriUpdate?.available || isUpdating) return

        try {
            setIsUpdating(true)

            // Wait for the update be downloaded
            await tauriUpdate.download()
            // Kill the currently running server
            await emit("kill-server")
            // Wait 1 second before installing the update
            setTimeout(async () => {
                await tauriUpdate.install()
                // Relaunch the app once the update is installed
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
                itemContentClass="text-brand-300"
            />
            <Modal
                open={updateModalOpen}
                onOpenChange={v => !isUpdating && setUpdateModalOpen(v)}
                contentClass="max-w-2xl"
            >
                <div
                    className="bg-[url(/pattern-2.svg)] z-[-1] w-full h-[4rem] absolute opacity-60 left-0 bg-no-repeat bg-right bg-cover"
                >
                    <div
                        className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>
                <div className="space-y-2">
                    <h3>Seanime {updateData.release.version} is out!</h3>
                    <p className="text-[--muted]">A new version of Seanime has been released.</p>

                    {!tauriUpdate?.available && (
                        <Alert intent="alert">
                            This update is not available for desktop clients.
                            Check the GitHub page for more information.
                        </Alert>
                    )}

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full !mt-4">
                        {tauriUpdate?.available && <Button
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
