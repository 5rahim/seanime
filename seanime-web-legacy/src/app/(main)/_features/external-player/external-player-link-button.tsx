import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"

type ExternalPlayerLinkButtonProps = {}

export const __externalPlayerLinkButton_linkAtom = atom<string | null>(null)

export function ExternalPlayerLinkButton(props: ExternalPlayerLinkButtonProps) {

    const {} = props

    const router = useRouter()

    const [link, setLink] = useAtom(__externalPlayerLinkButton_linkAtom)

    if (!link) return null

    return (
        <>
            <div className="fixed bottom-2 right-2 z-50">
                <SeaLink href={link} target="_blank" prefetch={false}>
                    <Button
                        rounded
                        size="lg"
                        className="animate-bounce"
                        onClick={() => {
                            React.startTransition(() => {
                                setLink(null)
                            })
                        }}
                    >
                        Open media in external player
                    </Button>
                </SeaLink>
            </div>
        </>
    )
}
