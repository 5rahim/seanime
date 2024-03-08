import { PlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-modal"
import { PlaylistsList } from "@/app/(main)/(library)/_containers/playlists/_components/playlists-list"
import { serverStatusAtom } from "@/atoms/server-status"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Drawer } from "@/components/ui/drawer"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"

type PlaylistsModalProps = {
    children?: React.ReactNode
}

export const __playlists_modalOpenAtom = atom(false)

export function PlaylistsModal(props: PlaylistsModalProps) {

    const {
        children,
        ...rest
    } = props

    const serverStatus = useAtomValue(serverStatusAtom)
    const [isOpen, setIsOpen] = useAtom(__playlists_modalOpenAtom)

    return (
        <>
            <Drawer
                open={isOpen}
                onOpenChange={v => setIsOpen(v)}
                size="lg"
                side="bottom"
            >
                <div className="space-y-4">
                    <div className="flex flex-col md:flex-row justify-between items-center gap-4">
                        <p>Playlists</p>
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
                                You need to enable the "auto-update progress" feature to use playlists.
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
