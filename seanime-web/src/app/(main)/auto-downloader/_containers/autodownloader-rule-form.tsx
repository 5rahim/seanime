import {
    AL_BaseMedia,
    Anime_AutoDownloaderRule,
    Anime_AutoDownloaderRuleEpisodeType,
    Anime_AutoDownloaderRuleTitleComparisonType,
    Anime_LibraryCollection,
} from "@/api/generated/types"
import { useCreateAutoDownloaderRule, useDeleteAutoDownloaderRule, useUpdateAutoDownloaderRule } from "@/api/hooks/auto_downloader.hooks"
import { useAnilistUserMedia } from "@/app/(main)/_hooks/anilist-collection.hooks"
import { useLibraryCollection } from "@/app/(main)/_hooks/anime-library.hooks"
import { CloseButton, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DangerZone, defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { useQueryClient } from "@tanstack/react-query"
import { uniq } from "lodash"
import Image from "next/image"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { BiPlus } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { LuTextCursorInput } from "react-icons/lu"
import { MdVerified } from "react-icons/md"
import { toast } from "sonner"

type AutoDownloaderRuleFormProps = {
    type: "create" | "edit"
    rule?: Anime_AutoDownloaderRule
    onRuleCreatedOrDeleted?: () => void
}

const schema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    mediaId: z.number().min(1),
    releaseGroups: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    resolutions: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    episodeNumbers: z.array(z.number()).transform(value => uniq(value.filter(Boolean))),
    comparisonTitle: z.string().min(1),
    titleComparisonType: z.string(),
    episodeType: z.string(),
    destination: z.string().min(1),
}))

export function AutoDownloaderRuleForm(props: AutoDownloaderRuleFormProps) {

    const {
        type,
        rule,
        onRuleCreatedOrDeleted,
    } = props

    const qc = useQueryClient()
    const userMedia = useAnilistUserMedia()
    const libraryCollection = useLibraryCollection()

    const allMedia = React.useMemo(() => {
        return userMedia ?? []
    }, [userMedia])

    // Upcoming & airing media
    const notFinishedMedia = React.useMemo(() => {
        return allMedia.filter(media => media.status !== "FINISHED")
    }, [allMedia])

    const { mutate: createRule, isPending: creatingRule } = useCreateAutoDownloaderRule()

    const { mutate: updateRule, isPending: updatingRule } = useUpdateAutoDownloaderRule()

    const { mutate: deleteRule, isPending: deletingRule } = useDeleteAutoDownloaderRule(rule?.dbId)

    const isPending = creatingRule || updatingRule || deletingRule

    function handleSave(data: InferType<typeof schema>) {
        if (data.episodeType === "selected" && data.episodeNumbers.length === 0) {
            return toast.error("You must specify at least one episode number")
        }
        if (type === "create") {
            createRule({
                ...data,
                titleComparisonType: data.titleComparisonType as Anime_AutoDownloaderRuleTitleComparisonType,
                episodeType: data.episodeType as Anime_AutoDownloaderRuleEpisodeType,
            }, {
                onSuccess: () => onRuleCreatedOrDeleted?.(),
            })
        }
        if (type === "edit" && rule?.dbId) {
            updateRule({
                rule: {
                    ...data,
                    dbId: rule.dbId || 0,
                    titleComparisonType: data.titleComparisonType as Anime_AutoDownloaderRuleTitleComparisonType,
                    episodeType: data.episodeType as Anime_AutoDownloaderRuleEpisodeType,
                },
            }, {
                onSuccess: () => onRuleCreatedOrDeleted?.(),
            })
        }
    }

    if (type === "create" && allMedia.length === 0) {
        return <div className="p-4 text-[--muted] text-center">No media found in your library</div>
    }

    return (
        <div className="space-y-4 mt-2">
            <Form
                schema={schema}
                onSubmit={handleSave}
                defaultValues={{
                    enabled: rule?.enabled ?? true,
                    mediaId: rule?.mediaId ?? notFinishedMedia[0]?.id,
                    releaseGroups: rule?.releaseGroups ?? [],
                    resolutions: rule?.resolutions ?? [],
                    comparisonTitle: rule?.comparisonTitle ?? "",
                    titleComparisonType: rule?.titleComparisonType ?? "likely",
                    episodeType: rule?.episodeType ?? "recent",
                    episodeNumbers: rule?.episodeNumbers ?? [],
                    destination: rule?.destination ?? "",
                }}
                onError={() => {
                    toast.error("An error occurred, verify the fields.")
                }}
            >
                {(f) => <RuleFormForm
                    form={f}
                    allMedia={allMedia}
                    type={type}
                    isPending={isPending}
                    notFinishedMedia={notFinishedMedia}
                    libraryCollection={libraryCollection}
                    rule={rule}
                />}
            </Form>
            {type === "edit" && <DangerZone
                actionText="Delete this rule"
                onDelete={() => {
                    if (rule?.dbId) {
                        deleteRule()
                    }
                }}
            />}
        </div>
    )
}

type RuleFormFormProps = {
    form: UseFormReturn<InferType<typeof schema>>
    allMedia: AL_BaseMedia[]
    type: "create" | "edit"
    isPending: boolean
    notFinishedMedia: AL_BaseMedia[]
    libraryCollection?: Anime_LibraryCollection | undefined
    rule?: Anime_AutoDownloaderRule
}

export function RuleFormForm(props: RuleFormFormProps) {

    const {
        form,
        allMedia,
        type,
        isPending,
        notFinishedMedia,
        libraryCollection,
        rule,
        ...rest
    } = props

    const selectedMedia = allMedia.find(media => media.id === Number(form.watch("mediaId")))

    React.useEffect(() => {
        const id = Number(form.watch("mediaId"))
        const destination = libraryCollection?.lists?.flatMap(list => list.entries)?.find(entry => entry?.media?.id === id)?.libraryData?.sharedPath
        if (!isNaN(id) && !rule?.comparisonTitle) {
            const media = allMedia.find(media => media.id === id)
            if (media) {
                form.setValue("comparisonTitle", media.title?.romaji || "")
            }
        }
        if (destination) {
            form.setValue("destination", destination)
        }
    }, [form.watch("mediaId"), selectedMedia, libraryCollection])

    if (!selectedMedia) {
        return <div className="p-4 text-[--muted] text-center">Media is not in your library</div>
    }
    // if (type === "create" && selectedMedia?.status != "RELEASING") {
    //     return <div className="p-4 text-[--muted] text-center">You can only create rules for airing anime</div>
    // }

    return (
        <>
            <Field.Switch name="enabled" label="Enabled" />
            <Separator />
            <div
                className={cn(
                    "space-y-3",
                    !form.watch("enabled") && "opacity-50 pointer-events-none",
                )}
            >
                <div className="flex gap-4 items-end">
                    <div
                        className="w-[6rem] h-[6rem] rounded-[--radius] flex-none object-cover object-center overflow-hidden relative bg-gray-800"
                    >
                        {!!selectedMedia?.coverImage?.large && <Image
                            src={selectedMedia.coverImage.large}
                            alt="banner"
                            fill
                            quality={80}
                            priority
                            sizes="20rem"
                            className="object-cover object-center"
                        />}
                    </div>
                    <Select
                        name="mediaId"
                        label="Library Entry"
                        options={notFinishedMedia.map(media => ({ label: media.title?.userPreferred || "N/A", value: String(media.id) }))}
                        value={String(form.watch("mediaId"))}
                        onValueChange={(v) => form.setValue("mediaId", parseInt(v))}
                        help="The anime must be airing or upcoming"
                        disabled={type === "edit"}
                    />
                </div>

                {selectedMedia?.status === "FINISHED" && <div className="py-2 text-red-300 text-center">This anime is no longer airing</div>}

                <Field.DirectorySelector
                    name="destination"
                    label="Destination"
                    help="Folder in your local library where the files will be saved"
                    leftIcon={<FcFolder />}
                    shouldExist={false}
                />

                <div className="border  rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Title</div>
                    <Field.Text
                        name="comparisonTitle"
                        label="Comparison title"
                        help="Used for comparison purposes. When using 'Exact match', use a title most likely to be used in a torrent name."
                    />
                    <Field.RadioCards
                        label="Type of search"
                        name="titleComparisonType"
                        options={[
                            {
                                label: <div className="w-full">
                                    <p className="mb-1 flex items-center"><MdVerified className="text-lg inline-block mr-2" />Most likely</p>
                                    <p className="font-normal text-sm text-[--muted]">The torrent name will be parsed and analyzed using a comparison
                                                                                      algorithm</p>
                                </div>,
                                value: "likely",
                            },
                            {
                                label: <div className="w-full">
                                    <p className="mb-1 flex items-center"><LuTextCursorInput className="text-lg inline-block mr-2" />Exact match</p>
                                    <p className="font-normal text-sm text-[--muted]">The torrent name must contain the comparison title you set (case
                                                                                      insensitive)</p>
                                </div>,
                                value: "contains",
                            },
                        ]}
                    />
                </div>
                <div
                    className={cn(
                        "border  rounded-[--radius] p-4 relative !mt-8 space-y-3",
                        (selectedMedia?.format === "MOVIE" || (!!selectedMedia.episodes && selectedMedia.episodes === 1)) && "opacity-50 pointer-events-none",
                    )}
                >
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Episodes</div>
                    <Field.RadioCards
                        name="episodeType"
                        label="Episodes to look for"
                        options={[
                            {
                                label: <div className="w-full">
                                    <p>Recent releases</p>
                                    <p className="font-normal text-sm text-[--muted]">New episodes you have not yet watched</p>
                                </div>,
                                value: "recent",
                            },
                            {
                                label: <div className="w-full">
                                    <p>Select</p>
                                    <p className="font-normal text-sm text-[--muted]">Only the specified episodes that aren't in your library</p>
                                </div>,
                                value: "selected",
                            },
                        ]}
                    />

                    {form.watch("episodeType") === "selected" && <TextArrayField
                        label="Episode numbers"
                        value={form.watch("episodeNumbers") || []}
                        onChange={(value) => form.setValue("episodeNumbers", value)}
                        type="number"
                    />}
                </div>

                <div className="border  rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Release Groups</div>
                    <p className="text-sm">
                        List of release groups to look for. If empty, any release group will be accepted.
                    </p>

                    <TextArrayField
                        value={form.watch("releaseGroups") || []}
                        onChange={(value) => form.setValue("releaseGroups", value)}
                        type="text"
                        placeholder="e.g. SubsPlease"
                    />
                </div>

                <div className="border  rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Resolutions</div>
                    <p className="text-sm">
                        List of resolutions to look for. If empty, the highest resolution will be accepted.
                    </p>

                    <TextArrayField
                        value={form.watch("resolutions") || []}
                        onChange={(value) => form.setValue("resolutions", value)}
                        type="text"
                        placeholder="e.g. 1080p"
                    />
                </div>

            </div>
            {type === "create" &&
                <Field.Submit role="create" loading={isPending} disableOnSuccess={false} showLoadingOverlayOnSuccess>Create</Field.Submit>}
            {type === "edit" && <Field.Submit role="update" loading={isPending}>Update</Field.Submit>}
        </>
    )
}

type TextArrayFieldProps<T extends string | number> = {
    value: T[]
    onChange: (value: T[]) => void
    type?: "text" | "number"
    label?: string
    placeholder?: string
}

export function TextArrayField<T extends string | number>(props: TextArrayFieldProps<T>) {

    return (
        <div className="space-y-2">
            {props.label && <div className="flex items-center">
                <div className="text-base font-semibold">{props.label}</div>
            </div>}
            {props.value.map((value, index) => (
                <div key={index} className="flex gap-2 items-center">
                    {props.type === "text" && <TextInput
                        value={value}
                        onChange={(e) => {
                            const newValue = [...props.value]
                            newValue[index] = e.target.value as any
                            props.onChange(newValue)
                        }}
                        placeholder={props.placeholder}
                    />}
                    {props.type === "number" && <TextInput
                        type="number"
                        value={value as number}
                        onChange={(e) => {
                            const newValue = [...props.value]
                            const intVal = parseInt(e.target.value) as number
                            newValue[index] = (isNaN(parseInt(e.target.value)) ? 1 : (intVal < 1 ? 1 : intVal)) as any
                            props.onChange(newValue)
                        }}
                    />}
                    <CloseButton
                        size="sm"
                        intent="alert-subtle"
                        onClick={() => {
                            const newValue = [...props.value]
                            newValue.splice(index, 1)
                            props.onChange(newValue)
                        }}
                    />
                </div>
            ))}
            <IconButton
                intent="success"
                className="rounded-full"
                onClick={() => {
                    props.onChange([...props.value as any, props.type === "number" ? 1 : ""])
                }}
                icon={<BiPlus />}
            />
        </div>
    )
}
