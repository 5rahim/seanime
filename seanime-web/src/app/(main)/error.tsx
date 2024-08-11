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
                title="Something went wrong!"
            >
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
