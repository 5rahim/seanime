const { contextBridge, ipcRenderer } = require("electron")

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld(
    'electron', {
        // Window Controls
        window: {
            minimize: () => ipcRenderer.send("window:minimize"),
            maximize: () => ipcRenderer.send("window:maximize"),
            close: () => ipcRenderer.send("window:close"),
            isMaximized: () => ipcRenderer.invoke("window:isMaximized"),
            isMinimizable: () => ipcRenderer.invoke("window:isMinimizable"),
            isMaximizable: () => ipcRenderer.invoke("window:isMaximizable"),
            isClosable: () => ipcRenderer.invoke("window:isClosable"),
            isFullscreen: () => ipcRenderer.invoke("window:isFullscreen"),
            setFullscreen: (fullscreen) => ipcRenderer.send("window:setFullscreen", fullscreen),
            toggleMaximize: () => ipcRenderer.send("window:toggleMaximize"),
            hide: () => ipcRenderer.send("window:hide"),
            show: () => ipcRenderer.send("window:show"),
            isVisible: () => ipcRenderer.invoke("window:isVisible"),
            setTitleBarStyle: (style) => ipcRenderer.send("window:setTitleBarStyle", style),
            getCurrentWindow: () => ipcRenderer.invoke("window:getCurrentWindow"),
            isMainWindow: () => ipcRenderer.send("window:isMainWindow"),
        },

        localServer: {
            getPort: () => ipcRenderer.invoke("get-local-server-port"),
            allowWebviewOrigin: (origin) => ipcRenderer.invoke("denshi:allowWebviewOrigin", origin),
        },

        startup: {
            ready: () => ipcRenderer.send("startup:renderer-ready")
        },

        // Event listeners
        on: (channel, callback) => {
            // Whitelist channels
            const validChannels = [
                "message",
                "crash",
                "window:minimized",
                "window:hidden",
                "window:maximized",
                "window:unmaximized",
                "window:fullscreen",
                "update-downloaded",
                "update-error",
                "update-available",
                "download-progress",
                "window:currentWindow",
                "window:isMainWindow",
                "cast:deviceFound",
                "cast:sessionUpdate",
                "cast:mediaStatus",
                "cast:receiverReady",
                "cast:error",
            ]
            if (validChannels.includes(channel)) {
                // Remove the event listener to avoid memory leaks
                ipcRenderer.removeAllListeners(channel)
                // Add the event listener
                ipcRenderer.on(channel, (_, ...args) => callback(...args))

                // Return a function to remove the listener
                return () => {
                    ipcRenderer.removeAllListeners(channel)
                }
            }
        },

        // Send events
        emit: (channel, data) => {
            // Whitelist channels
            const validChannels = [
                "restart-server",
                "kill-server",
                "macos-activation-policy-accessory",
                "macos-activation-policy-regular"
            ]
            if (validChannels.includes(channel)) {
                ipcRenderer.send(channel, data)
            }
        },

        // General send method for any whitelisted channel
        send: (channel, ...args) => {
            // Whitelist channels
            const validChannels = [
                "restart-app",
                "quit-app"
            ]
            if (validChannels.includes(channel)) {
                ipcRenderer.send(channel, ...args)
            }
        },

        // Platform
        platform: process.platform,

        // Clipboard
        clipboard: {
            writeText: (text) => ipcRenderer.invoke("clipboard:writeText", text)
        },

        // Update functions
        checkForUpdates: () => ipcRenderer.invoke("check-for-updates"),
        installUpdate: () => ipcRenderer.invoke("install-update"),
        killServer: () => ipcRenderer.invoke("kill-server"),

        // Denshi Settings
        denshiSettings: {
            get: () => ipcRenderer.invoke("denshi:getSettings"),
            set: (settings) => ipcRenderer.invoke("denshi:setSettings", settings),
        },

        mpvCore: {
            createTempSubtitle: (filename, content) => ipcRenderer.invoke("mpvcore:create-temp-subtitle", filename, content),
            writeConfigFile: (content) => ipcRenderer.invoke("mpvcore:write-config-file", content),
            createScreenshotPath: () => ipcRenderer.invoke("mpvcore:create-screenshot-path"),
            saveScreenshot: (filePath, base64Data) => ipcRenderer.invoke("mpvcore:save-screenshot", filePath, base64Data),
            setLoggingEnabled: (enabled) => ipcRenderer.invoke("mpvcore:setLoggingEnabled", enabled),
            getAnime4KDirectory: () => ipcRenderer.invoke("mpvcore:get-anime4k-directory"),
            scanAnime4KDirectory: (directory) => ipcRenderer.invoke("mpvcore:scan-anime4k-directory", directory),
            openAnime4KDirectory: (directory) => ipcRenderer.invoke("mpvcore:open-anime4k-directory", directory),
        },

        powerSaveBlocker: {
            start: () => ipcRenderer.invoke("power-save-blocker:start"),
            stop: (id) => ipcRenderer.invoke("power-save-blocker:stop", id),
        },

        // Chromecast
        cast: {
            discover: () => ipcRenderer.invoke("cast:discover"),
            stopDiscovery: () => ipcRenderer.invoke("cast:stopDiscovery"),
            getDevices: () => ipcRenderer.invoke("cast:getDevices"),
            connect: (deviceId) => ipcRenderer.invoke("cast:connect", deviceId),
            disconnect: () => ipcRenderer.invoke("cast:disconnect"),
            getStatus: () => ipcRenderer.invoke("cast:getStatus"),
            loadMedia: (opts) => ipcRenderer.invoke("cast:loadMedia", opts),
            play: () => ipcRenderer.invoke("cast:play"),
            pause: () => ipcRenderer.invoke("cast:pause"),
            seek: (time) => ipcRenderer.invoke("cast:seek", time),
            stop: () => ipcRenderer.invoke("cast:stop"),
            setVolume: (level) => ipcRenderer.invoke("cast:setVolume", level),
            setMuted: (muted) => ipcRenderer.invoke("cast:setMuted", muted),
            sendSubtitleEvents: (events) => ipcRenderer.invoke("cast:sendSubtitleEvents", events),
            sendSubtitleTracks: (tracks) => ipcRenderer.invoke("cast:sendSubtitleTracks", tracks),
            switchSubtitleTrack: (trackNumber) => ipcRenderer.invoke("cast:switchSubtitleTrack", trackNumber),
            sendFonts: (fontUrls, serverPort) => ipcRenderer.invoke("cast:sendFonts", fontUrls, serverPort),
            sendSubtitleHeader: (header) => ipcRenderer.invoke("cast:sendSubtitleHeader", header),
            disableSubtitles: () => ipcRenderer.invoke("cast:disableSubtitles"),
            getLanIP: () => ipcRenderer.invoke("cast:getLanIP"),
        },
    }
);

// Set __isElectronDesktop__ global variable
contextBridge.exposeInMainWorld('__isElectronDesktop__', true);
