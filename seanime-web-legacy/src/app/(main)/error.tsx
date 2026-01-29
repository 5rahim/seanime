"use client"

import { LuffyError } from "@/components/shared/luffy-error"
import { Button } from "@/components/ui/button"
import React from "react"

export default function Error({
    error,
    reset,
}: {
    error: Error & { digest?: string }
    reset: () => void
}) {
    React.useEffect(() => {
        console.error(error)
    }, [error])

    return (
        <div className="flex justify-center">
            <LuffyError
                title="Client side error"
            >
                <p className="max-w-xl text-sm text-[--muted] mb-4">
                    {error.message || "An unexpected error occurred."}
                </p>
                <Button
                    onClick={
                        () => reset()
                    }
                >
                    Try again
                </Button>
            </LuffyError>
        </div>
    )
}
