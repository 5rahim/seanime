import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import { VitePWA } from "vite-plugin-pwa"
import { tanstackRouter } from "@tanstack/router-vite-plugin"
import path from "path"

// https://vitejs.dev/config/
export default defineConfig(() => {
    const isDesktop = process.env.VITE_PUBLIC_PLATFORM === "desktop"
    const isElectronDesktop = process.env.VITE_PUBLIC_DESKTOP === "electron"
    const outDir = isDesktop ? (isElectronDesktop ? "out-denshi" : "out-desktop") : "out"

    return {
        envPrefix: "VITE_PUBLIC_",
        define: {
            "process.env": {},
        },
        plugins: [
            tanstackRouter({
                routesDirectory: "./src/routes",
                generatedRouteTree: "./src/routeTree.gen.ts",
                routeFileIgnorePattern: ".+((_components)|(_containers)|(_features)|(_hooks)|(_utils)|(_lib)|(_screens)|(_home)|(_atoms)|(_listeners)|(_tauri)|(_electron))",
                autoCodeSplitting: true,
            }),
            VitePWA({
                registerType: "autoUpdate",
                manifest: {
                    name: "Seanime",
                    short_name: "Seanime",
                    description: "Self-hosted, user-friendly media server for anime and manga.",
                    start_url: "/",
                    display: "standalone",
                    background_color: "#0F172A",
                    theme_color: "#0F172A",
                    icons: [
                        {
                            src: "/icons/android-chrome-192x192.png",
                            sizes: "192x192",
                            type: "image/png",
                            purpose: "maskable",
                        },
                        {
                            src: "/icons/android-chrome-512x512.png",
                            sizes: "512x512",
                            type: "image/png",
                            purpose: "maskable",
                        },
                        {
                            src: "/icons/apple-icon.png",
                            sizes: "180x180",
                            type: "image/png",
                            purpose: "any",
                        },
                    ],
                },
            }),
            react({
                babel: {
                    plugins: [
                        ["babel-plugin-react-compiler", { target: "18" }],
                    ],
                },
            }),
        ],
        resolve: {
            alias: {
                "@": path.resolve(__dirname, "./src"),
            },
        },
        build: {
            outDir,
            emptyOutDir: true,
            assetsDir: "static",
            rollupOptions: {
                output: {
                    manualChunks: (id) => {
                        if (id.includes("node_modules")) {
                            if (
                                id.includes("/node_modules/react/") ||
                                id.includes("/node_modules/react-dom/") ||
                                id.includes("/node_modules/react-compiler-runtime/") ||
                                id.includes("/node_modules/scheduler/") ||
                                id.includes("/node_modules/prop-types/")
                            ) {
                                return "react-vendor"
                            }

                            if (id.includes("@tanstack")) {
                                return "tanstack-vendor"
                            }

                            if (id.includes("jassub")) {
                                return "jassub-vendor"
                            }
                            if (id.includes("anime4k-webgpu")) {
                                return "anime4k-vendor"
                            }

                            if (id.includes("@vidstack") || id.includes("media-captions") || id.includes("media-icons")) {
                                return "vidstack-vendor"
                            }

                            if (id.includes("hls.js")) {
                                return "hls-vendor"
                            }

                            if (id.includes("@codemirror") || id.includes("@uiw") || id.includes("rehype")) {
                                return "code-editor-vendor"
                            }

                            if (
                                id.includes("@radix-ui") ||
                                id.includes("@headlessui") ||
                                id.includes("vaul") ||
                                id.includes("sonner") ||
                                id.includes("cmdk") ||
                                id.includes("class-variance-authority") ||
                                id.includes("clsx") ||
                                id.includes("tailwind-merge") ||
                                id.includes("floating-ui")
                            ) {
                                return "ui-vendor"
                            }

                            if (id.includes("motion") || id.includes("tailwindcss-animate")) {
                                return "animation-vendor"
                            }

                            if (id.includes("recharts")) {
                                return "charts-vendor"
                            }

                            if (
                                id.includes("date-fns") ||
                                id.includes("lodash") ||
                                id.includes("crypto-js") ||
                                id.includes("libphonenumber-js")
                            ) {
                                return "utils-vendor"
                            }

                            if (id.includes("react-icons")) {
                                return "icons-vendor"
                            }

                            if (id.includes("lucide-react")) {
                                return "lucide-icons-vendor"
                            }

                            return "vendor"
                        }
                    },
                },
            },
        },
    }
})
