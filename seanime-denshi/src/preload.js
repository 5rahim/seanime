const {contextBridge, ipcRenderer, shell} = require('electron');

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld(
    'electron', {
        // Window Controls
        window: {
            minimize: () => ipcRenderer.send('window:minimize'),
            maximize: () => ipcRenderer.send('window:maximize'),
            close: () => ipcRenderer.send('window:close'),
            isMaximized: () => ipcRenderer.invoke('window:isMaximized'),
            isMinimizable: () => ipcRenderer.invoke('window:isMinimizable'),
            isMaximizable: () => ipcRenderer.invoke('window:isMaximizable'),
            isClosable: () => ipcRenderer.invoke('window:isClosable'),
            isFullscreen: () => ipcRenderer.invoke('window:isFullscreen'),
            setFullscreen: (fullscreen) => ipcRenderer.send('window:setFullscreen', fullscreen),
            toggleMaximize: () => ipcRenderer.send('window:toggleMaximize'),
            hide: () => ipcRenderer.send('window:hide'),
            show: () => ipcRenderer.send('window:show'),
            isVisible: () => ipcRenderer.invoke('window:isVisible'),
            setTitleBarStyle: (style) => ipcRenderer.send('window:setTitleBarStyle', style)
        },

        // Event listeners
        on: (channel, callback) => {
            // Whitelist channels
            const validChannels = [
                'message',
                'crash',
                'window:maximized',
                'window:unmaximized',
                'window:fullscreen',
                'update-downloaded',
                'update-error',
                'update-available',
                'download-progress'
            ];
            if (validChannels.includes(channel)) {
                // Remove the event listener to avoid memory leaks
                ipcRenderer.removeAllListeners(channel);
                // Add the event listener
                ipcRenderer.on(channel, (_, ...args) => callback(...args));

                // Return a function to remove the listener
                return () => {
                    ipcRenderer.removeAllListeners(channel);
                };
            }
        },

        // Send events
        emit: (channel, data) => {
            // Whitelist channels
            const validChannels = [
                'restart-server',
                'kill-server',
                'macos-activation-policy-accessory',
                'macos-activation-policy-regular'
            ];
            if (validChannels.includes(channel)) {
                ipcRenderer.send(channel, data);
            }
        },

        // General send method for any whitelisted channel
        send: (channel, ...args) => {
            // Whitelist channels
            const validChannels = [
                'restart-app',
                'quit-app'
            ];
            if (validChannels.includes(channel)) {
                ipcRenderer.send(channel, ...args);
            }
        },

        // Platform
        platform: process.platform,

        // Shell functions
        shell: {
            open: (url) => shell.openExternal(url)
        },

        // Clipboard
        clipboard: {
            writeText: (text) => ipcRenderer.invoke('clipboard:writeText', text)
        },

        // Update functions
        checkForUpdates: () => ipcRenderer.invoke('check-for-updates'),
        installUpdate: () => ipcRenderer.invoke('install-update'),
        killServer: () => ipcRenderer.invoke('kill-server')
    }
);

// Set __isElectronDesktop__ global variable
contextBridge.exposeInMainWorld('__isElectronDesktop__', true);
