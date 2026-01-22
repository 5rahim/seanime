import { Anime_AutoDownloaderProfile, Anime_AutoDownloaderProfileRuleFormatAction } from "@/api/generated/types"
import { useCreateAutoDownloaderProfile, useUpdateAutoDownloaderProfile } from "@/api/hooks/auto_downloader.hooks"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Combobox } from "@/components/ui/combobox"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React from "react"
import { BiMenu, BiPlus, BiTrash } from "react-icons/bi"

type ConditionType = {
    id: string
    term: string
    isRegex: boolean
    action: string
    score: number
}

type ResolutionType = {
    id: string
    value: string
}

type FormData = {
    name: string
    global: boolean
    resolutions: ResolutionType[]
    conditions: ConditionType[]
    minimumScore: number
    minSeeders: number
    minSize: string
    maxSize: string
    providers: string[]
}

const formDataAtom = atomWithImmer<FormData>({
    name: "",
    global: false,
    resolutions: [],
    conditions: [],
    minimumScore: 0,
    minSeeders: 0,
    minSize: "",
    maxSize: "",
    providers: [],
})

type AutoDownloaderProfileFormProps = {
    profile?: Anime_AutoDownloaderProfile
    onSuccess?: () => void
}

export function AutoDownloaderProfileForm(props: AutoDownloaderProfileFormProps) {
    const { profile, onSuccess } = props

    const { mutate: createProfile, isPending: creating } = useCreateAutoDownloaderProfile()
    const { mutate: updateProfile, isPending: updating } = useUpdateAutoDownloaderProfile()

    const [formData, setFormData] = useAtom(formDataAtom)

    React.useEffect(() => {
        setFormData(draft => {
            draft.name = profile?.name ?? ""
            draft.global = profile?.global ?? false
            draft.resolutions = profile?.resolutions?.map(res => ({
                id: `resolution-${Date.now()}-${Math.random()}`,
                value: res,
            })) ?? []
            draft.conditions = profile?.conditions?.map(c => ({
                ...c,
                id: c.id || `condition-${Date.now()}-${Math.random()}`,
            })) ?? []
            draft.minimumScore = profile?.minimumScore ?? 0
            draft.minSeeders = profile?.minSeeders ?? 0
            draft.minSize = profile?.minSize ?? ""
            draft.maxSize = profile?.maxSize ?? ""
            draft.providers = profile?.providers ?? []
            return
        })
    }, [profile])

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        if (!formData.name) {
            return
        }

        const data = {
            ...formData,
            resolutions: formData.resolutions.map(r => r.value),
            conditions: formData.conditions.map(c => ({
                ...c,
                action: c.action as Anime_AutoDownloaderProfileRuleFormatAction,
            })),
        }

        if (profile) {
            updateProfile({
                ...profile,
                ...data,
            }, {
                onSuccess: () => onSuccess?.(),
            })
        } else {
            createProfile({
                dbId: 0,
                ...data,
            }, {
                onSuccess: () => onSuccess?.(),
            })
        }
    }

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
                <label className="text-sm font-medium">Profile Name</label>
                <TextInput
                    value={formData.name}
                    onValueChange={(v) => setFormData(draft => {
                        draft.name = v
                        return
                    })}
                    placeholder="Name"
                    required
                />
            </div>

            <div className="flex items-center gap-2">
                <Switch
                    value={formData.global}
                    onValueChange={(value) => setFormData(draft => {
                        draft.global = value
                        return
                    })}
                />
                <label className="text-sm">
                    Global
                    <span className="text-[--muted] block text-xs">Apply this profile to all rules automatically</span>
                </label>
            </div>

            <Separator />

            <ResolutionsSortableField />

            <ConditionsSortableField />

            <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Thresholds</div>
                <div className="space-y-2">
                    <label className="text-sm font-medium">Minimum Score</label>
                    <TextInput
                        type="number"
                        value={formData.minimumScore}
                        onChange={(e) => setFormData(draft => {
                            draft.minimumScore = parseInt(e.target.value) || 0
                            return
                        })}
                        placeholder="0"
                    />
                    <p className="text-sm text-[--muted]">Torrents with a score lower than this will be rejected</p>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Min Seeders</label>
                        <TextInput
                            type="number"
                            value={formData.minSeeders}
                            onChange={(e) => setFormData(draft => {
                                draft.minSeeders = parseInt(e.target.value) || 0
                                return
                            })}
                            placeholder="0"
                            min={0}
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Min Size</label>
                        <TextInput
                            value={formData.minSize}
                            onChange={(e) => setFormData(draft => {
                                draft.minSize = e.target.value
                                return
                            })}
                            placeholder="e.g., 100MB"
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Max Size</label>
                        <TextInput
                            value={formData.maxSize}
                            onChange={(e) => setFormData(draft => {
                                draft.maxSize = e.target.value
                                return
                            })}
                            placeholder="e.g., 2GB or 10GiB"
                        />
                    </div>
                </div>
            </div>

            <ProvidersFieldControlled />

            <div className="flex justify-end">
                <Button
                    type="submit"
                    intent={profile ? "primary" : "success"}
                    loading={creating || updating}
                >
                    {profile ? "Update" : "Create"}
                </Button>
            </div>
        </form>
    )
}

function ResolutionsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const resolutions = formData.resolutions

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = resolutions.findIndex(item => item.id === active.id)
            const newIndex = resolutions.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.resolutions.splice(oldIndex, 1)
                draft.resolutions.splice(newIndex, 0, movedItem)
                return
            })
        }
    }

    const handleAdd = (value: string) => {
        setFormData(draft => {
            draft.resolutions.push({
                id: `resolution-${Date.now()}-${Math.random()}`,
                value,
            })
            return
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            const index = draft.resolutions.findIndex(r => r.id === id)
            if (index !== -1) {
                draft.resolutions.splice(index, 1)
            }
            return
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.resolutions.findIndex(r => r.id === id)
            if (index !== -1) {
                draft.resolutions[index].value = value
            }
            return
        })
    }

    const suggestions = ["2160p", "1080p", "720p", "540p", "480p"]

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Resolutions</div>
            <p className="text-sm text-[--muted]">
                Drag and drop to reorder. The first matching resolution will be picked.
            </p>

            <div className="flex flex-wrap gap-2 mb-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => handleAdd(suggestion)}
                        disabled={resolutions.some(r => r.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext strategy={verticalListSortingStrategy} items={resolutions.map(r => r.id)}>
                    <div className="space-y-2">
                        {resolutions.map((item) => (
                            <SortableItem key={item.id} id={item.id}>
                                <div className="flex gap-2 items-center w-full">
                                    <TextInput
                                        value={item.value}
                                        onChange={(e) => handleUpdate(item.id, e.target.value)}
                                        className="flex-1"
                                    />
                                    <IconButton
                                        icon={<BiTrash />}
                                        intent="alert-subtle"
                                        onClick={() => handleRemove(item.id)}
                                        type="button"
                                    />
                                </div>
                            </SortableItem>
                        ))}
                    </div>
                </SortableContext>
            </DndContext>
            <Button
                intent="success-subtle"
                leftIcon={<BiPlus />}
                onClick={() => handleAdd("")}
                size="sm"
                type="button"
            >
                Add Resolution
            </Button>
        </div>
    )
}

function ConditionsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const conditions = formData.conditions

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = conditions.findIndex((item) => item.id === active.id)
            const newIndex = conditions.findIndex((item) => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.conditions.splice(oldIndex, 1)
                draft.conditions.splice(newIndex, 0, movedItem)
                return
            })
        }
    }

    const handleAppend = () => {
        const newCondition: ConditionType = {
            id: `condition-${Date.now()}-${Math.random()}`,
            term: "",
            isRegex: false,
            action: "score",
            score: 0,
        }
        setFormData(draft => {
            draft.conditions.push(newCondition)
            return
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            const index = draft.conditions.findIndex(c => c.id === id)
            if (index !== -1) {
                draft.conditions.splice(index, 1)
            }
            return
        })
    }

    const handleUpdateField = <K extends keyof ConditionType>(id: string, fieldName: K, value: ConditionType[K]) => {
        setFormData(draft => {
            const condition = draft.conditions.find(c => c.id === id)
            if (condition) {
                condition[fieldName] = value
            }
            return
        })
    }

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Conditions</div>
            <p className="text-sm text-[--muted]">
                Add conditions to filter torrents or adjust their score.
            </p>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext strategy={verticalListSortingStrategy} items={conditions.map((c) => c.id)}>
                    <div className="space-y-2">
                        {conditions.map((field) => (
                            <ConditionItem
                                key={field.id}
                                field={field}
                                onUpdateField={handleUpdateField}
                                onRemove={() => handleRemove(field.id)}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>
            <Button
                intent="success-subtle"
                leftIcon={<BiPlus />}
                onClick={handleAppend}
                size="sm"
                type="button"
            >
                Add Condition
            </Button>
        </div>
    )
}

type ConditionItemProps = {
    field: ConditionType
    onUpdateField: <K extends keyof ConditionType>(id: string, fieldName: K, value: ConditionType[K]) => void
    onRemove: () => void
}

function ConditionItem(props: ConditionItemProps) {
    const { field, onUpdateField, onRemove } = props

    return (
        <SortableItem id={field.id}>
            <div className="space-y-2 w-full">
                <TextInput
                    value={field.term}
                    onChange={(e) => onUpdateField(field.id, "term", e.target.value)}
                    placeholder="e.g., Blu-Ray, BluRay or \b(group)\bi"
                    className="w-full"
                    help="Comma-separated case insensitive values or regex pattern"
                />
                <div className="space-y-2">
                    <div className="flex items-center gap-4 flex-wrap">
                        <Select
                            value={field.action}
                            onValueChange={(value) => onUpdateField(field.id, "action", value)}
                            options={[
                                { label: "Score", value: "score" },
                                { label: "Block", value: "block" },
                                { label: "Require", value: "require" },
                            ]}
                            label="Action"
                            fieldClass="!flex !items-center gap-2 w-fit"
                            labelProps={{ className: "items-center" }}
                            className="w-32"
                        />
                        {field.action === "score" && (
                            <NumberInput
                                value={field.score}
                                onValueChange={(v) => onUpdateField(field.id, "score", v || 0)}
                                placeholder="Score"
                                label="Score"
                                fieldClass="!flex !items-center gap-2 w-fit"
                                labelProps={{ className: "items-center" }}
                                className="w-32"
                                formatOptions={{ useGrouping: false }}
                                min={-999999}
                                max={999999}
                            />
                        )}
                    </div>
                    <div className="flex items-center gap-4">
                        <Checkbox
                            value={field.isRegex}
                            onValueChange={(value) => onUpdateField(field.id, "isRegex", !!value)}
                            label="Regex"
                        />
                        <IconButton
                            icon={<BiTrash />}
                            intent="alert-subtle"
                            onClick={onRemove}
                            size="sm"
                            type="button"
                        />
                    </div>
                </div>
            </div>
        </SortableItem>
    )
}

function SortableItem({ id, children }: { id: string, children: React.ReactNode }) {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
    } = useSortable({ id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
    }

    return (
        <div ref={setNodeRef} style={style} className="flex items-start gap-2 bg-[--subtle] p-2 rounded-md">
            <div {...attributes} {...listeners} className="cursor-move text-[--muted] hover:text-[--foreground] mt-2">
                <BiMenu className="text-xl" />
            </div>
            {children}
        </div>
    )
}

function ProvidersFieldControlled() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const { data: extensions } = useAnimeListTorrentProviderExtensions()

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Providers</div>
            <p className="text-sm">
                Select specific providers to look for. If empty, the default provider will be used.
            </p>
            <Combobox
                value={formData.providers}
                onValueChange={(value) => setFormData(draft => {
                    draft.providers = value
                    return
                })}
                options={extensions?.map(ext => ({
                    label: ext.name,
                    textValue: ext.name,
                    value: ext.id,
                })) ?? []}
                multiple
                label="Select providers"
                emptyMessage="No providers found"
            />
        </div>
    )
}

