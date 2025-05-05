export const __isDesktop__ = process.env.NEXT_PUBLIC_PLATFORM === "desktop" // Tauri
export const __isElectronDesktop__ = process.env.NEXT_PUBLIC_DESKTOP === "electron"
export const __isTauriDesktop__ = process.env.NEXT_PUBLIC_DESKTOP === "tauri"
