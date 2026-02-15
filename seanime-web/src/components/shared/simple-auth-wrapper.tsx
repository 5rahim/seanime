import { useGetSettings } from "@/api/hooks/settings.hooks"
import { useGetStatus } from "@/api/hooks/status.hooks"
import { serverAuthTokenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { useAtomValue } from "jotai"
import React from "react"

type SimpleAuthWrapperProps = {
    children?: React.ReactNode
}

// standalone auth wrapper for routes outside of main
export function SimpleAuthWrapper({ children }: SimpleAuthWrapperProps) {
    const { data: serverStatus, isLoading: isStatusLoading } = useGetStatus()
    const password = useAtomValue(serverAuthTokenAtom)

    // check if we need to verify authentication
    const shouldVerifyAuth = React.useMemo(() => {
        if (!serverStatus) return false
        return serverStatus.serverHasPassword
    }, [serverStatus])

    const { isLoading: isSettingsLoading, isError: isSettingsError } = useGetSettings({
        enabled: shouldVerifyAuth && !!password,
    })

    const [redirecting, setRedirecting] = React.useState(false)

    React.useEffect(() => {
        if (serverStatus) {
            if (serverStatus.serverHasPassword && !password) {
                setRedirecting(true)
                window.location.href = "/public/auth"
            }
        }
    }, [serverStatus, password])

    if (isStatusLoading || !serverStatus) return <LoadingOverlayWithLogo />

    if (redirecting) return <LoadingOverlayWithLogo />

    // if verification is needed
    if (shouldVerifyAuth) {
        // if password missing -> handled by effect (redirecting), return loading
        if (!password) return <LoadingOverlayWithLogo />
        // if settings loading -> return loading
        if (isSettingsLoading) return <LoadingOverlayWithLogo />
        // if settings error -> return loading (useServerQuery handles redirect)
        if (isSettingsError) return <LoadingOverlayWithLogo />
    }

    return <>{children}</>
}
