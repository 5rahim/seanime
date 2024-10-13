"use client"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    return (
        <>
            {children}
        </>
    )

}


export const dynamic = "force-static"
