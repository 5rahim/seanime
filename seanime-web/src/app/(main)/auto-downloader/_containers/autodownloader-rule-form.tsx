import {
    AL_BaseAnime,
    Anime_AutoDownloaderRule,
    Anime_AutoDownloaderRuleEpisodeType,
    Anime_AutoDownloaderRuleTitleComparisonType,
    Anime_LibraryCollection,
} from "@/api/generated/types"
import { useCreateAutoDownloaderRule, useDeleteAutoDownloaderRule, useUpdateAutoDownloaderRule } from "@/api/hooks/auto_downloader.hooks"
import { useAnilistUserAnime } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useLibraryCollection } from "@/app/(main)/_hooks/anime-library-collection-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { CloseButton, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DangerZone, defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { upath } from "@/lib/helpers/upath"
import { uniq } from "lodash"
import Image from "next/image"
import React from "react"
import { useFieldArray, UseFormReturn, useWatch } from "react-hook-form"
import { BiPlus } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { LuTextCursorInput } from "react-icons/lu"
import { MdVerified } from "react-icons/md"
import { toast } from "sonner"

type AutoDownloaderRuleFormProps = {
    type: "create" | "edit"
    rule?: Anime_AutoDownloaderRule
    mediaId?: number
    onRuleCreatedOrDeleted?: () => void
}

const schema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    mediaId: z.number().min(1),
    releaseGroups: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    resolutions: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    episodeNumbers: z.array(z.number()).transform(value => uniq(value.filter(Boolean))),
    additionalTerms: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
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
        mediaId,
    } = props

    const userMedia = useAnilistUserAnime()
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
                    mediaId: mediaId ?? rule?.mediaId ?? notFinishedMedia[0]?.id,
                    releaseGroups: rule?.releaseGroups ?? [],
                    resolutions: rule?.resolutions ?? [],
                    comparisonTitle: rule?.comparisonTitle ?? "",
                    titleComparisonType: rule?.titleComparisonType ?? "likely",
                    episodeType: rule?.episodeType ?? "recent",
                    episodeNumbers: rule?.episodeNumbers ?? [],
                    destination: rule?.destination ?? "",
                    additionalTerms: rule?.additionalTerms ?? [],
                }}
                onError={() => {
                    toast.error("An error occurred, verify the fields.")
                }}
            >
                {(f) => <RuleFormFields
                    form={f}
                    allMedia={allMedia}
                    mediaId={mediaId}
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

type RuleFormFieldsProps = {
    form: UseFormReturn<InferType<typeof schema>>
    allMedia: AL_BaseAnime[]
    mediaId?: number
    type: "create" | "edit"
    isPending: boolean
    notFinishedMedia: AL_BaseAnime[]
    libraryCollection?: Anime_LibraryCollection | undefined
    rule?: Anime_AutoDownloaderRule
}

export function RuleFormFields(props: RuleFormFieldsProps) {

    const {
        form,
        allMedia,
        mediaId,
        type,
        isPending,
        notFinishedMedia,
        libraryCollection,
        rule,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const form_mediaId = useWatch({ name: "mediaId" }) as number
    const form_episodeType = useWatch({ name: "episodeType" }) as Anime_AutoDownloaderRuleEpisodeType

    const selectedMedia = allMedia.find(media => media.id === Number(form_mediaId))

    React.useEffect(() => {
        const id = Number(form_mediaId)
        const destination = libraryCollection?.lists?.flatMap(list => list.entries)?.find(entry => entry?.media?.id === id)?.libraryData?.sharedPath
        if (!isNaN(id) && !rule?.comparisonTitle) {
            const media = allMedia.find(media => media.id === id)
            if (media) {
                form.setValue("comparisonTitle", media.title?.romaji || "")
            }
        }
        // If no rule is passed, set the comparison title to the media title
        if (!rule) {
            if (destination) {
                form.setValue("destination", destination)
            } else if (type === "create") {
                // form.setValue("destination", "")
                const newDestination = upath.join(upath.normalizeSafe(serverStatus?.settings?.library?.libraryPath || ""),
                    sanitizeDirectoryName(selectedMedia?.title?.romaji || selectedMedia?.title?.english || ""))
                form.setValue("destination", newDestination)
            }
        }
    }, [form_mediaId, selectedMedia, libraryCollection, rule])

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
                    // !form.watch("enabled") && "opacity-50 pointer-events-none",
                )}
            >
                {!mediaId && <div className="flex gap-4 items-end">
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
                        options={notFinishedMedia.map(media => ({ label: media.title?.userPreferred || "N/A", value: String(media.id) }))
                            .toSorted((a, b) => a.label.localeCompare(b.label))}
                        value={String(form_mediaId)}
                        onValueChange={(v) => form.setValue("mediaId", parseInt(v))}
                        help={!mediaId ? "The anime must be airing or upcoming" : undefined}
                        disabled={type === "edit" || !!mediaId}
                    />
                </div>}

                {selectedMedia?.status === "FINISHED" && <div className="py-2 text-red-300 text-center">This anime is no longer airing</div>}

                <Field.DirectorySelector
                    name="destination"
                    label="Destination"
                    help="Folder in your local library where the files will be saved"
                    leftIcon={<FcFolder />}
                    shouldExist={false}
                />

                <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
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

                    {form_episodeType === "selected" && <TextArrayField
                        label="Episode numbers"
                        name="episodeNumbers"
                        control={form.control}
                        type="number"
                    />}
                </div>

                <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Release Groups</div>
                    <p className="text-sm">
                        List of release groups to look for. If empty, any release group will be accepted.
                    </p>

                    <TextArrayField
                        name="releaseGroups"
                        control={form.control}
                        type="text"
                        placeholder="e.g. SubsPlease"
                        separatorText="OR"
                    />
                </div>

                <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Resolutions</div>
                    <p className="text-sm">
                        List of resolutions to look for. If empty, the highest resolution will be accepted.
                    </p>

                    <TextArrayField
                        name="resolutions"
                        control={form.control}
                        type="text"
                        placeholder="e.g. 1080p"
                        separatorText="OR"
                    />
                </div>

                <Accordion type="single" collapsible className="!my-4" defaultValue={!!rule?.additionalTerms?.length ? "more" : undefined}>
                    <AccordionItem value="more">
                        <AccordionTrigger className="border rounded-[--radius] bg-gray-900">
                            More filters
                        </AccordionTrigger>
                        <AccordionContent className="pt-0">
                            <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Additional
                                                                                                                                         terms
                                </div>
                                <div>
                                    {/*<p className="text-sm">*/}
                                    {/*    List of video terms to look for. If any term is found in the torrent name, it will be accepted.*/}
                                    {/*</p>*/}
                                    <p className="text-sm -top-2 relative"><span className="text-red-100">
                                        All options must be included for the torrent to be accepted.</span> Within each option, you can
                                                                                                            include variations separated by
                                                                                                            commas. For example, adding
                                                                                                            "H265,H.265, H 265,x265" and
                                                                                                            "10bit,10-bit,10 bit" will match
                                        <code className="text-gray-400"> [Group] Torrent name [HEVC 10bit
                                                                         x265]</code> but not <code className="text-gray-400">[Group] Torrent name
                                                                                                                              [H265]</code>. Case
                                                                                                            insensitive.</p>
                                </div>

                                <TextArrayField
                                    name="additionalTerms"
                                    control={form.control}
                                    // value={additionalTerms}
                                    // onChange={(value) => form.setValue("additionalTerms", value)}
                                    type="text"
                                    placeholder="e.g. H265,H.265,H 265,x265"
                                    separatorText="AND"
                                />
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                </Accordion>

            </div>
            {type === "create" &&
                <Field.Submit role="create" loading={isPending} disableOnSuccess={false} showLoadingOverlayOnSuccess>Create</Field.Submit>}
            {type === "edit" && <Field.Submit role="update" loading={isPending}>Update</Field.Submit>}
        </>
    )
}

type TextArrayFieldProps<T extends string | number> = {
    name: string
    control: any
    type?: "text" | "number"
    label?: string
    placeholder?: string
    separatorText?: string
}

export function TextArrayField<T extends string | number>(props: TextArrayFieldProps<T>) {
    const { fields, append, remove } = useFieldArray({
        control: props.control,
        name: props.name,
    })

    return (
        <div className="space-y-2">
            {props.label && <div className="flex items-center">
                <div className="text-base font-semibold">{props.label}</div>
            </div>}
            {fields.map((field, index) => (
                <React.Fragment key={field.id}>
                    <div className="flex gap-2 items-center">
                        {props.type === "text" && (
                            <TextInput
                                {...props.control.register(`${props.name}.${index}`)}
                                placeholder={props.placeholder}
                            />
                        )}
                        {props.type === "number" && (
                            <TextInput
                                type="number"
                                {...props.control.register(`${props.name}.${index}`, {
                                    valueAsNumber: true,
                                    min: 1,
                                    validate: (value: number) => !isNaN(value),
                                })}
                            />
                        )}
                        <CloseButton
                            size="sm"
                            intent="alert-subtle"
                            onClick={() => remove(index)}
                        />
                    </div>
                    {(!!props.separatorText && index < fields.length - 1) && (
                        <p className="text-center text-[--muted]">{props.separatorText}</p>
                    )}
                </React.Fragment>
            ))}
            <IconButton
                intent="success"
                className="rounded-full"
                onClick={() => append(props.type === "number" ? 1 : "")}
                icon={<BiPlus />}
            />
        </div>
    )
}


function sanitizeDirectoryName(input: string): string {
    const disallowedChars = /[<>:"/\\|?*\x00-\x1F.!`]/g // Pattern for disallowed characters
    // Replace disallowed characters with an underscore
    const sanitized = input.replace(disallowedChars, " ")
    // Remove leading/trailing spaces and dots (periods) which are not allowed
    const trimmed = sanitized.trim().replace(/^\.+|\.+$/g, "").replace(/\s+/g, " ")
    // Ensure the directory name is not empty after sanitization
    return trimmed || "Untitled"
}
