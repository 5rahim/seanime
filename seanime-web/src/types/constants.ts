export const __isDesktop__ = import.meta.env.VITE_PUBLIC_PLATFORM === "desktop" // Tauri
export const __isElectronDesktop__ = import.meta.env.VITE_PUBLIC_DESKTOP === "electron"
export const __isTauriDesktop__ = import.meta.env.VITE_PUBLIC_DESKTOP === "tauri"
export const HIDE_IMAGES = false
