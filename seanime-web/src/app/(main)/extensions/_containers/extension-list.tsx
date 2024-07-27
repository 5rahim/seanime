import { Extension_Extension } from "@/api/generated/types"
import { AddExtensionModal } from "@/app/(main)/extensions/_containers/add-extension-modal"
import { ExtensionCard } from "@/app/(main)/extensions/_containers/extension-card"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { CgMediaPodcast } from "react-icons/cg"
import { GrInstallOption } from "react-icons/gr"
import { PiBookFill } from "react-icons/pi"
import { RiFolderDownloadFill } from "react-icons/ri"

type ExtensionListProps = {
    children?: React.ReactNode
    extensions: Extension_Extension[]
    isLoading: boolean
}

export function ExtensionList(props: ExtensionListProps) {

    const {
        children,
        extensions,
        isLoading,
        ...rest
    } = props

    function orderExtensions(extensions: Extension_Extension[]) {
        return extensions.sort((a, b) => a.name.localeCompare(b.name))
    }

    if (isLoading) return <LoadingSpinner />

    if (extensions.length === 0) return <LuffyError title="No extensions installed" />

    return (
        <AppLayoutStack>
            <div className="flex items-center">
                <h2>
                    Extensions
                </h2>

                <div className="flex flex-1"></div>

                <AddExtensionModal extensions={extensions}>
                    <Button
                        className="rounded-full"
                        intent="primary-subtle"
                        leftIcon={<GrInstallOption className="text-lg" />}
                    >
                        Add extension
                    </Button>
                </AddExtensionModal>
            </div>
            <h3 className="flex gap-3 items-center"><RiFolderDownloadFill />Torrent providers</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(extensions).filter(n => n.type === "anime-torrent-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><PiBookFill />Manga sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(extensions).filter(n => n.type === "manga-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><CgMediaPodcast /> Online streaming sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(extensions).filter(n => n.type === "onlinestream-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>
        </AppLayoutStack>
    )
}
