import { app, dialog, ipcMain, shell } from "electron"
import * as fs from "node:fs"
import * as net from "node:net"
import * as os from "node:os"
import * as path from "node:path"
import { Transform } from "node:stream"
import { pipeline } from "node:stream/promises"
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

type MpvCoreLogFile = {
    path: string
    name: string
}

type LiteralLogRedaction = {
    value: string
    replacement: string
    caseInsensitive?: boolean
    wholeToken?: boolean
    pathPrefix?: boolean
}

function escapeRegExp(value: string): string {
    return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}

function getPathVariants(value: string): string[] {
    const variants = new Set<string>()
    const normalized = value.trim().replace(/[\\/]+$/, "")
    if (!normalized) return []

    variants.add(normalized)
    variants.add(normalized.replace(/\\/g, "/"))
    variants.add(normalized.replace(/\//g, "\\"))
    variants.add(normalized.replace(/\\/g, "\\\\"))
    return [...variants]
}

function getLiteralLogRedactions(): LiteralLogRedaction[] {
    const rules: LiteralLogRedaction[] = []
    const seen = new Set<string>()
    const add = (
        value: string | undefined,
        replacement: string,
        caseInsensitive = false,
        wholeToken = false,
        pathPrefix = false,
    ): void => {
        if (!value) return
        const trimmed = value.trim()
        if (!trimmed) return

        const key = `${caseInsensitive ? "i" : "s"}:${wholeToken ? "t" : "l"}:${pathPrefix ? "p" : "v"}:${trimmed}`
        if (seen.has(key)) return
        seen.add(key)
        rules.push({ value: trimmed, replacement, caseInsensitive, wholeToken, pathPrefix })
    }

    const homeDirectories = new Set<string>([
        app.getPath("home"),
        os.homedir(),
        process.env.HOME,
        process.env.USERPROFILE,
    ].filter((value): value is string => Boolean(value)))

    for (const homeDirectory of homeDirectories) {
        for (const variant of getPathVariants(homeDirectory)) {
            add(variant, "<HOME>", process.platform === "win32", false, true)
        }
    }

    const usernames = new Set([
        process.env.USER,
        process.env.USERNAME,
        process.env.LOGNAME,
        (() => {
            try {
                return os.userInfo().username
            }
            catch {
                return undefined
            }
        })(),
    ].filter((value): value is string => Boolean(value)))

    for (const username of usernames) {
        if (username.length >= 3) add(username, "<USER>", process.platform === "win32", true)
    }

    const hostname = os.hostname()
    if (hostname.length >= 3) add(hostname, "<HOST>", true, true)

    return rules.sort((a, b) => b.value.length - a.value.length)
}

function replaceLiteralLogValue(text: string, rule: LiteralLogRedaction): string {
    const flags = rule.caseInsensitive ? "gi" : "g"
    const escapedValue = escapeRegExp(rule.value)
    if (rule.pathPrefix) {
        return text.replace(
            new RegExp(`${escapedValue}(?=$|[^A-Za-z0-9._-])`, flags),
            rule.replacement,
        )
    }
    if (!rule.wholeToken) {
        return text.replace(new RegExp(escapedValue, flags), rule.replacement)
    }

    return text.replace(
        new RegExp(`(^|[^A-Za-z0-9._-])${escapedValue}(?=$|[^A-Za-z0-9._-])`, flags),
        `$1${rule.replacement}`,
    )
}

function _redactIpAddresses(text: string): string {
    const _withIpv4Redacted = text.replace(/\b(?:\d{1,3}\.){3}\d{1,3}\b/g, candidate => {
        if (net.isIP(candidate) !== 4) return candidate
        if (candidate === "0.0.0.0" || candidate.startsWith("127.")) return candidate
        return "<IPV4>"
    })

    return _withIpv4Redacted.replace(
        /(^|[^0-9A-Fa-f:])((?:[0-9A-Fa-f]{0,4}:){2,7}[0-9A-Fa-f]{0,4})(?=$|[^0-9A-Fa-f:])/g,
        (match, prefix: string, candidate: string) => {
            if (net.isIP(candidate) !== 6 || candidate === "::" || candidate === "::1") return match
            return `${prefix}<IPV6>`
        },
    )
}

function toAnonymizedText(text: string, literalRules: LiteralLogRedaction[]): string {
    let redacted = text

    redacted = redacted.replace(
        /([a-z][a-z0-9+.-]*:\/\/)([^\s\/:@]+):([^\s\/@]+)@/gi,
        "$1<USER>:<REDACTED>@",
    )
    redacted = redacted.replace(
        /([?&](?:access_token|refresh_token|token|api_key|apikey|client_secret|password|passwd|session(?:id)?|auth)=)[^&#\s]+/gi,
        "$1<REDACTED>",
    )
    redacted = redacted.replace(
        /((?:["']?)(?:access[_-]?token|refresh[_-]?token|api[_-]?key|client[_-]?secret|password|passwd|cookie|session[_-]?id)(?:["']?)\s*[:=]\s*)(["']?)([^"'\s,;]+)\2/gi,
        "$1$2<REDACTED>$2",
    )
    redacted = redacted.replace(
        /((?:["']?)authorization(?:["']?)\s*[:=]\s*)(["']?)(?:bearer\s+)?([^"'\s,;]+)\2/gi,
        "$1$2<REDACTED>$2",
    )
    redacted = redacted.replace(
        /\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{5,}\b/g,
        "<JWT>",
    )
    redacted = redacted.replace(
        /\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b/gi,
        "<EMAIL>",
    )

    for (const rule of literalRules) {
        redacted = replaceLiteralLogValue(redacted, rule)
    }

    redacted = redacted.replace(
        /([A-Za-z]:[\\/]{1,2}(?:Users|Documents and Settings)[\\/]{1,2})([^\\/\s"'<>]+)/gi,
        "$1<USER>",
    )
    redacted = redacted.replace(
        /(\/(?:home|Users)\/)([^/\s"'<>]+)/g,
        "$1<USER>",
    )
    redacted = redacted.replace(
        /(\b(?:username|user|login|account)\b["']?\s*[:=]\s*["']?)([^"',;\s]+)/gi,
        "$1<USER>",
    )
    redacted = redacted.replace(
        /(\b(?:hostname|computer(?:name)?|machine(?:name)?|device(?:name)?)\b["']?\s*[:=]\s*["']?)([^"',;\s]+)/gi,
        "$1<HOST>",
    )
    redacted = redacted.replace(
        /\b(?:[0-9A-F]{2}[:-]){5}[0-9A-F]{2}\b/gi,
        "<MAC>",
    )

    return _redactIpAddresses(redacted)
}

class MpvCoreLogAnonymizer extends Transform {
    private static readonly MAX_CARRY_CHARACTERS = 64 * 1024
    private carry = ""

    constructor(private readonly literalRules: LiteralLogRedaction[]) {
        super({ decodeStrings: false })
    }

    override _transform(chunk: Buffer | string, _encoding: BufferEncoding, callback: (error?: Error | null) => void): void {
        try {
            const text = this.carry + chunk.toString()
            const lastLineFeed = text.lastIndexOf("\n")
            if (lastLineFeed < 0) {
                if (text.length > MpvCoreLogAnonymizer.MAX_CARRY_CHARACTERS) {
                    const splitAt = text.length - MpvCoreLogAnonymizer.MAX_CARRY_CHARACTERS
                    this.push(toAnonymizedText(text.slice(0, splitAt), this.literalRules))
                    this.carry = text.slice(splitAt)
                } else {
                    this.carry = text
                }
                callback()
                return
            }

            this.push(toAnonymizedText(text.slice(0, lastLineFeed + 1), this.literalRules))
            this.carry = text.slice(lastLineFeed + 1)
            callback()
        }
        catch (error) {
            callback(error instanceof Error ? error : new Error(String(error)))
        }
    }

    override _flush(callback: (error?: Error | null) => void): void {
        try {
            if (this.carry) this.push(toAnonymizedText(this.carry, this.literalRules))
            callback()
        }
        catch (error) {
            callback(error instanceof Error ? error : new Error(String(error)))
        }
    }
}

async function toAnonymizedLogs(files: MpvCoreLogFile[], directory: string): Promise<MpvCoreLogFile[]> {
    const literalRules = getLiteralLogRedactions()
    const anonymizedFiles: MpvCoreLogFile[] = []

    for (const file of files) {
        const outputPath = path.join(directory, path.basename(file.name))
        await pipeline(
            fs.createReadStream(file.path, { encoding: "utf8" }),
            new MpvCoreLogAnonymizer(literalRules),
            fs.createWriteStream(outputPath, { encoding: "utf8" }),
        )
        anonymizedFiles.push({ path: outputPath, name: file.name })
    }

    return anonymizedFiles
}

async function exportMpvCoreLogs(): Promise<string> {
    const prismLogs: MpvCoreLogFile[] = [
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
    const outputDir = path.join(app.getPath("downloads"), "Seanime")
    const outputPath = path.join(outputDir, `mpv-prism-logs_${timestamp}.zip`)
    const tempDir = fs.mkdtempSync(path.join(app.getPath("temp"), "seanime-mpvcore-logs-"))
    fs.mkdirSync(outputDir, { recursive: true })

    try {
        const anonymizedFiles = await toAnonymizedLogs(files, tempDir)

        await new Promise<void>((resolve, reject) => {
            const output = fs.createWriteStream(outputPath)
            const archive = archiver("zip", { zlib: { level: 9 } })

            output.on("close", resolve)
            output.on("error", reject)
            archive.on("warning", reject)
            archive.on("error", reject)
            archive.pipe(output)

            for (const file of anonymizedFiles) {
                archive.file(file.path, { name: file.name })
            }

            void archive.finalize()
        })
    }
    catch (error) {
        fs.rmSync(outputPath, { force: true })
        throw error
    }
    finally {
        fs.rmSync(tempDir, { recursive: true, force: true })
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
