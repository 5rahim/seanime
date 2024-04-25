import { PlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-modal"
import { PlaylistsList } from "@/app/(main)/(library)/_containers/playlists/_components/playlists-list"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Drawer } from "@/components/ui/drawer"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type PlaylistsModalProps = {}

export const __playlists_modalOpenAtom = atom(false)

export function PlaylistsModal(props: PlaylistsModalProps) {

    const {} = props

    const serverStatus = useServerStatus()
    const [isOpen, setIsOpen] = useAtom(__playlists_modalOpenAtom)

    return (
        <>
            <Drawer
                open={isOpen}
                onOpenChange={v => setIsOpen(v)}
                size="lg"
                side="bottom"
            >
                <div
                    className="!mt-0 bg-[url(/pattern-2.svg)] z-[-1] w-full h-[5rem] absolute opacity-30 top-0 left-0 bg-no-repeat bg-right bg-cover"
                >
                    <div
                        className="w-full absolute top-0 h-full bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>

                <div className="space-y-6">
                    <div className="flex flex-col md:flex-row justify-between items-center gap-4">
                        <h4>Playlists</h4>
                        <div className="flex gap-2 items-center md:pr-8">
                            <PlaylistModal
                                trigger={
                                    <Button intent="success" className="rounded-full">
                                        Add a playlist
                                    </Button>
                                }
                            />
                        </div>
                    </div>

                    {!serverStatus?.settings?.library?.autoUpdateProgress && <Alert
                        className="max-w-2xl mx-auto"
                        intent="warning-basic"
                        description={<>
                            <p>
                                You need to enable "Automatically update progress" to use playlists.
                            </p>
                        </>}
                    />}

                    <div className="">
                        <PlaylistsList />
                    </div>
                </div>
            </Drawer>
        </>
    )
}
