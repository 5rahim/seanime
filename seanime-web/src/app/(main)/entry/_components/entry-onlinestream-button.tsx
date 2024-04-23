import { Anime_MediaEntry } from "@/api/generated/types"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { Button } from "@/components/ui/button"
import { useAtomValue } from "jotai/react"
import Link from "next/link"
import React from "react"
import { FiPlayCircle } from "react-icons/fi"

type EntryOnlinestreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_MediaEntry | undefined
}

export function EntryOnlinestreamButton(props: EntryOnlinestreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const status = useAtomValue(serverStatusAtom)

    if (
        !entry ||
        entry.media?.status === "NOT_YET_RELEASED" ||
        !status?.settings?.library?.enableOnlinestream ||
        entry.media?.isAdult
    ) return null

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
