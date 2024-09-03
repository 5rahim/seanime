"use client"

import { LoadingOverlay } from "@/components/ui/loading-spinner"
import Image from "next/image"
import React from "react"

export default function Page() {

    return (
        <LoadingOverlay showSpinner={false}>
            <Image
                src="/logo_2.png"
                alt="Launching..."
                priority
                width={180}
                height={180}
                className="animate-pulse"
            />
            Launching...
        </LoadingOverlay>
    )

}
