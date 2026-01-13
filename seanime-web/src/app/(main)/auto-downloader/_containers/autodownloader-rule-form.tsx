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
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button, CloseButton, IconButton } from "@/components/ui/button"
import { Combobox } from "@/components/ui/combobox"
import { cn } from "@/components/ui/core/styling"
import { DangerZone, defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { upath } from "@/lib/helpers/upath"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { uniq } from "lodash"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React, { useMemo, useRef } from "react"
import { useFieldArray, UseFormReturn, useWatch } from "react-hook-form"
import { BiPlus } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { LuTextCursorInput } from "react-icons/lu"
import { MdFilterAlt, MdVerified } from "react-icons/md"
import { useMount } from "react-use"
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

export const _autoDownloader_listActiveMediaOnlyAtom = atomWithStorage<"airing" | "airing-upcoming" | "all">(
    "sea-auto-downloader-list-active-media-only",
    "airing-upcoming")
const listActiveMediaOptions: ("airing" | "airing-upcoming" | "all")[] = ["airing", "airing-upcoming", "all"]

export function useAutoDownloaderMediaList(allMedia: AL_BaseAnime[]) {
    const showReleasingOnly = useAtomValue(_autoDownloader_listActiveMediaOnlyAtom)

    return React.useMemo(() => {
        if (showReleasingOnly === "airing") {
            return allMedia.filter(media => media.status === "RELEASING")
        }
        if (showReleasingOnly === "airing-upcoming") {
            return allMedia.filter(media => media.status !== "FINISHED")
        }
        return allMedia
    }, [allMedia, showReleasingOnly])
}

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

    const mediaList = useAutoDownloaderMediaList(allMedia)

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
                    mediaId: mediaId ?? rule?.mediaId ?? mediaList[0]?.id,
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
                {(f) => (
                    <div className="space-y-4">
                        <RuleFormFields
                            form={f}
                            allMedia={allMedia}
                            mediaId={mediaId}
                            type={type}
                            isPending={isPending}
                            mediaList={mediaList}
                            libraryCollection={libraryCollection}
                            rule={rule}
                        />
                    </div>
                )}
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
    mediaList: AL_BaseAnime[]
    libraryCollection?: Anime_LibraryCollection | undefined

    rule?: Anime_AutoDownloaderRule
}

export function AutoDownloaderMediaCombobox(props: {
    mediaList: AL_BaseAnime[],
    value: number,
    onValueChange: (v: string[]) => void,
    type: "create" | "edit",
    mediaId?: number | undefined
}) {
    const [showReleasingOnly, setShowReleasingOnly] = useAtom(_autoDownloader_listActiveMediaOnlyAtom)

    return <Combobox
        name="mediaId"
        label={<div className="flex items-center gap-2">
            <p className="text-lg font-semibold">Media</p>
            {props.type !== "edit" && <Button
                leftIcon={<MdFilterAlt />} intent="gray-link" className="!text-[--muted] cursor-pointer hover:underline underline-offset-2 py-0 px-2"
                onClick={() => setShowReleasingOnly(prev => {
                    const currentIndex = listActiveMediaOptions.indexOf(prev)
                    const nextIndex = (currentIndex + 1) % listActiveMediaOptions.length
                    return listActiveMediaOptions[nextIndex]
                })}
            >
                {showReleasingOnly === "airing" && "Showing airing only"}
                {showReleasingOnly === "airing-upcoming" && "Showing airing & upcoming"}
                {showReleasingOnly === "all" && "Showing all"}
            </Button>}
        </div>}
        options={props.mediaList.map(media => ({
            label: <div className="flex items-center gap-2">
                <div className="size-10 rounded-full bg-gray-800 flex items-center justify-center relative overflow-hidden flex-none">
                    <Image
                        src={media.coverImage?.medium ?? "/no-cover.png"}
                        alt="cover"
                        sizes="2rem"
                        fill
                        className="object-cover object-center"
                    />
                </div>
                <p>{media.title?.userPreferred || "N/A"}</p>
                <p className="text-[--muted] text-sm">{capitalize(media.status)?.replaceAll("_", " ")}</p>
            </div>,
            value: String(media.id),
            textValue: media.title?.userPreferred || "N/A",
        })).toSorted((a, b) => a.textValue.localeCompare(b.textValue))}
        value={[String(props.value)]}
        onValueChange={props.onValueChange}
        disabled={props.type === "edit" || !!props.mediaId}
        multiple={false}
        emptyMessage="No media found"
    />
}

export function RuleFormFields(props: RuleFormFieldsProps) {

    const {
        form,
        allMedia,
        mediaId,
        type,
        isPending,
        mediaList,
        libraryCollection,
        rule,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    // Fallback to showing all media if editing so the current media is visible
    const [showReleasingOnly, setShowReleasingOnly] = useAtom(_autoDownloader_listActiveMediaOnlyAtom)
    const previousShowReleasingOnly = useRef(showReleasingOnly)
    React.useEffect(() => {
        console.warn("RuleFormFields: type changed", type)
        if (type === "edit" && showReleasingOnly !== "all") {
            previousShowReleasingOnly.current = showReleasingOnly
            setShowReleasingOnly("all")
        }
    }, [type, showReleasingOnly])
    useMount(() => {
        setShowReleasingOnly(previousShowReleasingOnly.current)
    })

    const form_mediaId = useWatch({ name: "mediaId" }) as number
    const form_episodeType = useWatch({ name: "episodeType" }) as Anime_AutoDownloaderRuleEpisodeType
    const destination = useWatch({ name: "destination" }) as string

    const selectedMedia = allMedia.find(media => media.id === Number(form_mediaId))

    const animeFolderName = useMemo(() => {
        return sanitizeDirectoryName(selectedMedia?.title?.userPreferred || "")
    }, [selectedMedia])

    const libraryPathSelectionProps = useLibraryPathSelection({
        destination,
        setDestination: path => form.setValue("destination", path),
        animeFolderName,
    })

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
                const newDestination = upath.join(upath.normalizeSafe(serverStatus?.settings?.library?.libraryPath || ""), animeFolderName)
                form.setValue("destination", newDestination)
            }
        }
    }, [form_mediaId, selectedMedia, libraryCollection, rule, animeFolderName])

    if (!selectedMedia) {
        return <div className="p-4 text-[--muted] text-center">Media is not in your library</div>
    }

    return (
        <>
            <div className="flex flex-col gap-2 md:flex-row justify-between items-center">
                <Field.Switch name="enabled" label="Enabled" />
            </div>
            <Separator />
            <div
                className={cn(
                    "space-y-3",
                )}
            >
                {!mediaId && <div className="flex gap-4 items-end">
                    <AutoDownloaderMediaCombobox
                        mediaList={mediaList}
                        value={form_mediaId}
                        onValueChange={(v) => form.setValue("mediaId", v[0] ? parseInt(v[0]) : mediaList[0]?.id)}
                        type={type}
                        mediaId={mediaId}
                    />
                </div>}

                {selectedMedia?.status === "FINISHED" && <div className="py-2 text-[--orange] text-center">No longer airing</div>}

                <Field.DirectorySelector
                    name="destination"
                    label="Destination"
                    help="Folder in your local library where the files will be saved"
                    leftIcon={<FcFolder />}
                    shouldExist={false}
                    libraryPathSelectionProps={libraryPathSelectionProps}
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
                intent="success-glass"
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
