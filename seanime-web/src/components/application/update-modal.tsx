"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { VerticalNav } from "@/components/ui/vertical-nav"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
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
            console.log(updateData)
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
        // let body = `- 🎉 New feature: Track progress on MyAnimeList - You can now link your MyAnimeList account to Seanime and automatically
        // update your progress - 🎉 New feature: Sync anime lists between AniList and MyAnimeList (Experimental) - New interface to sync your anime
        // lists when you link your MyAnimeList account - 🎉 New feature: Automatically download new episodes - Add rules (filters) that specify
        // which episodes to download based on parameters such as release group, resolution, episode numbers - Seanime will automatically parse the
        // Nyaa RSS feed and download new episodes based on your rules - ✨ Added scan summaries - You can now read detailed summaries of your latest
        // scan results, allowing you to see how files were matched - ✨ Added ability to automatically update progress without confirmation when you
        // finish an episode - ⚡️ Improved handling of AniList rate limits - Seanime will now pause and resume requests when rate limits are reached
        // without throwing errors. This fixes the largest issue pertaining to scanning. - ⚡️ AniList media with incorrect mapping to AniDB will be
        // accessible in a limited view (without metadata) instead of being hidden - ⚡️ Enhanced scanning mode is now stable and more accurate - 💄
        // UI improvements - 🦺 Fixed various UX issues - ⬆️ Updated dependencies`
        if (body.includes("---")) {
            body = body.split("---")[0]
        }
        return body.split(/\s+-\s+/).filter((line) => line.trim() !== "")
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
                <div>
                    <h3>Seanime {updateData.release.version} is out!</h3>
                    <p className="text-[--muted] mb-2">A new version of Seanime is available on the GitHub repository.</p>
                    <div className="bg-[--background-color] rounded-[--radius] p-2 px-4 space-y-1">
                        {body.map((line, index) => {
                            if (line.startsWith("##")) return <h5 key={index}>What's new?</h5>
                            if (line.includes("🚑️")) return <p key={index} className="text-red-300 font-semibold flex gap-2 items-center">{line}
                                <AiFillExclamationCircle /></p>
                            if (line.includes("🎉")) return <p key={index} className="text-rose-100">{line}</p>
                            if (line.includes("✨")) return <p key={index} className="text-orange-100">{line}</p>
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
