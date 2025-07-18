import { useGetDebridSettings, useSaveDebridSettings } from "@/api/hooks/debrid.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsIsDirty, SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
import { UseFormReturn } from "react-hook-form"

const debridSettingsSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean().default(false),
    provider: z.string().default(""),
    apiKey: z.string().optional().default(""),
    includeDebridStreamInLibrary: z.boolean().default(false),
    streamAutoSelect: z.boolean().default(false),
    streamPreferredResolution: z.string(),
}))

type DebridSettingsProps = {
    children?: React.ReactNode
}

export function DebridSettings(props: DebridSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const { data: settings, isLoading } = useGetDebridSettings()
    const { mutate, isPending } = useSaveDebridSettings()

    const formRef = React.useRef<UseFormReturn<any>>(null)

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4">

            <Form
                schema={debridSettingsSchema}
                mRef={formRef}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                                provider: data.provider === "-" ? "" : data.provider,
                                streamPreferredResolution: data.streamPreferredResolution === "-" ? "" : data.streamPreferredResolution,
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
                    enabled: settings?.enabled,
                    provider: settings?.provider || "-",
                    apiKey: settings?.apiKey,
                    includeDebridStreamInLibrary: settings?.includeDebridStreamInLibrary,
                    streamAutoSelect: settings?.streamAutoSelect ?? false,
                    streamPreferredResolution: settings?.streamPreferredResolution || "-",
                }}
                stackClass="space-y-4"
            >
                {(f) => (
                    <>
                        <SettingsIsDirty />
                        <SettingsCard>
                            <Field.Switch
                                side="right"
                                name="enabled"
                                label="Enable"
                            />
                            {(f.watch("enabled") && serverStatus?.settings?.autoDownloader?.enabled && !serverStatus?.settings?.autoDownloader?.useDebrid) && (
                                <Alert
                                    intent="info"
                                    title="Auto Downloader not using Debrid"
                                    description={<p>
                                        Auto Downloader is enabled but not using Debrid. Change the <SeaLink
                                        href="/auto-downloader"
                                        className="underline"
                                    >Auto Downloader settings</SeaLink> to use your Debrid service.
                                    </p>}
                                />
                            )}
                        </SettingsCard>


                        <SettingsCard>
                            <Field.Select
                                options={[
                                    { label: "None", value: "-" },
                                    { label: "TorBox", value: "torbox" },
                                    { label: "Real-Debrid", value: "realdebrid" },
                                ]}
                                name="provider"
                                label="Provider"
                            />

                            <Field.Text
                                name="apiKey"
                                label="API Key"
                            />
                        </SettingsCard>

                        <h3>
                            Debrid Streaming
                        </h3>

                        <SettingsCard title="My library">
                            <Field.Switch
                                side="right"
                                name="includeDebridStreamInLibrary"
                                label="Include in library"
                                help="Add non-downloaded shows that are in your currently watching list to 'My library' for streaming"
                            />
                        </SettingsCard>

                        <SettingsCard title="Auto-select">
                            <Field.Switch
                                side="right"
                                name="streamAutoSelect"
                                label="Enable"
                                help="Let Seanime find the best torrent automatically, based on cache and resolution."
                            />

                            {/*{f.watch("streamAutoSelect") && f.watch("provider") === "torbox" && (*/}
                            {/*    <Alert*/}
                            {/*        intent="warning-basic"*/}
                            {/*        title="Auto-select with TorBox"*/}
                            {/*        description={<p>*/}
                            {/*            Avoid using auto-select if you have a limited amount of downloads on your Debrid service.*/}
                            {/*        </p>}*/}
                            {/*    />*/}
                            {/*)}*/}

                            <Field.Select
                                name="streamPreferredResolution"
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


                        <SettingsSubmitButton isPending={isPending} />
                    </>
                )}
            </Form>

        </div>
    )
}
