"use client"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status"
import { anilistUserMediaAtom } from "@/app/(main)/_hooks/anilist-user-media"
import { AutoDownloaderItems } from "@/app/(main)/auto-downloader/_components/items"
import { RuleForm } from "@/app/(main)/auto-downloader/_components/rule-form"
import { AutoDownloaderItem, AutoDownloaderRule } from "@/app/(main)/auto-downloader/_lib/autodownloader.types"
import { tabsListClass, tabsTriggerClass } from "@/components/shared/styling/classnames"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useBoolean } from "@/hooks/use-disclosure"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { InferType } from "prop-types"
import React from "react"
import { BiChevronRight, BiPlus } from "react-icons/bi"
import { FaSquareRss } from "react-icons/fa6"
import { toast } from "sonner"

const settingsSchema = defineSchema(({ z }) => z.object({
    interval: z.number().min(2),
    enabled: z.boolean(),
    downloadAutomatically: z.boolean(),
}))

export const dynamic = "force-static"

export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()
    const userMedia = useAtomValue(anilistUserMediaAtom)

    const createRuleModal = useBoolean(false)

    const { mutate: runAutoDownloader, isPending: isRunning } = useSeaMutation<null, void>({
        mutationKey: ["run-auto-downloader"],
        endpoint: SeaEndpoints.RUN_AUTO_DOWNLOADER,
        method: "post",
        onSuccess: async () => {
            toast.success("Auto downloader started")
            setTimeout(() => {
                qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            }, 1000)
        },
    })

    const { mutate: updateSettings, isPending } = useSeaMutation<null, InferType<typeof settingsSchema>>({
        mutationKey: ["auto-downloader-settings"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_SETTINGS,
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["status"] })
            await qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            toast.success("Settings updated")
        },
    })

    const { data, isLoading } = useSeaQuery<AutoDownloaderRule[] | null>({
        queryKey: ["auto-downloader-rules"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_RULES,
    })

    const { data: items, isLoading: itemsLoading } = useSeaQuery<AutoDownloaderItem[]>({
        queryKey: ["auto-downloader-items"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEMS,
    })

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
                                        <Rule
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
                        <AutoDownloaderItems items={items} isLoading={itemsLoading} />
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
                            }}
                            stackClass="space-y-6"
                        >
                            {(f) => (
                                <>
                                    <Field.Switch
                                        label="Enabled"
                                        name="enabled"
                                    />

                                    <Separator />

                                    <div
                                        className={cn(
                                            "space-y-3",
                                            !f.watch("enabled") && "pointer-events-none opacity-50",
                                        )}
                                    >
                                        <Field.Checkbox
                                            label="Download episodes immediately"
                                            name="downloadAutomatically"
                                            help="If disabled, torrents will be added to the queue"
                                        />
                                        <Field.Number
                                            label="Interval"
                                            help="How often to check for new episodes"
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
                <RuleForm type="create" onRuleCreatedOrDeleted={() => createRuleModal.off()} />
            </Modal>
        </div>
    )

}

type RuleProps = {
    rule: AutoDownloaderRule
    userMedia: BaseMediaFragment[] | undefined
}

function Rule(props: RuleProps) {

    const {
        rule,
        userMedia,
        ...rest
    } = props

    const modal = useBoolean(false)

    const media = React.useMemo(() => {
        return userMedia?.find(media => media.id === rule.mediaId)
    }, [(userMedia?.length || 0), rule])

    return (
        <>
            <div className="rounded-[--radius] bg-gray-900 hover:bg-gray-800 transition-colors">
                <div className="flex justify-between p-3 gap-2 items-center cursor-pointer" onClick={() => modal.on()}>

                    <div className="space-y-1 w-full">
                        <p
                            className={cn(
                                "font-medium text-base tracking-wide line-clamp-1",
                            )}
                        ><span className="text-gray-400 italic font-normal pr-1">Rule for</span> "{rule.comparisonTitle}"</p>
                        <p className="text-sm text-gray-400 line-clamp-1 flex space-x-2 items-center divide-x divide-[--border] [&>span]:pl-2">
                            <FaSquareRss
                                className={cn(
                                    "text-xl",
                                    rule.enabled ? "text-green-500" : "text-gray-500",
                                    (media?.status === "FINISHED" || !media) && "text-red-300",
                                )}
                            />
                            {!!rule.releaseGroups.length && <span>{rule.releaseGroups.join(", ")}</span>}
                            {!!rule.resolutions.length && <span>{rule.resolutions.join(", ")}</span>}
                            {!!rule.episodeType && <span>{getEpisodeTypeName(rule.episodeType)}</span>}
                            {!!media ? (
                                <>
                                    {media.status === "FINISHED" && <span className="text-red-300">This anime is no longer airing</span>}
                                </>
                            ) : (
                                <span className="text-red-300">This anime is not in your library</span>
                            )}
                        </p>
                    </div>

                    <div>
                        <IconButton intent="white-basic" icon={<BiChevronRight />} size="sm" />
                    </div>
                </div>
            </div>
            <Modal
                open={modal.active}
                onOpenChange={modal.off}
                title="Edit rule"
                contentClass="max-w-3xl"

            >
                <RuleForm type="edit" rule={rule} />
            </Modal>
        </>
    )
}

function getEpisodeTypeName(episodeType: AutoDownloaderRule["episodeType"]) {
    switch (episodeType) {
        case "recent":
            return "Recent releases"
        case "selected":
            return "Select episodes"
    }
}
