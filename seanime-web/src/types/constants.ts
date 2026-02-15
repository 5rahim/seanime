export const __isElectronDesktop__ = import.meta.env.SEA_PUBLIC_DESKTOP === "electron"
export const __isDesktop__ = import.meta.env.SEA_PUBLIC_PLATFORM === "desktop" || __isElectronDesktop__
export const HIDE_IMAGES = false
