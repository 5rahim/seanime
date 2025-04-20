import { Extension_Extension } from "@/api/generated/types"
import { useGetAllExtensions, useInstallExternalExtension } from "@/api/hooks/extensions.hooks"
import { AddExtensionModal } from "@/app/(main)/extensions/_containers/add-extension-modal"
import { ExtensionCard } from "@/app/(main)/extensions/_containers/extension-card"
import { InvalidExtensionCard, UnauthorizedExtensionPluginCard } from "@/app/(main)/extensions/_containers/invalid-extension-card"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { atom, useSetAtom } from "jotai"
import { orderBy } from "lodash"
import { useRouter } from "next/navigation"
import React from "react"
import { BiDotsVerticalRounded } from "react-icons/bi"
import { CgMediaPodcast } from "react-icons/cg"
import { GrInstallOption } from "react-icons/gr"
import { LuBlocks, LuDownload } from "react-icons/lu"
import { PiBookFill } from "react-icons/pi"
import { RiFolderDownloadFill } from "react-icons/ri"
import { TbReload } from "react-icons/tb"
import { toast } from "sonner"

type ExtensionListProps = {
    children?: React.ReactNode
}

export const __extensions_currentPageAtom = atom<"installed" | "marketplace">("installed")

export function ExtensionList(props: ExtensionListProps) {

    const {
        children,
        ...rest
    } = props

    const router = useRouter()

    const [checkForUpdates, setCheckForUpdates] = React.useState(false)

    const { data: allExtensions, isPending: isLoading, refetch } = useGetAllExtensions(checkForUpdates)

    const setPage = useSetAtom(__extensions_currentPageAtom)

    const {
        mutate: installExtension,
        data: installResponse,
        isPending: isInstalling,
    } = useInstallExternalExtension()

    function orderExtensions(extensions: Extension_Extension[] | undefined) {
        return extensions ?
            orderBy(extensions, ["name", "manifestUri"])
            : []
    }

    function isExtensionInstalled(extensionID: string) {
        return !!allExtensions?.extensions?.find(n => n.id === extensionID) ||
            !!allExtensions?.invalidExtensions?.find(n => n.id === extensionID)
    }

    const pluginExtensions = orderExtensions(allExtensions?.extensions ?? []).filter(n => n.type === "plugin")
    const animeTorrentExtensions = orderExtensions(allExtensions?.extensions ?? []).filter(n => n.type === "anime-torrent-provider")
    const mangaExtensions = orderExtensions(allExtensions?.extensions ?? []).filter(n => n.type === "manga-provider")
    const onlinestreamExtensions = orderExtensions(allExtensions?.extensions ?? []).filter(n => n.type === "onlinestream-provider")

    const nonvalidExtensions = (allExtensions?.invalidExtensions ?? []).filter(n => n.code !== "plugin_permissions_not_granted")
        .sort((a, b) => a.id.localeCompare(b.id))
    const pluginPermissionsNotGrantedExtensions = (allExtensions?.invalidExtensions ?? []).filter(n => n.code === "plugin_permissions_not_granted")
        .sort((a, b) => a.id.localeCompare(b.id))

    if (isLoading) return <LoadingSpinner />

    if (!allExtensions) return <LuffyError>
        Could not get extensions.
    </LuffyError>

    return (
        <AppLayoutStack className="gap-6">
            <div className="flex items-center gap-2 flex-wrap">
                <div>
                    <h2>
                        Extensions
                    </h2>
                    <p className="text-[--muted] text-sm">
                        Manage your plugins and content providers.
                    </p>
                </div>

                <div className="flex flex-1"></div>

                <div className="flex items-center gap-2">
                    {!!allExtensions?.hasUpdate?.length && (
                        <Button
                            className="rounded-full animate-pulse"
                            intent="success"
                            leftIcon={<LuDownload className="text-lg" />}
                            loading={isInstalling}
                            onClick={() => {
                                toast.info("Installing updates...")
                                allExtensions?.hasUpdate?.forEach(update => {
                                    installExtension({
                                        manifestUri: update.manifestURI,
                                    })
                                })
                            }}
                        >
                            Update all
                        </Button>
                    )}
                    <Button
                        className="rounded-full"
                        intent="gray-outline"
                        leftIcon={<TbReload className="text-lg" />}
                        disabled={isLoading}
                        onClick={() => {
                            setCheckForUpdates(true)
                            // React.startTransition(() => {
                            //     refetch()
                            // })
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
                            Add an extension/plugin
                        </Button>
                    </AddExtensionModal>

                    <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" />}>

                        <DropdownMenuItem
                            onClick={() => {
                                router.push("/extensions/playground")
                            }}
                        >
                            <span>Playground</span>
                        </DropdownMenuItem>

                        <DropdownMenuItem
                            onClick={() => {
                                setPage("marketplace")
                            }}
                        >
                            <span>Marketplace</span>
                        </DropdownMenuItem>
                    </DropdownMenu>
                </div>
            </div>


            {!!pluginPermissionsNotGrantedExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center">Permissions required</h3>

                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {pluginPermissionsNotGrantedExtensions.map(extension => (
                            <UnauthorizedExtensionPluginCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}
            {!!nonvalidExtensions?.length && (
                <Card className="p-4 space-y-6 border-red-800">

                    <h3 className="flex gap-3 items-center">Invalid extensions</h3>

                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {nonvalidExtensions.map(extension => (
                            <InvalidExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {/*<Card className="p-4 space-y-6">*/}

            {!!pluginExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><LuBlocks /> Plugins</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {pluginExtensions.map(extension => (
                            <ExtensionCard
                                key={extension.id}
                                extension={extension}
                                updateData={allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                                isInstalled={isExtensionInstalled(extension.id)}
                                userConfigError={allExtensions?.invalidUserConfigExtensions?.find(n => n.id == extension.id)}
                                allowReload={true}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!animeTorrentExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><RiFolderDownloadFill />Anime torrents</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {animeTorrentExtensions.map(extension => (
                            <ExtensionCard
                                key={extension.id}
                                extension={extension}
                                updateData={allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                                isInstalled={isExtensionInstalled(extension.id)}
                                userConfigError={allExtensions?.invalidUserConfigExtensions?.find(n => n.id == extension.id)}
                                allowReload
                            />
                        ))}
                    </div>
                </Card>
            )}


            {!!mangaExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><PiBookFill />Manga</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {mangaExtensions.map(extension => (
                            <ExtensionCard
                                key={extension.id}
                                extension={extension}
                                updateData={allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                                isInstalled={isExtensionInstalled(extension.id)}
                                userConfigError={allExtensions?.invalidUserConfigExtensions?.find(n => n.id == extension.id)}
                                allowReload
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!onlinestreamExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><CgMediaPodcast /> Online streaming</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {onlinestreamExtensions.map(extension => (
                            <ExtensionCard
                                key={extension.id}
                                extension={extension}
                                updateData={allExtensions?.hasUpdate?.find(n => n.extensionID === extension.id)}
                                isInstalled={isExtensionInstalled(extension.id)}
                                userConfigError={allExtensions?.invalidUserConfigExtensions?.find(n => n.id == extension.id)}
                                allowReload
                            />
                        ))}
                    </div>
                </Card>
            )}

            {/*</Card>*/}
        </AppLayoutStack>
    )
}
