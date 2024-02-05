"use client"
import { ListSyncDiffs } from "@/app/(main)/list-sync/_containers/list-sync-diffs"
import { serverStatusAtom } from "@/atoms/server-status"
import { BetaBadge } from "@/components/application/beta-badge"
import { LuffyError } from "@/components/shared/luffy-error"
import { cn } from "@/components/ui/core"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { TabPanels } from "@/components/ui/tabs"
import { createTypesafeFormSchema, Field, TypesafeForm } from "@/components/ui/typesafe-form"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { ListSyncAnimeDiff, ListSyncOrigin } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { InferType } from "prop-types"
import React from "react"
import toast from "react-hot-toast"

const settingsSchema = createTypesafeFormSchema(({ z, presets }) => z.object({
    automatic: presets.checkbox,
    origin: z.string().min(1),
}))

export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()


    const { mutate: updateSettings, isPending } = useSeaMutation<null, InferType<typeof settingsSchema>>({
        mutationKey: ["list-sync-settings"],
        endpoint: SeaEndpoints.LIST_SYNC_SETTINGS,
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["status"] })
            await qc.refetchQueries({ queryKey: ["list-sync-anime-diffs"] })
            toast.success("Settings updated")
        },
    })

    const { mutate: clearCache, isPending: isDeletingCache } = useSeaMutation({
        endpoint: SeaEndpoints.LIST_SYNC_CACHE,
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["list-sync-anime-diffs"] })
            toast.success("List refreshed")
        },
    })

    const { data: animeDiffs, isLoading } = useSeaQuery<ListSyncAnimeDiff[] | string>({
        queryKey: ["list-sync-anime-diffs"],
        endpoint: SeaEndpoints.LIST_SYNC_ANIME_DIFFS,
    })

    function handleClearCache() {
        clearCache()
    }


    return (
        <div className="p-12 space-y-4">
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>List Sync <BetaBadge /></h2>
                    <p className="text-[--muted]">Sync your anime lists between different providers.</p>
                </div>
            </div>

            <div className="border border-[--border] rounded-[--radius] bg-[--paper] text-lg space-y-2">
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
                        <TabPanels.Tab>Lists</TabPanels.Tab>
                        <TabPanels.Tab>Settings</TabPanels.Tab>
                    </TabPanels.Nav>
                    <TabPanels.Container>
                        <TabPanels.Panel>
                            {!isLoading && <div className="p-4">
                                {typeof animeDiffs !== "string" &&
                                    <ListSyncDiffs diffs={animeDiffs ?? []} onClearCache={handleClearCache} isDeletingCache={isDeletingCache} />}
                                {typeof animeDiffs === "string" && <LuffyError>{animeDiffs}</LuffyError>}
                            </div>}
                            {isLoading && <LoadingSpinner />}
                        </TabPanels.Panel>
                        <TabPanels.Panel>
                            <div className="p-4">
                                <TypesafeForm
                                    schema={settingsSchema}
                                    onSubmit={data => {
                                        updateSettings(data)
                                    }}
                                    defaultValues={{
                                        automatic: serverStatus?.settings?.listSync?.automatic ?? false,
                                        origin: serverStatus?.settings?.listSync?.origin ?? "",
                                    }}
                                >
                                    <Field.RadioGroup
                                        label="Source"
                                        help="Select the source of truth for your anime list."
                                        options={[
                                            { value: ListSyncOrigin.ANILIST, label: "AniList" },
                                            ...(!!serverStatus?.mal ? [{ value: ListSyncOrigin.MAL, label: "MyAnimeList" }] : []),
                                        ]}
                                        name="origin"
                                        // fieldClassName="w-fit"
                                        // radioLabelClassName="font-semibold flex-none flex pr-8"
                                    />

                                    {/*<Field.Checkbox*/}
                                    {/*    label="Automatic background sync"*/}
                                    {/*    help="Automatically sync your lists with the source of truth."*/}
                                    {/*    name="automatic"*/}
                                    {/*/>*/}

                                    <Field.Submit role="save" isLoading={isPending} />
                                </TypesafeForm>
                            </div>
                        </TabPanels.Panel>
                    </TabPanels.Container>
                </TabPanels>
            </div>
        </div>
    )

}
