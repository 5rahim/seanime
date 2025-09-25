import { Models_HomeItem } from "@/api/generated/types"
import { useGetHomeItems, useUpdateHomeItems } from "@/api/hooks/status.hooks"
import { DEFAULT_HOME_ITEMS, HOME_ITEM_IDS, HOME_ITEMS } from "@/app/(main)/(library)/_home/home-items.utils"
import { uuidv4 } from "@/app/websocket-provider"
import { BetaBadge } from "@/components/shared/beta-badge"
import { GlowingEffect } from "@/components/shared/glowing-effect"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { arrayMove, SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { BiCog, BiPlus, BiStats, BiTrash } from "react-icons/bi"
import { FaRegCompass } from "react-icons/fa6"
import { IoHomeOutline, IoLibraryOutline } from "react-icons/io5"
import { LuBookOpen, LuCalendar, LuCalendarClock, LuCirclePlay, LuClock, LuGrid3X3 } from "react-icons/lu"
import { MdOutlineVideoLibrary } from "react-icons/md"
import { toast } from "sonner"

export const __home_settingsModalOpen = atom(false)

const HOME_ITEM_ICONS = {
    "anime-continue-watching": LuCirclePlay,
    "anime-continue-watching-header": LuCirclePlay,
    "anime-library": MdOutlineVideoLibrary,
    "local-anime-library": IoLibraryOutline,
    "library-upcoming-episodes": LuClock,
    "aired-recently": LuCalendarClock,
    "anime-schedule-calendar": LuCalendar,
    "local-anime-library-stats": BiStats,
    "discover-header": FaRegCompass,
    "anime-carousel": FaRegCompass,
    "manga-carousel": FaRegCompass,
    "manga-continue-reading": LuBookOpen,
    "manga-library": LuBookOpen,
} as const

export function HomeSettingsModal() {
    const [isModalOpen, setIsModalOpen] = useAtom(__home_settingsModalOpen)
    const [optionsModalOpen, setOptionsModalOpen] = React.useState<string | null>(null)

    const { data: _homeItems, isLoading: isLoadingHomeItems } = useGetHomeItems()
    const { mutate: updateHomeItems, isPending: isUpdatingHomeItems } = useUpdateHomeItems()

    const [currentItems, setCurrentItems] = React.useState<Models_HomeItem[]>(_homeItems || DEFAULT_HOME_ITEMS)
    const availableItems = HOME_ITEM_IDS.filter(type => {
        if (type === "anime-carousel" || type === "manga-carousel") {
            return true
        }
        return !currentItems.some(item => item.type === type)
    })

    const checkTimeRef = React.useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        const homeItems = _homeItems || DEFAULT_HOME_ITEMS
        setCurrentItems(homeItems)

        if (checkTimeRef.current) {
            clearTimeout(checkTimeRef.current)
            checkTimeRef.current = null
        }

        // Check if an item doesn't exist anymore and remove it
        checkTimeRef.current = setTimeout(() => {
            const newItems = normalizeHomeItems(currentItems)

            if (newItems.length !== homeItems.length) {
                setCurrentItems(newItems)
                updateHomeItems({ items: newItems }, {
                    onSuccess: () => {
                        console.log("Home items updated")
                    },
                })
            }
        }, 500)

        return () => {
            if (checkTimeRef.current) {
                clearTimeout(checkTimeRef.current)
                checkTimeRef.current = null
            }
        }
    }, [_homeItems])

    const handleDragEnd = React.useCallback((event: DragEndEvent) => {
        const { active, over } = event

        if (active.id !== over?.id) {
            const oldIndex = currentItems.findIndex(item => item.id === active.id)
            const newIndex = currentItems.findIndex(item => item.id === over?.id)

            const newItems = normalizeHomeItems(arrayMove(currentItems, oldIndex, newIndex))
            setCurrentItems(newItems)
            updateHomeItems({ items: newItems }, {
                onSuccess: () => {
                    // toast.success("Home items reordered")
                },
            })
        }
    }, [currentItems, updateHomeItems])

    function normalizeHomeItems(items: Models_HomeItem[]) {
        let newItems = items.filter(item => !!HOME_ITEMS[item.type as keyof typeof HOME_ITEMS])
        newItems = newItems.map(item => ({
            ...item,
            schemaVersion: HOME_ITEMS[item.type as keyof typeof HOME_ITEMS].schemaVersion,
        }))
        return newItems
    }

    const handleAddItem = (type: string) => {
        const newItem: Models_HomeItem = {
            id: uuidv4(),
            type,
            schemaVersion: HOME_ITEMS[type as keyof typeof HOME_ITEMS].schemaVersion,
        }

        const newItems = normalizeHomeItems([...currentItems, newItem])
        setCurrentItems(newItems)
        updateHomeItems({ items: newItems }, {
            onSuccess: () => {
                toast.success("Home item added")
            },
        })
    }

    const handleRemoveItem = (id: string) => {
        const newItems = normalizeHomeItems(currentItems.filter(item => item.id !== id))
        setCurrentItems(newItems)
        updateHomeItems({ items: newItems }, {
            onSuccess: () => {
                toast.success("Home item removed")
            },
        })
    }

    const handleUpdateItemOptions = (id: string, options: any) => {
        const newItems = normalizeHomeItems(currentItems.map(item =>
            item.id === id
                ? { ...item, options }
                : item,
        ))
        setCurrentItems(newItems)
        updateHomeItems({ items: newItems }, {
            onSuccess: () => {
                toast.success("Home layout updated")
                setOptionsModalOpen(null)
            },
        })
    }

    return (
        <>
            <Modal
                open={isModalOpen}
                onOpenChange={setIsModalOpen}
                title={<div className="flex items-center gap-2 w-full justify-center">
                    <IoHomeOutline className="size-5" />
                    Home <BetaBadge className="ml-0 mt-0.5" />
                </div>}
                contentClass="max-w-5xl bg-gray-950 bg-opacity-60 backdrop-blur-sm firefox:bg-opacity-100 firefox:backdrop-blur-none sm:rounded-3xl"
                overlayClass="bg-gray-950/70 backdrop-blur-sm"
            >
                <GlowingEffect
                    variant="classic"
                    spread={40}
                    glow={true}
                    disabled={false}
                    proximity={64}
                    inactiveZone={0.01}
                    className="opacity-50 !mt-0"
                />

                <div className="space-y-6">
                    <div>
                        <div className="flex items-center gap-2 mb-4">
                            <LuGrid3X3 className="size-5" />
                            <h4 className="text-lg font-semibold">Home Layout</h4>
                        </div>

                        {isLoadingHomeItems ? (
                            <div className="flex items-center justify-center py-8">
                                <LoadingSpinner />
                            </div>
                        ) : (
                            <DndContext
                                modifiers={[restrictToVerticalAxis]}
                                onDragEnd={handleDragEnd}
                            >
                                <SortableContext items={currentItems.map(item => item.id)} strategy={verticalListSortingStrategy}>
                                    <div className="space-y-2 bg-gray-900/30 rounded-xl p-4 border border-gray-800">
                                        {currentItems.length === 0 ? (
                                            <div className="text-center py-8 text-gray-400">
                                                No items added yet. Add some items below to customize your home screen.
                                            </div>
                                        ) : (
                                            currentItems.map((item) => (
                                                <SortableHomeItem
                                                    key={item.id}
                                                    item={item}
                                                    onRemove={handleRemoveItem}
                                                    onEditOptions={setOptionsModalOpen}
                                                    isUpdating={isUpdatingHomeItems}
                                                />
                                            ))
                                        )}
                                    </div>
                                </SortableContext>
                            </DndContext>
                        )}
                    </div>

                    <Separator />

                    <div>
                        <div className="flex items-center gap-2 mb-4">
                            <BiPlus className="size-5" />
                            <h4 className="text-lg font-semibold">Available Items</h4>
                        </div>

                        {availableItems.length === 0 ? (
                            <div className="text-center py-6 text-gray-400">
                                All available items have been added to your home screen.
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                {availableItems.map((itemType) => (
                                    <AvailableHomeItem
                                        key={itemType}
                                        id={itemType}
                                        type={itemType}
                                        onAdd={handleAddItem}
                                        isUpdating={isUpdatingHomeItems}
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </Modal>

            {optionsModalOpen && (
                <HomeItemOptionsModal
                    id={optionsModalOpen}
                    item={currentItems.find(item => item.id === optionsModalOpen)!}
                    isOpen={!!optionsModalOpen}
                    onClose={() => setOptionsModalOpen(null)}
                    onSave={handleUpdateItemOptions}
                    isUpdating={isUpdatingHomeItems}
                />
            )}
        </>
    )
}

interface SortableHomeItemProps {
    item: Models_HomeItem
    onRemove: (id: string) => void
    onEditOptions: (id: string) => void
    isUpdating: boolean
}

function SortableHomeItem({ item, onRemove, onEditOptions, isUpdating }: SortableHomeItemProps) {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
    } = useSortable({ id: item.id })

    const style = {
        transform: CSS.Transform.toString(transform ? { ...transform, scaleY: 1 } : null),
        transition,
    }

    const homeItemConfig = HOME_ITEMS[item.type as keyof typeof HOME_ITEMS]
    const Icon = HOME_ITEM_ICONS[item.type as keyof typeof HOME_ITEM_ICONS] || IoHomeOutline

    if (!homeItemConfig) return null

    return (
        <div
            ref={setNodeRef}
            style={style}
            {...attributes}
            {...listeners}
            className="flex items-center gap-3 p-3 bg-gray-800/50 rounded-xl border border-gray-700 hover:border-gray-600 transition-colors cursor-move"
        >

            <div className="p-2 bg-gray-700/50 rounded-lg">
                <Icon className="size-5 text-gray-300" />
            </div>

            <div className="flex-1">
                <div className="font-medium text-white">{homeItemConfig.name}{!!item.options?.name && `: ${item.options.name}`}</div>
                <p className="text-xs text-[--muted] line-clamp-1">
                    {homeItemConfig.description}
                </p>
                <div className="text-sm text-gray-400">
                    {homeItemConfig.kind.map(k => k.charAt(0).toUpperCase() + k.slice(1)).join(", ")}
                </div>
            </div>

            <div className="flex items-center gap-1">
                {(homeItemConfig as any).options && (
                    <IconButton
                        intent="gray-subtle"
                        size="sm"
                        onClick={() => onEditOptions(item.id)}
                        disabled={isUpdating}
                        className="hover:bg-blue-500/20 hover:text-blue-400"
                        icon={<BiCog className="size-4" />}
                        onPointerDown={(e) => e.stopPropagation()}
                    />
                )}

                <IconButton
                    intent="gray-subtle"
                    size="sm"
                    onClick={() => onRemove(item.id)}
                    disabled={isUpdating}
                    className="hover:bg-red-500/20 hover:text-red-400"
                    icon={<BiTrash className="size-4" />}
                    onPointerDown={(e) => e.stopPropagation()}
                />
            </div>
        </div>
    )
}

interface AvailableHomeItemProps {
    id: string
    type: string
    onAdd: (id: string) => void
    isUpdating: boolean
}

function AvailableHomeItem({ id, type, onAdd, isUpdating }: AvailableHomeItemProps) {
    const homeItemConfig = HOME_ITEMS[type as keyof typeof HOME_ITEMS]
    const Icon = HOME_ITEM_ICONS[type as keyof typeof HOME_ITEM_ICONS] || IoHomeOutline

    if (!homeItemConfig) return null

    return (
        <div className="flex items-center gap-3 p-3 bg-gray-900/30 rounded-xl border border-gray-800 hover:border-gray-700 transition-colors group">
            <div className="p-2 bg-gray-800/50 rounded-lg group-hover:bg-gray-700/50 transition-colors">
                <Icon className="size-5 text-gray-400 group-hover:text-gray-300 transition-colors" />
            </div>

            <div className="flex-1">
                <div className="font-medium text-white">{homeItemConfig.name}</div>
                <p className="text-xs text-[--muted]">
                    {homeItemConfig.description}
                </p>
                <div className="text-sm text-gray-400">
                    {homeItemConfig.kind.map(k => k.charAt(0).toUpperCase() + k.slice(1)).join(", ")}
                </div>
            </div>

            <Button
                intent="primary-subtle"
                size="sm"
                onClick={() => onAdd(type)}
                disabled={isUpdating}
                leftIcon={<BiPlus />}
            >
                Add
            </Button>
        </div>
    )
}

interface HomeItemOptionsModalProps {
    id: string
    item: Models_HomeItem
    isOpen: boolean
    onClose: () => void
    onSave: (id: string, options: any) => void
    isUpdating: boolean
}

function HomeItemOptionsModal({ id, item, isOpen, onClose, onSave, isUpdating }: HomeItemOptionsModalProps) {
    const homeItemConfig = HOME_ITEMS[item.type as keyof typeof HOME_ITEMS]
    const [formData, setFormData] = React.useState<Record<string, any>>(item.options || {})

    React.useEffect(() => {
        if (!homeItemConfig || homeItemConfig.schemaVersion !== item.schemaVersion) {
            setFormData({})
            return
        }
        setFormData(item.options || {})
    }, [item.options, homeItemConfig])

    if (!(homeItemConfig as any)?.options) return null

    const handleFieldChange = (fieldName: string, value: any) => {
        setFormData(prev => ({
            ...prev,
            [fieldName]: value,
        }))
    }

    const handleSave = () => {
        onSave(id, formData)
    }

    return (
        <Modal
            open={isOpen}
            onOpenChange={onClose}
            title={
                <div className="flex items-center gap-2">
                    <BiCog className="size-5" />
                    Configure {homeItemConfig.name}
                </div>
            }
            contentClass="max-w-2xl bg-gray-950 bg-opacity-60 backdrop-blur-sm firefox:bg-opacity-100 firefox:backdrop-blur-none sm:rounded-3xl"
            overlayClass="bg-gray-950/70 backdrop-blur-sm"
        >
            <div className="space-y-6">
                <div className="text-sm text-gray-400">
                    Customize the settings for this home item.
                </div>

                <div className="space-y-4">
                    {((homeItemConfig as any).options || []).map((option: any) => (
                        <OptionField
                            key={option.name}
                            option={option}
                            value={formData[option.name]}
                            onChange={(value) => handleFieldChange(option.name, value)}
                        />
                    ))}
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                    <Button
                        intent="gray-subtle"
                        onClick={onClose}
                        disabled={isUpdating}
                    >
                        Cancel
                    </Button>
                    <Button
                        intent="primary"
                        onClick={handleSave}
                        loading={isUpdating}
                    >
                        Save Changes
                    </Button>
                </div>
            </div>
        </Modal>
    )
}

interface OptionFieldProps {
    option: any
    value: any
    onChange: (value: any) => void
}

function OptionField({ option, value, onChange }: OptionFieldProps) {
    const { label, type, name, options, min, max } = option

    const handleMultiSelectChange = (selectedValue: string) => {
        const currentValues = Array.isArray(value) ? value : []
        const newValues = currentValues.includes(selectedValue)
            ? currentValues.filter((v: any) => v !== selectedValue)
            : [...currentValues, selectedValue]
        onChange(newValues)
    }

    switch (type) {
        case "text":
            return (
                <div className="space-y-2">
                    <label className="text-sm font-medium text-white">{label}</label>
                    <TextInput
                        value={value || ""}
                        onChange={(e) => onChange(e.target.value)}
                        placeholder={`Enter ${label.toLowerCase()}`}
                    />
                </div>
            )

        case "number":
            return (
                <div className="space-y-2">
                    <label className="text-sm font-medium text-white">{label}</label>
                    <NumberInput
                        value={value || min || 0}
                        onValueChange={(valueAsNumber) => onChange(valueAsNumber)}
                        min={min}
                        max={max}
                    />
                </div>
            )

        case "select":
            return (
                <div className="space-y-2">
                    <label className="text-sm font-medium text-white">{label}</label>
                    <Select
                        value={value || ""}
                        onValueChange={onChange}
                        placeholder={`Select ${label.toLowerCase()}`}
                        options={[
                            ...options,
                        ]}
                    />
                </div>
            )

        case "multi-select":
            const selectedValues = Array.isArray(value) ? value : []
            return (
                <div className="space-y-2">
                    <label className="text-sm font-medium text-white">{label}</label>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-2 max-h-48 overflow-y-auto p-3 bg-gray-900/30 rounded-lg border border-gray-800">
                        {options.map((opt: any) => (
                            <button
                                key={opt.value}
                                type="button"
                                onClick={() => handleMultiSelectChange(opt.value)}
                                className={cn(
                                    "p-2 text-sm rounded-md border transition-colors text-left",
                                    selectedValues.includes(opt.value)
                                        ? "bg-brand-500/20 border-brand-500 text-brand-300"
                                        : "bg-gray-800/50 border-gray-700 text-gray-300 hover:border-gray-600",
                                )}
                            >
                                {opt.label}
                            </button>
                        ))}
                    </div>
                    {selectedValues.length > 0 && (
                        <div className="text-xs text-gray-400">
                            {selectedValues.length} selected: {selectedValues.slice(0, 3).join(", ")}
                            {selectedValues.length > 3 && ` +${selectedValues.length - 3} more`}
                        </div>
                    )}
                </div>
            )

        default:
            return (
                <div className="text-sm text-gray-400">
                    Unsupported field type: {type}
                </div>
            )
    }
}
