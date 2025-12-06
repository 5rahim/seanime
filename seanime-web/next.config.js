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
        qualities: [10, 20, 30, 40, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100],
    },
    transpilePackages: ["@uiw/react-textarea-code-editor", "@replit/codemirror-vscode-keymap", "media-chrome", "anime4k-webgpu"],
    assetPrefix: isProd ? undefined : (isDesktop ? `http://${internalHost}:43210` : undefined),
    reactCompiler: true,
    devIndicators: false,
    allowedDevOrigins: ["127.0.0.1", "localhost"],
    experimental: {
        isolatedDevBuild: true,
        browserDebugInfoInTerminal: !isProd,
    },
}

module.exports = nextConfig
