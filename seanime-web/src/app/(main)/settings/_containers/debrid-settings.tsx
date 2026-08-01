import { Models_DummyDebridSettings } from "@/api/generated/types"
import { useGetDebridSettings, useGetDummyDebridSettings, useSaveDebridSettings, useSaveDummyDebridSettings } from "@/api/hooks/debrid.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets.ts"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AutoSelectProfileButton } from "@/app/(main)/settings/_components/autoselect-profile-form"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsIsDirty, SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { WSEvents } from "@/lib/server/ws-events.ts"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { HiOutlineServerStack } from "react-icons/hi2"
import { LuCirclePlay } from "react-icons/lu"
import { toast } from "sonner"

const debridSettingsSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean().default(false),
    provider: z.string().default(""),
    apiKey: z.string().optional().default(""),
    includeDebridStreamInLibrary: z.boolean().default(false),
    streamAutoSelect: z.boolean().default(false),
    streamPreferredResolution: z.string(),
}))

const dummyDebridSettingsSchema = defineSchema(({ z }) => z.object({
    enabled: z.boolean().default(true),
    profileName: z.string().default(""),
    fallbackFilePath: z.string().default(""),
    filesJson: z.string().superRefine((value, ctx) => {
        try {
            const parsed = JSON.parse(value || "[]")
            if (!Array.isArray(parsed)) {
                ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Expected a JSON array" })
            }
        }
        catch {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Invalid JSON" })
        }
    }),
    cached: z.boolean().default(true),
    readyDelayMs: z.number().min(0).default(0),
    progressIntervalMs: z.number().min(0).default(0),
    firstByteDelayMs: z.number().min(0).default(0),
    bandwidthBytesPerSecond: z.number().min(0).default(0),
    chunkSize: z.number().min(0).default(0),
    jitterMs: z.number().min(0).default(0),
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
    const { data: settings, isLoading, refetch } = useGetDebridSettings()
    const { mutate, isPending } = useSaveDebridSettings()
    const dummyDebridEnabled = !!serverStatus?.featureFlags?.dummyDebrid
    const [selectedProvider, setSelectedProvider] = React.useState(settings?.provider || "-")
    const providerOptions = React.useMemo(() => [
        { label: "None", value: "-" },
        { label: "TorBox", value: "torbox" },
        { label: "Real-Debrid", value: "realdebrid" },
        { label: "AllDebrid", value: "alldebrid" },
        { label: "Premiumize", value: "premiumize" },
        ...(dummyDebridEnabled ? [{ label: "Dummy Debrid", value: "dummy" }] : []),
    ], [dummyDebridEnabled])

    useWebsocketMessageListener({
        type: WSEvents.SETTINGS_CHANGED,
        onMessage: () => {
            refetch()
        },
    })

    const formRef = React.useRef<UseFormReturn<any>>(null)

    React.useEffect(() => {
        setSelectedProvider(settings?.provider || "-")
    }, [settings?.provider])

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4">

            <SettingsPageHeader
                title="Debrid Service"
                description="Configure your Debrid service integration"
                icon={HiOutlineServerStack}
            />

            <Form
                key={settings?.updatedAt ?? "debrid-settings"}
                schema={debridSettingsSchema}
                mRef={formRef}
                onSubmit={data => {
                    if (settings) {
                        mutate({
                                settings: {
                                    ...settings,
                                    ...data,
                                    provider: data.provider === "-" ? "" : data.provider,
                                    apiKey: data.provider === "dummy" ? "" : data.apiKey,
                                    streamPreferredResolution: data.streamPreferredResolution === "-" ? "" : data.streamPreferredResolution,
                                },
                            },
                            {
                                onSuccess: () => {
                                    formRef.current?.reset(formRef.current.getValues())
                                    toast.success("Settings saved")
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
                stackClass="space-y-8"
            >
                {(f) => (
                    <>
                        <ProviderWatch provider={f.watch("provider")} onChange={setSelectedProvider} />
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


                        <SettingsCard title="Provider">
                            <Field.Select
                                options={providerOptions}
                                name="provider"
                                label="Provider"
                            />

                            {f.watch("provider") !== "dummy" && (
                                <Field.Text
                                    name="apiKey"
                                    label="API Key"
                                    type="password"
                                />
                            )}
                        </SettingsCard>

                        <SettingsPageHeader
                            title="Debrid Streaming"
                            description="Configure how shows are streaming from your Debrid service"
                            icon={LuCirclePlay}
                        />

                        <SettingsCard title="Home Screen">
                            <Field.Switch
                                side="right"
                                name="includeDebridStreamInLibrary"
                                label="Include streaming in anime lists"
                                help="Show currently watching streaming titles in your anime lists."
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

                            <div className="pt-2">
                                <AutoSelectProfileButton />
                            </div>
                        </SettingsCard>


                        <SettingsSubmitButton isPending={isPending} />
                    </>
                )}
            </Form>

            {dummyDebridEnabled && selectedProvider === "dummy" && <DummyDebridProfileEditor />}

        </div>
    )
}

function ProviderWatch(props: { provider: string, onChange: (provider: string) => void }) {
    React.useEffect(() => {
        props.onChange(props.provider)
    }, [props.provider, props.onChange])

    return null
}

function DummyDebridProfileEditor() {
    const { data: settings, isLoading } = useGetDummyDebridSettings(true)
    const { mutate, isPending } = useSaveDummyDebridSettings()
    const formRef = React.useRef<UseFormReturn<any>>(null)

    if (isLoading) return <LoadingSpinner />
    if (!settings) return null

    return (
        <Form
            key={settings.updatedAt ?? "dummy-debrid-settings"}
            schema={dummyDebridSettingsSchema}
            mRef={formRef}
            onSubmit={data => {
                const { filesJson, ...values } = data
                const nextSettings: Models_DummyDebridSettings = {
                    ...settings,
                    ...values,
                    files: JSON.parse(filesJson || "[]") as Models_DummyDebridSettings["files"],
                }

                mutate({ settings: nextSettings }, {
                    onSuccess: () => {
                        formRef.current?.reset(formRef.current.getValues())
                        toast.success("Dummy profile saved")
                    },
                })
            }}
            defaultValues={dummyDebridDefaultValues(settings)}
            stackClass="space-y-4"
        >
            {() => (
                <>
                    <SettingsIsDirty />
                    <SettingsCard title="Dummy Debrid Profile">
                        <Field.Switch
                            side="right"
                            name="enabled"
                            label="Enable profile"
                        />
                        <Field.Text
                            name="profileName"
                            label="Profile name"
                        />
                        <Field.Text
                            name="fallbackFilePath"
                            label="Fallback MKV path"
                        />
                        <Field.Switch
                            side="right"
                            name="cached"
                            label="Cache available"
                        />
                        <Field.Textarea
                            name="filesJson"
                            label="Files"
                            className="min-h-[220px] font-mono text-sm"
                        />
                    </SettingsCard>

                    <SettingsCard title="Dummy Network">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <Field.Number
                                name="readyDelayMs"
                                label="Ready delay (ms)"
                                min={0}
                                step={100}
                            />
                            <Field.Number
                                name="progressIntervalMs"
                                label="Progress interval (ms)"
                                min={0}
                                step={50}
                            />
                            <Field.Number
                                name="firstByteDelayMs"
                                label="First byte delay (ms)"
                                min={0}
                                step={50}
                            />
                            <Field.Number
                                name="bandwidthBytesPerSecond"
                                label="Bandwidth (B/s)"
                                min={0}
                                step={1024}
                            />
                            <Field.Number
                                name="chunkSize"
                                label="Chunk size (bytes)"
                                min={0}
                                step={1024}
                            />
                            <Field.Number
                                name="jitterMs"
                                label="Jitter (ms)"
                                min={0}
                                step={10}
                            />
                        </div>
                    </SettingsCard>

                    <SettingsSubmitButton isPending={isPending} />
                </>
            )}
        </Form>
    )
}

function dummyDebridDefaultValues(settings: Models_DummyDebridSettings) {
    return {
        enabled: settings.enabled,
        profileName: settings.profileName,
        fallbackFilePath: settings.fallbackFilePath,
        filesJson: JSON.stringify(settings.files ?? [], null, 2),
        cached: settings.cached,
        readyDelayMs: settings.readyDelayMs,
        progressIntervalMs: settings.progressIntervalMs,
        firstByteDelayMs: settings.firstByteDelayMs,
        bandwidthBytesPerSecond: settings.bandwidthBytesPerSecond,
        chunkSize: settings.chunkSize,
        jitterMs: settings.jitterMs,
    }
}
