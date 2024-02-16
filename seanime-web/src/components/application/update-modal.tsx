"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { Update } from "@/lib/server/types"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"

type UpdateModalProps = {}


export const availableUpdateAtom = atom<boolean>(false)
export const updateModalOpenAtom = atom<boolean>(true)

export function UpdateModal(props: UpdateModalProps) {

    const {
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)
    const [availableUpdate, setAvailableUpdate] = useAtom(availableUpdateAtom)
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)

    const { data: updateData, isLoading } = useSeaQuery<Update>({
        queryKey: ["get-last-update"],
        endpoint: SeaEndpoints.LATEST_UPDATE,
        enabled: !!serverStatus,
    })

    React.useEffect(() => {
        if (updateData && updateData.release) {
            console.log(updateData)
            localStorage.setItem("latest-available-update", JSON.stringify(updateData.release.version))
            const latestVersionNotified = localStorage.getItem("notified-available-update")
            setAvailableUpdate(true)
            if (latestVersionNotified !== updateData.release.version) {
                setUpdateModalOpen(true)
            }
        }
    }, [updateData])

    const ignoreUpdate = () => {
        if (updateData && updateData.release) {
            localStorage.setItem("notified-available-update", updateData.release.version)
            setUpdateModalOpen(false)
        }
    }

    if (isLoading || !updateData || !updateData.release) return null

    return (
        <>
            <Modal
                isOpen={updateModalOpen}
                onClose={() => setUpdateModalOpen(false)}
                title={"Update Available"}
                size="xl"
            >
                <div
                    className="bg-[url(/pattern-2.svg)] z-[-1] w-full h-[10rem] absolute opacity-60 top-[-5rem] left-0 bg-no-repeat bg-right bg-contain"
                >
                    <div
                        className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-gray-900 to-transparent z-[-2]"
                    />
                </div>
                <div>
                    <h3>Seanime {updateData.release.version} is out!</h3>
                    <p className="text-[--muted] mb-2">A new version of Seanime is available on the GitHub repository.</p>
                    <p className="bg-[--background-color] rounded-[--radius] p-2 px-4">
                        {updateData.release.body.substring(0, updateData.release.body.indexOf("---")).split(/\s-\s/).map((line, index) => {
                            if (line.startsWith("##")) return <h5 key={index}>What's new?</h5>
                            if (line.includes("🚑️")) return <p key={index} className="text-red-300 font-semibold flex gap-2 items-center">{line}
                                <AiFillExclamationCircle /></p>
                            if (line.includes("🎉")) return <p key={index} className="text-brand-50">{line}</p>
                            if (line.includes("✨")) return <p key={index} className="text-yellow-100">{line}</p>
                            return (
                                <p key={index}>{line}<br /></p>
                            )
                        })}
                    </p>
                    <div className="flex gap-2 justify-end mt-2">
                        <Button intent="white-subtle" onClick={() => ignoreUpdate()}>Ignore</Button>
                        <Button intent="white" leftIcon={<BiDownload />} onClick={() => setUpdateModalOpen(false)}>Download now</Button>
                    </div>
                </div>
            </Modal>
        </>
    )

}
