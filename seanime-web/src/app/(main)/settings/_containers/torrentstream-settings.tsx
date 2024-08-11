import { Models_TorrentstreamSettings } from "@/api/generated/types"
import { useSaveTorrentstreamSettings, useTorrentstreamDropTorrent } from "@/api/hooks/torrentstream.hooks"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Button } from "@/components/ui/button"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { SiBittorrent } from "react-icons/si"

const torrentstreamSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    downloadDir: z.string(),
    autoSelect: z.boolean(),
    disableIPv6: z.boolean(),
    addToLibrary: z.boolean(),
    streamingServerPort: z.number(),
    streamingServerHost: z.string(),
    torrentClientPort: z.number(),
    preferredResolution: z.string(),
    fallbackToTorrentStreamingView: z.boolean(),
    includeInLibrary: z.boolean(),
}))


type TorrentstreamSettingsProps = {
    children?: React.ReactNode
    settings: Models_TorrentstreamSettings | undefined
}

export function TorrentstreamSettings(props: TorrentstreamSettingsProps) {

    const {
        children,
        settings,
        ...rest
    } = props

    const { mutate, isPending } = useSaveTorrentstreamSettings()

    const { mutate: dropTorrent, isPending: droppingTorrent } = useTorrentstreamDropTorrent()

    if (!settings) return null

    return (
        <>
            <Form
                schema={torrentstreamSchema}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                                preferredResolution: data.preferredResolution === "-" ? "" : data.preferredResolution,
                            },
                        })
                    }
                }}
                defaultValues={{
                    enabled: settings.enabled,
                    autoSelect: settings.autoSelect,
                    downloadDir: settings.downloadDir || "",
                    disableIPv6: settings.disableIPV6,
                    addToLibrary: settings.addToLibrary,
                    streamingServerPort: settings.streamingServerPort,
                    streamingServerHost: settings.streamingServerHost,
                    torrentClientPort: settings.torrentClientPort,
                    preferredResolution: settings.preferredResolution || "-",
                    fallbackToTorrentStreamingView: settings.fallbackToTorrentStreamingView,
                    includeInLibrary: settings.includeInLibrary,
                }}
                stackClass="space-y-6"
            >
                <Field.Switch
                    name="enabled"
                    label="Enable"
                />

                <Separator />

                <h3>
                    Integration
                </h3>

                <Field.Switch
                    name="fallbackToTorrentStreamingView"
                    label="Fall back to torrent streaming view"
                    help="If the anime is not in your library, default to the torrent streaming view."
                />

                <Field.Switch
                    name="includeInLibrary"
                    label="Include in library"
                    help="Make non-downloaded episodes and shows appear in your library for torrent streaming."
                />

                <Separator />

                <h3>
                    Torrent selection
                </h3>

                <Field.Switch
                    name="autoSelect"
                    label="Auto-select torrent"
                    help="Let Seanime find the best torrent automatically."
                />

                <Field.Select
                    name="preferredResolution"
                    label="Preferred resolution"
                    help="If auto-select is enabled, Seanime will try to find torrents with this resolution. 'Any' will prefer the highest resolution available."
                    options={[
                        { label: "Any", value: "-" },
                        { label: "480p", value: "480" },
                        { label: "720p", value: "720" },
                        { label: "1080p", value: "1080" },
                    ]}
                />

                {/*<Field.DirectorySelector*/}
                {/*    name="downloadDir"*/}
                {/*    label="Download directory"*/}
                {/*    leftIcon={<FcFolder />}*/}
                {/*    help="Directory"*/}
                {/*    shouldExist*/}
                {/*/>*/}

                {/*<Field.Switch*/}
                {/*    name="addToLibrary"*/}
                {/*    label="Add to library"*/}
                {/*    help="Keep completely downloaded files in corresponding library entries."*/}
                {/*/>*/}

                <Separator />

                <h3>Torrent client</h3>

                <p>
                    Seanime uses a built-in torrent client to download torrents.
                </p>


                <Field.Number
                    name="torrentClientPort"
                    label="Port"
                    formatOptions={{
                        useGrouping: false,
                    }}
                />
                <Field.Switch
                    name="disableIPv6"
                    label="Disable IPv6"
                />

                <Separator />

                <h3>
                    Streaming server
                </h3>

                <p>
                    Seanime will launch a separate server to stream torrents. You can configure the port and host it uses here.
                </p>

                <div className="flex items-center gap-3">
                    <Field.Number
                        name="streamingServerPort"
                        label="Port"
                        formatOptions={{
                            useGrouping: false,
                        }}
                    />

                    <Field.Text
                        name="streamingServerHost"
                        label="Host"
                    />
                </div>


                <div className="flex w-full items-center">
                    <SettingsSubmitButton isPending={isPending} />
                    <div className="flex flex-1"></div>
                    <Button leftIcon={<SiBittorrent />} intent="alert-subtle" onClick={() => dropTorrent()} disabled={droppingTorrent}>
                        Drop torrent
                    </Button>
                </div>
            </Form>
        </>
    )
}
