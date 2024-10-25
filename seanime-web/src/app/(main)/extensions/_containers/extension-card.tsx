import { Extension_Extension, Extension_InvalidExtension } from "@/api/generated/types"
import { useFetchExternalExtensionData, useInstallExternalExtension, useUninstallExternalExtension } from "@/api/hooks/extensions.hooks"
import { ExtensionDetails } from "@/app/(main)/extensions/_components/extension-details"
import { ExtensionCodeModal } from "@/app/(main)/extensions/_containers/extension-code"
import { ExtensionUserConfigModal } from "@/app/(main)/extensions/_containers/extension-user-config"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { BiCog } from "react-icons/bi"
import { FaCode } from "react-icons/fa"
import { GrUpdate } from "react-icons/gr"
import { HiOutlineAdjustments } from "react-icons/hi"
import { RiDeleteBinLine } from "react-icons/ri"
import { TbCloudDownload } from "react-icons/tb"
import { toast } from "sonner"

type ExtensionCardProps = {
    extension: Extension_Extension
    hasUpdate: boolean
    isInstalled: boolean
    userConfigError?: Extension_InvalidExtension | undefined
}

export function ExtensionCard(props: ExtensionCardProps) {

    const {
        extension,
        hasUpdate,
        isInstalled,
        userConfigError,
        ...rest
    } = props

    const isBuiltin = extension.manifestURI === "builtin"

    return (
        <div
            className={cn(
                "group/extension-card border border-[rgb(255_255_255_/_5%)] relative overflow-hidden",
                "bg-gray-950 rounded-md p-3",
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

            <div className="absolute top-3 right-3 flex flex-col gap-1 z-[2]">
                {!isBuiltin && (
                    <ExtensionSettings extension={extension} isInstalled={isInstalled}>
                        <IconButton
                            size="sm"
                            intent="gray-basic"
                            icon={<BiCog />}
                        />
                    </ExtensionSettings>
                )}
                {extension.manifestURI === "" && (
                    <ExtensionCodeModal extension={extension}>
                        <IconButton
                            size="sm"
                            intent="gray-basic"
                            icon={<FaCode />}
                        />
                    </ExtensionCodeModal>
                )}
                {!!extension.userConfig && (
                    <>
                        <ExtensionUserConfigModal extension={extension} userConfigError={userConfigError}>
                            <IconButton
                                size="sm"
                                intent={userConfigError ? "alert" : "gray-basic"}
                                icon={<HiOutlineAdjustments />}
                                className={cn(
                                    userConfigError && "animate-bounce",
                                )}
                            />
                        </ExtensionUserConfigModal>
                    </>
                )}
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-md size-12 bg-gray-900 overflow-hidden">
                        {!!extension.icon ? (
                            <Image
                                src={extension.icon}
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

                <div className="flex gap-2 flex-wrap">
                    {!!extension.version && <Badge className="rounded-md">
                        {extension.version}
                    </Badge>}
                    <Badge className="rounded-md" intent="unstyled">
                        {extension.author}
                    </Badge>
                    <Badge className="rounded-md">
                        {extension.lang.toUpperCase()}
                    </Badge>
                    <Badge className="rounded-md" intent="unstyled">
                        {capitalize(extension.language)}
                    </Badge>
                    {hasUpdate && <Badge className="rounded-md" intent="success">
                        Update available
                    </Badge>}
                </div>

            </div>
        </div>
    )
}

type ExtensionSettingsProps = {
    extension: Extension_Extension
    children?: React.ReactElement
    isInstalled: boolean
}

export function ExtensionSettings(props: ExtensionSettingsProps) {

    const {
        extension,
        children,
        isInstalled,
        ...rest
    } = props

    // If the extension is installed, it will not have a payload
    // const installed = !extension.payload

    const { mutate: uninstall, isPending: isUninstalling } = useUninstallExternalExtension()

    const { mutate: fetchExtensionData, data: fetchedExtensionData, isPending: isFetchingData, reset } = useFetchExternalExtensionData(extension.id)

    const confirmUninstall = useConfirmationDialog({
        title: `Remove ${extension.name}`,
        description: "This action cannot be undone.",
        onConfirm: () => {
            uninstall({
                id: extension.id,
            })
        },
    })

    const {
        mutate: installExtension,
        data: installResponse,
        isPending: isInstalling,
    } = useInstallExternalExtension()

    React.useEffect(() => {
        if (installResponse) {
            toast.success(installResponse.message)
            reset()
        }
    }, [installResponse])

    const checkingForUpdatesRef = React.useRef(false)

    function handleCheckUpdate() {
        fetchExtensionData({
            manifestUri: extension.manifestURI,
        })
        checkingForUpdatesRef.current = true
    }

    React.useEffect(() => {

        if (fetchedExtensionData && checkingForUpdatesRef.current) {
            checkingForUpdatesRef.current = false

            if (fetchedExtensionData.version !== extension.version) {
                toast.success("Update available")
            } else {
                toast.info("The extension is up to date")
            }
        }
    }, [fetchedExtensionData])

    return (
        <Modal
            trigger={children}
            contentClass="max-w-3xl"
        >
            {isUninstalling && <LoadingOverlay />}

            <ExtensionDetails extension={extension} />

            {isInstalled && (
                <div className="flex gap-2">
                    <>
                        {!!extension.manifestURI && <Button
                            intent="gray-outline"
                            leftIcon={<GrUpdate className="text-lg" />}
                            disabled={!extension.manifestURI}
                            onClick={handleCheckUpdate}
                            loading={isFetchingData}
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
                </div>
            )}


            {(!!fetchedExtensionData && fetchedExtensionData?.version !== extension.version) && (
                <AppLayoutStack>
                    <p className="">
                        Update available: <span className="font-bold text-white">{fetchedExtensionData.version}</span>
                    </p>
                    <Button
                        intent="white"
                        leftIcon={<TbCloudDownload className="text-lg" />}
                        loading={isInstalling}
                        onClick={() => {
                            installExtension({
                                manifestUri: fetchedExtensionData.manifestURI,
                            })
                        }}
                    >
                        Install update
                    </Button>
                </AppLayoutStack>
            )}

            <ConfirmationDialog {...confirmUninstall} />
        </Modal>
    )
}
