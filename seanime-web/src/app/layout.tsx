import { ClientProviders } from "@/components/client-providers"
import type { Metadata } from "next"
import { Inter } from "next/font/google"
import "./globals.css"
import React from "react"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
    title: "Seanime",
    description: "Manage your anime library.",
    icons: {
        icon: "/icons/favicon.ico",
    },
}

export default function RootLayout({ children }: {
    children: React.ReactNode
}) {
    return (
        <html lang="en" suppressHydrationWarning>
            <body className={inter.className} suppressHydrationWarning>
                <ClientProviders>
                    {children}
                </ClientProviders>
            </body>
        </html>
    )
}
