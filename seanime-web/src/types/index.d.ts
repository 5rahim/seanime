import "@total-typescript/ts-reset"

// audioTracks
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
        /** Not standardized, supported in some browsers (e.g., Chrome) */
        readonly audioTracks: AudioTrackList | undefined;
    }
}
