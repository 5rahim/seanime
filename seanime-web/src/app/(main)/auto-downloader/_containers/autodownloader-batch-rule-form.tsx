import {
    AL_BaseAnime,
    Anime_AutoDownloaderRuleEpisodeType,
    Anime_AutoDownloaderRuleTitleComparisonType,
    Anime_LibraryCollection,
} from "@/api/generated/types"
import { useCreateAutoDownloaderRule } from "@/api/hooks/auto_downloader.hooks"
import { useAnilistUserAnime } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useLibraryCollection } from "@/app/(main)/_hooks/anime-library-collection-loader"
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    AutoDownloaderMediaCombobox,
    TextArrayField,
    useAutoDownloaderMediaList,
} from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { CloseButton, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { upath } from "@/lib/helpers/upath"
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
}

const schema = defineSchema(({ z }) => z.object({
    enabled: z.boolean(),
    entries: z.array(z.object({
        mediaId: z.number(),
        destination: z.string(),
        comparisonTitle: z.string(),
    })).min(1),
    releaseGroups: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    resolutions: z.array(z.string()).transform(value => uniq(value.filter(Boolean))),
    additionalTerms: z.array(z.string()).optional().transform(value => !value?.length ? [] : uniq(value.filter(Boolean))),
    titleComparisonType: z.string(),
}))

export function AutoDownloaderBatchRuleForm(props: AutoDownloaderBatchRuleFormProps) {

    const {
        onRuleCreated,
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
                titleComparisonType: data.titleComparisonType as Anime_AutoDownloaderRuleTitleComparisonType,
                episodeType: "recent" as Anime_AutoDownloaderRuleEpisodeType,
                enabled: data.enabled,
                mediaId: entry.mediaId,
                releaseGroups: data.releaseGroups,
                resolutions: data.resolutions,
                additionalTerms: data.additionalTerms,
                comparisonTitle: entry.comparisonTitle,
                destination: entry.destination,
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
}

function RuleFormFields(props: RuleFormFieldsProps) {

    const {
        form,
        allMedia,
        isPending,
        mediaList,
        libraryCollection,
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
                    label="Library entries"
                    separatorText="AND"
                    form={form}
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

                <Accordion type="single" collapsible className="!my-4">
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
                                    type="text"
                                    placeholder="e.g. H265,H.265,H 265,x265"
                                    separatorText="AND"
                                />
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                </Accordion>

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
