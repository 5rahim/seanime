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
            };
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
        };

        __isElectronDesktop__?: boolean;
    }
}
