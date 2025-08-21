import { Extension_Extension, Extension_InvalidExtension, ExtensionRepo_UpdateData } from "@/api/generated/types"
import {
    useFetchExternalExtensionData,
    useInstallExternalExtension,
    useReloadExternalExtension,
    useUninstallExternalExtension,
} from "@/api/hooks/extensions.hooks"
import { ExtensionDetails } from "@/app/(main)/extensions/_components/extension-details"
import { ExtensionCodeModal } from "@/app/(main)/extensions/_containers/extension-code"
import { ExtensionUserConfigModal } from "@/app/(main)/extensions/_containers/extension-user-config"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { Tooltip } from "@/components/ui/tooltip"
import Image from "next/image"
import React from "react"
import { FaCode } from "react-icons/fa"
import { GrUpdate } from "react-icons/gr"
import { HiOutlineAdjustments } from "react-icons/hi"
import { LuEllipsisVertical, LuRefreshCcw } from "react-icons/lu"
import { RiDeleteBinLine } from "react-icons/ri"
import { TbCloudDownload } from "react-icons/tb"
import { toast } from "sonner"

type ExtensionCardProps = {
    extension: Extension_Extension
    updateData?: ExtensionRepo_UpdateData | undefined
    isInstalled: boolean
    userConfigError?: Extension_InvalidExtension | undefined
    allowReload?: boolean
}

export function ExtensionCard(props: ExtensionCardProps) {

    const {
        extension,
        updateData,
        isInstalled,
        userConfigError,
        allowReload,
        ...rest
    } = props

    const isBuiltin = extension.manifestURI === "builtin"

    const { mutate: reloadExternalExtension, isPending: isReloadingExtension } = useReloadExternalExtension()

    return (
        <div
            className={cn(
                "group/extension-card border border-[rgb(255_255_255_/_5%)] relative overflow-hidden",
                "bg-gray-950 rounded-md p-3",
                !!updateData && "border-[--green]",
            )}
        >
            <div
                className={cn(
                    "absolute z-[0] right-0 top-0 h-full w-full max-w-[150px] bg-gradient-to-l to-gray-950",
                    !isBuiltin && "max-w-[50%] from-indigo-950/20",
                )}
            ></div>

            <div className="absolute top-3 right-3 z-[2]">
                <div className=" flex flex-row gap-1 z-[2] flex-wrap justify-end">

                    {!!extension.userConfig && (
                        <>
                            <ExtensionUserConfigModal extension={extension} userConfigError={userConfigError}>
                                <div>
                                    <Tooltip
                                        side="top"
                                        trigger={<IconButton
                                            size="sm"
                                            intent={userConfigError ? "alert" : "gray-basic"}
                                            icon={<HiOutlineAdjustments />}
                                            className={cn(
                                                userConfigError && "animate-bounce",
                                            )}
                                        />}
                                    >Preferences</Tooltip>
                                </div>
                            </ExtensionUserConfigModal>
                        </>
                    )}

                    <ExtensionSettings extension={extension} isInstalled={isInstalled} updateData={updateData}>
                        <div>
                            <Tooltip
                                trigger={<IconButton
                                    size="sm"
                                    intent="gray-basic"
                                    icon={<LuEllipsisVertical />}
                                />}
                            >Info</Tooltip>
                        </div>
                    </ExtensionSettings>
                </div>
                <div className="flex flex-row gap-1 z-[2] flex-wrap">
                    {!isBuiltin && (
                        <ExtensionCodeModal extension={extension}>
                            <div>
                                <Tooltip
                                    trigger={<IconButton
                                        size="sm"
                                        intent="gray-basic"
                                        icon={<FaCode />}
                                    />}
                                    side="left"
                                >Code</Tooltip>
                            </div>
                        </ExtensionCodeModal>
                    )}

                    {(allowReload && !isBuiltin) && (
                        <div>
                            <Tooltip
                                side="right" trigger={<IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<LuRefreshCcw />}
                                onClick={() => {
                                    if (!extension.id) return toast.error("Extension has no ID")
                                    reloadExternalExtension({ id: extension.id })
                                }}
                                disabled={isReloadingExtension}
                            />}
                            >Reload</Tooltip>
                        </div>
                    )}
                </div>
            </div>

            <div className="z-[1] relative flex flex-col h-full">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-md size-12 flex-none bg-gray-900 overflow-hidden">
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
                        <Popover
                            className="text-sm cursor-pointer" trigger={<p className="opacity-30 mt-1 text-xs line-clamp-1 tracking-wide">
                            {extension.description}
                        </p>}
                        >
                            {extension.description}
                        </Popover>
                    </div>
                </div>

                {!!updateData && <Badge className="rounded-md absolute right-9 top-1" intent="success">
                    Update available
                </Badge>}

                <div className="flex gap-2 flex-wrap pt-4 flex-1 items-end">
                    {isBuiltin && <Badge className="rounded-md tracking-wide border-transparent px-0 italic opacity-50" intent="unstyled">
                        Built-in
                    </Badge>}
                    {!!extension.version && <Badge className="rounded-md tracking-wide">
                        {extension.version}
                    </Badge>}
                    {!isBuiltin && <Badge className="rounded-md" intent="unstyled">
                        {extension.author}
                    </Badge>}
                    <Badge className="rounded-md" intent="unstyled">
                        {/*{extension.lang.toUpperCase()}*/}
                        {LANGUAGES_LIST[extension.lang?.toLowerCase()]?.nativeName || extension.lang?.toUpperCase() || "Unknown"}
                    </Badge>
                    {/*<Badge className="rounded-md" intent="unstyled">*/}
                    {/*    {capitalize(extension.language)}*/}
                    {/*</Badge>*/}
                </div>

            </div>
        </div>
    )
}

type ExtensionSettingsProps = {
    extension: Extension_Extension
    children?: React.ReactElement
    isInstalled: boolean
    updateData?: ExtensionRepo_UpdateData | undefined
}

export function ExtensionSettings(props: ExtensionSettingsProps) {

    const {
        extension,
        children,
        isInstalled,
        updateData,
        ...rest
    } = props

    const isBuiltin = extension.manifestURI === "builtin"

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

            {!isBuiltin && (
                <>

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


                    {((!!fetchedExtensionData && fetchedExtensionData?.version !== extension.version) || !!updateData) && (
                        <AppLayoutStack>
                            <p className="">
                                Update available: <span className="font-bold text-white">{fetchedExtensionData?.version || updateData?.version}</span>
                            </p>
                            <Button
                                intent="white"
                                leftIcon={<TbCloudDownload className="text-lg" />}
                                loading={isInstalling}
                                onClick={() => {
                                    installExtension({
                                        manifestUri: fetchedExtensionData?.manifestURI || updateData?.manifestURI || "",
                                    })
                                }}
                            >
                                Install update
                            </Button>
                        </AppLayoutStack>
                    )}

                    <ConfirmationDialog {...confirmUninstall} />


                </>
            )}
        </Modal>
    )
}
