import { libraryCollectionAtom } from "@/atoms/collection"
import { CloseButton, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Divider } from "@/components/ui/divider"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { createTypesafeFormSchema, DangerZone, Field, InferType, TypesafeForm } from "@/components/ui/typesafe-form"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { AutoDownloaderRule, LibraryCollection } from "@/lib/server/types"
import { BiPlus } from "@react-icons/all-files/bi/BiPlus"
import { FcFolder } from "@react-icons/all-files/fc/FcFolder"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { uniq } from "lodash"
import Image from "next/image"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import toast from "react-hot-toast"

type RuleFormProps = {
    type: "create" | "edit"
    rule?: AutoDownloaderRule
    onRuleCreatedOrDeleted?: () => void
}

const schema = createTypesafeFormSchema(({ z }) => z.object({
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

export function RuleForm(props: RuleFormProps) {

    const {
        type,
        rule,
        onRuleCreatedOrDeleted,
    } = props

    const qc = useQueryClient()
    const libraryCollection = useAtomValue(libraryCollectionAtom)

    const allMedia = React.useMemo(() => {
        return libraryCollection?.lists?.flatMap(list => list.entries)?.flatMap(entry => entry.media)?.filter(Boolean) ?? []
    }, [(libraryCollection?.lists?.length || 0)])

    // Create a new rule
    const { mutate: createRule, isPending: creatingRule } = useSeaMutation<null, InferType<typeof schema>>({
        mutationKey: ["create-auto-downloader-rule"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_RULE,
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            toast.success("Rule created")
            onRuleCreatedOrDeleted?.()
        },
    })
    // Update a rule
    const { mutate: updateRule, isPending: updatingRule } = useSeaMutation<null, AutoDownloaderRule>({
        mutationKey: ["update-auto-downloader-rule"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_RULE,
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            toast.success("Rule updated")
        },
    })
    // Delete a rule
    const { mutate: deleteRule, isPending: deletingRule } = useSeaMutation({
        mutationKey: ["delete-auto-downloader-rule", rule?.dbId],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_RULE_DETAILS.replace("{id}", String(rule?.dbId)),
        method: "delete",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            toast.success("Rule deleted")
            onRuleCreatedOrDeleted?.()
        },
    })

    const isPending = creatingRule || updatingRule || deletingRule

    function handleSave(data: InferType<typeof schema>) {
        if (data.episodeType === "selected" && data.episodeNumbers.length === 0) {
            return toast.error("You must specify at least one episode number")
        }
        if (type === "create") {
            createRule(data)
        }
        if (type === "edit" && rule?.dbId) {
            updateRule({ ...data, dbId: rule?.dbId })
        }
    }

    if (type === "create" && allMedia.length === 0) {
        return <div className="p-4 text-[--muted] text-center">No media found in your library</div>
    }

    return (
        <div className="space-y-4 mt-8">
            <TypesafeForm
                schema={schema}
                onSubmit={handleSave}
                defaultValues={{
                    enabled: rule?.enabled ?? true,
                    mediaId: rule?.mediaId ?? libraryCollection?.lists?.[0]?.entries?.[0]?.media?.id,
                    releaseGroups: rule?.releaseGroups ?? [],
                    resolutions: rule?.resolutions ?? [],
                    comparisonTitle: rule?.comparisonTitle ?? "",
                    titleComparisonType: rule?.titleComparisonType ?? "likely",
                    episodeType: rule?.episodeType ?? "unwatched",
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
                    libraryCollection={libraryCollection}
                    rule={rule}
                />}
            </TypesafeForm>
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
    allMedia: BaseMediaFragment[]
    type: "create" | "edit"
    isPending: boolean
    libraryCollection: LibraryCollection | undefined
    rule?: AutoDownloaderRule
}

export function RuleFormForm(props: RuleFormFormProps) {

    const {
        form,
        allMedia,
        type,
        isPending,
        libraryCollection,
        rule,
        ...rest
    } = props

    const selectedMedia = allMedia.find(media => media.id === Number(form.watch("mediaId")))

    React.useEffect(() => {
        const id = Number(form.watch("mediaId"))
        const destination = libraryCollection?.lists?.flatMap(list => list.entries)?.find(entry => entry.media?.id === id)?.libraryData?.sharedPath
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

    return (
        <>
            <Field.Switch name="enabled" label="Enabled" />
            <Divider />
            <div
                className={cn(
                    "space-y-3",
                    !form.watch("enabled") && "opacity-50 pointer-events-none",
                )}
            >
                <div className={"flex gap-4 items-end"}>
                    <div
                        className="w-[6rem] h-[6rem] rounded-[--radius] flex-none object-cover object-center overflow-hidden relative bg-gray-800"
                    >
                        {!!selectedMedia?.coverImage?.large && <Image
                            src={selectedMedia.coverImage.large}
                            alt={"banner"}
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
                        options={allMedia.map(media => ({ label: media.title?.userPreferred || "N/A", value: String(media.id) }))}
                        value={String(form.watch("mediaId"))}
                        onChange={(e) => form.setValue("mediaId", parseInt(e.target.value))}
                        help="The anime must already be in your library"
                        isDisabled={type === "edit"}
                    />
                </div>

                <Field.DirectorySelector
                    name="destination"
                    label="Destination"
                    help="The directory to save the files to"
                    leftIcon={<FcFolder />}
                    shouldExist={false}
                />

                <div className="border border-[--border] rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-900 px-2">Title</div>
                    <Field.Text
                        name="comparisonTitle"
                        label="Comparison title"
                        help="The title to compare the torrent name with. Use a title most likely to be found in a torrent name. (e.g. the Romaji title)"
                    />
                    <Field.RadioCards
                        label="Type of search"
                        name="titleComparisonType"
                        options={[
                            { label: "Close match", value: "likely", help: "The torrent name must closely match the title. (More accurate)" },
                            { label: "Contains", value: "contains", help: "The torrent name must contain the title. (Less accurate)" },
                        ]}
                    />
                </div>
                <div className="border border-[--border] rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-900 px-2">Episodes</div>
                    <Field.RadioCards
                        name="episodeType"
                        label="Episodes to look for"
                        options={[
                            { label: "Unwatched", value: "unwatched" },
                            { label: "Any episode", value: "any" },
                            { label: "Specific", value: "selected" },
                        ]}
                        radioLabelClassName="font-normal text-sm flex-none flex"
                    />

                    {form.watch("episodeType") === "selected" && <TextArrayField
                        label="Episode numbers"
                        value={form.watch("episodeNumbers") || []}
                        onChange={(value) => form.setValue("episodeNumbers", value)}
                        type="number"
                    />}
                </div>

                <div className="border border-[--border] rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-900 px-2">Release Groups</div>
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

                <div className="border border-[--border] rounded-[--radius] p-4 relative !mt-8 space-y-3">
                    <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-900 px-2">Resolutions</div>
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
            {type === "create" && <Field.Submit role="create" isLoading={isPending} disableOnSuccess={false} showLoadingOverlayOnSuccess />}
            {type === "edit" && <Field.Submit role="update" isLoading={isPending} />}
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

    React.useEffect(() => {
        console.log(props.value)
    }, [props.value])

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
