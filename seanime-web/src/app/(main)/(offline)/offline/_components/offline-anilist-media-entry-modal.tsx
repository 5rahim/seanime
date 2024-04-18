"use client"
import { OfflineAssetMap, OfflineListData } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { userAtom } from "@/atoms/user"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { BaseMangaFragment, BaseMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"
import { normalizeDate } from "@/lib/helpers/date"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React, { Fragment } from "react"
import { AiFillEdit } from "react-icons/ai"
import { BiListPlus, BiStar } from "react-icons/bi"
import { useToggle } from "react-use"
import { toast } from "sonner"

type Props = {
    children?: React.ReactNode
    listData: OfflineListData | undefined
    assetMap: OfflineAssetMap | undefined
    media: BaseMediaFragment | BaseMangaFragment
    hideButton?: boolean
    type: "anime" | "manga"
}

type OfflineListData_QueryVariables = Partial<Omit<OfflineListData, "startedAt" | "completedAt">> & {
    startDate?: string
    endDate?: string
}

const mediaListDataSchema = defineSchema(({ z, presets }) => z.object({
    status: z.custom<MediaListStatus>().nullish(),
    score: z.number().min(0).max(1000).nullish(),
    progress: z.number().min(0).nullish(),
    startDate: presets.datePicker.nullish(),
    endDate: presets.datePicker.nullish(),
}))

export function useUpdateSnapshotEntryListData() {
    const qc = useQueryClient()
    return useSeaMutation<any, OfflineListData_QueryVariables & {
        mediaId: number,
        type: string
    }>({
        endpoint: SeaEndpoints.OFFLINE_SNAPSHOT_ENTRY,
        method: "patch",
        mutationKey: ["update-offline-anilist-list-entry"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-offline-snapshot"] })
            toast.success("Entry updated")
        },
    })
}


export const OfflineAnilistMediaEntryModal: React.FC<Props> = (props) => {
    const [open, toggle] = useToggle(false)

    const { children, media, listData, hideButton, assetMap, type = "anime", ...rest } = props

    const user = useAtomValue(userAtom)


    const { mutate, isPending, isSuccess } = useUpdateSnapshotEntryListData()

    if (!user) return null

    return (
        <>
            {!hideButton && <>
                {!!listData && <IconButton
                    intent="gray-subtle"
                    icon={<AiFillEdit />}
                    rounded
                    size="sm"
                    loading={isPending}
                    onClick={toggle}
                />}
            </>}

            <Modal
                open={open}
                onOpenChange={o => toggle(o)}
                title={media?.title?.userPreferred ?? undefined}
                titleClass="text-xl"
                contentClass="max-w-2xl overflow-hidden"
            >

                {media?.bannerImage && <div
                    className="h-24 w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]"
                >
                    <Image
                        src={offline_getAssetUrl(media?.bannerImage, assetMap) || ""}
                        alt="banner"
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-30"
                    />
                    <div
                        className="z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-[--background] to-transparent"
                    />
                </div>}

                {(!!listData) && <Form
                    schema={mediaListDataSchema}
                    onSubmit={data => {
                        mutate({
                            mediaId: media?.id || 0,
                            status: data.status || "PLANNING",
                            score: data.score || 0,
                            progress: data.progress || 0,
                            startDate: data.startDate ? data.startDate.toISOString() : undefined,
                            endDate: data.endDate ? data.endDate.toISOString() : undefined,
                            type,
                        })
                    }}
                    className={cn(
                        {
                            "mt-8": !!media?.bannerImage,
                        },
                    )}
                    onError={console.log}
                    defaultValues={{
                        status: listData?.status,
                        score: listData?.score,
                        progress: listData?.progress,
                        startDate: listData?.startedAt ? normalizeDate(listData?.startedAt) : undefined,
                        endDate: listData?.completedAt ? normalizeDate(listData?.completedAt) : undefined,
                    }}
                >
                    <div className="flex flex-col sm:flex-row gap-4">
                        <Field.Select
                            label="Status"
                            name="status"
                            options={[
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "CURRENT",
                                    label: type === "anime" ? "Watching" : "Reading",
                                } : undefined,
                                { value: "PLANNING", label: "Planning" },
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "PAUSED",
                                    label: "Paused",
                                } : undefined,
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "COMPLETED",
                                    label: "Completed",
                                } : undefined,
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "DROPPED",
                                    label: "Dropped",
                                } : undefined,
                            ].filter(Boolean)}
                        />
                        {media?.status !== "NOT_YET_RELEASED" && <>
                            <Field.Number
                                label="Score"
                                name="score"
                                min={0}
                                max={10}
                                formatOptions={{
                                    maximumFractionDigits: 0,
                                    minimumFractionDigits: 0,
                                    useGrouping: false,
                                }}
                                rightIcon={<BiStar />}
                            />
                            <Field.Number
                                label="Progress"
                                name="progress"
                                min={0}
                                formatOptions={{
                                    maximumFractionDigits: 0,
                                    minimumFractionDigits: 0,
                                    useGrouping: false,
                                }}
                                rightIcon={<BiListPlus />}
                            />
                        </>}
                    </div>
                    {media?.status !== "NOT_YET_RELEASED" && <div className="flex flex-col sm:flex-row gap-4">
                        <Field.DatePicker
                            label="Start date"
                            name="startDate"
                        />
                        <Field.DatePicker
                            label="Completion date"
                            name="endDate"
                        />
                    </div>}

                    <div className="flex w-full items-center justify-between mt-4">
                        <div>

                        </div>

                        <Field.Submit role="save" disableIfInvalid={true} loading={isPending}>
                            Save
                        </Field.Submit>
                    </div>
                </Form>}

            </Modal>
        </>
    )

}
