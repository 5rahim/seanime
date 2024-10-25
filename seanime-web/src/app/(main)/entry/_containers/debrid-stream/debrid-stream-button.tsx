import { Anime_Entry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __anime_debridStreamingViewActiveAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai/react"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { HiOutlineServerStack } from "react-icons/hi2"

type DebridStreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_Entry
}

export function DebridStreamButton(props: DebridStreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const [viewActive, setViewActive] = useAtom(__anime_debridStreamingViewActiveAtom)

    return (
        <>
            {serverStatus?.debridSettings?.enabled && (
                <Button
                    intent={viewActive ? "alert-subtle" : "white-subtle"}
                    className="w-full"
                    size="md"
                    leftIcon={viewActive ? <AiOutlineArrowLeft className="text-xl" /> : <HiOutlineServerStack className="text-2xl" />}
                    onClick={() => setViewActive(p => !p)}
                >
                    {viewActive ? "Close Debrid streaming" : "Stream with Debrid"}
                </Button>
            )}
        </>
    )
}
