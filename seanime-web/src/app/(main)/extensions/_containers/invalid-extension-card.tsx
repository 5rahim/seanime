import { Extension_InvalidExtension } from "@/api/generated/types"
import { ExtensionSettings } from "@/app/(main)/extensions/_containers/extension-card"
import { ExtensionCodeModal } from "@/app/(main)/extensions/_containers/extension-code"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { BiCog, BiInfoCircle } from "react-icons/bi"
import { FaCode } from "react-icons/fa"

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


    return (
        <div
            className={cn(
                "group/extension-card relative overflow-hidden",
                "bg-gray-950 border border-[rgb(255_255_255_/_5%)] rounded-md p-3",
            )}
        >
            <div
                className={cn(
                    "absolute z-[0] right-0 top-0 h-full w-full max-w-[150px] bg-gradient-to-l to-gray-950",
                    "max-w-[50%] from-red-950/20",
                )}
            ></div>

            <div className="absolute top-3 right-3 flex flex-col gap-2 z-[2]">
                <Modal
                    trigger={<IconButton
                        size="sm"
                        intent="alert-basic"
                        icon={<BiInfoCircle />}
                    />}
                    title="Error details"
                >
                    <p>
                        Seanime failed to load this extension.
                    </p>
                    <code className="code">
                        {extension.code}
                    </code>
                    <code className="code text-red-200">
                        {extension.reason}
                    </code>

                    <p className="whitespace-pre-wrap w-full max-w-full overflow-x-auto">
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

                        <ExtensionCodeModal extension={extension.extension}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<FaCode />}
                            />
                        </ExtensionCodeModal>
                    </>
                )}
                {(!!extension.extension?.id && !extension.extension?.manifestURI) && (
                    <>
                        <ExtensionCodeModal extension={extension.extension}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<FaCode />}
                            />
                        </ExtensionCodeModal>
                    </>
                )}
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-md size-12 bg-gray-900 overflow-hidden">
                        {!!extension.extension?.icon ? (
                            <Image
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
                        {extension.code === "invalid_payload" && "Invalid or obsolete code"}
                    </p>
                </div>

                <div className="flex gap-2">
                    {!!extension.extension?.version && <Badge className="rounded-md">
                        {extension.extension?.version}
                    </Badge>}
                    <Badge className="rounded-md" intent="unstyled">
                        {extension.extension?.author ?? "Unknown author"}
                    </Badge>
                    <Badge className="rounded-md" intent="unstyled">
                        {capitalize(extension.extension?.author ?? "?")}
                    </Badge>
                </div>

            </div>
        </div>
    )
}
