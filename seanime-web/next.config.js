/** @type {import('next').NextConfig} */
const nextConfig = {
    output: "export",
    distDir: "../web",
    reactStrictMode: false,
    images: {
        unoptimized: true,
    },
}

module.exports = nextConfig
