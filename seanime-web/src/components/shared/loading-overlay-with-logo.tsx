import { TextGenerateEffect } from "@/components/shared/styling/text-generate-effect"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import Image from "next/image"
import React from "react"

export function LoadingOverlayWithLogo() {
    return <LoadingOverlay showSpinner={false}>
        <Image
            src="/icons/android-chrome-192x192.png"
            alt="Loading..."
            priority
            width={80}
            height={80}
            className="animate-pulse"
        />
        <TextGenerateEffect className="text-lg mt-2 text-[--muted] animate-pulse" words={"S e a n i m e"} />
    </LoadingOverlay>
}
