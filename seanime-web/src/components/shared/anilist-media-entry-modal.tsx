"use client"
import { MediaEntryListData } from "@/app/(main)/(library)/_lib/anime-library.types"
import { MangaEntryListData } from "@/app/(main)/manga/_lib/manga.types"
import { userAtom } from "@/atoms/user"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { BaseMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React, { Fragment } from "react"
import { AiFillEdit } from "react-icons/ai"
import { BiListPlus, BiPlus, BiStar, BiTrash } from "react-icons/bi"
import { useToggle } from "react-use"
import { toast } from "sonner"
import { Modal } from "../ui/modal"

interface AnilistMediaEntryModalProps {
    children?: React.ReactNode
    listData?: MediaEntryListData | MangaEntryListData
    media?: BaseMediaFragment
    hideButton?: boolean
    type?: "anime" | "manga"
}

const entrySchema = defineSchema(({ z, presets }) => z.object({
    status: z.custom<MediaListStatus>().nullish(),
    score: z.number().min(0).max(1000).nullish(),
    progress: z.number().min(0).nullish(),
    startedAt: presets.datePicker.nullish().transform(value => value ? ({
        day: value.getDate(),
        month: value.getMonth() + 1,
        year: value.getFullYear(),
    }) : null),
    completedAt: presets.datePicker.nullish().transform(value => value ? ({
        day: value.getDate(),
        month: value.getMonth() + 1,
        year: value.getFullYear(),
    }) : null),
}))


export const AnilistMediaEntryModal: React.FC<AnilistMediaEntryModalProps> = (props) => {
    const [open, toggle] = useToggle(false)

    const { children, media, listData, hideButton, type = "anime", ...rest } = props

    const user = useAtomValue(userAtom)

    const qc = useQueryClient()

    const { mutate, isPending, isSuccess } = useSeaMutation<any, InferType<typeof entrySchema> & { mediaId: number, type: "anime" | "manga" }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY,
        mutationKey: ["update-anilist-list-entry"],
        onSuccess: async () => {
            toast.success("Entry updated")
            if (type === "anime") {
                await qc.refetchQueries({ queryKey: ["get-media-entry", media?.id] })
                await qc.refetchQueries({ queryKey: ["get-library-collection"] })
                await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            } else if (type === "manga") {
                await qc.refetchQueries({ queryKey: ["get-manga-entry", media?.id] })
                await qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            }
        },
    })

    const { mutate: deleteEntry, isPending: isDeleting } = useSeaMutation<any, { mediaId: number, type: "anime" | "manga" }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY,
        mutationKey: ["delete-anilist-list-entry"],
        method: "delete",
        onSuccess: async () => {
            toast.success("Entry removed")
            toggle(false)
            if (type === "anime") {
                await qc.refetchQueries({ queryKey: ["get-media-entry", media?.id] })
                await qc.refetchQueries({ queryKey: ["get-library-collection"] })
                await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            } else if (type === "manga") {
                await qc.refetchQueries({ queryKey: ["get-manga-entry", media?.id] })
                await qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            }
        },
    })

    if (!user) return null

    return (
        <>
            {!hideButton && <>
                {!!listData && <IconButton
                    intent="gray-subtle"
                    icon={<AiFillEdit />}
                    rounded
                    size="sm"
                    loading={isPending || isDeleting}
                    onClick={toggle}
                />}

                {(!listData) && <IconButton
                    intent="primary-subtle"
                    icon={<BiPlus />}
                    rounded
                    size="sm"
                    loading={isPending || isDeleting}
                    className={cn({ "hidden": isSuccess })} // Hide button when mutation is successful
                    onClick={() => mutate({
                        mediaId: media?.id || 0,
                        status: "PLANNING",
                        score: 0,
                        progress: 0,
                        startedAt: null,
                        completedAt: null,
                        type: type,
                    })}
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
                        src={media?.bannerImage!}
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
                    schema={entrySchema}
                    onSubmit={data => {
                        mutate({
                            mediaId: media?.id || 0,
                            status: data.status || "PLANNING",
                            score: data.score ? data.score * 10 : 0,
                            progress: data.progress || 0,
                            startedAt: data.startedAt,
                            completedAt: data.completedAt,
                            type: type,
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
                        //@ts-expect-error
                        startedAt: listData?.startedAt ? new Date(listData?.startedAt) : undefined,
                        //@ts-expect-error
                        completedAt: listData?.completedAt ? new Date(listData?.completedAt) : undefined,
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
                                max={!!media?.nextAiringEpisode?.episode ? media?.nextAiringEpisode?.episode - 1 : (media?.episodes
                                    ? media.episodes
                                    : undefined)}
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
                            name="startedAt"
                            // defaultValue={(state.startedAt && state.startedAt.year) ? parseAbsoluteToLocal(new Date(state.startedAt.year,
                            // (state.startedAt.month || 1)-1, state.startedAt.day || 1).toISOString()) : undefined}
                        />
                        <Field.DatePicker
                            label="Completion date"
                            name="completedAt"
                            // defaultValue={(state.completedAt && state.completedAt.year) ? parseAbsoluteToLocal(new Date(state.completedAt.year,
                            // (state.completedAt.month || 1)-1, state.completedAt.day || 1).toISOString()) : undefined}
                        />
                    </div>}

                    <div className="flex w-full items-center justify-between mt-4">
                        <div>
                            <Disclosure type="multiple" defaultValue={["item-2"]}>
                                <DisclosureItem value="item-1" className="flex items-center gap-1">
                                    <DisclosureTrigger>
                                        <IconButton
                                            intent="alert-subtle"
                                            icon={<BiTrash />}
                                            rounded
                                            size="md"
                                        />
                                    </DisclosureTrigger>
                                    <DisclosureContent>
                                        <Button
                                            intent="alert-basic"
                                            rounded
                                            size="md"
                                            loading={isDeleting}
                                            onClick={() => deleteEntry({
                                                mediaId: media?.id!,
                                                type: type,
                                            })}
                                        >Confirm</Button>
                                    </DisclosureContent>
                                </DisclosureItem>
                            </Disclosure>
                        </div>

                        <Field.Submit role="save" disableIfInvalid={true} loading={isPending} disabled={isDeleting}>
                            Save
                        </Field.Submit>
                    </div>
                </Form>}

            </Modal>
        </>
    )

}
