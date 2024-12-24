import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button } from "@/components/ui/button"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import Image from "next/image"
import React from "react"

export function LoadingOverlayWithLogo({ refetch }: { refetch?: () => void }) {
    return <LoadingOverlay showSpinner={false}>
        <Image
            src="/logo_2.png"
            alt="Loading..."
            priority
            width={180}
            height={180}
            className="animate-pulse"
        />
        <TextGenerateEffect className="text-lg mt-2 text-[--muted] animate-pulse" words={"S e a n i m e"} />

        {(process.env.NEXT_PUBLIC_PLATFORM === "desktop" && !!refetch) && (
            <Button
                onClick={() => window.location.reload()}
                className="mt-4"
                intent="gray-outline"
                size="sm"
            >Reload</Button>
        )}
    </LoadingOverlay>
}
