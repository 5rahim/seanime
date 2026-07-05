import "@total-typescript/ts-reset"

declare global {
    interface AudioTrack {
        id: string;
        kind: string;
        label: string;
        language: string;
        enabled: boolean;
    }

    interface AudioTrackList extends EventTarget {
        readonly length: number;
        onchange: ((this: AudioTrackList, ev: Event) => any) | null;
        onaddtrack: ((this: AudioTrackList, ev: TrackEvent) => any) | null;
        onremovetrack: ((this: AudioTrackList, ev: TrackEvent) => any) | null;

        [index: number]: AudioTrack;

        getTrackById(id: string): AudioTrack | null;
    }

    interface HTMLMediaElement {
        readonly audioTracks: AudioTrackList | undefined;
    }

    interface Window {
        electron?: {
            window: {
                minimize: () => void;
                maximize: () => void;
                close: () => void;
                isMaximized: () => Promise<boolean>;
                isMinimizable: () => Promise<boolean>;
                isMaximizable: () => Promise<boolean>;
                isClosable: () => Promise<boolean>;
                isFullscreen: () => Promise<boolean>;
                setFullscreen: (fullscreen: boolean) => void;
                toggleMaximize: () => void;
                hide: () => void;
                show: () => void;
                isVisible: () => Promise<boolean>;
                setTitleBarStyle: (style: string) => void;
                getCurrentWindow: () => Promise<string>;
                isMainWindow: () => Promise<boolean>;
            };
            localServer: {
                getPort: () => Promise<number>;
                allowWebviewOrigin?: (origin: string) => Promise<boolean>;
            },
            startup: {
                ready: () => void;
            },
            media?: {
                setMetadata: (metadata: any) => Promise<boolean>
                clearSession: () => Promise<boolean>
                stopAllMedia: () => Promise<boolean>
            }
            on: (channel: string, callback: (...args: any[]) => void) => (() => void) | undefined;
            // Send events
            emit: (channel: string, data?: any) => void;
            // General send method
            send: (channel: string, ...args: any[]) => void;
            platform: NodeJS.Platform;
            shell: {
                open: (url: string) => Promise<void>;
            };
            clipboard: {
                writeText: (text: string) => Promise<void>;
            };
            checkForUpdates: () => Promise<any>;
            installUpdate: () => Promise<any>;
            killServer: () => Promise<any>;
            denshiSettings: {
                get: () => Promise<DenshiSettings>;
                set: (settings: DenshiSettings) => Promise<DenshiSettings>;
            };
            mpvCore: {
                createTempSubtitle: (filename: string, content: string) => Promise<string>;
                writeConfigFile: (content: string) => Promise<string | null>;
                createScreenshotPath: () => Promise<string>;
                saveScreenshot: (filePath: string, base64Data: string) => Promise<boolean>;
                setLoggingEnabled: (enabled: boolean) => Promise<boolean>;
                getAnime4KDirectory: () => Promise<MpvCoreAnime4KDirectory>;
                scanAnime4KDirectory: (directory: string) => Promise<MpvCoreAnime4KDirectory>;
                openAnime4KDirectory: (directory?: string) => Promise<boolean>;
            };
            powerSaveBlocker?: {
                start: () => Promise<number>;
                stop: (id: number) => Promise<void>;
            };
            cast?: {
                discover: () => Promise<void>;
                stopDiscovery: () => Promise<void>;
                getDevices: () => Promise<CastDevice[]>;
                connect: (deviceId: string) => Promise<CastSessionState>;
                disconnect: () => Promise<void>;
                getStatus: () => Promise<CastStatus>;
                loadMedia: (opts: CastLoadMediaOptions) => Promise<number>;
                play: () => Promise<void>;
                pause: () => Promise<void>;
                seek: (time: number) => Promise<void>;
                stop: () => Promise<void>;
                setVolume: (level: number) => Promise<void>;
                setMuted: (muted: boolean) => Promise<void>;
                sendSubtitleEvents: (events: any[]) => Promise<void>;
                sendSubtitleTracks: (tracks: any[]) => Promise<void>;
                switchSubtitleTrack: (trackNumber: number) => Promise<void>;
                sendFonts: (fontUrls: string[], serverPort?: number) => Promise<void>;
                sendSubtitleHeader: (header: string) => Promise<void>;
                disableSubtitles: () => Promise<void>;
                getLanIP: () => Promise<string>;
            };
        };

        __isElectronDesktop__?: boolean;
    }

    interface CastDevice {
        id: string;
        name: string;
        host: string;
        port: number;
    }

    interface MpvCoreAnime4KDirectory {
        directory: string;
        shaders: Array<{
            name: string;
            path: string;
        }>;
    }

    interface CastSessionState {
        connected: boolean;
        device: CastDevice | null;
        sessionId: string | null;
    }

    interface CastStatus {
        connected: boolean;
        device: CastDevice | null;
        sessionId: string | null;
        mediaStatus: CastMediaStatus | null;
    }

    interface CastMediaStatus {
        mediaSessionId: number;
        playerState: "IDLE" | "BUFFERING" | "PLAYING" | "PAUSED";
        currentTime: number;
        duration?: number;
        volume?: { level: number; muted: boolean };
        idleReason?: string;
    }

    interface CastLoadMediaOptions {
        streamUrl: string;
        contentType: string;
        title?: string;
        subtitle?: string;
        imageUrl?: string;
        duration?: number;
        serverPort?: number;
    }

    interface DenshiSettings {
        minimizeToTray: boolean;
        openInBackground: boolean;
        openAtLaunch: boolean;
        updateChannel?: string;
    }
}
