import { LuffyError } from "@/components/shared/luffy-error"
import { useQueryClient } from "@tanstack/react-query"
import { useLocation, useRouter } from "@tanstack/react-router"
import React from "react"

interface AppErrorBoundaryProps {
    error: any
    reset?: () => void
    resetErrorBoundary?: () => void
}

export function AppErrorBoundary({ error, reset, resetErrorBoundary }: AppErrorBoundaryProps) {
    const router = useRouter()
    const queryClient = useQueryClient()
    const location = useLocation()

    React.useEffect(() => {
        if (resetErrorBoundary) {
            resetErrorBoundary()
        }
        if (reset) {
            reset()
        }
    }, [location.pathname])

    const handleReset = () => {
        if (resetErrorBoundary) {
            resetErrorBoundary()
        }
        if (reset) {
            reset()
        }
        router.invalidate()
        queryClient.invalidateQueries()
    }

    return (
        <LuffyError
            title="Client side error"
            reset={handleReset}
        >
            <p className="text-[--muted]">
                {(error as Error)?.message || "An unexpected error occurred."}
            </p>
        </LuffyError>
    )
}
