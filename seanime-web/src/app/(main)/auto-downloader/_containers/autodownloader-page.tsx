import { useGetAutoDownloaderItems, useGetAutoDownloaderRules, useRunAutoDownloader } from "@/api/hooks/auto_downloader.hooks"
import { useSaveAutoDownloaderSettings } from "@/api/hooks/settings.hooks"
import { __anilist_userMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AutoDownloaderRuleItem } from "@/app/(main)/auto-downloader/_components/autodownloader-rule-item"
import { AutoDownloaderItemList } from "@/app/(main)/auto-downloader/_containers/autodownloader-item-list"
import { AutoDownloaderRuleForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import { tabsListClass, tabsTriggerClass } from "@/components/shared/classnames"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useBoolean } from "@/hooks/use-disclosure"
import { useAtomValue } from "jotai/react"
import React from "react"
import { BiPlus } from "react-icons/bi"
import { FaSquareRss } from "react-icons/fa6"

const settingsSchema = defineSchema(({ z }) => z.object({
    interval: z.number().min(2),
    enabled: z.boolean(),
    downloadAutomatically: z.boolean(),
    enableEnhancedQueries: z.boolean(),
    enableSeasonCheck: z.boolean(),
    useDebrid: z.boolean(),
}))

export function AutoDownloaderPage() {
    const serverStatus = useServerStatus()
    const userMedia = useAtomValue(__anilist_userMediaAtom)

    const createRuleModal = useBoolean(false)

    const { mutate: runAutoDownloader, isPending: isRunning } = useRunAutoDownloader()

    const { mutate: updateSettings, isPending } = useSaveAutoDownloaderSettings()

    const { data, isLoading } = useGetAutoDownloaderRules()

    const { data: items, isLoading: itemsLoading } = useGetAutoDownloaderItems()

    return (
        <div className="space-y-4">

            <Tabs
                defaultValue="rules"
                triggerClass={tabsTriggerClass}
                listClass={tabsListClass}
            >
                <TabsList>
                    <TabsTrigger value="rules">Rules</TabsTrigger>
                    <TabsTrigger value="queue">
                        Queue
                        {!!items?.length && (
                            <Badge className="ml-1 font-bold" intent="alert">
                                {items.length}
                            </Badge>
                        )}
                    </TabsTrigger>
                    <TabsTrigger value="settings">Settings</TabsTrigger>
                </TabsList>
                <TabsContent value="rules">
                    <div className="pt-4">
                        {isLoading && <LoadingSpinner />}
                        {!isLoading && (
                            <div className="space-y-4">
                                <div className="w-full flex justify-between items-center gap-2">
                                    <Button
                                        className="rounded-full"
                                        intent="primary-subtle"
                                        leftIcon={<FaSquareRss />}
                                        onClick={() => {
                                            runAutoDownloader()
                                        }}
                                        loading={isRunning}
                                        disabled={!serverStatus?.settings?.autoDownloader?.enabled}
                                    >
                                        Check RSS feed
                                    </Button>
                                    <Button
                                        className="rounded-full"
                                        intent="success-subtle"
                                        leftIcon={<BiPlus />}
                                        onClick={() => {
                                            createRuleModal.on()
                                        }}
                                    >
                                        New Rule
                                    </Button>
                                </div>

                                <ul className="text-base text-[--muted]">
                                    <li><em className="font-semibold">Rules</em> allow you to programmatically download new episodes based on the
                                                                                 parameters you set.
                                    </li>
                                </ul>

                                {(!data?.length) && <div className="p-4 text-[--muted] text-center">No rules</div>}
                                {(!!data?.length) && <div className="space-y-4">
                                    {data?.map(rule => (
                                        <AutoDownloaderRuleItem
                                            key={rule.dbId}
                                            rule={rule}
                                            userMedia={userMedia}
                                        />
                                    ))}
                                </div>}
                            </div>
                        )}
                    </div>
                </TabsContent>


                <TabsContent value="queue">

                    <div className="pt-4">
                        <AutoDownloaderItemList items={items} isLoading={itemsLoading} />
                    </div>

                </TabsContent>

                <TabsContent value="settings">
                    <div className="pt-4">
                        <Form
                            schema={settingsSchema}
                            onSubmit={data => {
                                updateSettings(data)
                            }}
                            defaultValues={{
                                enabled: serverStatus?.settings?.autoDownloader?.enabled ?? false,
                                interval: serverStatus?.settings?.autoDownloader?.interval || 10,
                                downloadAutomatically: serverStatus?.settings?.autoDownloader?.downloadAutomatically ?? false,
                                enableEnhancedQueries: serverStatus?.settings?.autoDownloader?.enableEnhancedQueries ?? false,
                                enableSeasonCheck: serverStatus?.settings?.autoDownloader?.enableSeasonCheck ?? false,
                                useDebrid: serverStatus?.settings?.autoDownloader?.useDebrid ?? false,
                            }}
                            stackClass="space-y-6"
                        >
                            {(f) => (
                                <>
                                    <Field.Switch
                                        label="Enabled"
                                        name="enabled"
                                    />

                                    <Field.Switch
                                        label="Use Debrid service"
                                        name="useDebrid"
                                    />

                                    {f.watch("useDebrid") && !(serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider) && (
                                        <Alert
                                            intent="alert"
                                            title="Auto Downloader deactivated"
                                            description="Debrid service is not enabled or configured. Please enable it in the settings."
                                        />
                                    )}

                                    <Separator />

                                    <div
                                        className={cn(
                                            "space-y-3",
                                            !f.watch("enabled") && "pointer-events-none opacity-50",
                                        )}
                                    >
                                        <Field.Switch
                                            label="Use enhanced queries"
                                            name="enableEnhancedQueries"
                                            help="Seanime will use multiple custom queries instead of a single one. Enable this if you notice some missing downloads."
                                        />
                                        <Field.Switch
                                            label="Verify season"
                                            name="enableSeasonCheck"
                                            help="Seanime will perform an additional check to ensure the season number is correct. This is not needed in most cases."
                                        />
                                        <Field.Switch
                                            label="Download episodes immediately"
                                            name="downloadAutomatically"
                                            help="If disabled, torrents will be added to the queue."
                                        />
                                        <Field.Number
                                            label="Interval"
                                            help="How often to check for new episodes."
                                            name="interval"
                                            leftAddon="Every"
                                            rightAddon="minutes"
                                            size="sm"
                                            className="text-center w-20"
                                            min={2}
                                        />
                                    </div>

                                    <Field.Submit role="save" loading={isPending}>Save</Field.Submit>
                                </>
                            )}
                        </Form>
                    </div>
                </TabsContent>

            </Tabs>


            <Modal
                open={createRuleModal.active}
                onOpenChange={createRuleModal.off}
                title="Create a new rule"
                contentClass="max-w-3xl"

            >
                <AutoDownloaderRuleForm type="create" onRuleCreatedOrDeleted={() => createRuleModal.off()} />
            </Modal>
        </div>
    )

}


