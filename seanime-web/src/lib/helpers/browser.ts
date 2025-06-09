import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import copy from "copy-to-clipboard"


export function openTab(url: string) {
    if (__isTauriDesktop__) {
        const { open } = require("@tauri-apps/plugin-shell")
        open(url)
    } else {
        window.open(url, "_blank")
    }
}

export async function copyToClipboard(text: string) {
    if (__isTauriDesktop__) {
        const { writeText } = require("@tauri-apps/plugin-clipboard-manager")
        await writeText(text)
    } else if (__isElectronDesktop__ && (window as any).electron?.clipboard) {
        await (window as any).electron.clipboard.writeText(text)
    } else {
        copy(text)
    }
}
