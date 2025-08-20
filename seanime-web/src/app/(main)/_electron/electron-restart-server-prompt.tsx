import { isUpdateInstalledAtom, isUpdatingAtom } from "@/app/(main)/_tauri/tauri-update-modal"
import { websocketConnectedAtom, websocketConnectionErrorCountAtom } from "@/app/websocket-provider"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button } from "@/components/ui/button"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { useAtom, useAtomValue } from "jotai/react"
import React from "react"
import { toast } from "sonner"

export function ElectronRestartServerPrompt() {

    const [hasRendered, setHasRendered] = React.useState(false)

    const [isConnected, setIsConnected] = useAtom(websocketConnectedAtom)
    const connectionErrorCount = useAtomValue(websocketConnectionErrorCountAtom)
    const [hasClickedRestarted, setHasClickedRestarted] = React.useState(false)
    const isUpdatedInstalled = useAtomValue(isUpdateInstalledAtom)
    const isUpdating = useAtomValue(isUpdatingAtom)

    React.useEffect(() => {
        (async () => {
            if (window.electron) {
                await window.electron.window.getCurrentWindow() // TODO: Isn't called
                setHasRendered(true)
            }
        })()
    }, [])

    const handleRestart = async () => {
        setHasClickedRestarted(true)
        toast.info("Restarting server...")
        if (window.electron) {
            window.electron.emit("restart-server")
            React.startTransition(() => {
                setTimeout(() => {
                    setHasClickedRestarted(false)
                }, 5000)
            })
        }
    }

    // Try to reconnect automatically
    const tryAutoReconnectRef = React.useRef(true)
    React.useEffect(() => {
        if (!isConnected && connectionErrorCount >= 10 && tryAutoReconnectRef.current && !isUpdatedInstalled) {
            tryAutoReconnectRef.current = false
            console.log("Connection error count reached 10, restarting server automatically")
            handleRestart()
        }
    }, [connectionErrorCount])

    React.useEffect(() => {
        if (isConnected) {
            setHasClickedRestarted(false)
            tryAutoReconnectRef.current = true
        }
    }, [isConnected])

    if (!hasRendered) return null

    // Not connected for 10 seconds
    return (
        <>
            {(!isConnected && connectionErrorCount < 10 && !isUpdating && !isUpdatedInstalled) && (
                <LoadingOverlay className="fixed left-0 top-0 z-[9999]">
                    <p>
                        The server connection has been lost. Please wait while we attempt to reconnect.
                    </p>
                </LoadingOverlay>
            )}

            <Modal
                open={!isConnected && connectionErrorCount >= 10 && !isUpdatedInstalled}
                onOpenChange={() => {}}
                hideCloseButton
                contentClass="max-w-2xl"
            >
                <LuffyError>
                    <div className="space-y-4 flex flex-col items-center">
                        <p className="text-lg max-w-sm">
                            The background server process has stopped responding. Please restart it to continue.
                        </p>

                        <Button
                            onClick={handleRestart}
                            loading={hasClickedRestarted}
                            intent="white-outline"
                            size="lg"
                            className="rounded-full"
                        >
                            Restart server
                        </Button>
                        <p className="text-[--muted] text-sm max-w-xl">
                            If this message persists after multiple tries, please relaunch the application.
                        </p>
                    </div>
                </LuffyError>
            </Modal>
        </>
    )
}
