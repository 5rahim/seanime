"use client"

import { AnimeEntryPage } from "@/app/(main)/entry/_containers/anime-entry-page"
import React from "react"

export const dynamic = "force-static"

export default function Page() {
    return (
        <AnimeEntryPage />
    )
}
