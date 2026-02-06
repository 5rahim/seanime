/// <reference types="@rsbuild/core/types" />

interface ImportMetaEnv {
    readonly SEA_APP_TITLE: string
    readonly SEA_PUBLIC_PLATFORM: string
    readonly SEA_PUBLIC_DESKTOP: string
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}
