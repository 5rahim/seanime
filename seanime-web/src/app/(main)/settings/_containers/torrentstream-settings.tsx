import { Models_TorrentstreamSettings } from "@/api/generated/types"
import { useSaveTorrentstreamSettings, useTorrentstreamDropTorrent } from "@/api/hooks/torrentstream.hooks"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsIsDirty, SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { defineSchema, Field, Form } from "@/components/ui/form"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { FcFolder } from "react-icons/fc"
import { SiBittorrent } from "react-icons/si"

const torrentstreamSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    downloadDir: z.string(),
    autoSelect: z.boolean(),
    disableIPv6: z.boolean(),
    addToLibrary: z.boolean(),
    // streamingServerPort: z.number(),
    // streamingServerHost: z.string(),
    torrentClientHost: z.string().optional().default(""),
    torrentClientPort: z.number(),
    preferredResolution: z.string(),
    includeInLibrary: z.boolean(),
    streamUrlAddress: z.string().optional().default(""),
    slowSeeding: z.boolean().optional().default(false),
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

    const formRef = React.useRef<UseFormReturn<any>>(null)

    if (!settings) return null

    return (
        <>
            <Form
                schema={torrentstreamSchema}
                mRef={formRef}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                                preferredResolution: data.preferredResolution === "-" ? "" : data.preferredResolution,
                            },
                        },
                            {
                                onSuccess: () => {
                                    formRef.current?.reset(formRef.current.getValues())
                                },
                            },
                        )
                    }
                }}
                defaultValues={{
                    enabled: settings.enabled,
                    autoSelect: settings.autoSelect,
                    downloadDir: settings.downloadDir || "",
                    disableIPv6: settings.disableIPV6,
                    addToLibrary: settings.addToLibrary,
                    // streamingServerPort: settings.streamingServerPort,
                    // streamingServerHost: settings.streamingServerHost || "",
                    torrentClientHost: settings.torrentClientHost || "",
                    torrentClientPort: settings.torrentClientPort,
                    preferredResolution: settings.preferredResolution || "-",
                    includeInLibrary: settings.includeInLibrary,
                    streamUrlAddress: settings.streamUrlAddress || "",
                    slowSeeding: settings.slowSeeding,
                }}
                stackClass="space-y-4"
            >
                <SettingsIsDirty />
                <SettingsCard>
                    <Field.Switch
                        side="right"
                        name="enabled"
                        label="Enable"
                    />
                </SettingsCard>

                <SettingsCard title="Integration">
                    <Field.Switch
                        side="right"
                        name="includeInLibrary"
                        label="Include in library"
                        help="Shows that are currently being watched but haven't been downloaded will default to the streaming view and appear in your library."
                    />
                </SettingsCard>


                <SettingsCard title="Auto-select">
                    <Field.Switch
                        side="right"
                        name="autoSelect"
                        label="Enable"
                        help="Let Seanime find the best torrent automatically."
                    />

                    <Field.Select
                        name="preferredResolution"
                        label="Preferred resolution"
                        help="If auto-select is enabled, Seanime will try to find torrents with this resolution."
                        options={[
                            { label: "Highest", value: "-" },
                            { label: "480p", value: "480" },
                            { label: "720p", value: "720" },
                            { label: "1080p", value: "1080" },
                        ]}
                    />
                </SettingsCard>


                {/*<Field.Switch
                 side="right"*/}
                {/*    name="addToLibrary"*/}
                {/*    label="Add to library"*/}
                {/*    help="Keep completely downloaded files in corresponding library entries."*/}
                {/*/>*/}

                {/* <SettingsCard title="Torrent Client" description="Seanime uses a built-in torrent client to download torrents.">

                 </SettingsCard> */}

                <Accordion
                    type="single"
                    collapsible
                    className="border rounded-[--radius-md]"
                    triggerClass="dark:bg-[--paper]"
                    contentClass="!pt-2 dark:bg-[--paper]"
                >
                    <AccordionItem value="more">
                        <AccordionTrigger className="bg-gray-900 rounded-[--radius-md]">
                            Torrent Client
                        </AccordionTrigger>
                        <AccordionContent className="space-y-4">
                            <div className="flex items-center gap-3">

                                <Field.Text
                                    name="torrentClientHost"
                                    label="Host"
                                    help="Leave empty for default. The host to listen for new uTP and TCP BitTorrent connections."
                                />

                                <Field.Number
                                    name="torrentClientPort"
                                    label="Port"
                                    formatOptions={{
                                        useGrouping: false,
                                    }}
                                    help="Leave empty for default. Default is 43213."
                                />

                            </div>

                            <Field.Switch
                                side="right"
                                name="disableIPv6"
                                label="Disable IPv6"
                            />

                            <Field.Switch
                                side="right"
                                name="slowSeeding"
                                label="Slow seeding"
                                moreHelp="This can help avoid issues with your network."
                            />
                        </AccordionContent>
                    </AccordionItem>
                </Accordion>

                <Accordion
                    type="single"
                    collapsible
                    className="border rounded-[--radius-md]"
                    triggerClass="dark:bg-[--paper]"
                    contentClass="!pt-2 dark:bg-[--paper]"
                >
                    <AccordionItem value="more">
                        <AccordionTrigger className="bg-gray-900 rounded-[--radius-md]">
                            Advanced
                        </AccordionTrigger>
                        <AccordionContent className="pt-6 space-y-4">
                            <Field.Text
                                name="streamUrlAddress"
                                label="Stream URL address"
                                placeholder="e.g. 0.0.0.0:43211"
                                help="Modify the stream URL formatting. Leave empty for default."
                            />

                            <Field.DirectorySelector
                                name="downloadDir"
                                label="Cache directory"
                                leftIcon={<FcFolder />}
                                help="Where the torrents will be downloaded to while streaming. Leave empty to use the default cache directory."
                                shouldExist
                            />
                        </AccordionContent>
                    </AccordionItem>
                </Accordion>


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
