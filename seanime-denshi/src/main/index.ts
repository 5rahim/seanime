import {
    app,
    BrowserWindow,
    clipboard,
    dialog,
    ipcMain,
    Menu,
    nativeImage,
    net,
    powerSaveBlocker,
    protocol,
    screen,
    shell,
    Tray,
    webContents,
} from "electron"
import { autoUpdater } from "electron-updater"
import { spawn } from "node:child_process"
import fs from "node:fs"
import path from "node:path"
import { setupChromiumFlags } from "./chromium-flags"
import { DENSHI_SETTINGS_DEFAULTS, DenshiSettings, loadDenshiSettings, saveDenshiSettings } from "./denshi-settings"
import {
    allowedWebviewOrigins,
    getLocalServerPort,
    isAllowedLocalEmbedURL,
    isDesktopServerReachable,
    normalizeUpdateFeedURL,
    startLocalServer,
} from "./desktop-runtime"
import { log, setupLogging } from "./logging"
import { disposeMpvCore, initializeMpvCore, prepareMpvCore, registerMpvCoreIpc } from "./mpv-core"

let stripAnsi: ((str: string) => string) | undefined
import("strip-ansi").then(module => {
    stripAnsi = module.default
})

type ScopedLogger = {
    info: (...args: unknown[]) => void
    warn: (...args: unknown[]) => void
    error: (...args: unknown[]) => void
}

function createScopedLogger(scope: string): ScopedLogger {
    const prefix = `[${scope}]`
    return {
        info: (...args: unknown[]) => log.info(prefix, ...args),
        warn: (...args: unknown[]) => log.warn(prefix, ...args),
        error: (...args: unknown[]) => log.error(prefix, ...args),
    }
}

const logger = {
    app: createScopedLogger("App"),
    startup: createScopedLogger("Startup"),
    protocol: createScopedLogger("Protocol"),
    updater: createScopedLogger("Updater"),
    server: createScopedLogger("Server"),
    window: createScopedLogger("Window"),
    ipc: createScopedLogger("IPC"),
    settings: createScopedLogger("Settings"),
    power: createScopedLogger("Power"),
    cast: createScopedLogger("Cast"),
}

function formatError(value: unknown): string {
    if (value instanceof Error) {
        return value.stack || `${value.name}: ${value.message}`
    }

    if (typeof value === "string") {
        return value
    }

    try {
        return JSON.stringify(value) ?? String(value)
    }
    catch {
        return String(value)
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Settings
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const MAIN_WINDOW_DEFAULT_BOUNDS = {
    width: 800,
    height: 600,
}

const MIN_VISIBLE_WINDOW_EDGE = 120

let denshiSettings: DenshiSettings = { ...DENSHI_SETTINGS_DEFAULTS }
const mpvCoreSettings = {
    get: () => denshiSettings,
    updateLogging: (enabled: boolean) => {
        denshiSettings = { ...denshiSettings, mpvPrismLogging: enabled }
        saveDenshiSettings(denshiSettings)
    },
}
let shouldMaximizeMainWindow = false

// validates and returns safe window bounds based on the provided raw bounds and current display configurations
function getSafeMainWindowPlacement(rawBounds: Partial<Electron.Rectangle> | null | undefined) {
    const width = Number(rawBounds?.width)
    const height = Number(rawBounds?.height)
    if (!Number.isFinite(width) || width <= 0 || !Number.isFinite(height) || height <= 0) {
        return { bounds: { ...MAIN_WINDOW_DEFAULT_BOUNDS }, forceMaximize: false }
    }

    const x = Number(rawBounds?.x)
    const y = Number(rawBounds?.y)
    if (!app.isReady() || !Number.isFinite(x) || !Number.isFinite(y)) {
        return { bounds: { width, height }, forceMaximize: false }
    }

    const bounds = { x, y, width, height }
    if (screen.getAllDisplays().some(({ workArea }) => {
        const visibleWidth = Math.min(bounds.x + bounds.width, workArea.x + workArea.width) - Math.max(bounds.x, workArea.x)
        const visibleHeight = Math.min(bounds.y + bounds.height, workArea.y + workArea.height) - Math.max(bounds.y, workArea.y)
        return visibleWidth >= MIN_VISIBLE_WINDOW_EDGE && visibleHeight >= MIN_VISIBLE_WINDOW_EDGE
    })) {
        return { bounds, forceMaximize: false }
    }

    const { workArea } = screen.getPrimaryDisplay()
    const fallbackWidth = Math.min(width, workArea.width)
    const fallbackHeight = Math.min(height, workArea.height)

    return {
        bounds: {
            x: Math.round(workArea.x + (workArea.width - fallbackWidth) / 2),
            y: Math.round(workArea.y + (workArea.height - fallbackHeight) / 2),
            width: fallbackWidth,
            height: fallbackHeight,
        },
        forceMaximize: true,
    }
}

function saveMainWindowState() {
    if (!mainWindow || mainWindow.isDestroyed()) {
        return
    }

    const { x, y, width, height } = mainWindow.getNormalBounds()
    denshiSettings = {
        ...denshiSettings,
        windowBounds: { x, y, width, height },
        windowMaximized: mainWindow.isMaximized(),
    }
    saveDenshiSettings(denshiSettings)
}

function hideMainWindow() {
    if (!mainWindow || mainWindow.isDestroyed()) {
        return
    }

    // save window state before hiding so we can restore it later
    saveMainWindowState()
    mainWindow.hide()
    if (process.platform === "darwin") {
        app.dock?.hide()
    }
}

function showMainWindow() {
    if (!mainWindow || mainWindow.isDestroyed()) {
        return
    }

    let forceMaximize = shouldMaximizeMainWindow
    shouldMaximizeMainWindow = false

    if (app.isReady()) {
        const wasMaximized = mainWindow.isMaximized()
        const { bounds, forceMaximize: fallbackMaximize } = getSafeMainWindowPlacement(mainWindow.getNormalBounds())

        if (wasMaximized) {
            mainWindow.unmaximize()
        }

        const currentBounds = mainWindow.getBounds()
        const targetBounds = { ...currentBounds, ...bounds }
        if (currentBounds.x !== targetBounds.x
            || currentBounds.y !== targetBounds.y
            || currentBounds.width !== targetBounds.width
            || currentBounds.height !== targetBounds.height) {
            mainWindow.setBounds(targetBounds)
        }

        if (wasMaximized) {
            mainWindow.maximize()
        }

        forceMaximize = forceMaximize || fallbackMaximize
    }

    if (mainWindow.isMinimized()) {
        mainWindow.restore()
    }

    if ((denshiSettings.windowMaximized || forceMaximize) && !mainWindow.isMaximized()) {
        mainWindow.maximize()
    }

    mainWindow.show()
    mainWindow.focus()
    if (process.platform === "darwin") {
        app.dock?.show()
    }
}

setupLogging()
setupChromiumFlags()
const _development = process.env.NODE_ENV === "development"
const _isRsbuildFrontend = true
const DEFAULT_UPDATE_FEED_URL = "https://github.com/5rahim/seanime/releases/latest/download"

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom protocol for web content
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Custom protocol setup for serving web content in production
function setupCustomProtocol() {
    protocol.registerSchemesAsPrivileged([{
        scheme: "app", privileges: {
            standard: true,
            secure: true,
            allowServiceWorkers: true,
            supportFetchAPI: true,
            corsEnabled: true,
            stream: true,
        },
    },
    ])
}

// call before app.whenReady
setupCustomProtocol()

// Sets up the app protocol to serve the static files
function setupAppProtocol() {
    if (_development) return

    const webPath = path.join(app.getAppPath(), "web-denshi")

    if (!_isRsbuildFrontend) {
        protocol.handle("app", (request: Request) => {
            const requestUrl = new URL(request.url)
            let urlPath = requestUrl.pathname

            // next.js ssg: add .html to path
            if (!urlPath.endsWith(".html") && path.extname(urlPath) === "") {
                urlPath = urlPath + ".html"
            }

            // might not happen?
            if (urlPath === "/.html") {
                urlPath = "/index.html"
            }

            let filePath = path.join(webPath, urlPath)

            const resolvedPath = path.resolve(filePath)
            const resolvedWebPath = path.resolve(webPath)
            if (!resolvedPath.startsWith(resolvedWebPath)) {
                filePath = path.join(webPath, "index.html")
            }

            if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
                return net.fetch(`file://${filePath}`)
            }

            if (fs.existsSync(filePath) && fs.statSync(filePath).isDirectory()) {
                const indexPath = path.join(filePath, "index.html")
                if (fs.existsSync(indexPath)) {
                    return net.fetch(`file://${indexPath}`)
                }
            }

            // fallback to root index.html
            logger.protocol.warn("Route fallback; serving index.html", request.url)
            const fallbackPath = path.join(webPath, "index.html")
            return net.fetch(`file://${fallbackPath}`)
        })
    } else {
        protocol.handle("app", async (request: Request) => {
            const requestUrl = new URL(request.url)
            const urlPath = requestUrl.pathname
            let filePath = path.join(webPath, urlPath)

            if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
                const response = await net.fetch(`file://${filePath}`)
                const newHeaders = new Headers(response.headers)
                newHeaders.set("Cross-Origin-Opener-Policy", "same-origin")
                newHeaders.set("Cross-Origin-Embedder-Policy", "credentialless")
                return new Response(response.body, {
                    status: response.status,
                    statusText: response.statusText,
                    headers: newHeaders,
                })
            }

            const ext = path.extname(urlPath)
            if (!ext || ext === ".html") {
                const fallbackPath = path.join(webPath, "index.html")
                const response = await net.fetch(`file://${fallbackPath}`)
                const newHeaders = new Headers(response.headers)
                newHeaders.set("Cross-Origin-Opener-Policy", "same-origin")
                newHeaders.set("Cross-Origin-Embedder-Policy", "credentialless")
                return new Response(response.body, {
                    status: response.status,
                    statusText: response.statusText,
                    headers: newHeaders,
                })
            }

            logger.protocol.warn("Asset not found", urlPath)
            return new Response("Not Found", { status: 404 })
        })
    }
}

/////////////////
// Cast
/////////////////

const __CAST_ENABLED__ = false
const CastSender = __CAST_ENABLED__ ? require(path.join(app.getAppPath(), "src/cast/sender.js")).CastSender : null
let castSender: any = null

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Startup logs
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function logStartupEvent(stage: string, detail?: unknown) {
    if (detail === undefined || detail === "") {
        logger.startup.info(stage)
        return
    }

    logger.startup.info(`${stage}:`, detail)
}

// Global error handlers to catch unhandled exceptions
process.on("uncaughtException", (error: Error) => {
    logger.app.error("Uncaught exception", formatError(error))

    if (app.isReady()) {
        dialog.showErrorBox("An error occurred", `Uncaught Exception: ${error.message}\n\nCheck the logs for more details.`)
    }
})

process.on("unhandledRejection", (reason: unknown) => {
    logger.app.error("Unhandled rejection", formatError(reason))
})

// Dumps important environment information for debugging
function logEnvironmentInfo() {
    logStartupEvent("NODE_ENV", process.env.NODE_ENV || "not set")
    logStartupEvent("Platform", process.platform)
    logStartupEvent("Architecture", process.arch)
    logStartupEvent("Node version", process.version)
    logStartupEvent("Electron version", process.versions.electron)
    logStartupEvent("App path", app.getAppPath())
    logStartupEvent("Dir name", __dirname)
    logStartupEvent("User data path", app.getPath("userData"))
    logStartupEvent("Executable path", app.getPath("exe"))

    if (process.resourcesPath) {
        logStartupEvent("Resources path", process.resourcesPath)
        try {
            // const resourceFiles = fs.readdirSync(process.resourcesPath);
            // logStartupEvent('Resources directory contents', JSON.stringify(resourceFiles));

            // Check if binaries directory exists
            const binariesDir = path.join(process.resourcesPath, "binaries")
            if (fs.existsSync(binariesDir)) {
                const binariesFiles = fs.readdirSync(binariesDir)
                logStartupEvent("Binaries directory contents", JSON.stringify(binariesFiles))
            } else {
                logger.startup.warn("Binaries directory not found", binariesDir)
            }
        }
        catch (err: any) {
            logger.startup.error("Failed to read resources directory", formatError(err))
        }
    }

    // Check app directory structure
    try {
        const appPath = app.getAppPath()
        // logStartupEvent('App directory contents', JSON.stringify(fs.readdirSync(appPath)));

        const webPath = path.join(appPath, "web-denshi")
        if (!fs.existsSync(webPath)) {
            logger.startup.warn("web-denshi directory not found", webPath)
        }
    }
    catch (err: any) {
        logger.startup.error("Failed to inspect app directory", formatError(err))
    }
}

let mainWindow: Electron.BrowserWindow | null = null
let splashScreen: Electron.BrowserWindow | null = null
let crashScreen: Electron.BrowserWindow | null = null
let tray: Electron.Tray | null = null
let serverProcess: import("node:child_process").ChildProcess | null = null
let isShutdown = false
let serverStarted = false
let mainWindowStartupReady = false
let updateDownloaded = false
let serverRestartPromise: Promise<void> | null = null

app.on("child-process-gone", (event: Electron.Event, details: any) => {
    logger.app.warn("Child process gone", details)
})

// Setup autoUpdater events with improved error handling
autoUpdater.on("checking-for-update", () => {
    logger.updater.info("Checking for update")
})

autoUpdater.on("update-available", (info: any) => {
    logger.updater.info("Update available", info)
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-available", {
            version: info.version, releaseDate: info.releaseDate, files: info.files,
        })
    }
})

autoUpdater.on("update-not-available", (info: any) => {
    logger.updater.info("Update not available", info)
})

autoUpdater.on("download-progress", (progressObj: any) => {
    logger.updater.info("Download progress", `${Number(progressObj.percent).toFixed(1)}%`)
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("download-progress", {
            percent: progressObj.percent, bytesPerSecond: progressObj.bytesPerSecond, transferred: progressObj.transferred, total: progressObj.total,
        })
    }
})

autoUpdater.on("update-downloaded", (info: any) => {
    logger.updater.info("Update downloaded", info)
    updateDownloaded = true
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-downloaded", {
            version: info.version, releaseDate: info.releaseDate, files: info.files,
        })
    }
})

autoUpdater.on("error", (err: Error) => {
    logger.updater.error("Auto-updater error", formatError(err))
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-error", {
            code: (err as any).code || "unknown", message: err.message, stack: err.stack,
        })
    }
})

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Single instance
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const gotTheLock = _development ? true : app.requestSingleInstanceLock({ development: _development })

/**
 * Force single instance
 */
if (!gotTheLock) {
    if (!_development) {
        app.quit()
    }
} else {
    app.on("second-instance", (event: Electron.Event, commandLine: string[], workingDirectory: string, additionalData: any) => {
        if (additionalData && additionalData.development) return
        if (!serverStarted) return
        // tried to run a second instance, focus the window.
        if (mainWindow && !mainWindow.isDestroyed()) {
            if (mainWindow.isMinimized()) mainWindow.restore()
            if (!mainWindow.isVisible()) {
                showMainWindow()
            }
            mainWindow.focus()
        }
    })
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Tray
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function createTray() {
    const iconName = process.platform === "darwin" ? "18x18.png" : "icon.png"

    const iconPath = path.join(app.getAppPath(), "assets", iconName)
    const icon = nativeImage.createFromPath(iconPath)
    tray = new Tray(icon)

    const contextMenu = Menu.buildFromTemplate([{
        id: "toggle_visibility", label: "Toggle Visibility", click: () => {
            if (!serverStarted) return
            if (!mainWindow || mainWindow.isDestroyed()) return
            if (mainWindow.isVisible()) {
                hideMainWindow()
            } else {
                showMainWindow()
            }
        },
    }, ...(process.platform === "darwin" ? [{
        id: "accessory_mode", label: "Remove from Dock", click: () => {
            app.dock?.hide()
        },
    },
    ] : []), {
        id: "quit", label: "Quit Seanime", click: () => {
            cleanupAndExit()
        },
    },
    ])

    tray.setToolTip("Seanime")

    if (process.platform !== "darwin") {
        tray.setContextMenu(contextMenu)
    }

    tray.on("click", () => {
        if (!serverStarted) return
        if (!mainWindow || mainWindow.isDestroyed()) return
        if (mainWindow.isVisible()) {
            hideMainWindow()
        } else {
            showMainWindow()
        }
    })

    if (process.platform === "darwin") {
        tray.on("right-click", () => {
            tray?.popUpContextMenu(contextMenu)
        })
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Seanime server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

async function launchSeanimeServer(isRestart: boolean): Promise<void> {
    return new Promise<void>((resolve, reject) => {
        let startupResolved = false
        let startupPollInterval: NodeJS.Timeout | null = null
        let waitingForRenderer = false

        function clearStartupProbe() {
            if (startupPollInterval) {
                clearInterval(startupPollInterval)
                startupPollInterval = null
            }
        }

        function checkFinalizeStartup(source: string) {
            if (startupResolved) {
                return
            }

            if (!mainWindowStartupReady) {
                if (!waitingForRenderer) {
                    waitingForRenderer = true
                    logStartupEvent("WAITING FOR RENDERER", source)
                }
                return
            }

            finalizeStartup(source)
        }

        function finalizeStartup(source: string) {
            if (startupResolved) {
                return
            }

            startupResolved = true
            clearStartupProbe()

            logger.server.info("Startup finalized", { source, isRestart })
            serverStarted = true
            setTimeout(() => {
                if (splashScreen && !splashScreen.isDestroyed()) {
                    splashScreen.close()
                    splashScreen = null
                }
                setTimeout(() => {
                    if (mainWindow && !mainWindow.isDestroyed()) {
                        if (denshiSettings.openInBackground) {
                            // Don't maximize or show
                            logger.window.info("Opened in background")
                        } else {
                            showMainWindow()
                        }
                    }
                }, 1000)
                resolve()
            }, 2000)
        }

        async function probeServerStartup() {
            if (startupResolved || !serverProcess || serverProcess.killed) {
                return
            }

            if (await isDesktopServerReachable()) {
                checkFinalizeStartup("HTTP status probe")
            }
        }

        // TEST ONLY: Check for -no-binary flag
        if (process.argv.includes("-no-binary")) {
            logger.server.warn("Skipping launch because -no-binary is set")
            serverStarted = true // Assume server is "started" for UI flow
            // Resolve immediately to bypass server spawning
            if (splashScreen && !splashScreen.isDestroyed()) {
                splashScreen.close()
                splashScreen = null
            }
            if (mainWindow && !mainWindow.isDestroyed() && !denshiSettings.openInBackground) {
                showMainWindow()
            }
            return resolve()
        }

        // Determine the correct binary to use based on platform and architecture
        let binaryName = ""
        if (process.platform === "win32") {
            binaryName = "seanime-server-windows.exe"
        } else if (process.platform === "darwin") {
            const arch = process.arch === "arm64" ? "arm64" : "amd64"
            binaryName = `seanime-server-darwin-${arch}`
        } else if (process.platform === "linux") {
            const arch = process.arch === "arm64" ? "arm64" : "amd64"
            binaryName = `seanime-server-linux-${arch}`
        }

        let binaryPath: string

        if (_development) {
            // In development, look for binaries in the project directory
            binaryPath = path.join(app.getAppPath(), "binaries", binaryName)
        } else {
            // In production, use the resources path
            binaryPath = path.join(process.resourcesPath, "binaries", binaryName)
        }

        logStartupEvent("Using binary", `${binaryPath} (${process.arch})`)
        logStartupEvent("Resources path", process.resourcesPath)

        // Check if binary exists and is executable
        if (!fs.existsSync(binaryPath)) {
            const error = new Error(`Server binary not found at ${binaryPath}`)
            logger.server.error("Server binary not found", binaryPath)
            return reject(error)
        }

        // Make binary executable (for macOS/Linux)
        if (process.platform !== "win32") {
            try {
                fs.chmodSync(binaryPath, "755")
            }
            catch (error) {
                logger.server.error("Failed to make binary executable", formatError(error))
            }
        }

        // Arguments
        const args: string[] = []

        // Development mode
        if (_development) {
            args.push("-port", "43000")
            if (process.env.TEST_DATADIR) {
                logger.server.info("Using TEST_DATADIR", process.env.TEST_DATADIR)
                args.push("-datadir", process.env.TEST_DATADIR)
            } else {
                const devDataDir = path.join(app.getPath("appData"), "Seanime-dev")
                args.push("-datadir", devDataDir)
            }
        }

        args.push("-desktop-sidecar", "true")

        logger.server.info("Spawning process", { args, binaryPath, isRestart })

        // Spawn the process
        let proc: import("node:child_process").ChildProcess
        try {
            proc = spawn(binaryPath, args)
            serverProcess = proc
        }
        catch (spawnError) {
            logger.server.error("Failed to spawn process", formatError(spawnError))
            return reject(spawnError)
        }

        startupPollInterval = setInterval(() => {
            void probeServerStartup()
        }, 500)
        void probeServerStartup()

        if (proc.stdout) {
            proc.stdout.on("data", (data: any) => {
                const dataStr = data.toString()
                const lineStr = stripAnsi ? stripAnsi(dataStr) : dataStr

                // Check if the frontend is connected
                if (!serverStarted && lineStr.includes("Client connected")) {
                    checkFinalizeStartup("websocket client connection")
                }
            })
        }

        if (proc.stderr) {
            proc.stderr.on("data", (data: Buffer | string) => {
                const output = data.toString().trim()
                if (output) {
                    logger.server.error("stderr", stripAnsi ? stripAnsi(output) : output)
                }
            })
        }

        proc.on("close", (code: number | null) => {
            clearStartupProbe()
            logger.server.info("Process exited", { code })

            // If the server didn't start properly and we're not in the process of shutting down
            if (!startupResolved && !isShutdown) {
                logger.server.error("Process exited before startup completed", { code })
                reject(new Error(`Server process exited prematurely with code ${code} before starting.`))

                // close splash screen and main window
                if (splashScreen && !splashScreen.isDestroyed()) {
                    splashScreen.close()
                    splashScreen = null
                }

                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.close()
                    mainWindow.destroy()
                    mainWindow = null
                }

                // show crash screen
                if (crashScreen && !crashScreen.isDestroyed()) {
                    crashScreen.show()
                    crashScreen.webContents.send("crash", `Seanime server process terminated with status: ${code}. Closing in 10 seconds.`)

                    setTimeout(() => {
                        app.exit(1)
                    }, 10000)
                }
            }
        })

        // Handle spawn errors
        proc.on("error", (err: Error) => {
            clearStartupProbe()
            logger.server.error("Process error event", formatError(err))
            reject(err)
        })
    })
}

async function restartSeanimeServer() {
    if (serverRestartPromise) {
        logger.server.info("Restart already in progress; reusing existing request")
        return serverRestartPromise
    }

    serverRestartPromise = (async () => {
        if (await isDesktopServerReachable()) {
            logger.server.info("Restart skipped; server is already reachable")
            return
        }

        const currentServerProcess = serverProcess

        if (currentServerProcess && !currentServerProcess.killed) {
            logger.server.info("Stopping existing process before relaunch")

            await new Promise<void>((resolve) => {
                let settled = false

                function finish() {
                    if (settled) {
                        return
                    }

                    settled = true
                    currentServerProcess?.removeListener("close", finish)
                    currentServerProcess?.removeListener("error", finish)

                    if (serverProcess === currentServerProcess) {
                        serverProcess = null
                    }

                    resolve()
                }

                currentServerProcess.once("close", finish)
                currentServerProcess.once("error", finish)

                try {
                    currentServerProcess.kill()
                }
                catch (error) {
                    logger.server.error("Failed to stop process during restart", formatError(error))
                    finish()
                    return
                }

                setTimeout(() => {
                    logger.server.warn("Timed out waiting for process to exit during restart")
                    finish()
                }, 3000)
            })
        } else {
            serverProcess = null
        }

        await launchSeanimeServer(true)
    })().finally(() => {
        serverRestartPromise = null
    })

    return serverRestartPromise
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Main window
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function showEditableContextMenu(webContents: Electron.WebContents, params: Electron.ContextMenuParams) {
    if (!params.isEditable) return

    const template = [
        { role: "undo" as const, enabled: params.editFlags.canUndo },
        { role: "redo" as const, enabled: params.editFlags.canRedo },
        { type: "separator" as const },
        { role: "cut" as const, enabled: params.editFlags.canCut },
        { role: "copy" as const, enabled: params.editFlags.canCopy },
        { role: "paste" as const, enabled: params.editFlags.canPaste },
        { role: "selectAll" as const, enabled: params.editFlags.canSelectAll },
    ]

    Menu.buildFromTemplate(template).popup({
        window: BrowserWindow.fromWebContents(webContents) ?? undefined,
    })
}

function createMainWindow() {
    logStartupEvent("Creating main window")
    mainWindowStartupReady = false
    const savedPlacement = getSafeMainWindowPlacement(denshiSettings.windowBounds)
    shouldMaximizeMainWindow = savedPlacement.forceMaximize

    const windowOptions: Electron.BrowserWindowConstructorOptions = {
        ...savedPlacement.bounds, show: false,
        backgroundColor: "#111111",
        acceptFirstMouse: false,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            sandbox: false,
            preload: path.join(app.getAppPath(), "src/preload.mjs"),
            webSecurity: true,
            allowRunningInsecureContent: true,
            enableBlinkFeatures: "FontAccess, AudioVideoTracks",
            backgroundThrottling: false,
            webviewTag: true,
        },
    }

    // contextMenu({
    //     showInspectElement: true
    // });

    // Set title bar style based on platform
    if (process.platform === "darwin") {
        windowOptions.titleBarStyle = "hiddenInset"
    }

    if (process.platform === "win32") {
        windowOptions.titleBarStyle = "hidden"
    }

    const win = new BrowserWindow(windowOptions)
    mainWindow = win

    win.webContents.on("context-menu", (event: Electron.Event, params: Electron.ContextMenuParams) => {
        if (!params.isEditable) return

        event.preventDefault()
        showEditableContextMenu(win.webContents, params)
    })

    // Hide the title bar on Windows
    if (process.platform === "win32" || process.platform === "linux") {
        win.setMenuBarVisibility(false)
    }

    win.webContents.on("render-process-gone", (event: Electron.Event, details: Electron.RenderProcessGoneDetails) => {
        logger.window.error("Renderer process gone", details)
        if (crashScreen && !crashScreen.isDestroyed()) {
            crashScreen.show()
            crashScreen.webContents.send(
                "crash",
                `The desktop window stopped unexpectedly (${details.reason || "unknown reason"}${typeof details.exitCode === "number"
                    ? `, exit code ${details.exitCode}`
                    : ""}). The background server may still be running.`,
                { isRendererCrash: true },
            )
        }
    })

    win.webContents.on("will-attach-webview", (event: Electron.Event, webPreferences: Electron.WebPreferences, params: any) => {
        let isAllowed = false
        try {
            const parsed = new URL(params.src)
            if (isAllowedLocalEmbedURL(params.src) || allowedWebviewOrigins.has(parsed.origin)) {
                isAllowed = true
            }
        }
        catch (err) {
        }

        if (!isAllowed) {
            logger.window.warn("Blocked unexpected webview source", params.src)
            event.preventDefault()
            return
        }

        delete webPreferences.preload
        delete (webPreferences as any).preloadURL
        delete params.preload

        webPreferences.nodeIntegration = false
        webPreferences.contextIsolation = true
        webPreferences.sandbox = true
        webPreferences.webSecurity = true
        webPreferences.allowRunningInsecureContent = true

        params.allowpopups = params.allowpopups || false
    })


    win.webContents.setWindowOpenHandler(({ frameName, url }: Electron.HandlerDetails) => {
        // Allow DocumentPictureInPicture window requests which use about:blank
        if (url === "about:blank") {
            return { action: "allow" }
        }
        // Open external links in the default browser
        if (url.startsWith("http://") || url.startsWith("https://")) {
            shell.openExternal(url)
            return { action: "deny" }
        }
        // // Allow other URLs to open in the app
        // return {action: 'allow'};

        // For internal app:// (or file://) links, do not spawn a new renderer,
        // navigate the main window (or the opener) so it remains a single renderer.
        try {
            const openerId = Number.parseInt(frameName, 10)
            const opener = (!Number.isNaN(openerId) ? webContents.fromId(openerId) : null) || win.webContents
            // load in mainWindow instead of spawning new window
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.loadURL(url)
            }
        }
        catch (e) {
            logger.window.warn("Failed to route internal window URL", formatError(e))
        }

        return { action: "deny" }
    })

    // Load the web content
    if (_development) {
        // In development, load from the dev server
        logStartupEvent("Loading from dev server", "http://127.0.0.1:43210")
        win.loadURL("http://127.0.0.1:43210")
        // win.loadURL('chrome://gpu');
    } else {
        logStartupEvent("Loading production build with custom protocol")
        win.loadURL("app://-")
    }

    // Development tools
    // if (_development) {
    //     win.webContents.openDevTools();
    // }

    win.on("close", (event: Electron.Event) => {
        if (!isShutdown) {
            if (denshiSettings.minimizeToTray) {
                event.preventDefault()
                hideMainWindow()
            } else {
                // Close the app completely
                cleanupAndExit()
            }
        }
    })
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Splashscreen
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function createSplashScreen() {
    logStartupEvent("Creating splash screen")
    splashScreen = new BrowserWindow({
        width: 800, height: 600, frame: false, resizable: false, show: false, backgroundColor: "#070707", webPreferences: {
            nodeIntegration: false, contextIsolation: true, sandbox: true,
        },
    })

    function showSplashScreen() {
        if (denshiSettings.openInBackground || !splashScreen || splashScreen.isDestroyed() || splashScreen.isVisible()) {
            return
        }

        splashScreen.show()
    }

    splashScreen.once("ready-to-show", showSplashScreen)
    setTimeout(showSplashScreen, 500)

    logStartupEvent("Loading splash screen")
    splashScreen.loadFile(path.join(app.getAppPath(), "src/splash.html"))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Crash screen
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function createCrashScreen() {
    crashScreen = new BrowserWindow({
        width: 800, height: 600, frame: false, resizable: false, show: false, webPreferences: {
            nodeIntegration: false, contextIsolation: true, sandbox: true, preload: path.join(app.getAppPath(), "src/preload.js"),
        },
    })

    // Load the web content
    if (_development) {
        // In development, load from the dev server
        crashScreen.loadURL("http://127.0.0.1:43210/splashscreen/crash")
    } else {
        crashScreen.loadURL("app://-/splashscreen/crash")
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Exit
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function cleanupAndExit() {
    if (isShutdown) {
        return
    }

    isShutdown = true
    logger.app.info("Shutdown started")

    saveMainWindowState()
    disposeMpvCore()

    // Clean up cast
    if (__CAST_ENABLED__ && castSender) {
        castSender.destroy()
        castSender = null
    }

    // Kill server process first
    if (serverProcess) {
        logger.server.info("Stopping process for shutdown")
        try {
            serverProcess.kill()
            serverProcess = null
        }
        catch (err) {
            logger.server.error("Failed to stop process during shutdown", formatError(err))
        }
    }

    // Exit the app after a short delay to allow cleanup
    setTimeout(() => {
        app.exit(0)
    }, 500)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Initialization
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// returns true if github is ok OR url is unreachable
// returns false if github is down and fallback should be used
async function fetchGithubStatus() {
    try {
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 5000)

        const response = await net.fetch("https://seanime.app/api/github-status", {
            signal: controller.signal,
        })
        clearTimeout(timeoutId)

        if (!response.ok) {
            return { ok: true, fallback: "" }
        }

        const data = await response.json()

        // url is reachable, status is "down"
        if (data.status === "down") {
            logger.updater.warn("Using fallback update channel", { channel: data.fallback, reason: data.description })
            return { ok: false, fallback: data.fallback || "seanime" }
        }

        return { ok: true, fallback: "" }
    }
    catch (err) {
        return { ok: true, fallback: "" }
    }
}

// Initialize the app
app.whenReady().then(async () => {
    logStartupEvent("App ready")

    // Load denshi settings early so environment variables are registered prior to DLL import
    denshiSettings = loadDenshiSettings()

    prepareMpvCore(mpvCoreSettings)

    try {
        await initializeMpvCore()
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error)
        logger.app.error("Failed to initialize MpvCore", formatError(error))
        dialog.showErrorBox("Seanime Denshi could not start MpvCore", message)
        app.quit()
        return
    }
    if (_development) {
        denshiSettings.openInBackground = false
    }
    // Disregard openInBackground on Linux
    if (process.platform === "linux") {
        denshiSettings.openInBackground = false
    }
    logger.settings.info("Loaded", denshiSettings)

    let currentUpdateChannel = denshiSettings.updateChannel
    const { ok, fallback } = await fetchGithubStatus()
    // if github is down, use fallback channel
    if (!ok) {
        currentUpdateChannel = fallback
    }

    const updateConfig = {
        provider: "generic" as const,
        url: DEFAULT_UPDATE_FEED_URL,
        channel: "latest",
        allowPrerelease: false,
        verifyUpdateCodeSignature: false,
    }

    if (currentUpdateChannel === "seanime_nightly") {
        updateConfig.url = "https://seanime.app/api/updates/nightly/"
        updateConfig.allowPrerelease = true
    } else if (currentUpdateChannel === "seanime") {
        updateConfig.url = "https://seanime.app/api/updates/stable/"
        updateConfig.allowPrerelease = false
    }

    updateConfig.url = normalizeUpdateFeedURL(updateConfig.url, DEFAULT_UPDATE_FEED_URL)

    if (process.env.UPDATES_URL) {
        updateConfig.url = normalizeUpdateFeedURL(process.env.UPDATES_URL, updateConfig.url)
    }

    autoUpdater.setFeedURL(updateConfig as any)
    autoUpdater.autoDownload = true
    autoUpdater.autoInstallOnAppQuit = true
    autoUpdater.disableWebInstaller = true

    if (!_development && (process.platform === "darwin" || process.platform === "win32")) {
        app.setLoginItemSettings({
            openAtLogin: denshiSettings.openAtLaunch,
        })
    }

    // Log environment information
    logEnvironmentInfo()

    ipcMain.on("startup:renderer-ready", (event: Electron.IpcMainEvent) => {
        if (!mainWindow || event.sender !== mainWindow.webContents || mainWindowStartupReady) {
            return
        }

        mainWindowStartupReady = true
        logStartupEvent("RENDERER READY", "main window committed first frame")
    })

    // Setup IPC handlers for update functions
    ipcMain.handle("check-for-updates", async () => {
        try {
            logger.updater.info("Manual update check requested")
            const result = await autoUpdater.checkForUpdates()
            return {
                updateAvailable: !!result?.updateInfo,
                updateInfo: result?.updateInfo,
                updateDownloaded: updateDownloaded,
            }
        }
        catch (error) {
            logger.updater.error("Manual update check failed", formatError(error))
            throw error
        }
    })

    ipcMain.handle("install-update", async () => {
        try {
            if (!updateDownloaded) {
                logger.updater.info("Install requested before download completed; starting download")
                // Trigger download if not already downloaded
                await autoUpdater.checkForUpdatesAndNotify()
                throw new Error("Update download initiated. Please wait for download to complete.")
            }
            logger.updater.info("Installing update")
            autoUpdater.quitAndInstall(false, true)
            return true
        }
        catch (error) {
            logger.updater.error("Update installation failed", formatError(error))
            throw error
        }
    })

    ipcMain.handle("kill-server", async () => {
        if (serverProcess) {
            logger.server.info("Stopping process before update")
            serverProcess.kill()
            return true
        }
        return false
    })

    registerMpvCoreIpc(mpvCoreSettings)
    registerIpcHandlers()

    setupAppProtocol()
    startLocalServer()

    // Create windows
    createMainWindow()
    createSplashScreen()
    createCrashScreen()

    // Create tray
    createTray()

    // Launch server
    try {
        logStartupEvent("Attempting to launch server")
        await launchSeanimeServer(false)
        logStartupEvent("Server launched successfully")
        // Check for updates only after server launch and main window setup is successful
        autoUpdater.checkForUpdatesAndNotify()
    }
    catch (error) {
        logger.startup.error("Server launch failed", formatError(error))
        if (splashScreen && !splashScreen.isDestroyed()) {
            splashScreen.close()
            splashScreen = null
        }

        if (crashScreen && !crashScreen.isDestroyed()) {
            crashScreen.show()
            crashScreen.webContents.send("crash", `The server failed to start: ${error}. Closing in 10 seconds.`)

            setTimeout(() => {
                logger.app.error("Exiting because server startup failed")
                app.exit(1)
            }, 10000)
        }
    }

    function registerIpcHandlers() {
        // Register Window Control IPC handlers
        ipcMain.on("window:minimize", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.minimize()
            }
        })

        ipcMain.on("window:maximize", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.maximize()
            }
        })

        ipcMain.on("window:close", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.close()
            }
        })

        ipcMain.on("window:toggleMaximize", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                if (mainWindow.isMaximized()) {
                    mainWindow.unmaximize()
                } else {
                    mainWindow.maximize()
                }
            }
        })

        ipcMain.on("window:setFullscreen", (_: Electron.IpcMainEvent, fullscreen: boolean) => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.setFullScreen(fullscreen)
            }
        })

        ipcMain.on("window:hide", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                hideMainWindow()
            }
        })

        ipcMain.on("window:show", () => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                showMainWindow()
            }
        })

        ipcMain.handle("window:getCurrentWindow", () => {
            if (!mainWindow) return undefined
            const win = BrowserWindow.fromWebContents(mainWindow.webContents)
            return win?.id
        })

        ipcMain.handle("window:isMainWindow", (event: Electron.IpcMainInvokeEvent) => {
            const win = BrowserWindow.fromWebContents(event.sender)
            return win === mainWindow
        })

        // Window state query handlers
        ipcMain.handle("window:isMaximized", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.isMaximized() : false
        })

        ipcMain.handle("window:isMinimizable", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.minimizable : false
        })

        ipcMain.handle("window:isMaximizable", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.maximizable : false
        })

        ipcMain.handle("window:isClosable", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.closable : false
        })

        ipcMain.handle("window:isFullscreen", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.isFullScreen() : false
        })

        ipcMain.handle("window:isVisible", () => {
            return mainWindow && !mainWindow.isDestroyed() ? mainWindow.isVisible() : false
        })

        // Clipboard handler
        ipcMain.handle("clipboard:writeText", (_: Electron.IpcMainInvokeEvent, text: string) => {
            if (text) {
                clipboard.writeText(text)
                return true
            }
            return false
        })

        // Register server IPC handlers
        ipcMain.on("restart-server", () => {
            logger.ipc.info("restart-server")
            restartSeanimeServer().catch(error => {
                logger.server.error("Restart failed", formatError(error))
            })
        })

        ipcMain.on("kill-server", () => {
            logger.ipc.info("kill-server")
            if (serverProcess) {
                logger.server.info("Stopping process by IPC request")
                serverProcess.kill()
            }
        })

        // Watch for window events to notify renderer
        if (mainWindow) {
            mainWindow.on("minimize", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:minimized")
                }
            })

            mainWindow.on("hide", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:hidden")
                }
            })

            mainWindow.on("maximize", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:maximized")
                }
            })

            mainWindow.on("unmaximize", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:unmaximized")
                }
            })

            mainWindow.on("enter-full-screen", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:fullscreen", true)
                }
            })

            mainWindow.on("leave-full-screen", () => {
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("window:fullscreen", false)
                }
            })
        }

        // macOS specific events
        ipcMain.on("macos-activation-policy-accessory", () => {
            logger.ipc.info("macos-activation-policy-accessory")
            if (process.platform === "darwin" && mainWindow) {
                app.dock?.hide()
                mainWindow.show()
                mainWindow.setFullScreen(true)

                setTimeout(() => {
                    if (mainWindow) {
                        mainWindow.focus()
                        mainWindow.webContents.send("macos-activation-policy-accessory-done", "")
                    }
                }, 150)
            }
        })

        ipcMain.on("macos-activation-policy-regular", () => {
            logger.ipc.info("macos-activation-policy-regular")
            if (process.platform === "darwin") {
                app.dock?.show()
            }
        })

        // Quit app handler
        ipcMain.on("quit-app", () => {
            logger.ipc.info("quit-app")
            cleanupAndExit()
        })

        // Restart app handler
        ipcMain.on("restart-app", () => {
            logger.ipc.info("restart-app")
            if (crashScreen && !crashScreen.isDestroyed()) {
                crashScreen.hide()
            }
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.webContents.reload()
                mainWindow.show()
            } else {
                createMainWindow()
            }
        })
        ipcMain.handle("get-local-server-port", () => getLocalServerPort())

        ipcMain.handle("denshi:allowWebviewOrigin", (_: Electron.IpcMainInvokeEvent, origin: string) => {
            try {
                const parsed = new URL(origin)
                allowedWebviewOrigins.add(parsed.origin)
                logger.settings.info("Allowed webview origin", parsed.origin)
                return true
            }
            catch (err) {
                const message = err instanceof Error ? err.message : String(err)
                logger.settings.error("Failed to allow webview origin", { origin, message })
                return false
            }
        })

        ipcMain.handle("denshi:getSettings", () => {
            return { ...denshiSettings }
        })

        ipcMain.handle("denshi:setSettings", (_: Electron.IpcMainInvokeEvent, newSettings: Partial<DenshiSettings>) => {
            denshiSettings = { ...DENSHI_SETTINGS_DEFAULTS, ...denshiSettings, ...newSettings }
            saveDenshiSettings(denshiSettings)
            logger.settings.info("Updated", denshiSettings)

            // Apply openAtLaunch immediately (only supported on macOS and Windows)
            if (!_development && (process.platform === "darwin" || process.platform === "win32")) {
                app.setLoginItemSettings({
                    openAtLogin: denshiSettings.openAtLaunch,
                })
            }

            return { ...denshiSettings }
        })

        // Power save blocker
        ipcMain.handle("power-save-blocker:start", () => {
            try {
                const id = powerSaveBlocker.start("prevent-display-sleep")
                logger.power.info("Display sleep blocker started", { id })
                return id
            }
            catch (e) {
                logger.power.error("Failed to start display sleep blocker", formatError(e))
                throw e
            }
        })

        ipcMain.handle("power-save-blocker:stop", (_: Electron.IpcMainInvokeEvent, id: number) => {
            try {
                if (typeof id === "number" && powerSaveBlocker.isStarted(id)) {
                    powerSaveBlocker.stop(id)
                    logger.power.info("Display sleep blocker stopped", { id })
                }
            }
            catch (e) {
                logger.power.error("Failed to stop display sleep blocker", formatError(e))
            }
        })

        // Chromecast

        if (__CAST_ENABLED__) {

            function ensureCastSender() {
                if (!castSender) {
                    castSender = new CastSender()

                    // Forward events to the renderer
                    castSender.on("deviceFound", (device: any) => {
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            mainWindow.webContents.send("cast:deviceFound", device)
                        }
                    })
                    castSender.on("sessionUpdate", (state: any) => {
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            mainWindow.webContents.send("cast:sessionUpdate", state)
                        }
                    })
                    castSender.on("mediaStatus", (status: any) => {
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            mainWindow.webContents.send("cast:mediaStatus", status)
                        }
                    })
                    castSender.on("receiverReady", () => {
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            mainWindow.webContents.send("cast:receiverReady")
                        }
                    })
                    castSender.on("error", (err: Error) => {
                        logger.cast.error("Sender error", formatError(err))
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            mainWindow.webContents.send("cast:error", err)
                        }
                    })
                }
                return castSender
            }

            ipcMain.handle("cast:discover", async () => {
                const sender = ensureCastSender()
                sender.startDiscovery()
            })

            ipcMain.handle("cast:stopDiscovery", async () => {
                if (castSender) castSender.stopDiscovery()
            })

            ipcMain.handle("cast:getDevices", async () => {
                if (!castSender) return []
                return castSender.getDevices()
            })

            ipcMain.handle("cast:connect", async (_: Electron.IpcMainInvokeEvent, deviceId: string) => {
                const sender = ensureCastSender()
                return await sender.connect(deviceId)
            })

            ipcMain.handle("cast:disconnect", async () => {
                if (castSender) castSender.disconnect()
            })

            ipcMain.handle("cast:getStatus", async () => {
                if (!castSender) return { connected: false, device: null, sessionId: null, mediaStatus: null }
                return castSender.getStatus()
            })

            ipcMain.handle("cast:loadMedia", async (_: Electron.IpcMainInvokeEvent, opts: any) => {
                if (!castSender) throw new Error("Cast sender not initialized")
                return castSender.loadMedia(opts)
            })

            ipcMain.handle("cast:play", async () => {
                if (castSender) castSender.play()
            })

            ipcMain.handle("cast:pause", async () => {
                if (castSender) castSender.pause()
            })

            ipcMain.handle("cast:seek", async (_: Electron.IpcMainInvokeEvent, time: number) => {
                if (castSender) castSender.seek(time)
            })

            ipcMain.handle("cast:stop", async () => {
                if (castSender) castSender.stop()
            })

            ipcMain.handle("cast:setVolume", async (_: Electron.IpcMainInvokeEvent, level: number) => {
                if (castSender) castSender.setVolume(level)
            })

            ipcMain.handle("cast:setMuted", async (_: Electron.IpcMainInvokeEvent, muted: boolean) => {
                if (castSender) castSender.setMuted(muted)
            })

            ipcMain.handle("cast:sendSubtitleEvents", async (_: Electron.IpcMainInvokeEvent, events: any) => {
                if (castSender) castSender.sendSubtitleEvents(events)
            })

            ipcMain.handle("cast:sendSubtitleTracks", async (_: Electron.IpcMainInvokeEvent, tracks: any) => {
                if (castSender) castSender.sendSubtitleTracks(tracks)
            })

            ipcMain.handle("cast:switchSubtitleTrack", async (_: Electron.IpcMainInvokeEvent, trackNumber: number) => {
                if (castSender) castSender.switchSubtitleTrack(trackNumber)
            })

            ipcMain.handle("cast:sendFonts", async (_: Electron.IpcMainInvokeEvent, fontUrls: string[], serverPort: number) => {
                if (castSender) castSender.sendFonts(fontUrls, serverPort)
            })

            ipcMain.handle("cast:sendSubtitleHeader", async (_: Electron.IpcMainInvokeEvent, header: string) => {
                if (castSender) castSender.sendSubtitleHeader(header)
            })

            ipcMain.handle("cast:disableSubtitles", async () => {
                if (castSender) castSender.disableSubtitles()
            })

            ipcMain.handle("cast:getLanIP", async () => {
                const sender = ensureCastSender()
                return sender.getLanIP()
            })

        }
    }

    app.on("window-all-closed", () => {
        if (process.platform !== "darwin") {
            app.quit()
        }
    })

    // app.on('activate', () => {
    //     if (BrowserWindow.getAllWindows().length === 0) {
    //         createMainWindow();
    //     }
    // });

    app.on("before-quit", () => {
        logger.app.info("before-quit")
        cleanupAndExit()
    })
})
