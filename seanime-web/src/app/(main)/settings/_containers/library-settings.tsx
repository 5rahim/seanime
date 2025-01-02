import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { FcFolder } from "react-icons/fc"

type LibrarySettingsProps = {
    isPending: boolean
}

export function LibrarySettings(props: LibrarySettingsProps) {

    const {
        isPending,
        ...rest
    } = props


    return (
        <div className="space-y-4">

            <Field.DirectorySelector
                name="libraryPath"
                label="Library directory"
                leftIcon={<FcFolder />}
                help="Directory where your media is located. (Keep the casing consistent)"
                shouldExist
            />

            <Field.MultiDirectorySelector
                name="libraryPaths"
                label="Additional library directories"
                leftIcon={<FcFolder />}
                help="Include additional directories if your library is spread across multiple locations."
                shouldExist
            />

            <Separator />

            <Field.Switch
                name="autoScan"
                label="Automatically refresh library"
                help={<div>
                    <p>If enabled, your library will be refreshed in the background when new files are added/deleted. Make sure to
                       lock your files regularly.</p>
                    <p>
                        <em>Note:</em> This works best when single files are added/deleted. If you are adding a batch, not all
                                       files
                                       are guaranteed to be picked up.
                    </p>
                </div>}
            />

            <Field.Switch
                name="refreshLibraryOnStart"
                label="Refresh library on startup"
                help={<div>
                    <p>If enabled, your library will be refreshed in the background when the server starts. Make sure to
                       lock your files regularly.</p>
                    <p>
                        <em>Note:</em> Visit the scan summary page to see the results.
                    </p>
                </div>}
            />

            <Separator />

            <Accordion type="single" collapsible>
                <AccordionItem value="more">
                    <AccordionTrigger className="bg-gray-900 rounded-md">
                        Advanced
                    </AccordionTrigger>
                    <AccordionContent className="pt-6 flex flex-col md:flex-row gap-3">
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
                            help="The minimum score required for a file to be matched to an AniList entry."
                            formatOptions={{
                                minimumFractionDigits: 1,
                                maximumFractionDigits: 1,
                            }}
                            step={0.1}
                        />
                    </AccordionContent>
                </AccordionItem>
            </Accordion>

            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}
