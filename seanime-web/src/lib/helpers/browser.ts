import { __isElectronDesktop__ } from "@/types/constants"
import copy from "copy-to-clipboard"


export function openTab(url: string) {
    window.open(url, "_blank")
}

export async function copyToClipboard(text: string) {
    if (__isElectronDesktop__ && window.electron?.clipboard) {
        await window.electron.clipboard.writeText(text)
    } else {
        copy(text)
    }
}
