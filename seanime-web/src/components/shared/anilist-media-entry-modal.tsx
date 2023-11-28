"use client"
import React, { Fragment } from "react"
import { useToggle } from "react-use"
import { createTypesafeFormSchema, Field, InferType, TypesafeForm } from "@/components/ui/typesafe-form"
import { BaseMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"
import { userAtom } from "@/atoms/user"
import { useAtomValue } from "jotai/react"
import { MediaEntryListData } from "@/lib/server/types"
import { Button, IconButton } from "@/components/ui/button"
import { AiFillEdit } from "@react-icons/all-files/ai/AiFillEdit"
import { BiPlus } from "@react-icons/all-files/bi/BiPlus"
import { Modal } from "../ui/modal"
import Image from "next/image"
import { cn } from "@/components/ui/core"
import { BiListPlus } from "@react-icons/all-files/bi/BiListPlus"
import { BiStar } from "@react-icons/all-files/bi/BiStar"
import { Disclosure } from "@headlessui/react"
import { BiTrash } from "@react-icons/all-files/bi/BiTrash"
import { useQueryClient } from "@tanstack/react-query"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { SeaEndpoints } from "@/lib/server/endpoints"
import toast from "react-hot-toast"

interface AnilistMediaEntryModalProps {
    children?: React.ReactNode
    listData?: MediaEntryListData
    media?: BaseMediaFragment
}

const entrySchema = createTypesafeFormSchema(({ z, presets }) => z.object({
    status: z.custom<MediaListStatus>().nullish(),
    score: z.number().min(0).max(1000).nullish(),
    progress: z.number().min(0).nullish(),
    startedAt: presets.datePicker.nullish().transform(value => value ? ({
        day: value.getUTCDate(),
        month: value.getUTCMonth() + 1,
        year: value.getUTCFullYear(),
    }) : undefined),
    completedAt: presets.datePicker.nullish().transform(value => value ? ({
        day: value.getUTCDate(),
        month: value.getUTCMonth() + 1,
        year: value.getUTCFullYear(),
    }) : undefined),
}))


export const AnilistMediaEntryModal: React.FC<AnilistMediaEntryModalProps> = (props) => {

    const { children, media, listData, ...rest } = props

    const user = useAtomValue(userAtom)

    const qc = useQueryClient()

    const { mutate, isPending } = useSeaMutation<any, InferType<typeof entrySchema> & { mediaId: number }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY,
        mutationKey: ["update-anilist-list-entry"],
        onSuccess: async () => {
            toast.success("Entry updated")
            await qc.refetchQueries({ queryKey: ["get-media-entry", media?.id] })
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
        },
    })

    const [open, toggle] = useToggle(false)

    if (!user || !listData) return null

    return (
        <>
            <IconButton
                intent={"gray-subtle"}
                icon={!!listData ? <AiFillEdit/> : <BiPlus/>}
                rounded
                size={"sm"}
                onClick={toggle}
            />

            {(!listData) && <IconButton
                intent={"primary-subtle"}
                icon={<BiPlus/>}
                rounded
                size={"sm"}
                onClick={() => mutate({
                    mediaId: media?.id || 0,
                    status: "PLANNING",
                })}
            />}

            <Modal
                isOpen={open}
                onClose={toggle}
                title={media?.title?.userPreferred ?? undefined}
                isClosable
                size={"xl"}
                titleClassName={"text-xl"}
            >

                {media?.bannerImage && <div
                    className="h-24 w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]">
                    <Image
                        src={media?.bannerImage!}
                        alt={"banner"}
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-30"
                    />
                    <div
                        className={"z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-gray-900 to-transparent"}
                    />
                </div>}

                {(!!listData) && <TypesafeForm
                    schema={entrySchema}
                    onSubmit={data => {
                        console.log(data.startedAt)
                        mutate({
                            mediaId: media?.id || 0,
                            status: data.status,
                            score: data.score ? data.score * 10 : 0,
                            progress: data.progress,
                            startedAt: data.startedAt,
                            completedAt: data.completedAt,
                        })
                    }}
                    className={cn(
                        {
                            "mt-16": !!media?.bannerImage,
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
                    <div className={"flex flex-col sm:flex-row gap-4"}>
                        <Field.Select
                            label={"Status"}
                            name={"status"}
                            options={[
                                media?.status !== "NOT_YET_RELEASED" ? {
                                    value: "CURRENT",
                                    label: "Watching",
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
                                label={"Score"}
                                name={"score"}
                                discrete
                                min={0}
                                max={10}
                                maxFractionDigits={0}
                                minFractionDigits={0}
                                precision={1}
                                rightIcon={<BiStar/>}
                            />
                            <Field.Number
                                label={"Progress"}
                                name={"progress"}
                                discrete
                                min={0}
                                // max={anilist_getCurrentEpisodeCeilingFromMedia(media)}
                                maxFractionDigits={0}
                                minFractionDigits={0}
                                precision={1}
                                rightIcon={<BiListPlus/>}
                            />
                        </>}
                    </div>
                    {media?.status !== "NOT_YET_RELEASED" && <div className={"flex flex-col sm:flex-row gap-4"}>
                        <Field.DatePicker
                            label={"Start date"}
                            name={"startedAt"}
                            // defaultValue={(state.startedAt && state.startedAt.year) ? parseAbsoluteToLocal(new Date(state.startedAt.year, (state.startedAt.month || 1)-1, state.startedAt.day || 1).toISOString()) : undefined}
                        />
                        <Field.DatePicker
                            label={"Completion date"}
                            name={"completedAt"}
                            // defaultValue={(state.completedAt && state.completedAt.year) ? parseAbsoluteToLocal(new Date(state.completedAt.year, (state.completedAt.month || 1)-1, state.completedAt.day || 1).toISOString()) : undefined}
                        />
                    </div>}

                    <div className={"flex w-full items-center justify-between mt-4"}>
                        <div className={"flex items-center gap-1"}>
                            <Disclosure>
                                <Disclosure.Button as={Fragment}>
                                    <IconButton
                                        intent={"alert-subtle"}
                                        icon={<BiTrash/>}
                                        rounded
                                        size={"md"}
                                    />
                                </Disclosure.Button>
                                <Disclosure.Panel>
                                    <Button
                                        intent={"alert-basic"}
                                        rounded
                                        size={"md"}
                                        // onClick={() => deleteEntry({
                                        //     mediaListEntryId: state.id,
                                        //     status: state.status!,
                                        // })}
                                    >Confirm</Button>
                                </Disclosure.Panel>
                            </Disclosure>
                        </div>

                        <Field.Submit role={"save"} disableIfInvalid={true} isLoading={isPending}/>
                    </div>
                </TypesafeForm>}

            </Modal>
        </>
    )

}