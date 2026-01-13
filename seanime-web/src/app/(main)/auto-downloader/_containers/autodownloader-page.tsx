import {
    useDeleteAutoDownloaderRule,
    useGetAutoDownloaderItems,
    useGetAutoDownloaderRules,
    useRunAutoDownloader,
} from "@/api/hooks/auto_downloader.hooks"
import { useSaveAutoDownloaderSettings } from "@/api/hooks/settings.hooks"
import { __anilist_userAnimeMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AutoDownloaderRuleItem } from "@/app/(main)/auto-downloader/_components/autodownloader-rule-item"
import { AutoDownloaderBatchRuleForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-batch-rule-form"
import { AutoDownloaderItemList } from "@/app/(main)/auto-downloader/_containers/autodownloader-item-list"
import { AutoDownloaderRuleForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useBoolean } from "@/hooks/use-disclosure"
import { useAtomValue } from "jotai/react"
import React from "react"
import { BiDotsVerticalRounded } from "react-icons/bi"
import { FaSquareRss } from "react-icons/fa6"
import { LuTrash } from "react-icons/lu"
import { MdOutlineAdd } from "react-icons/md"
import { toast } from "sonner"

const tabContentClass = cn(
    "space-y-4 animate-in fade-in-0 duration-300",
)

const settingsSchema = defineSchema(({ z }) => z.object({
    interval: z.number().transform(n => {
        if (n < 15) {
            toast.info("Interval changed to be at least 15 minutes")
            return 15
        }
        return n
    }),
    enabled: z.boolean(),
    downloadAutomatically: z.boolean(),
    enableEnhancedQueries: z.boolean(),
    enableSeasonCheck: z.boolean(),
    useDebrid: z.boolean(),
}))

export function AutoDownloaderPage() {
    const serverStatus = useServerStatus()
    const userMedia = useAtomValue(__anilist_userAnimeMediaAtom)

    const createRuleModal = useBoolean(false)
    const createBatchRuleModal = useBoolean(false)

    const { mutate: runAutoDownloader, isPending: isRunning } = useRunAutoDownloader()

    const { mutate: updateSettings, isPending } = useSaveAutoDownloaderSettings()

    const { data, isLoading } = useGetAutoDownloaderRules()

    const { data: items, isLoading: itemsLoading } = useGetAutoDownloaderItems()

    const { mutate: deleteNoLongerAiring, isPending: deletingRule } = useDeleteAutoDownloaderRule(-1)

    const confirmDeleteNoLongerAiring = useConfirmationDialog({
        title: "Remove no longer airing media",
        description: "This action will remove all rules that no longer have media airing (finished). Are you sure you want to continue?",
        onConfirm: () => {
            deleteNoLongerAiring()
        },
    })

    return (
        <div className="space-y-4">
            <ConfirmationDialog {...confirmDeleteNoLongerAiring} />

            <Tabs
                defaultValue="rules"
                triggerClass={"text-base px-6 h-auto py-2 rounded-[--radius-md] w-fit md:w-full border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white"}
                listClass={"w-full flex flex-wrap md:flex-nowrap h-fit"}
            >
                <TabsList className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-xl">
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
                <TabsContent value="rules" className={tabContentClass}>
                    <div className="pt-4">
                        {isLoading && <LoadingSpinner />}
                        {!isLoading && (
                            <div className="space-y-4">

                                <Card className="p-4 space-y-4">
                                    <ul className="text-base text-[--muted]">
                                        <li><em className="font-semibold">Rules</em> allow you to programmatically download new episodes based on the
                                                                                     parameters you set.
                                        </li>
                                    </ul>

                                    <div className="w-full flex items-center gap-2">
                                        <DropdownMenu
                                            trigger={<Button
                                                className="rounded-full"
                                                intent="white-subtle"
                                                leftIcon={<MdOutlineAdd className="text-lg" />}

                                            >
                                                New Rule
                                            </Button>}
                                        >
                                            <DropdownMenuItem onClick={createRuleModal.on}>
                                                Single rule
                                            </DropdownMenuItem>
                                            <DropdownMenuItem onClick={createBatchRuleModal.on}>
                                                Multiple rules
                                            </DropdownMenuItem>
                                        </DropdownMenu>
                                        <div className="flex flex-1"></div>
                                        <Button
                                            className=""
                                            intent="gray-basic"
                                            leftIcon={<FaSquareRss />}
                                            onClick={() => {
                                                runAutoDownloader()
                                            }}
                                            loading={isRunning}
                                            disabled={!serverStatus?.settings?.autoDownloader?.enabled}
                                        >
                                            Check RSS feed
                                        </Button>
                                        <DropdownMenu
                                            trigger={<IconButton
                                                className=""
                                                intent="gray-basic"
                                                icon={<BiDotsVerticalRounded className="text-lg" />}
                                            />}
                                        >
                                            <DropdownMenuItem
                                                onClick={confirmDeleteNoLongerAiring.open}
                                                className="text-[--red]"
                                                disabled={deletingRule}
                                            >
                                                <LuTrash /> Remove no longer airing media
                                            </DropdownMenuItem>
                                        </DropdownMenu>
                                    </div>

                                    {(!data?.length) && <div className="p-4 text-[--muted] text-center">No rules</div>}
                                    {(!!data?.length) && <div className="space-y-2">
                                        {data?.map(rule => (
                                            <AutoDownloaderRuleItem
                                                key={rule.dbId}
                                                rule={rule}
                                                userMedia={userMedia}
                                            />
                                        ))}
                                    </div>}
                                </Card>
                            </div>
                        )}
                    </div>
                </TabsContent>


                <TabsContent value="queue" className={tabContentClass}>

                    <div className="pt-4">
                        <AutoDownloaderItemList items={items} isLoading={itemsLoading} />
                    </div>

                </TabsContent>

                <TabsContent value="settings" className={tabContentClass}>
                    <div className="pt-4">
                        <Form
                            schema={settingsSchema}
                            onSubmit={data => {
                                updateSettings(data)
                            }}
                            defaultValues={{
                                enabled: serverStatus?.settings?.autoDownloader?.enabled ?? false,
                                interval: serverStatus?.settings?.autoDownloader?.interval || 15,
                                downloadAutomatically: serverStatus?.settings?.autoDownloader?.downloadAutomatically ?? false,
                                enableEnhancedQueries: serverStatus?.settings?.autoDownloader?.enableEnhancedQueries ?? false,
                                enableSeasonCheck: serverStatus?.settings?.autoDownloader?.enableSeasonCheck ?? false,
                                useDebrid: serverStatus?.settings?.autoDownloader?.useDebrid ?? false,
                            }}
                            stackClass="space-y-4"
                        >
                            {(f) => (
                                <>
                                    <SettingsCard>
                                        <Field.Switch
                                            side="right"
                                            label="Enabled"
                                            name="enabled"
                                        />

                                        <Field.Switch
                                            side="right"
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
                                    </SettingsCard>

                                    <SettingsCard
                                        className={cn(
                                            !f.watch("enabled") && "pointer-events-none opacity-50",
                                        )}
                                    >
                                        <Field.Switch
                                            side="right"
                                            label="Use enhanced queries"
                                            name="enableEnhancedQueries"
                                            help="Seanime will use multiple custom queries instead of a single one. Enable this if you notice some missing downloads."
                                        />
                                        <Field.Switch
                                            side="right"
                                            label="Verify season"
                                            name="enableSeasonCheck"
                                            help="Seanime will perform an additional check to ensure the season number is correct. This is not needed in most cases."
                                        />
                                        <Field.Switch
                                            side="right"
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
                                            min={15}
                                        />
                                    </SettingsCard>

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
                contentClass="max-w-4xl"
            >
                <AutoDownloaderRuleForm type="create" onRuleCreatedOrDeleted={() => createRuleModal.off()} />
            </Modal>


            <Drawer
                open={createBatchRuleModal.active}
                onOpenChange={createBatchRuleModal.off}
                title="Create new rules"
                size="xl"
            >
                <p className="text-[--muted] py-4">
                    Create multiple rules at once. Each rule will be created with the same parameters, except for the destination folder.
                    By default, the episode type will be "Recent releases".
                </p>
                <AutoDownloaderBatchRuleForm onRuleCreated={() => createBatchRuleModal.off()} />
            </Drawer>
        </div>
    )

}


