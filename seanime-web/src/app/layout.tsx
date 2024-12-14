import { TauriManager } from "@/app/(main)/_tauri/tauri-manager"
import { ClientProviders } from "@/app/client-providers"
import type { Metadata } from "next"
import { Inter } from "next/font/google"
import "./globals.css"
import React from "react"

export const dynamic = "force-static"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
    title: "Seanime",
    description: "Self-hosted, user-friendly media server for anime and manga.",
    icons: {
        icon: "/icons/favicon.ico",
    },
}

export default function RootLayout({ children }: {
    children: React.ReactNode
}) {
    return (
        <html lang="en" suppressHydrationWarning>
        {/*<head>*/}
        {/*    {process.env.NODE_ENV === "development" && <script src="https://unpkg.com/react-scan/dist/auto.global.js" async></script>}*/}
        {/*</head>*/}
        <body className={inter.className} suppressHydrationWarning>
        {/*{process.env.NODE_ENV === "development" && <script src="http://localhost:8097"></script>}*/}
        <ClientProviders>
            {process.env.NEXT_PUBLIC_PLATFORM === "desktop" && <TauriManager />}
            {children}
        </ClientProviders>
        </body>
        </html>
    )
}


