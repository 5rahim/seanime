import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import React from "react"

export function ElectronCrashScreenError() {
    const [msg, setMsg] = React.useState("")
    const [isRendererCrash, setIsRendererCrash] = React.useState(false)

    React.useEffect(() => {
        if (window.electron) {
            const u = window.electron.on("crash", (msg: string, info?: { isRendererCrash?: boolean }) => {
                console.log("Received crash event", msg, info)
                setMsg(msg)
                if (info?.isRendererCrash) {
                    setIsRendererCrash(true)
                }
            })
            return () => {
                u?.()
            }
        }
    }, [])

    React.useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                if (window.electron) {
                    window.electron.send("quit-app")
                }
            } else if (e.key === "Enter") {
                if (window.electron) {
                    if (isRendererCrash) {
                        window.electron.send("restart-app")
                    } else {
                        window.electron.send("quit-app")
                    }
                }
            }
        }
        window.addEventListener("keydown", handleKeyDown)
        return () => {
            window.removeEventListener("keydown", handleKeyDown)
        }
    }, [isRendererCrash])

    return (
        <div className="px-4 space-y-4 px-10">
            <p>
                {msg || "An error occurred. Closing in 10 seconds."}
            </p>

            <Alert
                intent="warning"
                description={isRendererCrash 
                    ? "You can try reloading the window to resume your session." 
                    : "Make sure another instance of Seanime is not running or check the logs for more details."
                }
            />

            <div className="flex justify-center gap-3 pt-2">
                {isRendererCrash && (
                    <Button
                        intent="primary"
                        onClick={() => {
                            if (window.electron) {
                                window.electron.send("restart-app")
                            }
                        }}
                    >
                        Reload
                    </Button>
                )}
                <Button
                    intent="gray-outline"
                    onClick={() => {
                        if (window.electron) {
                            window.electron.send("quit-app")
                        }
                    }}
                >
                    Close
                </Button>
            </div>
        </div>
    )
}
