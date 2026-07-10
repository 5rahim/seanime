import { app, dialog, ipcMain, shell } from "electron"
import * as fs from "node:fs"
import * as path from "node:path"
import type { DenshiSettings } from "./denshi-settings"
import { log } from "./logging"

const archiver: any = require("archiver")

const MPVCORE_TEMP_SUBTITLE_EXTENSIONS = new Set([".srt", ".ass", ".ssa", ".vtt", ".ttml", ".stl", ".txt"])
const MPVCORE_MAX_SUBTITLE_BYTES = 20 * 1024 * 1024
const MPVCORE_ANIME4K_MAX_SHADERS = 512
const MPVCORE_MAX_CONFIG_BYTES = 1024 * 1024

const forcedLogging = [
    "MPV_PRISM_DEBUG_VIDEO",
    "MPV_PRISM_DEBUG_NATIVE",
    "MPV_PRISM_MPV_LOG_FILE",
    "MPV_PRISM_NATIVE_LOG_FILE",
].some(name => process.env[name] && process.env[name] !== "0")

type SettingsAccess = {
    get: () => DenshiSettings
    updateLogging: (enabled: boolean) => void
}

type Shader = {
    name: string
    path: string
}

let mpvCoreLogsReset = false
let mpvPrismMain: { dispose: () => void } | null = null

function getMpvCoreTempDirectory(): string {
    return path.join(app.getPath("temp"), "seanime-mpvcore")
}

function getMpvCoreConfigDirectory(): string {
    return path.join(app.getPath("userData"), "mpvcore")
}

function getMpvCoreConfigFilePath(): string {
    return path.join(getMpvCoreConfigDirectory(), "mpv.conf")
}

function getMpvCoreMpvLogPath(): string {
    const filePath = path.join(app.getPath("userData"), "mpv-prism-libmpv.log")
    return process.platform === "win32" ? filePath.replace(/\\/g, "/") : filePath
}

function getMpvCoreNativeLogPath(): string {
    const filePath = path.join(app.getPath("userData"), "mpv-prism-native.log")
    return process.platform === "win32" ? filePath.replace(/\\/g, "/") : filePath
}

function createFile(filePath: string): void {
    fs.mkdirSync(path.dirname(filePath), { recursive: true })
    fs.closeSync(fs.openSync(filePath, "a"))
}

function resetMpvCoreLogs(): void {
    if (mpvCoreLogsReset) return
    mpvCoreLogsReset = true

    const filePaths = new Set([
        process.env.MPV_PRISM_MPV_LOG_FILE,
        process.env.MPV_PRISM_NATIVE_LOG_FILE,
    ].filter((filePath): filePath is string => Boolean(filePath)))

    for (const filePath of filePaths) {
        try {
            createFile(filePath)
            fs.truncateSync(filePath, 0)
        }
        catch (error) {
            log.error(`[MpvCore] Failed to reset log file ${filePath}:`, error)
        }
    }
}

async function exportMpvCoreLogs(): Promise<string> {
    const prismLogs = [
        { path: process.env.MPV_PRISM_MPV_LOG_FILE || getMpvCoreMpvLogPath(), name: "mpv-prism-libmpv.log" },
        { path: process.env.MPV_PRISM_NATIVE_LOG_FILE || getMpvCoreNativeLogPath(), name: "mpv-prism-native.log" },
    ].filter(file => fs.existsSync(file.path))

    if (!prismLogs.length) {
        throw new Error("No MpvCore logs found. Enable logging and start playback first.")
    }

    const hasContent = prismLogs.some(file => {
        try {
            return fs.statSync(file.path).size > 0
        }
        catch {
            return false
        }
    })

    if (!hasContent) {
        throw new Error("The log files are empty. Please reproduce the issue before exporting them.")
    }

    const files = [...prismLogs]
    const denshiLogPath = log.transports.file.getFile().path
    if (denshiLogPath && fs.existsSync(denshiLogPath)) {
        files.push({ path: denshiLogPath, name: "seanime-denshi.log" })
    }

    const timestamp = new Date().toISOString().slice(0, 19).replace("T", "_").replaceAll(":", "-")
    const outputDirectory = path.join(app.getPath("downloads"), "Seanime")
    const outputPath = path.join(outputDirectory, `mpv-prism-logs_${timestamp}.zip`)
    fs.mkdirSync(outputDirectory, { recursive: true })

    try {
        await new Promise<void>((resolve, reject) => {
            const output = fs.createWriteStream(outputPath)
            const archive = archiver("zip", { zlib: { level: 9 } })

            output.on("close", resolve)
            output.on("error", reject)
            archive.on("warning", reject)
            archive.on("error", reject)
            archive.pipe(output)

            for (const file of files) {
                archive.file(file.path, { name: file.name })
            }

            void archive.finalize()
        })
    }
    catch (error) {
        fs.rmSync(outputPath, { force: true })
        throw error
    }

    shell.showItemInFolder(outputPath)
    return outputPath
}

function setMpvCoreLoggingEnabled(enabled: boolean, settings: SettingsAccess): boolean {
    if (settings.get().mpvPrismLogging !== enabled) {
        settings.updateLogging(enabled)

        if (process.platform === "win32") {
            void dialog.showMessageBox({
                type: "question",
                buttons: ["Restart Now", "Later"],
                defaultId: 0,
                title: "Restart Required",
                message: "A restart is required for MpvCore logging changes to take effect. Would you like to restart now?",
                cancelId: 1,
            }).then(({ response }) => {
                if (response === 0) {
                    app.relaunch()
                    app.exit(0)
                }
            })
        }
    }

    const shouldEnable = enabled || forcedLogging
    if (!shouldEnable) {
        delete process.env.MPV_PRISM_DEBUG_VIDEO
        delete process.env.MPV_PRISM_DEBUG_NATIVE
        delete process.env.MPV_PRISM_MPV_LOG_FILE
        delete process.env.MPV_PRISM_NATIVE_LOG_FILE
        log.info("[MpvCore] Logging disabled")
        return false
    }

    process.env.MPV_PRISM_DEBUG_VIDEO ||= "1"
    process.env.MPV_PRISM_DEBUG_NATIVE ||= "1"
    process.env.MPV_PRISM_MPV_LOG_FILE ||= getMpvCoreMpvLogPath()
    process.env.MPV_PRISM_NATIVE_LOG_FILE ||= getMpvCoreNativeLogPath()

    createFile(process.env.MPV_PRISM_MPV_LOG_FILE)
    createFile(process.env.MPV_PRISM_NATIVE_LOG_FILE)

    log.info("[MpvCore] Logging enabled:", JSON.stringify({
        forced: forcedLogging,
        mpvLogFile: process.env.MPV_PRISM_MPV_LOG_FILE,
        nativeLogFile: process.env.MPV_PRISM_NATIVE_LOG_FILE,
    }))
    return true
}

function createMpvCoreAnime4KDirectory(): string {
    const directory = path.join(app.getPath("userData"), "mpvcore-shaders")
    fs.mkdirSync(directory, { recursive: true })

    const readmePath = path.join(directory, "README.txt")
    if (!fs.existsSync(readmePath)) {
        fs.writeFileSync(readmePath, [
            "Seanime MpvCore shaders",
            "",
            "Place your custom shaders (e.g. .glsl, .hook) in this folder.",
            "",
            "You can enable individual custom shaders in the settings, or select one of the built-in Anime4K/CNN upscaler profiles.",
        ].join("\n"), "utf8")
    }

    try {
        const hasShaders = fs.readdirSync(directory).some(file => file.endsWith(".glsl") || file.endsWith(".hook"))
        if (!hasShaders) {
            const embeddedDirectory = path.join(app.getAppPath(), "assets/shaders")
            if (fs.existsSync(embeddedDirectory)) {
                for (const file of fs.readdirSync(embeddedDirectory)) {
                    if (file.endsWith(".glsl") || file.endsWith(".hook")) {
                        fs.copyFileSync(path.join(embeddedDirectory, file), path.join(directory, file))
                    }
                }
                log.info("[MpvCore] Copied embedded shaders to mpvcore-shaders")
            }
        }
    }
    catch (error) {
        log.error("[MpvCore] Failed to copy embedded shaders:", error)
    }

    return directory
}

function scanMpvCoreAnime4KDirectory(dir?: string): { directory: string, shaders: Shader[] } {
    const directory = path.resolve(dir?.trim() || createMpvCoreAnime4KDirectory())
    if (!fs.statSync(directory).isDirectory()) {
        throw new Error("Shader path must be a directory")
    }

    const shaders: Shader[] = []
    const walk = (currentDirectory: string, depth: number): void => {
        if (depth > 8 || shaders.length >= MPVCORE_ANIME4K_MAX_SHADERS) return

        for (const entry of fs.readdirSync(currentDirectory, { withFileTypes: true })) {
            if (shaders.length >= MPVCORE_ANIME4K_MAX_SHADERS) break
            const fullPath = path.join(currentDirectory, entry.name)
            if (entry.isDirectory()) {
                walk(fullPath, depth + 1)
                continue
            }
            if (!entry.isFile()) continue

            const extension = path.extname(entry.name).toLowerCase()
            if (extension !== ".glsl" && extension !== ".hook") continue
            shaders.push({
                name: path.relative(directory, fullPath).split(path.sep).join("/"),
                path: fullPath,
            })
        }
    }

    walk(directory, 0)
    shaders.sort((a, b) => a.name.localeCompare(b.name))
    return { directory, shaders }
}

function cleanupMpvCoreTempDirectory(): void {
    try {
        fs.rmSync(getMpvCoreTempDirectory(), { recursive: true, force: true })
    }
    catch (error) {
        log.warn("[MpvCore] Failed to clean temporary subtitle directory:", error)
    }
}

function sanitizeMpvCoreFilename(filename: string): { extension: string, stem: string } {
    const extension = path.extname(String(filename || "")).toLowerCase()
    if (!MPVCORE_TEMP_SUBTITLE_EXTENSIONS.has(extension)) {
        throw new Error("Unsupported subtitle file type")
    }

    const stem = path.basename(String(filename), extension)
        .replace(/[^a-zA-Z0-9._-]+/g, "-")
        .replace(/^[.-]+|[.-]+$/g, "")
        .slice(0, 80) || "subtitle"

    return { extension, stem }
}

function createUniqueMpvCoreFilename(stem: string, extension: string): string {
    const random = Math.random().toString(36).slice(2, 10)
    return `${stem}-${Date.now()}-${random}${extension}`
}

export function prepareMpvCore(settings: SettingsAccess): void {
    if (forcedLogging || settings.get().mpvPrismLogging) {
        setMpvCoreLoggingEnabled(true, settings)
        resetMpvCoreLogs()
    }
}

export async function initializeMpvCore(): Promise<void> {
    // @ts-ignore
    const { registerMpvPrismIpc } = await import("@mpv-prism/electron/main")
    mpvPrismMain = registerMpvPrismIpc({
        loader: { baseDirectory: app.getAppPath() },
    })
    cleanupMpvCoreTempDirectory()
}

export function disposeMpvCore(): void {
    mpvPrismMain?.dispose()
    mpvPrismMain = null
    cleanupMpvCoreTempDirectory()
}

export function registerMpvCoreIpc(settings: SettingsAccess): void {
    ipcMain.handle("mpvcore:create-temp-subtitle", async (_, filename: string, content: string) => {
        if (typeof content !== "string") throw new Error("Subtitle content must be text")
        if (Buffer.byteLength(content, "utf8") > MPVCORE_MAX_SUBTITLE_BYTES) {
            throw new Error("Subtitle file exceeds the 20 MiB limit")
        }

        const { extension, stem } = sanitizeMpvCoreFilename(filename)
        const directory = getMpvCoreTempDirectory()
        fs.mkdirSync(directory, { recursive: true })
        const target = path.join(directory, createUniqueMpvCoreFilename(stem, extension))
        fs.writeFileSync(target, content, "utf8")
        return target
    })

    ipcMain.handle("mpvcore:write-config-file", async (_, content: string) => {
        if (typeof content !== "string") throw new Error("MPV config must be text")

        const filePath = getMpvCoreConfigFilePath()
        if (!content.trim()) {
            fs.rmSync(filePath, { force: true })
            return null
        }
        if (Buffer.byteLength(content, "utf8") > MPVCORE_MAX_CONFIG_BYTES) {
            throw new Error("MPV config exceeds the 1 MiB limit")
        }

        fs.mkdirSync(getMpvCoreConfigDirectory(), { recursive: true })
        fs.writeFileSync(filePath, content, "utf8")
        return filePath
    })

    ipcMain.handle("mpvcore:create-screenshot-path", async () => {
        const directory = path.join(app.getPath("pictures"), "Seanime")
        fs.mkdirSync(directory, { recursive: true })
        return path.join(directory, createUniqueMpvCoreFilename("seanime", ".png"))
    })

    ipcMain.handle("mpvcore:save-screenshot", async (_, filePath: string, base64Data: string) => {
        fs.writeFileSync(filePath, Buffer.from(base64Data, "base64"))
        return true
    })

    ipcMain.handle("mpvcore:setLoggingEnabled", async (_, enabled: boolean) => {
        return setMpvCoreLoggingEnabled(enabled, settings)
    })
    ipcMain.handle("mpvcore:export-logs", exportMpvCoreLogs)
    ipcMain.handle("mpvcore:get-anime4k-directory", async () => {
        return scanMpvCoreAnime4KDirectory(createMpvCoreAnime4KDirectory())
    })
    ipcMain.handle("mpvcore:scan-anime4k-directory", async (_, directory: string) => {
        return scanMpvCoreAnime4KDirectory(directory)
    })
    ipcMain.handle("mpvcore:open-anime4k-directory", async (_, dir?: string) => {
        const directory = path.resolve(dir?.trim() || createMpvCoreAnime4KDirectory())
        if (!fs.statSync(directory).isDirectory()) throw new Error("Anime4K path must be a directory")
        const error = await shell.openPath(directory)
        if (error) throw new Error(error)
        return true
    })
}
