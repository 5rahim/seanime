const isProd = process.env.NODE_ENV === 'production';
const isDesktop = process.env.NEXT_PUBLIC_PLATFORM === 'desktop';
const internalHost = process.env.TAURI_DEV_HOST || '127.0.0.1';


/** @type {import('next').NextConfig} */
const nextConfig = {
    output: "export",
    distDir: isDesktop ? "out-desktop" : undefined,
    reactStrictMode: false,
    images: {
        unoptimized: true,
    },
    transpilePackages: ["@uiw/react-textarea-code-editor", "@replit/codemirror-vscode-keymap"],
    // Configure assetPrefix or else the server won't properly resolve your assets.
    assetPrefix: isProd ? undefined : (isDesktop ? `http://${internalHost}:43210` : undefined),
}

module.exports = nextConfig
