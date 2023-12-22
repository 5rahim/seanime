/** @type {import('next').NextConfig} */
const nextConfig = {
    output: "export",
    distDir: 'web',
    cleanDistDir: true,
    images: {
        unoptimized: true,
    },
}

module.exports = nextConfig