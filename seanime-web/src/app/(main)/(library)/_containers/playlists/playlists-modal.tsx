import { PlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-modal"
import { PlaylistsList } from "@/app/(main)/(library)/_containers/playlists/_components/playlists-list"
import { serverStatusAtom } from "@/atoms/server-status"
import { Vaul, VaulContent, VaulHeader, VaulTitle } from "@/components/shared/vaul"
import { Alert } from "@/components/ui/alert"
import { Button, CloseButton } from "@/components/ui/button"
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
            <Vaul open={isOpen} onOpenChange={v => setIsOpen(v)}>
                <VaulContent className="h-full mt-24 lg:mt-72 max-h-[90%]">
                    <CloseButton className="absolute top-2 right-2" onClick={() => setIsOpen(false)} />
                    <div className="w-full p-4 lg:p-8 space-y-4 overflow-y-auto" data-vaul-no-drag>
                        <VaulHeader className="flex flex-col md:flex-row justify-between items-center gap-4">
                            <VaulTitle>Playlists</VaulTitle>
                            <div className="flex gap-2 items-center">

                                <PlaylistModal
                                    trigger={
                                        <Button intent="success" className="rounded-full">
                                            Add a playlist
                                        </Button>
                                    }
                                />
                            </div>
                        </VaulHeader>

                        {!serverStatus?.settings?.library?.autoUpdateProgress && <Alert
                            className="max-w-2xl mx-auto"
                            intent="warning"
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
                </VaulContent>
            </Vaul>
        </>
    )
}
