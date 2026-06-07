export const __isElectronDesktop__ = import.meta.env.SEA_PUBLIC_DESKTOP === "electron"
export const __isDesktop__ = import.meta.env.SEA_PUBLIC_PLATFORM === "desktop" || __isElectronDesktop__
export const __clientPlatform__ = __isElectronDesktop__
    ? "denshi"
    : import.meta.env.SEA_PUBLIC_PLATFORM === "web"
        ? "web"
        : import.meta.env.SEA_PUBLIC_PLATFORM === "mobile"
            ? "mobile"
            : ""
export const HIDE_IMAGES = false

export const __CAST_ENABLED__ = false

