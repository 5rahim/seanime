import { useRouter } from "@/lib/navigation.ts"
import Mousetrap from "mousetrap"
import React from "react"

// This is only rendered on the Electron Desktop client
export function ElectronManager() {
    const { back, forward } = useRouter()

    React.useEffect(() => {
        if (!window.electron) return
        const isMac = window.electron?.platform === "darwin"
        const modifier = isMac ? "command" : "alt"

        Mousetrap.bind(`${modifier}+left`, (e) => {
            e.preventDefault()
            back()
            return false
        })

        Mousetrap.bind(`${modifier}+right`, (e) => {
            e.preventDefault()
            forward()
            return false
        })

        return () => {
            Mousetrap.unbind(`${modifier}+left`)
            Mousetrap.unbind(`${modifier}+right`)
        }
    }, [])

    return null
}
