const isProd = process.env.NODE_ENV === 'production';
const isDesktop = process.env.NEXT_PUBLIC_PLATFORM === 'desktop';
const isTauriDesktop = process.env.NEXT_PUBLIC_DESKTOP === 'tauri';
const isElectronDesktop = process.env.NEXT_PUBLIC_DESKTOP === 'electron';
const internalHost = process.env.TAURI_DEV_HOST || '127.0.0.1';


/** @type {import('next').NextConfig} */
const nextConfig = {
    ...(isProd && {output: "export"}),
    distDir: isDesktop ? (isElectronDesktop ? "out-denshi" : "out-desktop") : undefined,
    cleanDistDir: true,
    reactStrictMode: false,
    images: {
        unoptimized: true,
    },
    transpilePackages: ["@uiw/react-textarea-code-editor", "@replit/codemirror-vscode-keymap"],
    assetPrefix: isProd ? undefined : (isDesktop ? `http://${internalHost}:43210` : undefined),
    experimental: {
        reactCompiler: true,
    },
    devIndicators: false,
}

module.exports = nextConfig
