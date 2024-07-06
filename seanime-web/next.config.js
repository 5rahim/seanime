/** @type {import('next').NextConfig} */
const nextConfig = {
    output: "export",
    distDir: "out",
    reactStrictMode: false,
    images: {
        unoptimized: true,
    },
}

module.exports = nextConfig
