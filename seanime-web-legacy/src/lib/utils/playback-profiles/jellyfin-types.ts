/**
 *
 * @export
 * @interface DirectPlayProfile
 */
export interface DirectPlayProfile {
    /**
     *
     * @type {string}
     * @memberof DirectPlayProfile
     */
    "Container"?: string | null;
    /**
     *
     * @type {string}
     * @memberof DirectPlayProfile
     */
    "AudioCodec"?: string | null;
    /**
     *
     * @type {string}
     * @memberof DirectPlayProfile
     */
    "VideoCodec"?: string | null;
    /**
     *
     * @type {DlnaProfileType}
     * @memberof DirectPlayProfile
     */
    "Type"?: DlnaProfileType;
}

export const DlnaProfileType = {
    Audio: "Audio",
    Video: "Video",
    Photo: "Photo",
    Subtitle: "Subtitle",
} as const

export type DlnaProfileType = typeof DlnaProfileType[keyof typeof DlnaProfileType];

/**
 * Delivery method to use during playback of a specific subtitle format.
 * @export
 * @enum {string}
 */

export const SubtitleDeliveryMethod = {
    Encode: "Encode",
    Embed: "Embed",
    External: "External",
    Hls: "Hls",
    Drop: "Drop",
} as const

export type SubtitleDeliveryMethod = typeof SubtitleDeliveryMethod[keyof typeof SubtitleDeliveryMethod];

/**
 *
 * @export
 * @interface SubtitleProfile
 */
export interface SubtitleProfile {
    /**
     *
     * @type {string}
     * @memberof SubtitleProfile
     */
    "Format"?: string | null;
    /**
     *
     * @type {SubtitleDeliveryMethod}
     * @memberof SubtitleProfile
     */
    "Method"?: SubtitleDeliveryMethod;
    /**
     *
     * @type {string}
     * @memberof SubtitleProfile
     */
    "DidlMode"?: string | null;
    /**
     *
     * @type {string}
     * @memberof SubtitleProfile
     */
    "Language"?: string | null;
    /**
     *
     * @type {string}
     * @memberof SubtitleProfile
     */
    "Container"?: string | null;
}

/**
 *
 * @export
 * @enum {string}
 */

export const ProfileConditionType = {
    Equals: "Equals",
    NotEquals: "NotEquals",
    LessThanEqual: "LessThanEqual",
    GreaterThanEqual: "GreaterThanEqual",
    EqualsAny: "EqualsAny",
} as const

export type ProfileConditionType = typeof ProfileConditionType[keyof typeof ProfileConditionType];

/**
 *
 * @export
 * @enum {string}
 */

export const EncodingContext = {
    Streaming: "Streaming",
    Static: "Static",
} as const

export type EncodingContext = typeof EncodingContext[keyof typeof EncodingContext];

/**
 *
 * @export
 * @enum {string}
 */

export const ProfileConditionValue = {
    AudioChannels: "AudioChannels",
    AudioBitrate: "AudioBitrate",
    AudioProfile: "AudioProfile",
    Width: "Width",
    Height: "Height",
    Has64BitOffsets: "Has64BitOffsets",
    PacketLength: "PacketLength",
    VideoBitDepth: "VideoBitDepth",
    VideoBitrate: "VideoBitrate",
    VideoFramerate: "VideoFramerate",
    VideoLevel: "VideoLevel",
    VideoProfile: "VideoProfile",
    VideoTimestamp: "VideoTimestamp",
    IsAnamorphic: "IsAnamorphic",
    RefFrames: "RefFrames",
    NumAudioStreams: "NumAudioStreams",
    NumVideoStreams: "NumVideoStreams",
    IsSecondaryAudio: "IsSecondaryAudio",
    VideoCodecTag: "VideoCodecTag",
    IsAvc: "IsAvc",
    IsInterlaced: "IsInterlaced",
    AudioSampleRate: "AudioSampleRate",
    AudioBitDepth: "AudioBitDepth",
    VideoRangeType: "VideoRangeType",
} as const

export type ProfileConditionValue = typeof ProfileConditionValue[keyof typeof ProfileConditionValue];

/**
 *
 * @export
 * @interface ProfileCondition
 */
export interface ProfileCondition {
    /**
     *
     * @type {ProfileConditionType}
     * @memberof ProfileCondition
     */
    "Condition"?: ProfileConditionType;
    /**
     *
     * @type {ProfileConditionValue}
     * @memberof ProfileCondition
     */
    "Property"?: ProfileConditionValue;
    /**
     *
     * @type {string}
     * @memberof ProfileCondition
     */
    "Value"?: string | null;
    /**
     *
     * @type {boolean}
     * @memberof ProfileCondition
     */
    "IsRequired"?: boolean;
}

/**
 *
 * @export
 * @enum {string}
 */

export const TranscodeSeekInfo = {
    Auto: "Auto",
    Bytes: "Bytes",
} as const

export type TranscodeSeekInfo = typeof TranscodeSeekInfo[keyof typeof TranscodeSeekInfo];

/**
 *
 * @export
 * @interface TranscodingProfile
 */
export interface TranscodingProfile {
    /**
     *
     * @type {string}
     * @memberof TranscodingProfile
     */
    "Container"?: string;
    /**
     *
     * @type {DlnaProfileType}
     * @memberof TranscodingProfile
     */
    "Type"?: DlnaProfileType;
    /**
     *
     * @type {string}
     * @memberof TranscodingProfile
     */
    "VideoCodec"?: string;
    /**
     *
     * @type {string}
     * @memberof TranscodingProfile
     */
    "AudioCodec"?: string;
    /**
     *
     * @type {string}
     * @memberof TranscodingProfile
     */
    "Protocol"?: string;
    /**
     *
     * @type {boolean}
     * @memberof TranscodingProfile
     */
    "EstimateContentLength"?: boolean;
    /**
     *
     * @type {boolean}
     * @memberof TranscodingProfile
     */
    "EnableMpegtsM2TsMode"?: boolean;
    /**
     *
     * @type {TranscodeSeekInfo}
     * @memberof TranscodingProfile
     */
    "TranscodeSeekInfo"?: TranscodeSeekInfo;
    /**
     *
     * @type {boolean}
     * @memberof TranscodingProfile
     */
    "CopyTimestamps"?: boolean;
    /**
     *
     * @type {EncodingContext}
     * @memberof TranscodingProfile
     */
    "Context"?: EncodingContext;
    /**
     *
     * @type {boolean}
     * @memberof TranscodingProfile
     */
    "EnableSubtitlesInManifest"?: boolean;
    /**
     *
     * @type {string}
     * @memberof TranscodingProfile
     */
    "MaxAudioChannels"?: string | null;
    /**
     *
     * @type {number}
     * @memberof TranscodingProfile
     */
    "MinSegments"?: number;
    /**
     *
     * @type {number}
     * @memberof TranscodingProfile
     */
    "SegmentLength"?: number;
    /**
     *
     * @type {boolean}
     * @memberof TranscodingProfile
     */
    "BreakOnNonKeyFrames"?: boolean;
    /**
     *
     * @type {Array<ProfileCondition>}
     * @memberof TranscodingProfile
     */
    "Conditions"?: Array<ProfileCondition>;
}

/**
 *
 * @export
 * @enum {string}
 */

export const CodecType = {
    Video: "Video",
    VideoAudio: "VideoAudio",
    Audio: "Audio",
} as const

export type CodecType = typeof CodecType[keyof typeof CodecType];

/**
 *
 * @export
 * @interface CodecProfile
 */
export interface CodecProfile {
    /**
     *
     * @type {CodecType}
     * @memberof CodecProfile
     */
    "Type"?: CodecType;
    /**
     *
     * @type {Array<ProfileCondition>}
     * @memberof CodecProfile
     */
    "Conditions"?: Array<ProfileCondition> | null;
    /**
     *
     * @type {Array<ProfileCondition>}
     * @memberof CodecProfile
     */
    "ApplyConditions"?: Array<ProfileCondition> | null;
    /**
     *
     * @type {string}
     * @memberof CodecProfile
     */
    "Codec"?: string | null;
    /**
     *
     * @type {string}
     * @memberof CodecProfile
     */
    "Container"?: string | null;
}
