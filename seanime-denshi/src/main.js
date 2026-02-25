const { app, BrowserWindow, Menu, Tray, ipcMain, shell, dialog, remote, net, protocol, nativeImage } = require("electron")
const path = require("path")
const { spawn } = require("child_process")
const fs = require("fs")
const http = require("http")
let stripAnsi
import("strip-ansi").then(module => {
    stripAnsi = module.default
})
const { autoUpdater } = require("electron-updater")
const log = require("electron-log/main")
log.initialize()

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Settings
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const DENSHI_SETTINGS_DEFAULTS = {
    minimizeToTray: true,
    openInBackground: false,
    openAtLaunch: false,
}

function getDenshiSettingsPath() {
    return path.join(app.getPath("userData"), "denshi-settings.json")
}

function loadDenshiSettings() {
    try {
        const settingsPath = getDenshiSettingsPath()
        if (fs.existsSync(settingsPath)) {
            const data = JSON.parse(fs.readFileSync(settingsPath, "utf-8"))
            return { ...DENSHI_SETTINGS_DEFAULTS, ...data }
        }
    } catch (err) {
        log.error("[Denshi] Failed to load settings:", err)
    }
    return { ...DENSHI_SETTINGS_DEFAULTS }
}

function saveDenshiSettings(settings) {
    try {
        const settingsPath = getDenshiSettingsPath()
        fs.writeFileSync(settingsPath, JSON.stringify(settings, null, 2), "utf-8")
    } catch (err) {
        log.error("[Denshi] Failed to save settings:", err)
    }
}

let denshiSettings = DENSHI_SETTINGS_DEFAULTS

const _isRsbuildFrontend = true

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Chromium flags
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function setupChromiumFlags() {
    // Bypass CSP and security
    app.commandLine.appendSwitch("bypasscsp-schemes")
    app.commandLine.appendSwitch("no-sandbox")
    app.commandLine.appendSwitch("no-zygote")

    app.commandLine.appendSwitch("autoplay-policy", "no-user-gesture-required")
    app.commandLine.appendSwitch("force_high_performance_gpu")

    app.commandLine.appendSwitch("disk-cache-size", (400 * 1000 * 1000).toString())
    app.commandLine.appendSwitch("force-effective-connection-type", "4g")

    // Disable features that can interfere with playback
    app.commandLine.appendSwitch("disable-features", [
        "WidgetLayering",
        "ColorProviderRedirection",
        "WebContentsForceDarkMode",
        "HardwareMediaKeyHandling"
    ].join(","))

    // Hardware acceleration and GPU optimizations
    app.commandLine.appendSwitch("force-high-performance-gpu")
    // app.commandLine.appendSwitch('enable-gpu-rasterization');
    app.commandLine.appendSwitch("enable-zero-copy")
    app.commandLine.appendSwitch("enable-hardware-overlays", "single-fullscreen,single-on-top,underlay")
    app.commandLine.appendSwitch("ignore-gpu-blocklist")

    // Video-specific optimizations
    app.commandLine.appendSwitch("enable-accelerated-video-decode")

    // Enable advanced features
    app.commandLine.appendSwitch("enable-features", [
        "WebAssemblyLazyCompilation",
        "ThrottleDisplayNoneAndVisibilityHiddenCrossOriginIframes",
        "CanvasOopRasterization",
        "UseSkiaRenderer",
        "RawDraw",
        "PlatformEncryptedDolbyVision",
    ].join(","))

    app.commandLine.appendSwitch("enable-unsafe-webgpu")
    app.commandLine.appendSwitch("enable-gpu-rasterization")
    app.commandLine.appendSwitch("enable-oop-rasterization")

    // Background processing optimizations
    app.commandLine.appendSwitch("disable-background-timer-throttling")
    app.commandLine.appendSwitch("disable-backgrounding-occluded-windows")
    app.commandLine.appendSwitch("disable-renderer-backgrounding")
    app.commandLine.appendSwitch("disable-background-media-suspend")

    app.commandLine.appendSwitch("double-buffer-compositing")
    app.commandLine.appendSwitch("disable-direct-composition-video-overlays")

    if (process.platform === "linux") {
        log.info(`Passing --gtk-version=3 to Electron`)
        app.commandLine.appendSwitch("gtk-version", "3")
    }
}

// Setup update events for logging
autoUpdater.logger = log
log.transports.file.level = "debug"

// Redirect console logging to electron-log
console.log = log.info
console.error = log.error


const _development = process.env.NODE_ENV === "development"

// const _development = false;

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local server for youtube player embeds
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

let localServerPort

// Start local server for youtube player embeds
// Used by webviews inside React to load youtube embed and bypass 153 errors
function startLocalServer() {
    const server = http.createServer((req, res) => {
        // serve /player/:id
        const match = req.url.match(/^\/player\/([\w-]+)/)
        if (match) {
            const id = match[1]
            let url = `https://www.youtube-nocookie.com/embed/${id}?autoplay=1&enablejsapi=1&autoplay=1&playsinline=1&modestbranding=1&rel=0e`
            if (id.startsWith("compact_")) {
                url = `https://www.youtube-nocookie.com/embed/${id.substring(8)}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&vq=medium&playlist=${id.substring(8)}&cc_lang_pref=ja&enablejsapi=true`
            }
            if (id.startsWith("banner_")) {
                url = `https://www.youtube-nocookie.com/embed/${id.substring(7)}?autoplay=1&controls=0&mute=1&disablekb=1&loop=1&playlist=${id.substring(7)}&cc_lang_pref=ja&enablejsapi=true`
            }
            const html = `
        <!DOCTYPE html>
        <html lang="en">
          <head>
    <style>
      html, body {
        margin: 0;
        padding: 0;
        height: 100%;
        background-color: black;
      }
      iframe {
        position: absolute;
        inset: 0;       /* top:0; right:0; bottom:0; left:0 */
        width: 100%;
        height: 100%;
        border: none;
      }
    </style>
  </head>
        <body style="margin:0;background:black;">
          <iframe
            src="${url}"
            allow="autoplay; encrypted-media; picture-in-picture"
            allowfullscreen
            >
          </iframe>
        </body>
        </html>
      `
            res.writeHead(200, { "Content-Type": "text/html" })
            res.end(html)
            return
        }

        res.writeHead(404)
        res.end("Not found")
    })

    server.listen(0) // random free port
    const port = server.address().port
    console.log(`Local server running at http://localhost:${port}`)
    localServerPort = port
    return port
}

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
        }
    }
    ])
}

// call before app.whenReady
setupCustomProtocol()

// Sets up the app protocol to serve the static files
function setupAppProtocol() {
    if (_development) return

    const webPath = path.join(__dirname, "../web-denshi")

    if (!_isRsbuildFrontend) {
        protocol.handle("app", (request) => {
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
            log.warn(`[Protocol] Fallback for ${request.url}, serving index.html`) // Added a log
            const fallbackPath = path.join(webPath, "index.html")
            return net.fetch(`file://${fallbackPath}`)
        })
    } else {
        protocol.handle("app", async (request) => {
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

            console.error(`[App Protocol] 404 Not Found: ${urlPath}`)
            return new Response("Not Found", { status: 404 })
        })
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Startup logs
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function logStartupEvent(stage, detail = "") {
    const message = `[STARTUP] ${stage}: ${detail}`
    log.info(message)
    // console.info(message);
}

// Global error handlers to catch unhandled exceptions
process.on("uncaughtException", (error) => {
    log.error("Uncaught Exception:", error)

    if (app.isReady()) {
        dialog.showErrorBox("An error occurred", `Uncaught Exception: ${error.message}\n\nCheck the logs for more details.`)
    }

    logStartupEvent("UNCAUGHT EXCEPTION", error.stack || error.message)
})

process.on("unhandledRejection", (reason, promise) => {
    log.error("Unhandled Rejection at:", promise, "reason:", reason)
    logStartupEvent("UNHANDLED REJECTION", reason?.stack || reason?.message || JSON.stringify(reason))
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
                logStartupEvent("ERROR", "Binaries directory not found")
            }
        } catch (err) {
            logStartupEvent("ERROR reading resources", err.message)
        }
    }

    // Check app directory structure
    try {
        const appPath = app.getAppPath()
        // logStartupEvent('App directory contents', JSON.stringify(fs.readdirSync(appPath)));

        const webPath = path.join(appPath, "web-denshi")
        if (fs.existsSync(webPath)) {
            // logStartupEvent("Web directory contents", JSON.stringify(fs.readdirSync(webPath)))
        } else {
            logStartupEvent("ERROR", "web-denshi directory not found in app path")
        }
    } catch (err) {
        logStartupEvent("ERROR reading app directory", err.message)
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Updater
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const updateConfig = {
    provider: "generic",
    url: "https://github.com/5rahim/seanime/releases/latest/download",
    channel: "latest",
    allowPrerelease: false,
    verifyUpdateCodeSignature: false,
}

// Override with environment variable if set (for testing)
if (process.env.UPDATES_URL) {
    updateConfig.url = process.env.UPDATES_URL
}

// Configure the updater
autoUpdater.setFeedURL(updateConfig)

// Enable automatic download
autoUpdater.autoDownload = true
autoUpdater.autoInstallOnAppQuit = true

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


// App state
let mainWindow = null
let splashScreen = null
let crashScreen = null
let tray = null
let serverProcess = null
let isShutdown = false
let serverStarted = false
let updateDownloaded = false

// Setup autoUpdater events with improved error handling
autoUpdater.on("checking-for-update", () => {
    autoUpdater.logger.info("Checking for update...")
})

autoUpdater.on("update-available", (info) => {
    autoUpdater.logger.info("Update available:", info)
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-available", {
            version: info.version, releaseDate: info.releaseDate, files: info.files
        })
    }
})

autoUpdater.on("update-not-available", (info) => {
    autoUpdater.logger.info("Update not available:", info)
})

autoUpdater.on("download-progress", (progressObj) => {
    autoUpdater.logger.info(`Download progress: ${progressObj.percent}%`)
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("download-progress", {
            percent: progressObj.percent, bytesPerSecond: progressObj.bytesPerSecond, transferred: progressObj.transferred, total: progressObj.total
        })
    }
})

autoUpdater.on("update-downloaded", (info) => {
    autoUpdater.logger.info("Update downloaded:", info)
    updateDownloaded = true
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-downloaded", {
            version: info.version, releaseDate: info.releaseDate, files: info.files
        })
    }
})

autoUpdater.on("error", (err) => {
    autoUpdater.logger.error("Error in auto-updater:", err)
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send("update-error", {
            code: err.code || "unknown", message: err.message, stack: err.stack
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
    app.on("second-instance", (event, commandLine, workingDirectory, additionalData) => {
        if (additionalData && additionalData.development) return
        if (!serverStarted) return
        // tried to run a second instance, focus the window.
        if (mainWindow && !mainWindow.isDestroyed()) {
            if (mainWindow.isMinimized()) mainWindow.restore()
            if (!mainWindow.isVisible()) {
                if (!mainWindow.isMaximized()) mainWindow.maximize()
                mainWindow.show()
                if (process.platform === "darwin") {
                    app.dock.show()
                }
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

    let iconPath = path.join(__dirname, "../assets", iconName)
    if (_development) {
        iconPath = path.join(app.getAppPath(), "assets", iconName)
    }
    const icon = nativeImage.createFromPath(iconPath)
    tray = new Tray(icon)

    const contextMenu = Menu.buildFromTemplate([{
        id: "toggle_visibility", label: "Toggle Visibility", click: () => {
            if (!serverStarted) return
            if (mainWindow.isVisible()) {
                mainWindow.hide()
                if (process.platform === "darwin") {
                    app.dock.hide()
                }
            } else {
                if (!mainWindow.isMaximized()) mainWindow.maximize()
                mainWindow.show()
                mainWindow.focus()
                if (process.platform === "darwin") {
                    app.dock.show()
                }
            }
        }
    }, ...(process.platform === "darwin" ? [{
        id: "accessory_mode", label: "Remove from Dock", click: () => {
            app.dock.hide()
        }
    }
    ] : []), {
        id: "quit", label: "Quit Seanime", click: () => {
            cleanupAndExit()
        }
    }
    ])

    tray.setToolTip("Seanime")

    if (process.platform !== "darwin") {
        tray.setContextMenu(contextMenu)
    }

    tray.on("click", () => {
        if (!serverStarted) return
        if (mainWindow.isVisible()) {
            mainWindow.hide()
            if (process.platform === "darwin") {
                app.dock.hide()
            }
        } else {
            if (!mainWindow.isMaximized()) mainWindow.maximize()
            mainWindow.show()
            mainWindow.focus()
            if (process.platform === "darwin") {
                app.dock.show()
            }
        }
    })

    if (process.platform === "darwin") {
        tray.on("right-click", () => {
            tray.popUpContextMenu(contextMenu)
        })
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Seanime server
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

async function launchSeanimeServer(isRestart) {
    return new Promise((resolve, reject) => {
        // TEST ONLY: Check for -no-binary flag
        if (process.argv.includes("-no-binary")) {
            logStartupEvent("SKIPPING SERVER LAUNCH", "Detected -no-binary flag")
            console.log("[Main] Skipping server launch due to -no-binary flag")
            serverStarted = true // Assume server is "started" for UI flow
            // Resolve immediately to bypass server spawning
            if (splashScreen && !splashScreen.isDestroyed()) {
                splashScreen.close()
                splashScreen = null
            }
            if (mainWindow && !mainWindow.isDestroyed() && !denshiSettings.openInBackground) {
                mainWindow.maximize()
                mainWindow.show()
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

        let binaryPath

        if (_development) {
            // In development, look for binaries in the project directory
            binaryPath = path.join(__dirname, "../binaries", binaryName)
        } else {
            // In production, use the resources path
            binaryPath = path.join(process.resourcesPath, "binaries", binaryName)
        }

        logStartupEvent("Using binary", `${binaryPath} (${process.arch})`)
        logStartupEvent("Resources path", process.resourcesPath)

        // Check if binary exists and is executable
        if (!fs.existsSync(binaryPath)) {
            const error = new Error(`Server binary not found at ${binaryPath}`)
            logStartupEvent("ERROR", error.message)
            return reject(error)
        }

        // Make binary executable (for macOS/Linux)
        if (process.platform !== "win32") {
            try {
                fs.chmodSync(binaryPath, "755")
            } catch (error) {
                console.error(`Failed to make binary executable: ${error}`)
            }
        }

        // Arguments
        const args = []

        // Development mode
        if (_development) {
            args.push("-port", "43000")
            if (process.env.TEST_DATADIR) {
                console.log("[Main] TEST_DATADIR", process.env.TEST_DATADIR)
                args.push("-datadir", process.env.TEST_DATADIR)
            } else {
                const devDataDir = path.join(app.getPath("appData"), "Seanime-dev")
                args.push("-datadir", devDataDir)
            }
        }

        args.push("-desktop-sidecar", "true")

        console.log("\x1b[32m[Main] Spawning server process\x1b[0m", { args, binaryPath })

        // Spawn the process
        try {
            serverProcess = spawn(binaryPath, args)
        } catch (spawnError) {
            console.error("[Main] Failed to spawn server process synchronously:", spawnError)
            return reject(spawnError)
        }

        serverProcess.stdout.on("data", (data) => {
            const dataStr = data.toString()
            const lineStr = stripAnsi ? stripAnsi(dataStr) : dataStr

            // // Check if mainWindow exists and is not destroyed
            // if (mainWindow && !mainWindow.isDestroyed()) {
            //     mainWindow.webContents.send('message', lineStr);
            // }

            // Check if the frontend is connected
            if (!serverStarted && lineStr.includes("Client connected")) {
                console.log("[Main] Server started")
                serverStarted = true
                setTimeout(() => {
                    console.log("[Main] Server started timeout")
                    if (splashScreen && !splashScreen.isDestroyed()) {
                        splashScreen.close()
                        splashScreen = null
                    }
                    console.log("[Main] Server started close splash screen")
                    setTimeout(() => {
                        if (mainWindow && !mainWindow.isDestroyed()) {
                            if (denshiSettings.openInBackground) {
                                // Don't maximize or show
                                log.info("[Denshi] Opened in background")
                            } else {
                                mainWindow.maximize()
                                mainWindow.show()
                            }
                        }
                    }, 1000)
                    resolve()
                }, 2000)
            }
        })

        serverProcess.stderr.on("data", (data) => {
            console.error(data.toString())
        })

        serverProcess.on("close", (code) => {
            console.log(`[Main] Server process exited with code ${code}`)

            // If the server didn't start properly and we're not in the process of shutting down
            if (!serverStarted && !isShutdown) {
                console.log("[Main] Server process exited before starting")
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
        serverProcess.on("error", (err) => {
            console.error("[Main] Server process spawn error event:", err)
            reject(err)
        })
    })
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Main window
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function createMainWindow() {
    logStartupEvent("Creating main window")

    const windowOptions = {
        width: 800, height: 600, show: false,
        backgroundColor: "#111111",
        acceptFirstMouse: false,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            preload: path.join(__dirname, "preload.js"),
            webSecurity: false,
            allowRunningInsecureContent: true,
            enableBlinkFeatures: "FontAccess, AudioVideoTracks",
            backgroundThrottling: true,
            webviewTag: true,
        }
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

    mainWindow = new BrowserWindow(windowOptions)

    // Hide the title bar on Windows
    if (process.platform === "win32" || process.platform === "linux") {
        mainWindow.setMenuBarVisibility(false)
    }

    mainWindow.on("render-process-gone", (event, details) => {
        console.log("[Main] Render process gone", details)
        if (crashScreen && !crashScreen.isDestroyed()) {
            crashScreen.show()
        }
    })


    mainWindow.webContents.setWindowOpenHandler(({ webContents, frameName, url }) => {
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
            const opener = webContents.fromId(frameName || 0) || mainWindow.webContents
            // load in mainWindow instead of spawning new window
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.loadURL(url)
            }
        } catch (e) {
            console.warn("setWindowOpenHandler fallback loadURL failed", e)
        }

        return { action: "deny" }
    })

    // Load the web content
    if (_development) {
        // In development, load from the dev server
        logStartupEvent("Loading from dev server", "http://127.0.0.1:43210")
        mainWindow.loadURL("http://127.0.0.1:43210")
        // mainWindow.loadURL('chrome://gpu');
    } else {
        logStartupEvent("Loading production build with custom protocol")
        mainWindow.loadURL("app://-")
    }

    // Development tools
    // if (_development) {
    //     mainWindow.webContents.openDevTools();
    // }

    mainWindow.on("close", (event) => {
        if (!isShutdown) {
            if (denshiSettings.minimizeToTray) {
                event.preventDefault()
                mainWindow.hide()
                if (process.platform === "darwin") {
                    app.dock.hide()
                }
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
        width: 800, height: 600, frame: false, resizable: false, show: !denshiSettings.openInBackground, backgroundColor: "#0c0c0c", webPreferences: {
            nodeIntegration: false, contextIsolation: true, preload: path.join(__dirname, "preload.js")
        }
    })

    // Load the web content
    if (_development) {
        // In development, load from the dev server
        logStartupEvent("Loading splash from dev server", "http://127.0.0.1:43210/splashscreen")
        splashScreen.loadURL("http://127.0.0.1:43210/splashscreen")
    } else {
        logStartupEvent("Loading splash screen with custom protocol")
        splashScreen.loadURL("app://-/splashscreen")
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Crash screen
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function createCrashScreen() {
    crashScreen = new BrowserWindow({
        width: 800, height: 600, frame: false, resizable: false, show: false, webPreferences: {
            nodeIntegration: false, contextIsolation: true, preload: path.join(__dirname, "preload.js")
        }
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
    console.log("[Main] Cleaning up and exiting")
    isShutdown = true

    // Kill server process first
    if (serverProcess) {
        console.log("[Main] Killing server process")
        try {
            serverProcess.kill()
            serverProcess = null
        } catch (err) {
            console.error("[Main] Error killing server process:", err)
        }
    }

    // Exit the app after a short delay to allow cleanup
    setTimeout(() => {
        app.exit(0)
    }, 500)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// app.on("web-contents-created", (event, contents) => {
//     console.log("[WCC] created id=", contents.id, "type=", contents.getType())
//
//     contents.on("did-start-navigation", (e, url, isInPlace, isMainFrame) => {
//         console.log("[WCC] did-start-navigation", contents.id, { url, isMainFrame, isInPlace })
//     })
//
//     contents.on("did-navigate", (e, url) => {
//         console.log("[WCC] did-navigate", contents.id, url)
//     })
//
//     contents.on("did-frame-finish-load", () => {
//         try {
//             console.log("[WCC] URL after load", contents.id, contents.getURL())
//         } catch (e) {
//         }
//     })
//
//     contents.on("destroyed", () => console.log("[WCC] destroyed", contents.id))
//
//     try {
//         const owner = contents.getOwnerBrowserWindow && contents.getOwnerBrowserWindow()
//         if (owner) console.log("[WCC] owner window id=", owner.id, "title=", owner.getTitle && owner.getTitle())
//     } catch (e) {
//     }
// })

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Initialization
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Initialize the app
app.whenReady().then(async () => {
    logStartupEvent("App ready")

    // Load denshi settings early
    denshiSettings = loadDenshiSettings()
    if (_development) {
        denshiSettings.openInBackground = false
    }
    // Disregard openInBackground on Linux
    if (process.platform === "linux") {
        denshiSettings.openInBackground = false
    }
    log.info("[Denshi] Loaded settings:", JSON.stringify(denshiSettings))

    // Apply openAtLaunch setting (only supported on macOS and Windows)
    if (!_development && (process.platform === "darwin" || process.platform === "win32")) {
        app.setLoginItemSettings({
            openAtLogin: denshiSettings.openAtLaunch,
        })
    }

    // Set up Chromium flags for better video playback
    setupChromiumFlags()

    // Log environment information
    logEnvironmentInfo()

    // Setup IPC handlers for update functions
    ipcMain.handle("check-for-updates", async () => {
        try {
            console.log("[Main] Checking for updates...")
            const result = await autoUpdater.checkForUpdates()
            return {
                updateAvailable: !!result?.updateInfo,
                updateInfo: result?.updateInfo,
                updateDownloaded: updateDownloaded
            }
        } catch (error) {
            console.error("[Main] Error checking for updates:", error)
            throw error // Let the renderer handle the error
        }
    })

    ipcMain.handle("install-update", async () => {
        try {
            if (!updateDownloaded) {
                console.log("[Main] Update not downloaded yet, triggering download...")
                // Trigger download if not already downloaded
                await autoUpdater.checkForUpdatesAndNotify()
                throw new Error("Update download initiated. Please wait for download to complete.")
            }
            console.log("[Main] Installing update...")
            autoUpdater.quitAndInstall(false, true)
            return true
        } catch (error) {
            console.error("[Main] Error installing update:", error)
            throw error
        }
    })

    ipcMain.handle("kill-server", async () => {
        if (serverProcess) {
            console.log("[Main] Killing server before update...")
            serverProcess.kill()
            return true
        }
        return false
    })

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
    } catch (error) {
        logStartupEvent("Server launch failed", error.message)
        console.error("[Main] Failed to start server:", error)
        if (splashScreen && !splashScreen.isDestroyed()) {
            splashScreen.close()
            splashScreen = null
        }

        if (crashScreen && !crashScreen.isDestroyed()) {
            crashScreen.show()
            crashScreen.webContents.send("crash", `The server failed to start: ${error}. Closing in 10 seconds.`)

            setTimeout(() => {
                console.error("[Main] Exiting due to server start failure.")
                app.exit(1)
            }, 10000)
        }
    }

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

    ipcMain.on("window:setFullscreen", (_, fullscreen) => {
        if (mainWindow && !mainWindow.isDestroyed()) {
            mainWindow.setFullScreen(fullscreen)
        }
    })

    ipcMain.on("window:hide", () => {
        if (mainWindow && !mainWindow.isDestroyed()) {
            mainWindow.hide()
        }
    })

    ipcMain.on("window:show", () => {
        if (mainWindow && !mainWindow.isDestroyed()) {
            mainWindow.show()
        }
    })

    ipcMain.handle("window:getCurrentWindow", () => {
        const win = BrowserWindow.fromWebContents(mainWindow.webContents)
        return win?.id
    })

    ipcMain.handle("window:isMainWindow", (event) => {
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
    ipcMain.handle("clipboard:writeText", (_, text) => {
        if (text) {
            return require("electron").clipboard.writeText(text)
        }
        return false
    })

    // Register server IPC handlers
    ipcMain.on("restart-server", () => {
        console.log("EVENT restart-server")
        if (serverProcess) {
            console.log("Killing existing server process")
            serverProcess.kill()
        }
        // devnote: don't set this to false or it will trigger the crashscreen
        // serverStarted = false;
        launchSeanimeServer(true).catch(console.error)
    })

    ipcMain.on("kill-server", () => {
        console.log("EVENT kill-server")
        if (serverProcess) {
            console.log("Killing server process")
            serverProcess.kill()
        }
    })

    // Watch for window events to notify renderer
    if (mainWindow) {
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
        console.log("EVENT macos-activation-policy-accessory")
        if (process.platform === "darwin") {
            app.dock.hide()
            mainWindow.show()
            mainWindow.setFullScreen(true)

            setTimeout(() => {
                mainWindow.focus()
                mainWindow.webContents.send("macos-activation-policy-accessory-done", "")
            }, 150)
        }
    })

    ipcMain.on("macos-activation-policy-regular", () => {
        console.log("EVENT macos-activation-policy-regular")
        if (process.platform === "darwin") {
            app.dock.show()
        }
    })

    // Quit app handler
    ipcMain.on("quit-app", () => {
        console.log("EVENT quit-app")
        cleanupAndExit()
    })

    ipcMain.handle("get-local-server-port", () => localServerPort)

    // Denshi settings IPC handlers
    ipcMain.handle("denshi:getSettings", () => {
        return { ...denshiSettings }
    })

    ipcMain.handle("denshi:setSettings", (_, newSettings) => {
        denshiSettings = { ...DENSHI_SETTINGS_DEFAULTS, ...newSettings }
        saveDenshiSettings(denshiSettings)
        log.info("[Denshi] Settings updated:", JSON.stringify(denshiSettings))

        // Apply openAtLaunch immediately (only supported on macOS and Windows)
        if (!_development && (process.platform === "darwin" || process.platform === "win32")) {
            app.setLoginItemSettings({
                openAtLogin: denshiSettings.openAtLaunch,
            })
        }

        return { ...denshiSettings }
    })

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
        console.log("EVENT before-quit")
        cleanupAndExit()
    })
})
