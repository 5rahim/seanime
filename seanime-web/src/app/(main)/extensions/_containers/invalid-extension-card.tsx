import { Extension_InvalidExtension } from "@/api/generated/types"
import { useGrantPluginPermissions, useReloadExternalExtension } from "@/api/hooks/extensions.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { ExtensionSettings } from "@/app/(main)/extensions/_containers/extension-card"
import { ExtensionCodeModal } from "@/app/(main)/extensions/_containers/extension-code"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { clientIdAtom } from "@/app/websocket-provider"
import { SeaImage } from "@/components/shared/sea-image"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { BiCog, BiInfoCircle } from "react-icons/bi"
import { FaCode } from "react-icons/fa"
import { LuRefreshCcw, LuShieldCheck } from "react-icons/lu"
import { toast } from "sonner"

type InvalidExtensionCardProps = {
    extension: Extension_InvalidExtension
    isInstalled: boolean
}

export function InvalidExtensionCard(props: InvalidExtensionCardProps) {

    const {
        extension,
        isInstalled,
        ...rest
    } = props

    const { mutate: reloadExternalExtension, isPending: isReloadingExtension } = useReloadExternalExtension()

    return (
        <div
            className={cn(
                "group/extension-card relative overflow-hidden",
                "bg-gray-900 border border-red-400/10 rounded-xl p-3",
            )}
        >
            <div
                className={cn(
                    "absolute z-[0] left-0 top-0 h-full w-full max-w-[150px] bg-gradient-to-r to-gray-900",
                    "max-w-[50%] from-red-900/10",
                )}
            ></div>

            <div className="absolute top-3 right-3 grid grid-cols-2 gap-1 p-1 rounded-[--radius-md] bg-gray-900 z-[2]">
                <Modal
                    trigger={<IconButton
                        size="sm"
                        intent="alert-basic"
                        icon={<BiInfoCircle />}
                    />}
                    title="Error details"
                    contentClass="max-w-2xl"
                >
                    <p>
                        Seanime failed to load this extension. If you aren't sure what this means, please contact the author.
                    </p>
                    <p>
                        Code: <strong>{extension.code}</strong>
                    </p>
                    <code className="code text-red-200">
                        {extension.reason}
                    </code>

                    <p className="whitespace-pre-wrap w-full max-w-full overflow-x-auto text-xs text-center tracking-wide text-[--muted]">
                        {extension.path}
                    </p>
                </Modal>
                {/*Show settings if extension has an ID and manifest URI*/}
                {/*This will allow the user to fetch updates or uninstall the extension*/}
                {(!!extension.extension?.id && !!extension.extension?.manifestURI) && (
                    <>
                        <ExtensionSettings extension={extension?.extension} isInstalled={isInstalled}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<BiCog />}
                            />
                        </ExtensionSettings>
                    </>
                )}

                <ExtensionCodeModal extension={extension.extension}>
                    <IconButton
                        size="sm"
                        intent="gray-basic"
                        icon={<FaCode />}
                    />
                </ExtensionCodeModal>

                <IconButton
                    size="sm"
                    intent="gray-basic"
                    icon={<LuRefreshCcw />}
                    onClick={() => {
                        if (!extension.extension?.id) return toast.error("Extension has no ID")
                        reloadExternalExtension({ id: extension.extension?.id ?? "" })
                    }}
                    disabled={isReloadingExtension}
                />
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-[--radius-md] size-12 bg-gray-900 overflow-hidden">
                        {!!extension.extension?.icon ? (
                            <SeaImage
                                src={extension.extension?.icon}
                                alt="extension icon"
                                crossOrigin="anonymous"
                                fill
                                quality={100}
                                priority
                                className="object-cover"
                            />
                        ) : <div className="w-full h-full flex items-center justify-center">
                            <p className="text-2xl font-bold">
                                {(extension.extension?.name?.[0] ?? "?").toUpperCase()}
                            </p>
                        </div>}
                    </div>

                    <div>
                        <p className="font-semibold line-clamp-1">
                            {extension.extension?.name ?? "Unknown"}
                        </p>
                        <p className="text-[--muted] text-sm line-clamp-1 italic">
                            {extension.extension?.id ?? "Invalid ID"}
                        </p>
                    </div>
                </div>

                <div>
                    <p className="text-red-400 text-sm">
                        {extension.code === "invalid_manifest" && "Manifest error"}
                        {extension.code === "invalid_semver_constraint" && "Incompatible with this version of Seanime"}
                        {extension.code === "invalid_payload" && "Invalid or incompatible code"}
                    </p>
                </div>

                <div className="flex gap-2">
                    {!!extension.extension?.version && <Badge className="rounded-[--radius-md]">
                        {extension.extension?.version}
                    </Badge>}
                    <Badge className="rounded-[--radius-md]" intent="unstyled">
                        {extension.extension?.author ?? "-"}
                    </Badge>
                    {extension.extension?.lang && <Badge className="rounded-[--radius-md]" intent="unstyled">
                        {extension.extension?.lang?.toUpperCase?.()}
                    </Badge>}
                </div>

            </div>
        </div>
    )
}

type UnauthorizedExtensionPluginCardProps = {
    extension: Extension_InvalidExtension
    isInstalled: boolean
}

const shouldGrantPluginPermissionsAtom = atom<string[]>([])

export function UnauthorizedExtensionPluginCard(props: UnauthorizedExtensionPluginCardProps) {

    const {
        extension,
        isInstalled,
        ...rest
    } = props

    const clientId = useAtomValue(clientIdAtom)
    const { mutate: grantPluginPermissions, isPending: isGrantingPluginPermissions } = useGrantPluginPermissions()
    const { mutate: reloadExternalExtension, isPending: isReloadingExtension } = useReloadExternalExtension()

    const [shouldGrantPluginPermissions, setShouldGrantPluginPermissions] = useAtom(shouldGrantPluginPermissionsAtom)

    useWebsocketMessageListener({
        type: "grant-plugin-permission-check",
        onMessage: (message: string) => {
            // message format: {extensionId}$$${code}
            if (message.startsWith(extension.extension?.id) && shouldGrantPluginPermissions.includes(extension.extension?.id ?? "")) {
                setShouldGrantPluginPermissions(p => p.filter(id => id !== extension.extension?.id ?? ""))
                grantPluginPermissions({ id: extension.extension?.id ?? "", clientId: "CODE:" + message.split("$$$")?.[1] || "" })
            }
        },
    })

    return (
        <div
            className={cn(
                "group/extension-card relative overflow-hidden",
                "bg-gray-900 border  rounded-xl p-3 border-yellow-400/10",
            )}
        >
            <div
                className={cn(
                    "absolute z-[0] left-0 top-0 h-full w-full max-w-[150px] bg-gradient-to-r to-gray-900",
                    "max-w-[50%] from-yellow-900/10",
                )}
            ></div>

            <div className="absolute top-3 right-3 flex flex-col gap-1 p-1 rounded-xl bg-gray-900 z-[2]">
                <Modal
                    trigger={<Button
                        size="sm"
                        intent="warning-basic"
                        leftIcon={<LuShieldCheck />}
                        className="animate-bounce"
                    >Grant</Button>}
                    title="Permissions required"
                    contentClass="max-w-2xl"
                >
                    <p>
                        The plugin <span className="font-bold">{extension.extension?.name}</span> is requesting the following permissions:
                    </p>

                    <p className="whitespace-pre-wrap w-full max-w-full overflow-x-auto text-md leading-relaxed text-left bg-[--subtle] p-2 rounded-md">
                        {extension.pluginPermissionDescription}
                    </p>

                    <p className="whitespace-pre-wrap w-full max-w-full overflow-x-auto text-sm text-center text-[--muted]">
                        {extension.path}
                    </p>

                    <ExtensionCodeModal extension={extension.extension} readOnly>
                        <Button
                            size="md"
                            intent="gray-subtle"
                        >
                            View code
                        </Button>
                    </ExtensionCodeModal>

                    <Button
                        size="md"
                        intent="success-subtle"
                        leftIcon={<LuShieldCheck className="size-5" />}
                        onClick={() => {
                            if (!extension.extension?.id) return toast.error("Extension has no ID")
                            setShouldGrantPluginPermissions(p => [...p, extension.extension?.id!])
                            React.startTransition(() => {
                                grantPluginPermissions({ id: extension.extension?.id ?? "", clientId: clientId || "" })
                            })
                        }}
                        loading={isGrantingPluginPermissions}
                    >
                        Grant permissions
                    </Button>
                </Modal>
                {/*Show settings if extension has an ID and manifest URI*/}
                {/*This will allow the user to fetch updates or uninstall the extension*/}
                {(!!extension.extension?.id && !!extension.extension?.manifestURI) && (
                    <>
                        <ExtensionSettings extension={extension?.extension} isInstalled={isInstalled}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<BiCog />}
                            />
                        </ExtensionSettings>
                    </>
                )}

                {/* <ExtensionCodeModal extension={extension.extension}>
                 <IconButton
                 size="sm"
                 intent="gray-basic"
                 icon={<FaCode />}
                 />
                 </ExtensionCodeModal>

                 <IconButton
                 size="sm"
                 intent="gray-basic"
                 icon={<LuRefreshCcw />}
                 onClick={() => {
                 if (!extension.extension?.id) return toast.error("Extension has no ID")
                 reloadExternalExtension({ id: extension.extension?.id ?? "" })
                 }}
                 disabled={isReloadingExtension}
                 /> */}
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-[--radius-md] size-12 bg-gray-900 overflow-hidden">
                        {!!extension.extension?.icon ? (
                            <SeaImage
                                src={extension.extension?.icon}
                                alt="extension icon"
                                crossOrigin="anonymous"
                                fill
                                quality={100}
                                priority
                                className="object-cover"
                            />
                        ) : <div className="w-full h-full flex items-center justify-center">
                            <p className="text-2xl font-bold">
                                {(extension.extension?.name?.[0] ?? "?").toUpperCase()}
                            </p>
                        </div>}
                    </div>

                    <div>
                        <p className="font-semibold line-clamp-1">
                            {extension.extension?.name ?? "Unknown"}
                        </p>
                        <p className="text-[--muted] text-xs line-clamp-1 italic">
                            {extension.extension?.id ?? "Invalid ID"}
                        </p>
                    </div>
                </div>

                <div>
                    <p className="text-red-400 text-sm">
                        {extension.code === "invalid_manifest" && "Manifest error"}
                        {extension.code === "invalid_payload" && "Invalid or incompatible code"}
                    </p>
                </div>

                <div className="flex gap-2">
                    {!!extension.extension.version && <Badge className="rounded-md tracking-wide" intent={"unstyled"}>
                        {extension.extension.version}
                    </Badge>}
                    <Badge className="rounded-md" intent="unstyled">
                        {extension.extension.author}
                    </Badge>
                    <Badge className="border-transparent rounded-md" intent="unstyled">
                        {/*{extension.extension.lang.toUpperCase()}*/}
                        {LANGUAGES_LIST[extension.extension.lang?.toLowerCase()]?.nativeName || extension.extension.lang?.toUpperCase() || "Unknown"}
                    </Badge>
                </div>

            </div>
        </div>
    )
}
