import {
    AL_BaseAnime,
    Anime_AutoDownloaderRule,
    Anime_AutoDownloaderRuleEpisodeType,
    Anime_AutoDownloaderRuleTitleComparisonType,
    Anime_LibraryCollection,
} from "@/api/generated/types"
import { useCreateAutoDownloaderRule } from "@/api/hooks/auto_downloader.hooks"
import { __anilist_userAnimeListDataAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useAnilistUserAnime } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useLibraryCollection } from "@/app/(main)/_hooks/anime-library-collection-loader"
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    _autoDownloader_listActiveMediaOnlyAtom,
    AutoDownloaderMediaCombobox,
    useAutoDownloaderMediaList,
} from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import {
    AdditionalTermsField,
    ExcludeTermsField,
    ProfileSelectField,
    ProvidersField,
    ReleaseGroupsField,
    ResolutionsField,
} from "@/app/(main)/auto-downloader/_containers/autodownloader-shared-fields"
import { Button, CloseButton, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { upath } from "@/lib/helpers/upath"
import { useAtom, useAtomValue } from "jotai/react"
import { uniq } from "lodash"
import React from "react"
import { useFieldArray, UseFormReturn, useWatch } from "react-hook-form"
import { BiPlus } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { LuTextCursorInput } from "react-icons/lu"
import { MdVerified } from "react-icons/md"
import { toast } from "sonner"

type AutoDownloaderBatchRuleFormProps = {
    onRuleCreated: () => void
    rules: Anime_AutoDownloaderRule[]
}

const schema = defineSchema(({ z, presets }) => z.object({
    enabled: z.boolean(),
    entries: z.array(z.object({
        mediaId: z.number(),
        destination: z.string(),
        comparisonTitle: z.string(),
    })).min(1),
    releaseGroups: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    resolutions: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    additionalTerms: z.array(z.string()).optional().transform(value => !value?.length ? [] : uniq(value.filter(Boolean))),
    excludeTerms: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    titleComparisonType: z.string(),
    minSeeders: z.number().min(0).optional().default(0),
    minSize: z.string().optional().default(""),
    maxSize: z.string().optional().default(""),
    providers: z.array(z.string()).optional().transform(value => !value?.length ? [] : uniq(value.filter(Boolean))),
    profileId: presets.multiSelect,
}))

export function AutoDownloaderBatchRuleForm(props: AutoDownloaderBatchRuleFormProps) {

    const {
        onRuleCreated,
        rules,
    } = props

    const userMedia = useAnilistUserAnime()
    const libraryCollection = useLibraryCollection()

    const allMedia = React.useMemo(() => {
        return userMedia ?? []
    }, [userMedia])

    const mediaList = useAutoDownloaderMediaList(allMedia)

    const { mutate: createRule, isPending: creatingRule } = useCreateAutoDownloaderRule()

    const isPending = creatingRule

    function handleSave(data: InferType<typeof schema>) {
        for (const entry of data.entries) {
            if (entry.destination === "" || entry.mediaId === 0) {
                continue
            }
            createRule({
                rule: {
                    dbId: 0,
                    titleComparisonType: data.titleComparisonType as Anime_AutoDownloaderRuleTitleComparisonType,
                    episodeType: "recent" as Anime_AutoDownloaderRuleEpisodeType,
                    enabled: data.enabled,
                    mediaId: entry.mediaId,
                    releaseGroups: data.releaseGroups,
                    resolutions: data.resolutions,
                    additionalTerms: data.additionalTerms,
                    excludeTerms: data.excludeTerms,
                    comparisonTitle: entry.comparisonTitle,
                    destination: entry.destination,
                    minSeeders: data.minSeeders,
                    minSize: data.minSize,
                    maxSize: data.maxSize,
                    providers: data.providers,
                    profileId: !!data.profileId?.[0] ? Number(data.profileId[0]) : undefined,
                },
            })
        }
        onRuleCreated?.()
    }

    if (allMedia.length === 0) {
        return <div className="p-4 text-[--muted] text-center">No media found in your library</div>
    }

    return (
        <div className="space-y-4 mt-2">
            <Form
                schema={schema}
                onSubmit={handleSave}
                onError={errors => {
                    console.log(errors)
                    toast.error("An error occurred, verify the fields.")
                }}
                defaultValues={{
                    enabled: true,
                    titleComparisonType: "likely",
                    minSeeders: 0,
                    profileId: [],
                }}
            >
                {(f) => (
                    <div className="space-y-4">
                        <RuleFormFields
                            form={f}
                            allMedia={allMedia}
                            isPending={isPending}
                            mediaList={mediaList}
                            libraryCollection={libraryCollection}
                            rules={rules}
                        />
                    </div>
                )}
            </Form>
        </div>
    )
}

type RuleFormFieldsProps = {
    form: UseFormReturn<InferType<typeof schema>>
    allMedia: AL_BaseAnime[]
    isPending: boolean
    mediaList: AL_BaseAnime[]
    libraryCollection?: Anime_LibraryCollection | undefined
    rules: Anime_AutoDownloaderRule[]
}

function RuleFormFields(props: RuleFormFieldsProps) {

    const {
        form,
        allMedia,
        isPending,
        mediaList,
        libraryCollection,
        rules,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    return (
        <>
            <div className="flex flex-col gap-2 md:flex-row justify-between items-center">
                <Field.Switch name="enabled" label="Enabled" />
            </div>
            <Separator />
            <div
                className={cn(
                    "space-y-3",
                    // !form.watch("enabled") && "opacity-50 pointer-events-none",
                )}
            >

                <MediaArrayField
                    allMedia={mediaList}
                    libraryPath={serverStatus?.settings?.library?.libraryPath || ""}
                    name="entries"
                    control={form.control}
                    label="Anime"
                    separatorText="AND"
                    form={form}
                    libraryCollection={libraryCollection}
                    rules={rules}
                />

                <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Title</div>
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

                <ProfileSelectField name="profileId" />

                <ReleaseGroupsField name="releaseGroups" control={form.control} />

                <ResolutionsField name="resolutions" control={form.control} />

                <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Constraints</div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        <Field.Number
                            name="minSeeders"
                            label="Min Seeders"
                            min={0}
                            fieldClass="w-full"
                        />
                        <Field.Text
                            name="minSize"
                            label="Min Size"
                            placeholder="e.g. 100MB"
                            fieldClass="w-full"
                        />
                        <Field.Text
                            name="maxSize"
                            label="Max Size"
                            placeholder="e.g. 2GB or 10GiB"
                            fieldClass="w-full"
                        />
                    </div>
                </div>

                <ProvidersField name="providers" control={form.control} />

                <AdditionalTermsField name="additionalTerms" control={form.control} />

                <ExcludeTermsField name="excludeTerms" control={form.control} />

            </div>
            <div className="flex gap-2">
                <Field.Submit role="create" loading={isPending} disableOnSuccess={false} showLoadingOverlayOnSuccess>Create</Field.Submit>
            </div>
        </>
    )
}

type MediaArrayFieldProps = {
    name: string
    control: any
    allMedia: AL_BaseAnime[]
    libraryPath: string
    label?: string
    separatorText?: string
    form: UseFormReturn<InferType<typeof schema>>
    libraryCollection?: Anime_LibraryCollection | undefined
    rules: Anime_AutoDownloaderRule[]
}

interface MediaEntry {
    mediaId: number
    destination: string
    comparisonTitle: string
}

interface FormValues {
    [key: string]: MediaEntry[]
}

export function MediaArrayField(props: MediaArrayFieldProps) {
    const { fields, append, remove, update } = useFieldArray<FormValues>({
        control: props.control,
        name: props.name,
    })

    const userMedia = useAnilistUserAnime()

    const entriesAdded = useWatch({ name: "entries" }) as any[]
    const anilistListData = useAtomValue(__anilist_userAnimeListDataAtom)
    const [showReleasingOnly, setShowReleasingOnly] = useAtom(_autoDownloader_listActiveMediaOnlyAtom)

    const handleFieldChange = (index: number, updatedValues: Partial<MediaEntry>, field: MediaEntry) => {
        if ("mediaId" in updatedValues) {
            const mediaId = updatedValues.mediaId!
            const sanitizedTitle = sanitizeDirectoryName(props.allMedia.find(m => m.id === mediaId)?.title?.userPreferred || "")

            update(index, {
                ...field,
                ...updatedValues,
                destination: upath.join(props.libraryPath, sanitizedTitle),
                comparisonTitle: sanitizedTitle,
            })
        } else {
            update(index, { ...field, ...updatedValues })
        }
    }

    function handleAddCurrentlyWatching() {
        // Get media ids that already have rules
        const existingRuleMediaIds = new Set(props.rules.map(rule => rule.mediaId))

        // Filter media that are currently watching and don't have rules
        const currentlyWatchingMedia = userMedia?.filter(media => {
            const listData = anilistListData[String(media.id)]
            return listData?.status === "CURRENT" && !existingRuleMediaIds.has(media.id)
        }) ?? []

        setShowReleasingOnly("all")

        currentlyWatchingMedia.forEach(media => {
            const sanitizedTitle = sanitizeDirectoryName(media.title?.userPreferred || "")
            append({
                mediaId: media.id,
                destination: upath.join(props.libraryPath, sanitizedTitle),
                comparisonTitle: sanitizedTitle,
            })
        })
    }

    function handleAddUpcoming() {
        // Get media ids that already have rules
        const existingRuleMediaIds = new Set(props.rules.map(rule => rule.mediaId))

        // Filter media that are upcoming and don't have rules
        const upcomingMedia = userMedia?.filter(media => {
            return media.status === "NOT_YET_RELEASED" && !existingRuleMediaIds.has(media.id)
        }) ?? []

        setShowReleasingOnly("all")

        upcomingMedia.forEach(media => {
            const sanitizedTitle = sanitizeDirectoryName(media.title?.userPreferred || "")
            append({
                mediaId: media.id,
                destination: upath.join(props.libraryPath, sanitizedTitle),
                comparisonTitle: sanitizedTitle,
            })
        })
    }

    return (
        <div className="space-y-2">
            {props.label && (
                <div className="flex items-center">
                    <div className="text-base font-semibold">{props.label}</div>
                </div>
            )}
            {fields.map((field, index) => (
                <MediaFieldItem
                    key={field.id}
                    field={field}
                    index={index}
                    allMedia={props.allMedia}
                    libraryPath={props.libraryPath}
                    form={props.form}
                    onFieldChange={handleFieldChange}
                    onRemove={() => remove(index)}
                    separatorText={index < fields.length - 1 ? props.separatorText : undefined}
                />
            ))}
            <div className="flex gap-2 flex-wrap">
                <IconButton
                    intent="success"
                    className="rounded-full"
                    onClick={() => append({
                        mediaId: 0,
                        destination: props.libraryPath,
                        comparisonTitle: "",
                    })}
                    icon={<BiPlus />}
                />
                {!entriesAdded?.length && <Button
                    intent="gray-subtle"
                    className="rounded-full"
                    onClick={handleAddCurrentlyWatching}
                    leftIcon={<BiPlus />}
                >
                    All Currently Watching
                </Button>}
                {!entriesAdded?.length && <Button
                    intent="gray-subtle"
                    className="rounded-full"
                    onClick={handleAddUpcoming}
                    leftIcon={<BiPlus />}
                >
                    All Upcoming
                </Button>}
            </div>
        </div>
    )
}

type MediaFieldItemProps = {
    field: MediaEntry & { id: string }
    index: number
    allMedia: AL_BaseAnime[]
    libraryPath: string
    form: UseFormReturn<InferType<typeof schema>>
    onFieldChange: (index: number, updatedValues: Partial<MediaEntry>, field: MediaEntry) => void
    onRemove: () => void
    separatorText?: string
}

function MediaFieldItem(props: MediaFieldItemProps) {
    const {
        field,
        index,
        allMedia,
        libraryPath,
        form,
        onFieldChange,
        onRemove,
        separatorText,
    } = props

    const selectedMedia = React.useMemo(() => {
        return allMedia.find(m => m.id === field.mediaId)
    }, [allMedia, field.mediaId])

    const animeFolderName = React.useMemo(() => {
        return sanitizeDirectoryName(selectedMedia?.title?.userPreferred || "")
    }, [selectedMedia])

    const destination = useWatch({ name: `entries.${index}.destination` }) as string

    const libraryPathSelectionProps = useLibraryPathSelection({
        destination,
        setDestination: path => onFieldChange(index, { destination: path }, field),
        animeFolderName,
    })

    return (
        <div>
            <div className="flex gap-4 items-center w-full">
                <div className="flex flex-col gap-2 w-full">
                    <div className="border rounded-[--radius] p-4 relative space-y-3">
                        <div className="flex gap-4 items-center">
                            <AutoDownloaderMediaCombobox
                                mediaList={allMedia}
                                value={field.mediaId}
                                onValueChange={(v) => onFieldChange(index,
                                    { mediaId: v[0] ? parseInt(v[0]) : allMedia[0]?.id },
                                    field)}
                                type={"create"}
                            />
                        </div>
                        <Field.DirectorySelector
                            name={`entries.${index}.destination`}
                            label="Destination"
                            help="Folder in your local library where the files will be saved"
                            leftIcon={<FcFolder />}
                            shouldExist={false}
                            value={field.destination}
                            defaultValue={libraryPath}
                            libraryPathSelectionProps={libraryPathSelectionProps}
                        />
                        <TextInput
                            label="Comparison title"
                            help="Used for comparison purposes. When using 'Exact match', use a title most likely to be used in a torrent name."
                            {...form.register(`entries.${index}.comparisonTitle`)}
                        />
                    </div>
                </div>
                <CloseButton
                    size="sm"
                    intent="alert-subtle"
                    onClick={onRemove}
                />
            </div>
            {!!separatorText && (
                <p className="text-center text-[--muted] my-4">{separatorText}</p>
            )}
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
