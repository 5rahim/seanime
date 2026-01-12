"use client"

import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { DataSettings } from "@/app/(main)/settings/_containers/data-settings"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import {
    DEFAULT_TITLE_PREFERENCE,
    parsePreferredTitleLanguage,
    serializePreferredTitleLanguage,
    TitleLanguage,
    TITLE_LANGUAGE_OPTIONS,
} from "@/lib/helpers/title-preference"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { arrayMove, SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import React from "react"
import { useFormContext, useWatch } from "react-hook-form"
import { FcFolder } from "react-icons/fc"
import { LuGripVertical } from "react-icons/lu"

type LibrarySettingsProps = {
    isPending: boolean
}

export function AnimeLibrarySettings(props: LibrarySettingsProps) {

    const {
        isPending,
        ...rest
    } = props


    return (
        <div className="space-y-4">

            <SettingsCard>
                <Field.DirectorySelector
                    name="libraryPath"
                    label="Library directory"
                    leftIcon={<FcFolder />}
                    help="Path of the directory where your media files ared located. (Keep the casing consistent)"
                    shouldExist
                />

                <Field.MultiDirectorySelector
                    name="libraryPaths"
                    label="Additional library directories"
                    leftIcon={<FcFolder />}
                    help="Include additional directory paths if your library is spread across multiple locations."
                    shouldExist
                />
            </SettingsCard>

            <SettingsCard>

                <Field.Switch
                    side="right"
                    name="autoScan"
                    label="Automatically refresh library"
                    moreHelp={<p>
                        When adding batches, not all files are guaranteed to be picked up.
                    </p>}
                />

                <Field.Switch
                    side="right"
                    name="refreshLibraryOnStart"
                    label="Refresh library on startup"
                />
            </SettingsCard>

            <SettingsCard title="Naming Preferences">
                <p className="text-sm text-[--muted]">
                    Choose the order of title languages used for download folder names.
                    Drag to reorder.
                </p>
                <TitleLanguageSortableList />
            </SettingsCard>

            {/*<SettingsCard title="Advanced">*/}

            <Accordion
                type="single"
                collapsible
                className="border rounded-[--radius-md]"
                triggerClass="dark:bg-[--paper]"
                contentClass="!pt-2 dark:bg-[--paper]"
            >
                <AccordionItem value="more">
                    <AccordionTrigger className="bg-gray-900 rounded-[--radius-md]">
                        Advanced
                    </AccordionTrigger>
                    <AccordionContent className="space-y-4">
                        <div className="flex flex-col md:flex-row gap-3">

                            <Field.Select
                                options={[
                                    { value: "-", label: "Levenshtein + Sorensen-Dice (Default)" },
                                    { value: "sorensen-dice", label: "Sorensen-Dice" },
                                    { value: "jaccard", label: "Jaccard" },
                                ]}
                                name="scannerMatchingAlgorithm"
                                label="Matching algorithm"
                                help="Choose the algorithm used to match files to AniList entries."
                            />
                            <Field.Number
                                name="scannerMatchingThreshold"
                                label="Matching threshold"
                                placeholder="0.5"
                                help="The minimum score required for a file to be matched to an AniList entry. Default is 0.5."
                                formatOptions={{
                                    minimumFractionDigits: 1,
                                    maximumFractionDigits: 1,
                                }}
                                max={1.0}
                                step={0.1}
                            />
                        </div>

                        <Separator />

                        <DataSettings />
                    </AccordionContent>
                </AccordionItem>
            </Accordion>

            {/*</SettingsCard>*/}

            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}

function TitleLanguageSortableList() {
    const form = useFormContext()
    const preferredTitleLanguage = useWatch({ name: "preferredTitleLanguage" }) as string | undefined

    const items = React.useMemo(() => {
        return parsePreferredTitleLanguage(preferredTitleLanguage)
    }, [preferredTitleLanguage])

    const handleDragEnd = React.useCallback((event: DragEndEvent) => {
        const { active, over } = event

        if (active.id !== over?.id) {
            const oldIndex = items.findIndex(item => item === active.id)
            const newIndex = items.findIndex(item => item === over?.id)
            const newItems = arrayMove(items, oldIndex, newIndex)
            form.setValue("preferredTitleLanguage", serializePreferredTitleLanguage(newItems), { shouldDirty: true })
        }
    }, [items, form])

    return (
        <DndContext
            modifiers={[restrictToVerticalAxis]}
            onDragEnd={handleDragEnd}
        >
            <SortableContext items={items} strategy={verticalListSortingStrategy}>
                <div className="space-y-2">
                    {items.map((item, index) => (
                        <SortableTitleItem key={item} id={item} index={index} />
                    ))}
                </div>
            </SortableContext>
        </DndContext>
    )
}

interface SortableTitleItemProps {
    id: TitleLanguage
    index: number
}

function SortableTitleItem({ id, index }: SortableTitleItemProps) {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
    } = useSortable({ id })

    const style = {
        transform: CSS.Transform.toString(transform ? { ...transform, scaleY: 1 } : null),
        transition,
    }

    const label = TITLE_LANGUAGE_OPTIONS.find(opt => opt.id === id)?.label ?? id

    return (
        <div
            ref={setNodeRef}
            style={style}
            {...attributes}
            {...listeners}
            className={cn(
                "flex items-center gap-3 p-3 bg-gray-900 rounded-md border border-gray-700 hover:border-gray-600 transition-colors cursor-move",
            )}
        >
            <LuGripVertical className="text-gray-500" />
            <span className="text-sm text-[--muted] w-4">{index + 1}.</span>
            <span className="font-medium">{label}</span>
        </div>
    )
}
