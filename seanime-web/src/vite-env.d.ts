/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_APP_TITLE: string
    readonly VITE_PUBLIC_PLATFORM: string
    readonly VITE_PUBLIC_DESKTOP: string
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}
