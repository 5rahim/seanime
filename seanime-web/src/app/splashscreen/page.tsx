"use client"

import { LoadingOverlay } from "@/components/ui/loading-spinner"
import Image from "next/image"
import React from "react"

export default function Page() {

    return (
        <LoadingOverlay showSpinner={false}>
            <Image
                src="/seanime-logo.png"
                alt="Launching..."
                priority
                width={100}
                height={100}
                className="animate-pulse"
            />
            Launching...
        </LoadingOverlay>
    )

}
