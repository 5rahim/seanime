import { MKVParser_SubtitleEvent, NativePlayer_SubtitleEventsPayload } from "@/api/generated/types"

export function isSubtitleBatchCurrent(
    batch: NativePlayer_SubtitleEventsPayload,
    playbackId: string,
    generationId: number,
) {
    return batch.playbackId === playbackId && batch.generationId >= generationId
}

export function getSubtitleEvents(
    batches: NativePlayer_SubtitleEventsPayload[],
    playbackId: string,
    generationId: number,
): MKVParser_SubtitleEvent[] {
    return batches
        .filter(batch => batch.playbackId === playbackId && batch.generationId === generationId)
        .flatMap(batch => batch.events ?? [])
}
