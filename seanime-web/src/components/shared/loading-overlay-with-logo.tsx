import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button } from "@/components/ui/button"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { __isDesktop__ } from "@/types/constants"
import Image from "next/image"
import React from "react"

export function LoadingOverlayWithLogo({ refetch, title }: { refetch?: () => void, title?: string }) {
    return <LoadingOverlay showSpinner={false}>
        <Image
            src="/seanime-logo.png"
            alt="Loading..."
            priority
            width={100}
            height={100}
            className="animate-pulse"
        />
        <TextGenerateEffect className="text-lg mt-2 text-[--muted] animate-pulse" words={title ?? "S e a n i m e"} />

        {(__isDesktop__ && !!refetch) && (
            <Button
                onClick={() => window.location.reload()}
                className="mt-4"
                intent="gray-outline"
                size="sm"
            >Reload</Button>
        )}
    </LoadingOverlay>
}
