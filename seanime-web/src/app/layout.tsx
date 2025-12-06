import { ElectronManager } from "@/app/(main)/_electron/electron-manager"
import { TauriManager } from "@/app/(main)/_tauri/tauri-manager"
import { ClientProviders } from "@/app/client-providers"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
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
        apple: "/icons/apple-icon.png",
    },
    appleWebApp: {
        capable: true,
        statusBarStyle: "black-translucent",
        title: "Seanime",
    },
    formatDetection: {
        telephone: false,
    },
    other: {
        "mobile-web-app-capable": "yes",
        "apple-mobile-web-app-capable": "yes",
        "apple-mobile-web-app-status-bar-style": "black-translucent",
        "apple-mobile-web-app-title": "Seanime",
    },
}

export default function RootLayout({ children }: {
    children: React.ReactNode
}) {
    return (
        <html lang="en" suppressHydrationWarning>
        {process.env.NODE_ENV === "development" && <head>
            <script src="https://unpkg.com/react-scan/dist/auto.global.js" async></script>
        </head>}
        <body className={inter.className} suppressHydrationWarning>
        {/* {process.env.NODE_ENV === "development" && <script src="http://localhost:8097"></script>} */}
        <ClientProviders>
            {__isTauriDesktop__ && <TauriManager />}
            {__isElectronDesktop__ && <ElectronManager />}
            {children}
        </ClientProviders>
        </body>
        </html>
    )
}


