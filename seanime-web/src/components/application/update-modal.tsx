"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { VerticalNav } from "@/components/ui/vertical-nav"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { Update } from "@/lib/server/types"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"

type UpdateModalProps = {}


export const updateModalOpenAtom = atom<boolean>(false)

export function UpdateModal(props: UpdateModalProps) {
    const serverStatus = useAtomValue(serverStatusAtom)
    const [updateModalOpen, setUpdateModalOpen] = useAtom(updateModalOpenAtom)

    const { data: updateData, isLoading } = useSeaQuery<Update>({
        queryKey: ["get-last-update"],
        endpoint: SeaEndpoints.LATEST_UPDATE,
        enabled: !!serverStatus,
    })

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

    const body = React.useMemo(() => {
        if (!updateData || !updateData.release) return []
        let body = updateData.release.body
        if (body.includes("---")) {
            body = body.split("---")[0]
        }
        return body.split(/\s+-\s+/).filter((line) => line.trim() !== "").map(n => (n.startsWith("-") || n.startsWith("##")) ? n : "- " + n)
    }, [updateData])

    if (isLoading || !updateData || !updateData.release) return null

    return (
        <>
            <VerticalNav
                items={[
                    {
                        icon: AiFillExclamationCircle,
                        name: "Update available",
                        onClick: () => setUpdateModalOpen(true),
                    },
                ]}
                iconClassName="text-brand-300"
            />
            <Modal
                isOpen={updateModalOpen}
                onClose={() => ignoreUpdate()}
                size="xl"
                isClosable
            >
                <div
                    className="bg-[url(/pattern-2.svg)] z-[-1] w-full h-[10rem] absolute opacity-60 top-[-5rem] left-0 bg-no-repeat bg-right bg-contain"
                >
                    <div
                        className="w-full absolute bottom-0 h-[4rem] bg-gradient-to-t from-gray-900 to-transparent z-[-2]"
                    />
                </div>
                <div className="space-y-2">
                    <h3>Seanime {updateData.release.version} is out!</h3>
                    <p className="text-[--muted]">A new version of Seanime is available on the GitHub repository.</p>
                    {body.some(n => n.includes("🚑️")) &&
                        <p className="text-red-300 font-semibold flex gap-2 items-center">This update includes a critical patch</p>}
                    <div className="bg-[--background-color] rounded-[--radius] p-2 px-4 space-y-1.5">
                        {body.map((line, index) => {
                            if (line.startsWith("##")) return <h5 key={index}>What's new?</h5>
                            if (line.includes("🚑️")) return <p key={index} className="text-red-300 font-semibold flex gap-2 items-center">{line}
                                <AiFillExclamationCircle /></p>
                            if (line.includes("🎉")) return <p key={index} className="text-white">{line}</p>
                            if (line.includes("✨")) return <p key={index} className="text-white">{line}</p>
                            if (line.includes("⚡️")) return <p key={index} className="">{line}</p>
                            if (line.includes("💄")) return <p key={index} className="">{line}</p>
                            if (line.includes("🦺")) return <p key={index} className="">{line}</p>
                            if (line.includes("⬆️")) return <p key={index} className="">{line}</p>
                            if (line.includes("🏗️")) return <p key={index} className="">{line}</p>
                            if (line.includes("🚀")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔧")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔍")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔒")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔑")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔗")) return <p key={index} className="">{line}</p>
                            if (line.includes("🔨")) return <p key={index} className="">{line}</p>

                            return (
                                <p key={index} className="text-[--muted] pl-4 text-sm">{line}<br /></p>
                            )
                        })}
                    </div>
                    <div className="flex gap-2 justify-end mt-2">
                        <Button intent="white" leftIcon={<BiDownload />} onClick={() => setUpdateModalOpen(false)}>Download now</Button>
                        <Button intent="white-subtle" onClick={() => ignoreUpdate()}>Ignore</Button>
                    </div>
                </div>
            </Modal>
        </>
    )

}
