import { useRouter } from "next/navigation"
import React from "react"

type OfflineDetectorProps = {}

export function OfflineDetector(props: OfflineDetectorProps) {

    const {
        ...rest
    } = props

    const router = useRouter()
    const [isOnline, setIsOnline] = React.useState(window.navigator.onLine)

    function handleIsOffline() {
        // router.push("/offline")

    }

    React.useEffect(() => {
        if (!window.navigator.onLine) {
            handleIsOffline()
        }
        const handleOnlineStatusChange = () => {
            setIsOnline(window.navigator.onLine)
            if (!window.navigator.onLine) {
                handleIsOffline()
            }
        }

        // Add event listeners for online/offline status changes
        window.addEventListener("online", handleOnlineStatusChange)
        window.addEventListener("offline", handleOnlineStatusChange)

        // Cleanup: Remove event listeners on component unmount
        return () => {
            window.removeEventListener("online", handleOnlineStatusChange)
            window.removeEventListener("offline", handleOnlineStatusChange)
        }
    }, [])

    return (
        <>

        </>
    )
}
