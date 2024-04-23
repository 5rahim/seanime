"use client"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { ListSyncDiffs } from "@/app/(main)/list-sync/_containers/list-sync-diffs"
import { ListSyncAnimeDiff, ListSyncOrigin } from "@/app/(main)/list-sync/_lib/list-sync.types"
import { BetaBadge } from "@/components/application/beta-badge"
import { LuffyError } from "@/components/shared/luffy-error"
import { tabsListClass, tabsTriggerClass } from "@/components/shared/styling/classnames"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { InferType } from "prop-types"
import React from "react"
import { toast } from "sonner"

const settingsSchema = defineSchema(({ z }) => z.object({
    automatic: z.boolean(),
    origin: z.string().min(1),
}))

export const dynamic = "force-static"

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

    if (!serverStatus?.mal) return (
        <PageWrapper
            className="p-4 sm:p-8 space-y-4"
        >
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>List Sync <BetaBadge /></h2>
                    <p className="text-[--muted]">Sync your anime lists between different providers.</p>
                </div>
            </div>
            <LuffyError title="Nothing to see">
                Link your MyAnimeList account to use this feature.
            </LuffyError>
        </PageWrapper>
    )


    return (
        <PageWrapper
            className="p-4 sm:p-8 space-y-4"
        >
            <div className="flex justify-between items-center w-full relative">
                <div>
                    <h2>List Sync <BetaBadge /></h2>
                    <p className="text-[--muted]">Sync your anime lists between different providers.</p>
                </div>
            </div>

                <Tabs
                    defaultValue="list"
                    triggerClass={tabsTriggerClass}
                    listClass={tabsListClass}
                >
                    <TabsList>
                        <TabsTrigger value="list">Lists</TabsTrigger>
                        <TabsTrigger value="settings">Settings</TabsTrigger>
                    </TabsList>

                    <TabsContent value="list" className="pt-4">
                        {(!isLoading && !serverStatus?.settings?.listSync) && (
                            <p className="text-[--muted] text-center p-4">
                                List sync is not enabled. Enable it in the settings tab.
                            </p>
                        )}
                        {(!isLoading && !!serverStatus?.settings?.listSync) && <div className="">
                            {typeof animeDiffs !== "string" &&
                                <ListSyncDiffs diffs={animeDiffs ?? []} onClearCache={handleClearCache} isDeletingCache={isDeletingCache} />}
                            {typeof animeDiffs === "string" && <LuffyError>{animeDiffs}</LuffyError>}
                        </div>}
                        {isLoading && <LoadingSpinner />}
                    </TabsContent>
                    <TabsContent value="settings" className="pt-4">
                        <Form
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
                                // fieldClass="w-fit"
                                // radioLabelClass="font-semibold flex-none flex pr-8"
                            />

                            {/*<Field.Checkbox*/}
                            {/*    label="Automatic background sync"*/}
                            {/*    help="Automatically sync your lists with the source of truth."*/}
                            {/*    name="automatic"*/}
                            {/*/>*/}

                            <Field.Submit role="save" loading={isPending}>
                                {!serverStatus?.settings?.listSync ? "Enable" : "Save"}
                            </Field.Submit>
                        </Form>
                    </TabsContent>
                </Tabs>
        </PageWrapper>
    )

}
