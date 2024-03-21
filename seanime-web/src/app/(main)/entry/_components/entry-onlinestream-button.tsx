import { Button } from "@/components/ui/button"
import { MediaEntry } from "@/lib/server/types"
import Link from "next/link"
import React from "react"
import { FiPlayCircle } from "react-icons/fi"

type EntryOnlinestreamButtonProps = {
    children?: React.ReactNode
    entry: MediaEntry | undefined
}

export function EntryOnlinestreamButton(props: EntryOnlinestreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    if (!entry || entry.media?.status === "NOT_YET_RELEASED") return null

    return (
        <>
            <Link href={`/onlinestream?id=${entry?.mediaId}`}>
                <Button
                    intent="primary-subtle"
                    leftIcon={<FiPlayCircle />}
                >
                    Stream online
                </Button>
            </Link>
        </>
    )
}
