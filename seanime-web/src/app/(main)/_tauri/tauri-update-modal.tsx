"use client"
import { useGetLatestUpdate } from "@/api/hooks/releases.hooks"
import { UpdateChangelogBody } from "@/app/(main)/_features/update/update-helper"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { check, Update } from "@tauri-apps/plugin-updater"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"
import { BiLinkExternal } from "react-icons/bi"
import { GrInstall } from "react-icons/gr"

type UpdateModalProps = {
    collapsed?: boolean
}


const updateModalOpenAtom = atom<boolean>(false)

export function TauriUpdateModal(props: UpdateModalProps) {
    const serverStatus = useServerStatus()
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)

    const { data: updateData, isLoading } = useGetLatestUpdate(!!serverStatus && !serverStatus?.settings?.library?.disableUpdateCheck)

    const { update: tauriUpdate } = useTauriUpdate()

    const isPending = false

    function handleInstallUpdate() {

    }

    if (serverStatus?.settings?.library?.disableUpdateCheck) return null

    if (isLoading || !updateData || !updateData.release || !tauriUpdate?.available) return null

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
                onOpenChange={v => setUpdateModalOpen(v)}
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

                    <UpdateChangelogBody updateData={updateData} />

                    <div className="flex gap-2 w-full !mt-4">
                        <Modal
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

                                <Button
                                    className="w-full"
                                    onClick={handleInstallUpdate}
                                    disabled={isPending}
                                >
                                    Install
                                </Button>
                            </div>
                        </Modal>
                        <div className="flex flex-1" />
                        <Link href={updateData?.release?.html_url || ""} target="_blank">
                            <Button intent="white-subtle" rightIcon={<BiLinkExternal />}>See on GitHub</Button>
                        </Link>
                    </div>
                </div>
            </Modal>
        </>
    )
}

export const useTauriUpdate = () => {
    const [update, setUpdate] = React.useState<Update | null>(null)
    const [error, setError] = React.useState("")

    React.useEffect(() => {
        (async () => {
            try {
                const update = await check()

                setUpdate(update)
            }
            catch (error) {
                console.error(error)
                setError(JSON.stringify(error))
            }
        })()
    }, [])

    return { update, error }
}
