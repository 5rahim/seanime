import Image from "next/image"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import React from "react"

export function LoadingOverlayWithLogo() {
    return <LoadingOverlay hideSpinner>
        <Image
            src={"/icons/android-chrome-192x192.png"}
            alt={"Loading..."}
            priority
            width={80}
            height={80}
            className={"animate-bounce"}
        />
    </LoadingOverlay>
}