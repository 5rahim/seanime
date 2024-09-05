import { Updater_Update } from "@/api/generated/types"
import React from "react"
import { AiFillExclamationCircle } from "react-icons/ai"

type UpdateChangelogBodyProps = {
    updateData: Updater_Update | undefined
    children?: React.ReactNode
}

export function UpdateChangelogBody(props: UpdateChangelogBodyProps) {

    const {
        updateData,
        children,
        ...rest
    } = props

    const { body } = useUpdateChangelogBody(updateData)

    return (
        <>
            {body.some(n => n.includes("🚑️")) &&
                <p className="text-red-300 font-semibold flex gap-2 items-center">This update includes a critical patch</p>}
            <div className="rounded-[--radius] space-y-1.5">
                <h5>What's new?</h5>
                {body.map((line, index) => {
                    if (line.includes("🚑️")) return <p key={index} className="text-red-300 font-semibold flex gap-2 items-center">{line}
                        <AiFillExclamationCircle /></p>
                    if (line.includes("🎉")) return <p key={index} className="text-white">{line}</p>
                    if (line.includes("✨")) return <p key={index} className="text-white">{line}</p>
                    if (line.includes("⚡️")) return <p key={index} className="">{line}</p>
                    if (line.includes("💄")) return <p key={index} className="">{line}</p>
                    if (line.includes("🦺")) return <p key={index} className="">{line}</p>
                    if (line.includes("⬆️")) return <p key={index} className="">{line}</p>
                    if (line.includes("🏗️")) return <p key={index} className="">{line}</p>
                    if (line.includes("🚀")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔧")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔍")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔒")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔑")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔗")) return <p key={index} className="">{line}</p>
                    if (line.includes("🔨")) return <p key={index} className="">{line}</p>

                    return (
                        <p key={index} className="text-[--muted] pl-4 text-sm">{line}<br /></p>
                    )
                })}
            </div>
        </>
    )
}


export function useUpdateChangelogBody(updateData: Updater_Update | undefined) {
    const body = React.useMemo(() => {
        if (!updateData || !updateData.release) return []
        let body = updateData.release.body
        if (body.includes("---")) {
            body = body.split("---")[0]
        }
        return body.split(/\n/).filter((line) => line.trim() !== "" && line.trim().startsWith("-"))
    }, [updateData])

    return {
        body,
    }
}
