import { Extension_Extension } from "@/api/generated/types"
import { useUninstallExternalExtension } from "@/api/hooks/extensions.hooks"
import { ExtensionDetails } from "@/app/(main)/extensions/_components/extension-details"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { BiCog } from "react-icons/bi"
import { GrUpdate } from "react-icons/gr"
import { RiDeleteBinLine } from "react-icons/ri"
import { TbCloudDownload } from "react-icons/tb"

type ExtensionCardProps = {
    extension: Extension_Extension
}

export function ExtensionCard(props: ExtensionCardProps) {

    const {
        extension,
        ...rest
    } = props

    const isBuiltin = extension.manifestURI === "builtin"


    return (
        <div
            className={cn(
                "group/extension-card relative overflow-hidden",
                "bg-gray-950 border rounded-md p-3",
            )}
        >
            <div
                className={cn(
                    "absolute z-[0] right-0 top-0 h-full w-full max-w-[150px] bg-gradient-to-l to-gray-950",
                    !isBuiltin && "max-w-[50%] from-indigo-950/20",
                )}
            ></div>

            {isBuiltin && <p className="text-[--muted] text-xs absolute italic top-2 right-3">
                Built-in
            </p>}

            <div className="absolute top-3 right-3 flex flex-col gap-2 z-[2]">
                {!isBuiltin && (
                    <ExtensionSettings extension={extension}>
                        <IconButton
                            size="sm"
                            intent="gray-basic"
                            icon={<BiCog />}
                        />
                    </ExtensionSettings>
                )}
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-md size-12 bg-gray-900 overflow-hidden">
                        {!!extension.meta.icon ? (
                            <Image
                                src={extension.meta.icon}
                                alt="extension icon"
                                crossOrigin="anonymous"
                                fill
                                quality={100}
                                priority
                                className="object-cover"
                            />
                        ) : <div className="w-full h-full flex items-center justify-center">
                            <p className="text-2xl font-bold">
                                {(extension.name[0]).toUpperCase()}
                            </p>
                            {/*<p className="text-2xl opacity-20">*/}
                            {/*    {extension.type === "anime-torrent-provider" && <RiFolderDownloadFill />}*/}
                            {/*    {extension.type === "manga-provider" && <PiBookFill />}*/}
                            {/*    {extension.type === "onlinestream-provider" && <CgMediaPodcast />}*/}
                            {/*</p>*/}
                        </div>}
                    </div>

                    <div>
                        <p className="font-semibold line-clamp-1">
                            {extension.name}
                        </p>
                        <p className="text-[--muted] text-sm line-clamp-1 italic">
                            {extension.id}
                        </p>
                    </div>
                </div>

                <div className="flex gap-2">
                    {!!extension.version && <Badge className="rounded-md">
                        {extension.version}
                    </Badge>}
                    <Badge className="rounded-md" intent="unstyled">
                        {extension.author}
                    </Badge>
                    <Badge className="rounded-md" intent="unstyled">
                        {capitalize(extension.language)}
                    </Badge>
                </div>

            </div>
        </div>
    )
}

type ExtensionSettingsProps = {
    extension: Extension_Extension
    children?: React.ReactElement
}

export function ExtensionSettings(props: ExtensionSettingsProps) {

    const {
        extension,
        children,
        ...rest
    } = props

    // If the extension is installed, it will not have a payload
    const installed = !extension.payload

    const { mutate: uninstall, isPending: isUninstalling } = useUninstallExternalExtension()

    const confirmUninstall = useConfirmationDialog({
        title: `Remove ${extension.name}`,
        description: "This action cannot be undone.",
        onConfirm: () => {
            uninstall({
                id: extension.id,
            })
        },
    })

    return (
        <Modal
            trigger={children}
            contentClass="max-w-3xl"
        >
            {isUninstalling && <LoadingOverlay />}

            <ExtensionDetails extension={extension} />

            <div className="flex gap-2">

                {!installed && (
                    <>
                        <Button intent="primary-outline" leftIcon={<TbCloudDownload className="text-xl" />}>
                            Install
                        </Button>
                    </>
                )}

                {installed && (
                    <>
                        {<Button
                            intent="gray-outline"
                            leftIcon={<GrUpdate className="text-lg" />}
                            disabled={!extension.manifestURI}
                        >
                            Check for updates
                        </Button>}

                        <Button
                            intent="alert-subtle"
                            leftIcon={<RiDeleteBinLine className="text-xl" />}
                            onClick={confirmUninstall.open}
                        >
                            Uninstall
                        </Button>
                    </>
                )}

            </div>

            <ConfirmationDialog {...confirmUninstall} />
        </Modal>
    )
}
