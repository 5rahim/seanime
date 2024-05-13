import { Models_TorrentstreamSettings } from "@/api/generated/types"
import { useSaveTorrentstreamSettings } from "@/api/hooks/torrentstream.hooks"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"

const torrentstreamSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    downloadDir: z.string(),
    autoSelect: z.boolean(),
    disableIPv6: z.boolean(),
    addToLibrary: z.boolean(),
    streamingServerPort: z.number(),
    streamingServerHost: z.string(),
    torrentClientPort: z.number(),
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
                }}
                stackClass="space-y-6"
            >
                <Field.Switch
                    name="enabled"
                    label="Enable"
                    help="Enable torrent streaming."
                />
                <Field.Switch
                    name="autoSelect"
                    label="Auto select torrents"
                />

                {/*<Field.DirectorySelector*/}
                {/*    name="downloadDir"*/}
                {/*    label="Download directory"*/}
                {/*    leftIcon={<FcFolder />}*/}
                {/*    help="Directory"*/}
                {/*    shouldExist*/}
                {/*/>*/}

                <Field.Switch
                    name="addToLibrary"
                    label="Add to library"
                    help="Keep completely downloaded files in corresponding library entries."
                />

                <Separator />

                <h4>Torrent client</h4>

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

                <h4>
                    Streaming server
                </h4>

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
