export function openTab(url: string) {
    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop") {
        const { open } = require("@tauri-apps/plugin-shell")
        open(url)
    } else {
        window.open(url, "_blank")
    }
}
