"use client"

import { Button } from "@/components/ui/button"
import React from "react"

export function ElectronCrashScreenError() {

    return (
        <div className="flex items-center justify-center py-10">
            <div className="space-y-4">
                <Button
                    intent="primary-outline"
                    onClick={() => {
                        if ((window as any).electron) {
                            (window as any).electron.send("restart-app")
                        }
                    }}
                >
                    Restart Seanime
                </Button>
                <Button
                    intent="alert-subtle"
                    onClick={() => {
                        if ((window as any).electron) {
                            (window as any).electron.send("quit-app")
                        }
                    }}
                >
                    Quit Seanime
                </Button>
            </div>
        </div>
    )
}
