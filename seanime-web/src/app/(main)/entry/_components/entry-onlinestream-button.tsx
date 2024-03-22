import { serverStatusAtom } from "@/atoms/server-status"
import { Button } from "@/components/ui/button"
import { MediaEntry } from "@/lib/server/types"
import { useAtomValue } from "jotai/react"
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

    const status = useAtomValue(serverStatusAtom)

    if (!entry || entry.media?.status === "NOT_YET_RELEASED" || !status?.settings?.library?.enableOnlinestream) return null

    return (
        <>
            <Link href={`/onlinestream?id=${entry?.mediaId}`}>
                <Button
                    intent="primary-subtle"
                    leftIcon={<FiPlayCircle className="text-xl" />}
                >
                    Stream online
                </Button>
            </Link>
        </>
    )
}
