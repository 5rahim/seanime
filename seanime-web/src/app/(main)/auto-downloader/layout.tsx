"use client"
import { Card } from "@/components/ui/card"
import React from "react"


export default function Layout({ children }: { children: React.ReactNode }) {

    return (
        <div className="p-4 sm:p-8 space-y-4">
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>Auto Downloader</h2>
                    <p className="text-[--muted]">
                        Add and manage auto-downloading rules for your favorite anime.
                    </p>
                </div>
            </div>

            <Card className="p-0 overflow-hidden">
                {children}
            </Card>
        </div>
    )

}
