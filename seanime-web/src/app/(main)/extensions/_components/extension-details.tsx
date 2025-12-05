import { Extension_Extension } from "@/api/generated/types"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { SeaImage } from "@/components/shared/sea-image"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import capitalize from "lodash/capitalize"
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
                    <SeaImage
                        src={extension.icon}
                        alt="extension icon"
                        crossOrigin="anonymous"
                        fill
                        quality={100}
                        isExternal
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

                <div className="flex gap-2 flex-wrap">
                    {isBuiltin && <Badge className="rounded-md tracking-wide border-transparent px-0 italic opacity-50" intent="unstyled">
                        Built-in
                    </Badge>}
                    {<Badge className="rounded-md tracking-wide" intent={"unstyled"}>
                        ID: {extension.id}
                    </Badge>}
                    {!isBuiltin && <Badge className="rounded-md" intent="unstyled">
                        Author: {extension.author}
                    </Badge>}
                    {<Badge className="rounded-md" intent="unstyled">
                        {/*{extension.lang.toUpperCase()}*/}
                        Language: {LANGUAGES_LIST[extension.lang?.toLowerCase()]?.nativeName || extension.lang?.toUpperCase() || "Unknown"}
                    </Badge>}
                    {<Badge className="rounded-md" intent="unstyled">
                        {/*{extension.lang.toUpperCase()}*/}
                        {capitalize(extension.language)}
                    </Badge>}
                </div>

                {(!!extension.manifestURI && !isBuiltin) && <p className="text-sm w-full tracking-wide">
                    <span className="text-[--muted]">Manifest URL:</span> <span className="select-all break-all">{extension.manifestURI}</span>
                </p>}

                {(!!extension.notes) && <div className="text-md w-full tracking-wide space-y-1 py-2">
                    <p className="text-[--muted] text-sm">Notes:</p>
                    <div className="text-pretty space-y-1">{extension.notes.split("\n").map((line, i) => {
                        // return <p>
                        //     {line.replaceAll("\t", "    ")}
                        // </p>
                        return <p
                            key={i} className="flex flex-wrap" dangerouslySetInnerHTML={{
                            __html: line.replaceAll("<", "&lt;")
                                .replaceAll(">", "&gt;").replaceAll("\t", "<span class='w-3 relative block'></span>"),
                        }}
                        ></p>
                    })}</div>
                </div>}

                {/*<p className="text-md line-clamp-1">*/}
                {/*    <span className="text-[--muted]">ID:</span> <span className="">{extension.id}</span>*/}
                {/*</p>*/}
                {/*<p className="text-md line-clamp-1">*/}
                {/*    <span className="text-[--muted]">Author:</span> <span className="">{extension.author}</span>*/}
                {/*</p>*/}
                {/*<p className="text-md line-clamp-1">*/}
                {/*    <span className="text-[--muted]">Language: </span>*/}
                {/*    <span className="">{LANGUAGES_LIST[extension.lang?.toLowerCase()]?.nativeName || extension.lang}</span>*/}
                {/*</p>*/}
                {/*<p className="text-md line-clamp-1">*/}
                {/*    <span className="text-[--muted]">Programming language:</span> <span className="">{capitalize(extension.language)}</span>*/}
                {/*</p>*/}
                {/*{(!!extension.manifestURI && !isBuiltin) && <p className="text-md w-full">*/}
                {/*    <span className="text-[--muted]">Manifest URL:</span> <span className="">{extension.manifestURI}</span>*/}
                {/*</p>}*/}
            </div>
        </>
    )
}
