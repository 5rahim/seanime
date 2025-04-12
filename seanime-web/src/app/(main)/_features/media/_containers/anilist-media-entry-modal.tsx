"use client"
import { AL_BaseAnime, AL_BaseManga, AL_MediaListStatus, Anime_EntryListData, Manga_EntryListData } from "@/api/generated/types"
import { useDeleteAnilistListEntry, useEditAnilistListEntry } from "@/api/hooks/anilist.hooks"
import { useUpdateAnimeEntryRepeat } from "@/api/hooks/anime_entries.hooks"
import { useCurrentUser } from "@/app/(main)/_hooks/use-server-status"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Tooltip } from "@/components/ui/tooltip"
import { normalizeDate } from "@/lib/helpers/date"
import { getImageUrl } from "@/lib/server/assets"
import Image from "next/image"
import React, { Fragment } from "react"
import { AiFillEdit } from "react-icons/ai"
import { BiListPlus, BiPlus, BiStar, BiTrash } from "react-icons/bi"
import { useToggle } from "react-use"

type AnilistMediaEntryModalProps = {
    children?: React.ReactNode
    listData?: Anime_EntryListData | Manga_EntryListData
    media?: AL_BaseAnime | AL_BaseManga
    hideButton?: boolean
    type?: "anime" | "manga"
}

export const mediaListDataSchema = defineSchema(({ z, presets }) => z.object({
    status: z.custom<AL_MediaListStatus>().nullish(),
    score: z.number().min(0).max(100).nullish(),
    progress: z.number().min(0).nullish(),
    startedAt: presets.datePicker.nullish(),
    completedAt: presets.datePicker.nullish(),
}))


export const AnilistMediaEntryModal: React.FC<AnilistMediaEntryModalProps> = (props) => {
    const [open, toggle] = useToggle(false)

    const { children, media, listData, hideButton, type = "anime", ...rest } = props

    const user = useCurrentUser()

    const { mutate, isPending: _isPending1, isSuccess } = useEditAnilistListEntry(media?.id, type)
    const { mutate: mutateRepeat, isPending: _isPending2 } = useUpdateAnimeEntryRepeat(media?.id)
    const isPending = _isPending1
    const { mutate: deleteEntry, isPending: isDeleting } = useDeleteAnilistListEntry(media?.id, type, () => {
        toggle(false)
    })

    const [repeat, setRepeat] = React.useState(0)

    React.useEffect(() => {
        setRepeat(listData?.repeat || 0)
    }, [listData])

    if (!user) return null

    return (
        <>
            {!hideButton && <>
                {!!listData && <IconButton
                    data-anilist-media-entry-modal-edit-button
                    intent="white-subtle"
                    icon={<AiFillEdit />}
                    rounded
                    size="sm"
                    loading={isPending || isDeleting}
                    onClick={toggle}
                />}

                {(!listData) && <Tooltip
                    trigger={<IconButton
                        data-anilist-media-entry-modal-add-button
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
                            startedAt: undefined,
                            completedAt: undefined,
                            type: type,
                        })}
                    />}
                >
                    Add to list
                </Tooltip>}
            </>}

            <Modal
                open={open}
                onOpenChange={o => toggle(o)}
                title={media?.title?.userPreferred ?? undefined}
                titleClass="text-xl"
                contentClass="max-w-3xl overflow-hidden"
            >

                {media?.bannerImage && <div
                    data-anilist-media-entry-modal-banner-image-container
                    className="h-24 w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]"
                >
                    <Image
                        data-anilist-media-entry-modal-banner-image
                        src={getImageUrl(media?.bannerImage!)}
                        alt="banner"
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-15"
                    />
                    <div
                        data-anilist-media-entry-modal-banner-image-bottom-gradient
                        className="z-[5] absolute bottom-0 w-full h-[60%] bg-gradient-to-t from-[--background] to-transparent"
                    />
                </div>}

                {(!!listData) && <Form
                    data-anilist-media-entry-modal-form
                    schema={mediaListDataSchema}
                    onSubmit={data => {
                        if (repeat !== (listData?.repeat ?? 0)) {
                            // Update repeat count
                            mutateRepeat({
                                mediaId: media?.id || 0,
                                repeat: repeat,
                            })
                        }
                        mutate({
                            mediaId: media?.id || 0,
                            status: data.status || "PLANNING",
                            score: data.score ? data.score * 10 : 0, // should be 0-100
                            progress: data.progress || 0,
                            startedAt: data.startedAt ? {
                                // @ts-ignore
                                day: data.startedAt.getDate(),
                                month: data.startedAt.getMonth() + 1,
                                year: data.startedAt.getFullYear(),
                            } : undefined,
                            completedAt: data.completedAt ? {
                                // @ts-ignore
                                day: data.completedAt.getDate(),
                                month: data.completedAt.getMonth() + 1,
                                year: data.completedAt.getFullYear(),
                            } : undefined,
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
                        score: listData?.score ? listData?.score / 10 : undefined, // Returned score is 0-100
                        progress: listData?.progress,
                        startedAt: listData?.startedAt ? (normalizeDate(listData?.startedAt)) : undefined,
                        completedAt: listData?.completedAt ? (normalizeDate(listData?.completedAt)) : undefined,
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
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "REPEATING",
                                    label: "Repeating",
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
                                    maximumFractionDigits: 1,
                                    minimumFractionDigits: 0,
                                    useGrouping: false,
                                }}
                                rightIcon={<BiStar />}
                            />
                            <Field.Number
                                label="Progress"
                                name="progress"
                                min={0}
                                max={type === "anime" ? (!!(media as AL_BaseAnime)?.nextAiringEpisode?.episode
                                    ? (media as AL_BaseAnime)?.nextAiringEpisode?.episode! - 1
                                    : ((media as AL_BaseAnime)?.episodes
                                        ? (media as AL_BaseAnime).episodes
                                        : undefined)) : (media as AL_BaseManga)?.chapters}
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

                        <NumberInput
                            name="repeat"
                            label={type === "anime" ? "Total rewatches" : "Total rereads"}
                            min={0}
                            max={1000}
                            value={repeat}
                            onValueChange={setRepeat}
                            formatOptions={{
                                maximumFractionDigits: 0,
                                minimumFractionDigits: 0,
                                useGrouping: false,
                            }}
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
