import { defineConfig, loadEnv, RsbuildPluginAPI } from "@rsbuild/core"
import { pluginBabel } from "@rsbuild/plugin-babel"
import { pluginReact } from "@rsbuild/plugin-react"
import { RsdoctorRspackPlugin } from "@rsdoctor/rspack-plugin"
import { TanStackRouterRspack } from "@tanstack/router-plugin/rspack"
import { buildSync } from "esbuild"
import * as fs from "node:fs"
import path from "path"

const { publicVars } = loadEnv({ prefixes: ["SEA_"] })

const isElectronDesktop = process.env.SEA_PUBLIC_DESKTOP === "electron"
const distPath = isElectronDesktop ? "out-denshi" : "out"

export default defineConfig({
    plugins: [
        pluginReact(),
        { // run stuff before build
            name: "before-build",
            setup(api: RsbuildPluginAPI) {
                // api.onBeforeStartDevServer(processJassub)
                api.onBeforeBuild(processJassub)

                function processJassub() {
                    console.log("Running transpilation...")
                    const source = path.resolve(__dirname, "node_modules/jassub/dist/worker/worker.js")
                    const outDir = path.resolve(__dirname, "public", "jassub")
                    const outFile = path.join(outDir, "jassub-worker.js")

                    if (!fs.existsSync(outDir)) fs.mkdirSync(outDir, { recursive: true })

                    // transpile using esbuild (goated)
                    buildSync({
                        entryPoints: [source],
                        outfile: outFile,
                        bundle: true,
                        format: "iife",
                        define: {
                            "import.meta.url": "self.location.href",
                        },
                        minify: false,
                    })

                    // copy wasm files
                    const wasmSource = path.resolve(__dirname, "node_modules/jassub/dist/wasm/jassub-worker.wasm")
                    const wasmModernSource = path.resolve(__dirname, "node_modules/jassub/dist/wasm/jassub-worker-modern.wasm")
                    fs.copyFileSync(wasmSource, path.join(outDir, "jassub-worker.wasm"))
                    fs.copyFileSync(wasmModernSource, path.join(outDir, "jassub-worker-modern.wasm"))
                    console.log("Finished transpiling")
                }
            },
        },
        pluginBabel({
            include: /\.(?:jsx|tsx)$/,
            babelLoaderOptions(opts) {
                opts.plugins ??= []
                opts.plugins.push(["babel-plugin-react-compiler", { target: "18" }])
            },
        }),
    ].filter(Boolean),
    source: {
        entry: {
            index: "./src/main.tsx",
        },
        define: publicVars,
    },
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: { // dev server
        port: 43210,
        host: "0.0.0.0",
        headers: {
            "Cross-Origin-Embedder-Policy": "credentialless",
            "Cross-Origin-Opener-Policy": "same-origin",
        },
    },
    output: {
        cleanDistPath: true,
        sourceMap: !!process.env.RSDOCTOR,
        distPath: {
            root: distPath,
        },
        filename: {
            js: "[name].[contenthash:8].js",
            css: "[name].[contenthash:8].css",
        },
    },
    html: {
        template: "./index.html",
        title: "Seanime",
    },
    performance: {
        chunkSplit: {
            forceSplitting: {
                "hls": /hls\.js/,
            },
        },
    },
    tools: {
        // swc: {
        //   minify: true,
        // },
        rspack: {
            experiments: {
                outputModule: true,
            },
            output: { // redundant?
                chunkFilename: "static/js/async/[name].[contenthash:8].js",
            },
            optimization: {
                chunkIds: !!process.env.RSDOCTOR ? "named" : undefined,
            },
            plugins: [
                TanStackRouterRspack({
                    routesDirectory: "./src/routes",
                    generatedRouteTree: "./src/routeTree.gen.ts",
                    autoCodeSplitting: true,
                }),
                process.env.RSDOCTOR && new RsdoctorRspackPlugin({}),
            ].filter(Boolean),
            module: {
                rules: [
                    { // stops circular deps warning
                        test: /jassub\/dist\/.*\.js$/,
                        parser: {
                            worker: false,
                        },
                    },
                    { // don't emit these again
                        test: /\.wasm$/,
                        include: /node_modules[\\/]jassub/,
                        type: "asset/resource",
                        generator: {
                            emit: false,
                        },
                    },
                ],
            },
        },
    },
})
