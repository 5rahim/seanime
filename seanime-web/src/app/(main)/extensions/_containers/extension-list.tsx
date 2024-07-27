import { Extension_Extension, ExtensionRepo_AllExtensions } from "@/api/generated/types"
import { AddExtensionModal } from "@/app/(main)/extensions/_containers/add-extension-modal"
import { ExtensionCard } from "@/app/(main)/extensions/_containers/extension-card"
import { InvalidExtensionCard } from "@/app/(main)/extensions/_containers/invalid-extension-card"
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
    allExtensions?: ExtensionRepo_AllExtensions
    isLoading: boolean
}

export function ExtensionList(props: ExtensionListProps) {

    const {
        children,
        allExtensions,
        isLoading,
        ...rest
    } = props

    function orderExtensions(extensions: Extension_Extension[] | undefined) {
        return extensions?.sort((a, b) => b.name.localeCompare(a.name))?.sort((a, b) => {
            if (a.manifestURI === "builtin") return -1
            return 0
        }) ?? []
    }

    if (isLoading) return <LoadingSpinner />

    if (!allExtensions) return <LuffyError>
        Could not get extensions.
    </LuffyError>

    return (
        <AppLayoutStack>
            <div className="flex items-center">
                <h2>
                    Extensions
                </h2>

                <div className="flex flex-1"></div>

                <AddExtensionModal extensions={allExtensions.extensions}>
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
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "anime-torrent-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><PiBookFill />Manga sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "manga-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><CgMediaPodcast /> Online streaming sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "onlinestream-provider").map(extension => (
                    <ExtensionCard key={extension.id} extension={extension} />
                ))}
            </div>

            {!!allExtensions.invalidExtensions?.length && (
                <>
                    <Separator />

                    <h3 className="flex gap-3 items-center">Invalid extensions</h3>

                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {allExtensions.invalidExtensions.map(extension => (
                            <InvalidExtensionCard key={extension.id} extension={extension} />
                        ))}
                    </div>

                </>
            )}

        </AppLayoutStack>
    )
}
