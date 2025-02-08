import copy from "copy-to-clipboard"


export function openTab(url: string) {
    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop") {
        const { open } = require("@tauri-apps/plugin-shell")
        open(url)
    } else {
        window.open(url, "_blank")
    }
}

export async function copyToClipboard(text: string) {
    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop") {
        const { writeText } = require("@tauri-apps/plugin-clipboard-manager")
        await writeText(text)
    } else {
        copy(text)
    }
}
