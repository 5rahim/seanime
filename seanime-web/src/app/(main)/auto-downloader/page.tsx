"use client"
import { AutoDownloaderItems } from "@/app/(main)/auto-downloader/_containers/items"
import { RuleForm } from "@/app/(main)/auto-downloader/_containers/rule-form"
import { userMediaAtom } from "@/atoms/collection"
import { serverStatusAtom } from "@/atoms/server-status"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Divider } from "@/components/ui/divider"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { TabPanels } from "@/components/ui/tabs"
import { createTypesafeFormSchema, Field, TypesafeForm } from "@/components/ui/typesafe-form"
import { useBoolean } from "@/hooks/use-disclosure"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { AutoDownloaderItem, AutoDownloaderRule } from "@/lib/server/types"
import { BiChevronRight } from "@react-icons/all-files/bi/BiChevronRight"
import { BiPlus } from "@react-icons/all-files/bi/BiPlus"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { InferType } from "prop-types"
import React from "react"
import toast from "react-hot-toast"
import { FaSquareRss } from "react-icons/fa6"

const settingsSchema = createTypesafeFormSchema(({ z }) => z.object({
    interval: z.number().min(2),
    enabled: z.boolean(),
    downloadAutomatically: z.boolean(),
}))

export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()
    const userMedia = useAtomValue(userMediaAtom)

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
        refetchInterval: 10000,
        refetchIntervalInBackground: true,
    })

    return (
        <div className="space-y-4">

            <TabPanels
                navClassName="border-[--border]"
                tabClassName={cn(
                    "text-sm rounded-none border-b border-b-2 data-[selected=true]:text-white data-[selected=true]:border-brand-400",
                    "hover:bg-transparent dark:hover:bg-transparent hover:text-white",
                    "dark:border-transparent dark:hover:border-b-transparent dark:data-[selected=true]:border-brand-400 dark:data-[selected=true]:text-white",
                    "dark:data-[selected=true]:bg-[--highlight]",
                )}
            >
                <TabPanels.Nav>
                    <TabPanels.Tab>Rules</TabPanels.Tab>
                    <TabPanels.Tab>
                        Queue
                        {!!items?.length && (
                            <Badge className="ml-1 font-bold" intent="alert">
                                {items.length}
                            </Badge>
                        )}
                    </TabPanels.Tab>
                    <TabPanels.Tab>Settings</TabPanels.Tab>
                </TabPanels.Nav>
                <TabPanels.Container>

                    <TabPanels.Panel>
                        <div className="p-4">
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
                                            isLoading={isRunning}
                                        >
                                            Run
                                        </Button>
                                        <Button
                                            className="rounded-full"
                                            intent="success-subtle"
                                            leftIcon={<BiPlus />}
                                            onClick={() => {
                                                createRuleModal.on()
                                            }}
                                        >
                                            Add Rule
                                        </Button>
                                    </div>

                                    <ul className="text-base text-[--muted] list-disc pl-4">
                                        <li>The only provider currently supported is <em className="font-semibold">Nyaa.si</em></li>
                                        <li>Auto Downloader uses the <em className="font-semibold">qBittorrent</em> integration to download new files
                                        </li>
                                        <li><em className="font-semibold">Rules</em> are parameters that define which episodes and which files to
                                                                                     download for a specific anime
                                        </li>
                                        <li>The anime must already be present in your library</li>
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
                    </TabPanels.Panel>


                    <TabPanels.Panel>

                        <div className="p-4">
                            <AutoDownloaderItems items={items} isLoading={itemsLoading} />
                        </div>

                    </TabPanels.Panel>

                    <TabPanels.Panel>
                        <div className="p-4">
                            <TypesafeForm
                                schema={settingsSchema}
                                onSubmit={data => {
                                    updateSettings(data)
                                }}
                                defaultValues={{
                                    enabled: serverStatus?.settings?.autoDownloader?.enabled ?? false,
                                    interval: serverStatus?.settings?.autoDownloader?.interval ?? 10,
                                    downloadAutomatically: serverStatus?.settings?.autoDownloader?.downloadAutomatically ?? false,
                                }}
                            >
                                {(f) => (
                                    <>
                                        <Field.Switch
                                            label="Enabled"
                                            name="enabled"
                                        />

                                        <Divider />

                                        <div
                                            className={cn(
                                                "space-y-2",
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
                                                discrete
                                                size="sm"
                                                className="text-center w-20"
                                                min={2}
                                            />
                                        </div>

                                        <Field.Submit role="save" isLoading={isPending} />
                                    </>
                                )}
                            </TypesafeForm>
                        </div>
                    </TabPanels.Panel>

                </TabPanels.Container>
            </TabPanels>


            <Modal
                isOpen={createRuleModal.active}
                onClose={createRuleModal.off}
                title="Create a new rule"
                size="2xl"
                isClosable
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
            <div className="rounded-[--radius] bg-[--background-color] hover:bg-gray-800 transition-colors">
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
                isOpen={modal.active}
                onClose={modal.off}
                title="Edit rule"
                size="2xl"
                isClosable
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
