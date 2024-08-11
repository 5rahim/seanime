import { Anime_AnimeEntry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import Link from "next/link"
import React from "react"
import { FiPlayCircle } from "react-icons/fi"

type EntryOnlinestreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_AnimeEntry | undefined
}

export function EntryOnlinestreamButton(props: EntryOnlinestreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const status = useServerStatus()

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
