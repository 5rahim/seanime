import { Anime_AutoSelectPreference, Anime_AutoSelectProfile } from "@/api/generated/types"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useDeleteAutoSelectProfile, useGetAutoSelectProfile, useSaveAutoSelectProfile } from "@/api/hooks/torrent_search.hooks"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Combobox } from "@/components/ui/combobox"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"
import { BiMenu, BiPlus, BiTrash } from "react-icons/bi"
import { LuSettings2 } from "react-icons/lu"
import { toast } from "sonner"

type SortableItem = {
    id: string
    value: string
}

type FormData = {
    providers: SortableItem[]
    releaseGroups: SortableItem[]
    resolutions: SortableItem[]
    excludeTerms: SortableItem[]
    preferredLanguages: SortableItem[]
    preferredCodecs: SortableItem[]
    preferredSources: SortableItem[]
    multipleAudioPreference: Anime_AutoSelectPreference
    multipleSubsPreference: Anime_AutoSelectPreference
    batchPreference: Anime_AutoSelectPreference
    bestReleasePreference: Anime_AutoSelectPreference
    requireLanguage: boolean
    requireCodec: boolean
    requireSource: boolean
    minSeeders: number
    minSize: string
    maxSize: string
}

const formDataAtom = atomWithImmer<FormData>({
    providers: [],
    releaseGroups: [],
    resolutions: [],
    excludeTerms: [],
    preferredLanguages: [],
    preferredCodecs: [],
    preferredSources: [],
    bestReleasePreference: "neutral",
    multipleAudioPreference: "neutral",
    multipleSubsPreference: "neutral",
    batchPreference: "neutral",
    requireLanguage: false,
    requireCodec: false,
    requireSource: false,
    minSeeders: 0,
    minSize: "",
    maxSize: "",
})

export function AutoSelectProfileButton() {
    const [isOpen, setIsOpen] = React.useState(false)
    const { data: profile, isLoading } = useGetAutoSelectProfile()

    const hasProfile = !!profile?.dbId

    return (
        <>
            <div className="space-y-2">
                <Button
                    intent={hasProfile ? "white-subtle" : "primary-subtle"}
                    size="md"
                    className="rounded-full"
                    onClick={() => setIsOpen(true)}
                    loading={isLoading}
                    leftIcon={<LuSettings2 />}
                >
                    Customize auto-select
                </Button>
                {hasProfile && (
                    <div className="text-sm text-[--muted] space-y-1">
                        {profile.resolutions && profile.resolutions.length > 0 && (
                            <p><strong>Resolutions:</strong> {profile.resolutions.join(", ")}</p>
                        )}
                        {profile.releaseGroups && profile.releaseGroups.length > 0 && (
                            <p><strong>Groups:</strong> {profile.releaseGroups.slice(0, 3).join(", ")}{profile.releaseGroups.length > 3 ? "..." : ""}
                            </p>
                        )}
                        {profile.providers && profile.providers.length > 0 && (
                            <p><strong>Providers:</strong> {profile.providers.join(", ")}</p>
                        )}
                        {profile.preferredCodecs && profile.preferredCodecs.length > 0 && (
                            <p><strong>Preferred Codecs:</strong> {profile.preferredCodecs.join(", ")}</p>
                        )}
                        {profile.preferredSources && profile.preferredSources.length > 0 && (
                            <p><strong>Preferred Sources:</strong> {profile.preferredSources.join(", ")}</p>
                        )}
                        {profile.preferredLanguages && profile.preferredLanguages.length > 0 && (
                            <p><strong>Preferred Languages:</strong> {profile.preferredLanguages.join(", ")}</p>
                        )}
                        {profile.multipleAudioPreference && profile.multipleAudioPreference !== "neutral" && (
                            <p><strong>Multi Audio:</strong> {capitalize(profile.multipleAudioPreference)}</p>
                        )}
                        {profile.multipleSubsPreference && profile.multipleSubsPreference !== "neutral" && (
                            <p><strong>Multi Subs:</strong> {capitalize(profile.multipleSubsPreference)}</p>
                        )}
                        {profile.bestReleasePreference && profile.bestReleasePreference !== "neutral" && (
                            <p><strong>Best Releases:</strong> {capitalize(profile.bestReleasePreference)}</p>
                        )}
                        {profile.batchPreference && profile.batchPreference !== "neutral" && (
                            <p><strong>Batches:</strong> {capitalize(profile.batchPreference)}</p>
                        )}
                    </div>
                )}
            </div>

            <Modal
                open={isOpen}
                onOpenChange={setIsOpen}
                title="Auto-select"
                contentClass="max-w-3xl"
            >
                <AutoSelectProfileForm
                    profile={profile}
                    onSuccess={() => {
                        setIsOpen(false)
                        toast.success("Settings saved")
                    }}
                />
            </Modal>
        </>
    )
}

type AutoSelectProfileFormProps = {
    profile?: Anime_AutoSelectProfile
    onSuccess?: () => void
}

function AutoSelectProfileForm(props: AutoSelectProfileFormProps) {
    const { profile, onSuccess } = props

    const { mutate: saveProfile, isPending: saving } = useSaveAutoSelectProfile()
    const { mutate: deleteProfile, isPending: deleting } = useDeleteAutoSelectProfile()

    const [formData, setFormData] = useAtom(formDataAtom)

    React.useEffect(() => {
        if (profile) {
            setFormData(draft => {
                draft.providers = profile.providers?.map(p => ({
                    id: `provider-${Date.now()}-${Math.random()}`,
                    value: p,
                })) || []
                draft.releaseGroups = profile.releaseGroups?.map(rg => ({
                    id: `releaseGroup-${Date.now()}-${Math.random()}`,
                    value: rg,
                })) || []
                draft.resolutions = profile.resolutions?.map(res => ({
                    id: `resolution-${Date.now()}-${Math.random()}`,
                    value: res,
                })) || []
                draft.excludeTerms = profile.excludeTerms?.map(term => ({
                    id: `excludeTerm-${Date.now()}-${Math.random()}`,
                    value: term,
                })) || []
                draft.preferredLanguages = profile.preferredLanguages?.map(lang => ({
                    id: `language-${Date.now()}-${Math.random()}`,
                    value: lang,
                })) || []
                draft.preferredCodecs = profile.preferredCodecs?.map(codec => ({
                    id: `codec-${Date.now()}-${Math.random()}`,
                    value: codec,
                })) || []
                draft.preferredSources = profile.preferredSources?.map(source => ({
                    id: `source-${Date.now()}-${Math.random()}`,
                    value: source,
                })) || []
                draft.bestReleasePreference = profile.bestReleasePreference || "neutral"
                draft.multipleAudioPreference = profile.multipleAudioPreference || "neutral"
                draft.multipleSubsPreference = profile.multipleSubsPreference || "neutral"
                draft.batchPreference = profile.batchPreference || "neutral"
                draft.requireLanguage = profile.requireLanguage || false
                draft.requireCodec = profile.requireCodec || false
                draft.requireSource = profile.requireSource || false
                draft.minSeeders = profile.minSeeders || 0
                draft.minSize = profile.minSize || ""
                draft.maxSize = profile.maxSize || ""
            })
        }
    }, [profile, setFormData])

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        // Filter out empty values
        const filterEmpty = (items: SortableItem[]) => items.filter(item => item.value.trim() !== "").map(item => item.value)

        saveProfile({
            profile: {
                dbId: profile?.dbId || 0,
                providers: filterEmpty(formData.providers),
                releaseGroups: filterEmpty(formData.releaseGroups),
                resolutions: filterEmpty(formData.resolutions),
                excludeTerms: filterEmpty(formData.excludeTerms),
                preferredLanguages: filterEmpty(formData.preferredLanguages),
                preferredCodecs: filterEmpty(formData.preferredCodecs),
                preferredSources: filterEmpty(formData.preferredSources),
                bestReleasePreference: formData.bestReleasePreference,
                multipleAudioPreference: formData.multipleAudioPreference,
                multipleSubsPreference: formData.multipleSubsPreference,
                batchPreference: formData.batchPreference,
                requireLanguage: formData.requireLanguage,
                requireCodec: formData.requireCodec,
                requireSource: formData.requireSource,
                minSeeders: formData.minSeeders,
                minSize: formData.minSize,
                maxSize: formData.maxSize,
            },
        }, {
            onSuccess: () => {
                onSuccess?.()
            },
        })
    }

    const handleDelete = () => {
        deleteProfile(undefined, {
            onSuccess: () => {
                onSuccess?.()
            },
        })
    }

    return (
        <form onSubmit={handleSubmit} className="space-y-4">

            <Separator />

            <ProvidersFieldControlled />

            <Separator />

            <div className="border rounded-[--radius] p-4 relative space-y-3">
                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">
                    Release Preferences
                </div>

                <ReleaseGroupsSortableField />
                <ResolutionsSortableField />
                <ExcludeTermsSortableField />
            </div>

            <div className="border rounded-[--radius] p-4 relative space-y-3">
                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">
                    Metadata Preferences
                </div>

                <PreferredLanguagesSortableField />

                <div className="flex items-center gap-2">
                    <Checkbox
                        value={formData.requireLanguage}
                        onValueChange={(checked) => setFormData(draft => {
                            draft.requireLanguage = checked === true
                        })}
                        label="Require language match"
                        help="Reject if no preferred language is found"
                    />
                </div>

                <Separator />

                <PreferredCodecsSortableField />

                <div className="flex items-center gap-2">
                    <Checkbox
                        value={formData.requireCodec}
                        onValueChange={(checked) => setFormData(draft => {
                            draft.requireCodec = checked === true
                        })}
                        label="Require codec match"
                        help="Reject if no preferred codec is found"
                    />
                </div>

                <Separator />

                <PreferredSourcesSortableField />

                <div className="flex items-center gap-2">
                    <Checkbox
                        value={formData.requireSource}
                        onValueChange={(checked) => setFormData(draft => {
                            draft.requireSource = checked === true
                        })}
                        label="Require source match"
                        help="Reject if no preferred source is found"
                    />
                </div>
            </div>

            <div className="border rounded-[--radius] p-4 relative space-y-3">
                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">
                    Special Preferences
                </div>

                <PreferenceField
                    label="Multiple Audio"
                    value={formData.multipleAudioPreference}
                    onChange={(value) => setFormData(draft => {
                        draft.multipleAudioPreference = value
                    })}
                />

                <PreferenceField
                    label="Multiple Subtitles"
                    value={formData.multipleSubsPreference}
                    onChange={(value) => setFormData(draft => {
                        draft.multipleSubsPreference = value
                    })}
                />

                <PreferenceField
                    label="Batches"
                    value={formData.batchPreference}
                    onChange={(value) => setFormData(draft => {
                        draft.batchPreference = value
                    })}
                />
                <PreferenceField
                    label="Best Releases"
                    without={["only"]}
                    value={formData.bestReleasePreference}
                    onChange={(value) => setFormData(draft => {
                        draft.bestReleasePreference = value
                    })}
                />
            </div>

            <div className="border rounded-[--radius] p-4 relative space-y-3">
                <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">
                    Thresholds
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Min Seeders</label>
                        <NumberInput
                            value={formData.minSeeders}
                            onChange={(e) => setFormData(draft => {
                                draft.minSeeders = parseInt(e.target.value) || 0
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
                            })}
                            placeholder="e.g. 100MB"
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Max Size</label>
                        <TextInput
                            value={formData.maxSize}
                            onChange={(e) => setFormData(draft => {
                                draft.maxSize = e.target.value
                            })}
                            placeholder="e.g. 2GB or 10GiB"
                        />
                    </div>
                </div>
            </div>

            <div className="flex justify-between gap-2">
                {profile && (
                    <Button
                        type="button"
                        intent="alert-basic"
                        onClick={handleDelete}
                        loading={deleting}
                    >
                        Reset all
                    </Button>
                )}
                <div className="flex-1" />
                <Button
                    type="submit"
                    intent="success"
                    loading={saving}
                >
                    Save
                </Button>
            </div>
        </form>
    )
}

// Sortable components
function SortableTextItem(props: {
    id: string
    value: string
    onUpdate: (id: string, value: string) => void
    onRemove: (id: string) => void
    placeholder?: string
}) {
    const { id, value, onUpdate, onRemove } = props
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.5 : 1,
    }

    return (
        <div
            ref={setNodeRef}
            style={style}
            className="flex items-center gap-2 bg-gray-900 p-2 rounded-lg"
        >
            <IconButton
                {...attributes}
                {...listeners}
                icon={<BiMenu />}
                size="sm"
                intent="gray-basic"
                className="cursor-grab active:cursor-grabbing"
            />
            <TextInput
                value={value}
                onChange={(e) => onUpdate(id, e.target.value)}
                className="flex-1"
                placeholder={props.placeholder}
            />
            <IconButton
                icon={<BiTrash />}
                size="sm"
                intent="alert-basic"
                onClick={() => onRemove(id)}
            />
        </div>
    )
}

function SortableItem(props: { id: string; value: string; onRemove: (id: string) => void }) {
    const { id, value, onRemove } = props
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.5 : 1,
    }

    return (
        <div
            ref={setNodeRef}
            style={style}
            className="flex items-center gap-2 bg-gray-900 p-2 rounded-lg"
        >
            <IconButton
                {...attributes}
                {...listeners}
                icon={<BiMenu />}
                size="sm"
                intent="gray-basic"
                className="cursor-grab active:cursor-grabbing"
            />
            <span className="flex-1">{value}</span>
            <IconButton
                icon={<BiTrash />}
                size="sm"
                intent="alert-basic"
                onClick={() => onRemove(id)}
            />
        </div>
    )
}

function ProvidersFieldControlled() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const { data: extensions } = useAnimeListTorrentProviderExtensions()
    const [comboboxValue, setComboboxValue] = React.useState<string[]>([])

    // Filter to only include providers that are still installed
    const availableProviderIds = React.useMemo(() => {
        return new Set(extensions?.map(e => e.id) ?? [])
    }, [extensions])

    const items = formData.providers.filter(p => availableProviderIds.has(p.value))

    const availableOptions = React.useMemo(() => {
        return (extensions?.map(e => ({ label: e.name, textValue: e.name, value: e.id }))
            ?.filter(o => !items?.map(p => p.id)?.includes(o.value)) ?? [])
    }, [extensions, items])

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.providers.splice(oldIndex, 1)
                draft.providers.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = (values: string[]) => {
        const newValue = values[0]
        if (newValue && !items.some(item => item.value === newValue)) {
            if (items.length < 3) {
                setFormData(draft => {
                    draft.providers.push({
                        id: `provider-${Date.now()}-${Math.random()}`,
                        value: newValue,
                    })
                })
                setComboboxValue([])
            }
        }
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.providers = draft.providers.filter(item => item.id !== id)
        })
    }

    // Filter out providers that are no longer installed
    React.useEffect(() => {
        setFormData(draft => {
            draft.providers = draft.providers.filter(p => availableProviderIds.has(p.value))
        })
    }, [availableProviderIds, setFormData])

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Providers (ordered, max 3)</label>
            <p className="text-sm text-[--muted]">Select up to 3 providers in order of priority</p>
            <Combobox
                value={comboboxValue}
                onValueChange={handleAdd}
                options={availableOptions}
                multiple
                label="Add provider"
                disabled={items.length >= 3}
                emptyMessage="No providers found"
            />
            {items.length > 0 && (
                <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                    <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                        <div className="space-y-2">
                            {items.map((item) => (
                                <SortableItem
                                    key={item.id}
                                    id={item.id}
                                    value={extensions?.find(e => e.id === item.value)?.name || "N/A"}
                                    onRemove={handleRemove}
                                />
                            ))}
                        </div>
                    </SortableContext>
                </DndContext>
            )}
        </div>
    )
}

function ReleaseGroupsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.releaseGroups

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.releaseGroups.splice(oldIndex, 1)
                draft.releaseGroups.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.releaseGroups.push({
                id: `releaseGroup-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.releaseGroups = draft.releaseGroups.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.releaseGroups.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.releaseGroups[index].value = value
            }
        })
    }

    const suggestions = ["SubsPlease", "Erai-raws", "EMBER", "Judas", "ASW"]

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Release Groups (ordered)</label>
            <p className="text-sm text-[--muted]">Preferred groups in order of priority</p>

            <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => {
                            setFormData(draft => {
                                draft.releaseGroups.push({
                                    id: `releaseGroup-${Date.now()}-${Math.random()}`,
                                    value: suggestion,
                                })
                            })
                        }}
                        disabled={items.some(item => item.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function ResolutionsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.resolutions

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.resolutions.splice(oldIndex, 1)
                draft.resolutions.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.resolutions.push({
                id: `resolution-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.resolutions = draft.resolutions.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.resolutions.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.resolutions[index].value = value
            }
        })
    }

    const suggestions = ["1080p", "720p", "480p"]

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Resolutions (ordered)</label>
            <p className="text-sm text-[--muted]">Preferred resolutions in order of priority</p>

            <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => {
                            setFormData(draft => {
                                draft.resolutions.push({
                                    id: `resolution-${Date.now()}-${Math.random()}`,
                                    value: suggestion,
                                })
                            })
                        }}
                        disabled={items.some(item => item.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function ExcludeTermsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.excludeTerms

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.excludeTerms.splice(oldIndex, 1)
                draft.excludeTerms.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.excludeTerms.push({
                id: `excludeTerm-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.excludeTerms = draft.excludeTerms.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.excludeTerms.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.excludeTerms[index].value = value
            }
        })
    }

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Exclude Terms</label>
            <p className="text-sm text-[--muted]">Exclude torrents with these terms</p>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                placeholder="e.g. CamRip, Cam RIP"
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function PreferredLanguagesSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.preferredLanguages

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.preferredLanguages.splice(oldIndex, 1)
                draft.preferredLanguages.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.preferredLanguages.push({
                id: `language-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.preferredLanguages = draft.preferredLanguages.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.preferredLanguages.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.preferredLanguages[index].value = value
            }
        })
    }

    const suggestions = ["fr, french", "it, ita, italian", "de, ger, german", "es, spa, spanish"]

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Preferred Languages (ordered)</label>
            <p className="text-sm text-[--muted]">Ordered list of preferred languages</p>

            <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => {
                            setFormData(draft => {
                                draft.preferredLanguages.push({
                                    id: `language-${Date.now()}-${Math.random()}`,
                                    value: suggestion,
                                })
                            })
                        }}
                        disabled={items.some(item => item.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function PreferredCodecsSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.preferredCodecs

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.preferredCodecs.splice(oldIndex, 1)
                draft.preferredCodecs.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.preferredCodecs.push({
                id: `codec-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.preferredCodecs = draft.preferredCodecs.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.preferredCodecs.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.preferredCodecs[index].value = value
            }
        })
    }

    const suggestions = ["HEVC, x265, H.265, 10-bit, 10 bit, 10bit"]

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Preferred Codecs (ordered)</label>
            <p className="text-sm text-[--muted]">Ordered list of preferred codecs (comma-separated alternatives)</p>

            <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => {
                            setFormData(draft => {
                                draft.preferredCodecs.push({
                                    id: `codec-${Date.now()}-${Math.random()}`,
                                    value: suggestion,
                                })
                            })
                        }}
                        disabled={items.some(item => item.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function PreferredSourcesSortableField() {
    const [formData, setFormData] = useAtom(formDataAtom)
    const items = formData.preferredSources

    const onDragEnd = (event: DragEndEvent) => {
        const { active, over } = event
        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item.id === active.id)
            const newIndex = items.findIndex(item => item.id === over?.id)

            setFormData(draft => {
                const [movedItem] = draft.preferredSources.splice(oldIndex, 1)
                draft.preferredSources.splice(newIndex, 0, movedItem)
            })
        }
    }

    const handleAdd = () => {
        setFormData(draft => {
            draft.preferredSources.push({
                id: `source-${Date.now()}-${Math.random()}`,
                value: "",
            })
        })
    }

    const handleRemove = (id: string) => {
        setFormData(draft => {
            draft.preferredSources = draft.preferredSources.filter(item => item.id !== id)
        })
    }

    const handleUpdate = (id: string, value: string) => {
        setFormData(draft => {
            const index = draft.preferredSources.findIndex(item => item.id === id)
            if (index !== -1) {
                draft.preferredSources[index].value = value
            }
        })
    }

    const suggestions = ["BDRip, BD RIP, BluRay, Blu-Ray, Blu Ray, BD", "AT-X"]

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">Preferred Sources (ordered)</label>
            <p className="text-sm text-[--muted]">Ordered list of preferred sources (comma-separated alternatives)</p>

            <div className="flex flex-wrap gap-2">
                {suggestions.map((suggestion) => (
                    <Button
                        key={suggestion}
                        intent="gray-subtle"
                        size="sm"
                        className="rounded-full"
                        onClick={() => {
                            setFormData(draft => {
                                draft.preferredSources.push({
                                    id: `source-${Date.now()}-${Math.random()}`,
                                    value: suggestion,
                                })
                            })
                        }}
                        disabled={items.some(item => item.value === suggestion)}
                        type="button"
                    >
                        {suggestion}
                    </Button>
                ))}
            </div>

            <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={onDragEnd}>
                <SortableContext items={items.map(item => item.id)} strategy={verticalListSortingStrategy}>
                    <div className="space-y-2">
                        {items.map((item) => (
                            <SortableTextItem
                                key={item.id}
                                id={item.id}
                                value={item.value}
                                onUpdate={handleUpdate}
                                onRemove={handleRemove}
                            />
                        ))}
                    </div>
                </SortableContext>
            </DndContext>

            <Button
                type="button"
                intent="success-subtle"
                size="sm"
                onClick={handleAdd}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

function PreferenceField(props: {
    label: string
    without?: string[]
    value: Anime_AutoSelectPreference
    onChange: (value: Anime_AutoSelectPreference) => void
}) {
    const { label, value, onChange, without = [] } = props

    const options = [
        { label: "Neutral", value: "neutral" },
        { label: "Prefer", value: "prefer" },
        { label: "Avoid", value: "avoid" },
        { label: "Only", value: "only" },
        { label: "Never", value: "never" },
    ].filter(({ value }) => !without.includes(value as Anime_AutoSelectPreference))

    return (
        <div className="space-y-2">
            <label className="text-sm font-medium">{label}</label>
            <Select
                value={value}
                onValueChange={(v) => onChange(v as Anime_AutoSelectPreference)}
                options={options}
            />
        </div>
    )
}
