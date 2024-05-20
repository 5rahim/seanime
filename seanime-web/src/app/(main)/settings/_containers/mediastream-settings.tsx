import { Models_MediastreamSettings } from "@/api/generated/types"
import { useClearFileCacheMediastreamVideoFiles, useGetFileCacheMediastreamVideoFilesTotalSize } from "@/api/hooks/filecache.hooks"
import { useSaveMediastreamSettings } from "@/api/hooks/mediastream.hooks"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Button } from "@/components/ui/button"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { FcFolder } from "react-icons/fc"

const mediastreamSchema = defineSchema(({ z }) => z.object({
    transcodeEnabled: z.boolean(),
    transcodeHwAccel: z.string(),
    transcodeThreads: z.number(),
    transcodePreset: z.string().min(2),
    transcodeTempDir: z.string(),
    preTranscodeEnabled: z.boolean(),
    preTranscodeLibraryDir: z.string(),
    ffmpegPath: z.string().min(0),
    ffprobePath: z.string().min(0),
}))

const MEDIASTREAM_HW_ACCEL_OPTIONS = [
    { label: "CPU (Disabled)", value: "cpu" },
    { label: "NVIDIA (NVENC)", value: "nvidia" },
    { label: "Intel (QSV)", value: "qsv" },
    { label: "VAAPI", value: "vaapi" },
]

const MEDIASTREAM_PRESET_OPTIONS = [
    { label: "Ultrafast", value: "ultrafast" },
    { label: "Superfast", value: "superfast" },
    { label: "Veryfast", value: "veryfast" },
    { label: "Fast", value: "fast" },
    { label: "Medium", value: "medium" },
]

type MediastreamSettingsProps = {
    children?: React.ReactNode
    settings: Models_MediastreamSettings | undefined
}

export function MediastreamSettings(props: MediastreamSettingsProps) {

    const {
        children,
        settings,
        ...rest
    } = props

    const { mutate, isPending } = useSaveMediastreamSettings()

    const { data: totalSize, mutate: getTotalSize, isPending: isFetchingSize } = useGetFileCacheMediastreamVideoFilesTotalSize()

    const { mutate: clearCache, isPending: isClearing } = useClearFileCacheMediastreamVideoFiles(() => {
        getTotalSize()
    })

    if (!settings) return null

    return (
        <>
            <h2>Transcoding</h2>

            <Form
                schema={mediastreamSchema}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                            },
                        })
                    }
                }}
                defaultValues={{
                    transcodeEnabled: settings?.transcodeEnabled,
                    transcodeHwAccel: settings?.transcodeHwAccel || "cpu",
                    transcodeThreads: settings?.transcodeThreads,
                    transcodePreset: settings?.transcodePreset,
                    transcodeTempDir: settings?.transcodeTempDir,
                    preTranscodeEnabled: settings?.preTranscodeEnabled,
                    preTranscodeLibraryDir: settings?.preTranscodeLibraryDir,
                    ffmpegPath: settings?.ffmpegPath,
                    ffprobePath: settings?.ffprobePath,
                }}
                stackClass="space-y-6"
            >
                <Field.Switch
                    name="transcodeEnabled"
                    label="Enable real-time transcoding"
                    help="Enable transcoding for media files."
                />

                <Field.DirectorySelector
                    name="transcodeTempDir"
                    label="Transcode directory"
                    leftIcon={<FcFolder />}
                    help="Directory where transcoded files are temporarily stored. This directory should be different from your library directory."
                    shouldExist
                />

                <Field.Select
                    options={MEDIASTREAM_HW_ACCEL_OPTIONS}
                    name="transcodeHwAccel"
                    label="Hardware acceleration"
                    help="Hardware acceleration is highly recommended for a smoother transcoding experience."
                />

                <Field.Select
                    options={MEDIASTREAM_PRESET_OPTIONS}
                    name="transcodePreset"
                    label="Transcode preset"
                    help="'Fast' is recommended. VAAPI does not support presets."
                />

                <div className="flex gap-3 items-center">
                    <Field.Text
                        name="ffmpegPath"
                        label="FFmpeg path"
                        help="Path to the FFmpeg binary. Leave empty if binary is already in your PATH."
                    />

                    <Field.Text
                        name="ffprobePath"
                        label="FFprobe path"
                        help="Path to the FFprobe binary. Leave empty if binary is already in your PATH."
                    />
                </div>

                <SettingsSubmitButton isPending={isPending} />
            </Form>

            <Separator />

            <h2>Cache</h2>

            <div className="space-y-4">
                <div className="flex gap-2 items-center">
                    <Button intent="white-subtle" size="sm" onClick={() => getTotalSize()} disabled={isFetchingSize}>
                        Show total size
                    </Button>
                    {!!totalSize && (
                        <p>
                            {totalSize}
                        </p>
                    )}
                </div>
                <div className="flex gap-2 flex-wrap items-center">
                    <Button intent="alert-subtle" size="sm" onClick={() => clearCache()} disabled={isClearing}>
                        Clear cache
                    </Button>
                </div>
            </div>
        </>
    )
}
