import { Extension_Extension } from "@/api/generated/types"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React from "react"
import { FaLink } from "react-icons/fa"

type ExtensionDetailsProps = {
    extension: Extension_Extension
}

export function ExtensionDetails(props: ExtensionDetailsProps) {

    const {
        extension,
        ...rest
    } = props

    const isBuiltin = extension.manifestURI === "builtin"

    return (
        <>
            <div className="relative rounded-[--radius-md] size-12 bg-gray-900 overflow-hidden">
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

            <div className="space-y-2">
                <div className="flex items-center flex-wrap">
                    <p className="text-md font-semibold flex gap-2 flex-wrap">
                        {extension.name} {!!extension.version && <Badge className="rounded-[--radius-md] text-md">
                        {extension.version}
                    </Badge>}</p>

                    <div className="flex flex-1"></div>

                    {!!extension.website && <SeaLink
                        href={extension.website}
                        target="_blank"
                        className="inline-block"
                    >
                        <Button
                            size="sm"
                            intent="gray-outline"
                            leftIcon={<FaLink />}
                        >
                            Website
                        </Button>
                    </SeaLink>}
                </div>

                <p className="text-[--muted] text-sm text-pretty">
                    {extension.description}
                </p>

                <p className="text-md line-clamp-1">
                    <span className="text-[--muted]">ID:</span> <span className="">{extension.id}</span>
                </p>
                <p className="text-md line-clamp-1">
                    <span className="text-[--muted]">Author:</span> <span className="">{extension.author}</span>
                </p>
                <p className="text-md line-clamp-1">
                    <span className="text-[--muted]">Language:</span>
                    <span className="">{LANGUAGES_LIST[extension.lang?.toLowerCase()]?.nativeName || extension.lang}</span>
                </p>
                <p className="text-md line-clamp-1">
                    <span className="text-[--muted]">Programming language:</span> <span className="">{capitalize(extension.language)}</span>
                </p>
                {(!!extension.manifestURI && !isBuiltin) && <p className="text-md w-full">
                    <span className="text-[--muted]">Manifest URL:</span> <span className="">{extension.manifestURI}</span>
                </p>}
            </div>
        </>
    )
}
