"use client"
import { HomeScreen } from "@/app/(main)/(library)/_home/home-screen"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export const dynamic = "force-static"

export default function Page() {

    const ts = useThemeSettings()

    return (
        <>
            <HomeScreen />
        </>
    )

}
