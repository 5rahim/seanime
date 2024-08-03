import { Models_MediastreamSettings } from "@/api/generated/types"
import { useSaveMediastreamSettings } from "@/api/hooks/mediastream.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Alert } from "@/components/ui/alert"
import { Checkbox } from "@/components/ui/checkbox"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { FcFolder } from "react-icons/fc"
import { MdOutlineDevices } from "react-icons/md"

const mediastreamSchema = defineSchema(({ z }) => z.object({
    transcodeEnabled: z.boolean(),
    transcodeHwAccel: z.string(),
    transcodeThreads: z.number(),
    transcodePreset: z.string().min(2),
    transcodeTempDir: z.string().min(2),
    preTranscodeEnabled: z.boolean(),
    preTranscodeLibraryDir: z.string(),
    disableAutoSwitchToDirectPlay: z.boolean(),
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
    isLoading: boolean
}

export function MediastreamSettings(props: MediastreamSettingsProps) {

    const {
        children,
        settings,
        isLoading,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const { mutate, isPending } = useSaveMediastreamSettings()

    const { activeOnDevice, setActiveOnDevice } = useMediastreamActiveOnDevice()

    // const { data: totalSize, mutate: getTotalSize, isPending: isFetchingSize } = useGetFileCacheMediastreamVideoFilesTotalSize()

    // const { mutate: clearCache, isPending: isClearing } = useClearFileCacheMediastreamVideoFiles(() => {
    //     getTotalSize()
    // })

    if (!settings) return <LoadingSpinner />

    return (
        <>
            <Form
                schema={mediastreamSchema}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                                transcodeTempDir: data.transcodeTempDir,
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
                    disableAutoSwitchToDirectPlay: settings?.disableAutoSwitchToDirectPlay,
                    ffmpegPath: settings?.ffmpegPath,
                    ffprobePath: settings?.ffprobePath,
                }}
                stackClass="space-y-6"
            >
                <Field.Switch
                    name="transcodeEnabled"
                    label="Enable"
                />

                <div className="flex gap-4 items-center border rounded-md p-2 lg:p-4">
                    <MdOutlineDevices className="text-4xl" />
                    <div className="space-y-1">
                        <Checkbox
                            name=""
                            value={activeOnDevice ?? false}
                            onValueChange={v => setActiveOnDevice((prev) => typeof v === "boolean" ? v : prev)}
                            label="Use media streaming on this device"
                            help="Enable this option if you want to use media streaming on this device."
                        />
                        <p className="text-gray-200">
                            Current client: {serverStatus?.clientDevice}, {serverStatus?.clientPlatform}
                        </p>
                    </div>
                </div>

                {activeOnDevice && (
                    <Alert
                        intent="info" description={<>
                        If media streaming is enabled, your downloaded files will be transcoded and played back using the built-in player.
                    </>}
                    />
                )}

                <Separator />

                <Field.Switch
                    name="disableAutoSwitchToDirectPlay"
                    label="Don't auto switch to direct play"
                    help="By default, Seanime will automatically switch to direct play if the media codec is supported by the client."
                />

                <Separator />

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

            {/*<Separator />*/}

            {/*<h2>Cache</h2>*/}

            {/*<div className="space-y-4">*/}
            {/*    <div className="flex gap-2 items-center">*/}
            {/*        <Button intent="white-subtle" size="sm" onClick={() => getTotalSize()} disabled={isFetchingSize}>*/}
            {/*            Show total size*/}
            {/*        </Button>*/}
            {/*        {!!totalSize && (*/}
            {/*            <p>*/}
            {/*                {totalSize}*/}
            {/*            </p>*/}
            {/*        )}*/}
            {/*    </div>*/}
            {/*    <div className="flex gap-2 flex-wrap items-center">*/}
            {/*        <Button intent="alert-subtle" size="sm" onClick={() => clearCache()} disabled={isClearing}>*/}
            {/*            Clear cache*/}
            {/*        </Button>*/}
            {/*    </div>*/}
            {/*</div>*/}
        </>
    )
}
