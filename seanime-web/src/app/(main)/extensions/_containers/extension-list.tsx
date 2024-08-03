import { Extension_Extension } from "@/api/generated/types"
import { useGetAllExtensions } from "@/api/hooks/extensions.hooks"
import { AddExtensionModal } from "@/app/(main)/extensions/_containers/add-extension-modal"
import { ExtensionCard } from "@/app/(main)/extensions/_containers/extension-card"
import { InvalidExtensionCard } from "@/app/(main)/extensions/_containers/invalid-extension-card"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import { orderBy } from "lodash"
import React from "react"
import { CgMediaPodcast } from "react-icons/cg"
import { GrInstallOption } from "react-icons/gr"
import { PiBookFill } from "react-icons/pi"
import { RiFolderDownloadFill } from "react-icons/ri"
import { TbReload } from "react-icons/tb"

type ExtensionListProps = {
    children?: React.ReactNode
}

export function ExtensionList(props: ExtensionListProps) {

    const {
        children,
        ...rest
    } = props

    const [checkForUpdates, setCheckForUpdates] = React.useState(false)

    const { data: allExtensions, isPending: isLoading, refetch } = useGetAllExtensions(checkForUpdates)

    function orderExtensions(extensions: Extension_Extension[] | undefined) {
        return extensions ?
            orderBy(extensions, ["name", "manifestUri"])
            : []
    }

    function isExtensionInstalled(extensionID: string) {
        return !!allExtensions?.extensions?.find(n => n.id === extensionID) ||
            !!allExtensions?.invalidExtensions?.find(n => n.id === extensionID)
    }

    if (isLoading) return <LoadingSpinner />

    if (!allExtensions) return <LuffyError>
        Could not get extensions.
    </LuffyError>

    return (
        <AppLayoutStack>
            <div className="flex items-center gap-2 flex-wrap">
                <h2>
                    Extensions
                </h2>

                <div className="flex flex-1"></div>

                <div className="flex items-center gap-2">
                    <Button
                        className="rounded-full"
                        intent="gray-outline"
                        leftIcon={<TbReload className="text-lg" />}
                        disabled={isLoading}
                        onClick={() => {
                            setCheckForUpdates(true)
                            React.startTransition(() => {
                                refetch()
                            })
                        }}
                    >
                        Check for updates
                    </Button>
                    <AddExtensionModal extensions={allExtensions.extensions}>
                        <Button
                            className="rounded-full"
                            intent="primary-subtle"
                            leftIcon={<GrInstallOption className="text-lg" />}
                        >
                            Add an extension
                        </Button>
                    </AddExtensionModal>
                </div>
            </div>
            <h3 className="flex gap-3 items-center"><RiFolderDownloadFill />Torrent providers</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "anime-torrent-provider").map(extension => (
                    <ExtensionCard
                        key={extension.id}
                        extension={extension}
                        hasUpdate={!!allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                        isInstalled={isExtensionInstalled(extension.id)}
                    />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><PiBookFill />Manga sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "manga-provider").map(extension => (
                    <ExtensionCard
                        key={extension.id}
                        extension={extension}
                        hasUpdate={!!allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                        isInstalled={isExtensionInstalled(extension.id)}
                    />
                ))}
            </div>
            <Separator />
            <h3 className="flex gap-3 items-center"><CgMediaPodcast /> Online streaming sources</h3>
            <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                {orderExtensions(allExtensions.extensions).filter(n => n.type === "onlinestream-provider").map(extension => (
                    <ExtensionCard
                        key={extension.id}
                        extension={extension}
                        hasUpdate={!!allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                        isInstalled={isExtensionInstalled(extension.id)}
                    />
                ))}
            </div>

            {!!allExtensions.invalidExtensions?.length && (
                <>
                    <Separator />

                    <h3 className="flex gap-3 items-center">Invalid extensions</h3>

                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {allExtensions.invalidExtensions.map(extension => (
                            <InvalidExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>

                </>
            )}

        </AppLayoutStack>
    )
}
