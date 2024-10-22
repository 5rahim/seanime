import { useGetDebridSettings, useSaveDebridSettings } from "@/api/hooks/debrid.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import React from "react"

const debridSettingsSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean().default(false),
    provider: z.string().default(""),
    apiKey: z.string().optional().default(""),
    fallbackToDebridStreamingView: z.boolean().default(false),
    includeDebridStreamInLibrary: z.boolean().default(false),
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

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4">

            <Form
                schema={debridSettingsSchema}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                            settings: {
                                ...settings,
                                ...data,
                                provider: data.provider === "-" ? "" : data.provider,
                            },
                        })
                    }
                }}
                defaultValues={{
                    enabled: settings?.enabled,
                    provider: settings?.provider || "-",
                    apiKey: settings?.apiKey,
                    fallbackToDebridStreamingView: settings?.fallbackToDebridStreamingView,
                    includeDebridStreamInLibrary: settings?.includeDebridStreamInLibrary,
                }}
                stackClass="space-y-6"
            >
                {(f) => (
                    <>
                        <Field.Switch
                            name="enabled"
                            label="Enable"
                        />

                        {(f.watch("enabled") && serverStatus?.settings?.autoDownloader?.enabled && !serverStatus?.settings?.autoDownloader?.useDebrid) && (
                            <Alert
                                intent="info-basic"
                                title="Auto Downloader not using Debrid"
                                description={<p>
                                    Auto Downloader is enabled but not using Debrid. Change the <SeaLink
                                    href="/auto-downloader"
                                    className="underline"
                                >Auto Downloader settings</SeaLink> to use your Debrid service.
                                </p>}
                            />
                        )}

                        <Field.Select
                            options={[
                                { label: "None", value: "-" },
                                { label: "TorBox", value: "torbox" },
                            ]}
                            name="provider"
                            label="Provider"
                        />

                        <Field.Text
                            name="apiKey"
                            label="API Key"
                        />

                        <Separator />

                        <h3>
                            Streaming integration
                        </h3>

                        <Field.Switch
                            name="fallbackToDebridStreamingView"
                            label="Default to Debrid streaming view"
                            help="If the anime is not downloaded, the Debrid streaming view will be shown by default."
                        />

                        <Field.Switch
                            name="includeDebridStreamInLibrary"
                            label="Include in library"
                            help="Make non-downloaded episodes and shows appear in your library for torrent streaming."
                        />


                        <SettingsSubmitButton isPending={isPending} />
                    </>
                )}
            </Form>

        </div>
    )
}
