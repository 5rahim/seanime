import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { DataSettings } from "@/app/(main)/settings/_containers/data-settings"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { FcFolder } from "react-icons/fc"

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
