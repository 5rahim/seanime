import { useSuperUpdateLocalFiles } from "@/api/generated/library_explorer.hooks"
import { Anime_LocalFileMetadata, Anime_LocalFileType, LibraryExplorer_FileTreeNodeJSON } from "@/api/generated/types"
import {
    libraryExplorer_selectedPathsAtom,
    libraryExplorer_superUpdateDrawerOpenAtom,
} from "@/app/(main)/_features/library-explorer/library-explorer.atoms"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { Vaul, VaulContent } from "@/components/vaul"
import { useAtom, useAtomValue } from "jotai"
import React, { useMemo, useState } from "react"
import { AiOutlineExclamationCircle } from "react-icons/ai"
import { BiListCheck, BiPlus, BiTrash } from "react-icons/bi"
import { toast } from "sonner"

const superUpdateSchema = defineSchema(({ z }) => z.object({
    searchText: z.string(),
    replaceText: z.string(),
    useRegex: z.boolean().default(false),
    caseSensitive: z.boolean().default(false),
    matchAllOccurrences: z.boolean().default(false),
    enumerateItems: z.boolean().default(false),
    editMetadata: z.boolean().default(false),
}))

type SuperUpdateFormData = typeof superUpdateSchema._type

type MetadataEditType = "episode" | "anidb" | "type"

type MetadataEdit = {
    id: string
    type: MetadataEditType
    searchText: string
    anidbSearchText?: string // For AniDB episode filtering
    replaceText: string
    useRegex?: boolean // Only for anidb type
    caseSensitive?: boolean // Only for anidb type
}

type TextFormattingOption = "none" | "lowercase" | "uppercase" | "titlecase" | "capitalize"

type PreviewItem = {
    originalPath: string
    originalName: string
    newName: string
    willChange: boolean
    originalMetadata?: Anime_LocalFileMetadata
    newMetadata?: Anime_LocalFileMetadata
    metadataWillChange: boolean
}

type LibraryExplorerSuperUpdateDrawerProps = {
    fileNodes: LibraryExplorer_FileTreeNodeJSON[]
}

function applyEnumerationPattern(text: string, index: number): string {
    // Replace ${} with simple counter starting from 0
    text = text.replace(/\$\{\}/g, index.toString())

    // Handle combination of parameters
    text = text.replace(/\$\{([^}]+)\}/g, (match, params) => {
        // Parse parameters from the pattern
        const paramMap: { [key: string]: number } = {}
        const paramPairs = params.split(";")

        for (const pair of paramPairs) {
            const [key, value] = pair.split("=")
            if (key && value) {
                const trimmedKey = key.trim()
                const numValue = parseInt(value.trim())
                if (!isNaN(numValue)) {
                    paramMap[trimmedKey] = numValue
                }
            }
        }

        // Extract parameters with defaults
        const padding = paramMap.padding || 0
        const increment = paramMap.increment || 1
        const start = paramMap.start || 0

        // Calculate the final number
        const num = start + (index * increment)

        // Apply padding if specified
        if (padding > 0) {
            return num.toString().padStart(padding, "0")
        } else {
            return num.toString()
        }
    })

    return text
}

function validateFileName(fileName: string, originalName: string): string {
    if (!fileName || fileName.trim() === "") {
        return originalName // Don't allow empty names
    }

    // Separate the filename and extension
    const lastDotIndex = fileName.lastIndexOf(".")
    let nameWithoutExt = fileName
    let extension = ""

    if (lastDotIndex > 0 && lastDotIndex < fileName.length - 1) {
        nameWithoutExt = fileName.substring(0, lastDotIndex)
        extension = fileName.substring(lastDotIndex) // includes the dot
    }

    // Remove or replace invalid characters for most file systems
    const sanitized = nameWithoutExt
        .replace(/[<>:"/\\|?*]/g, "_") // Replace invalid characters with underscore
        .replace(/\.\./g, "_") // Replace double dots
        .trim()

    // Ensure the name isn't just dots or spaces
    if (sanitized === "" || /^[.\s]*$/.test(sanitized)) {
        return originalName
    }

    // Prevent names that are too long
    const maxNameLength = 250 - extension.length
    if (sanitized.length > maxNameLength && maxNameLength > 0) {
        const truncated = sanitized.substring(0, maxNameLength).trim()
        if (truncated === "" || /^[.\s]*$/.test(truncated)) {
            return originalName
        }
        return truncated + extension
    }

    return sanitized + extension
}

function validateEpisodeNumber(episode: number | string): number {
    if (typeof episode === "string") {
        const parsed = parseInt(episode)
        return isNaN(parsed) ? 0 : Math.max(0, parsed) // Default to 0, minimum 0
    }
    if (typeof episode === "number") {
        return isNaN(episode) ? 0 : Math.max(0, Math.floor(episode)) // Default to 0, minimum 0, integer only
    }
    return 0
}

function validateAniDBEpisode(anidbEpisode: string): string {
    // AniDB episodes can be empty, but trim whitespace
    return anidbEpisode ? anidbEpisode.trim() : ""
}

function validateFileType(type: string): Anime_LocalFileType {
    const validTypes: Anime_LocalFileType[] = ["main", "special", "nc"]
    return validTypes.includes(type as Anime_LocalFileType) ? type as Anime_LocalFileType : "nc"
}

// parse AniDB episode format (e.g., "1", "12", "S1", "C2", "T1")
function parseAniDBEpisode(anidbEpisode: string): { prefix: string, number: number } {
    const match = anidbEpisode.match(/^([a-zA-Z]*)(\d+)$/)
    if (match) {
        return {
            prefix: match[1].toUpperCase(),
            number: parseInt(match[2], 10),
        }
    }
    // If no match, try to parse as just a number
    const num = parseInt(anidbEpisode, 10)
    if (!isNaN(num)) {
        return { prefix: "", number: num }
    }
    return { prefix: "", number: 0 }
}

function parseEpisodeSearch(searchText: string): {
    episodeConditions: Array<{ operator: string, value: number }>,
    anidbConditions: Array<{ operator: string, value: number, prefix?: string }>,
    types?: Anime_LocalFileType[],
    excludeTypes?: Anime_LocalFileType[]
} {
    const conditions = searchText.split(";").map(s => s.trim())
    const result: {
        episodeConditions: Array<{ operator: string, value: number }>,
        anidbConditions: Array<{ operator: string, value: number, prefix?: string }>,
        types?: Anime_LocalFileType[],
        excludeTypes?: Anime_LocalFileType[]
    } = {
        episodeConditions: [],
        anidbConditions: [],
    }

    for (const condition of conditions) {
        // Handle type conditions
        if (condition.startsWith("type=")) {
            const typeString = condition.substring(5)
            result.types = typeString.split("|").filter(t =>
                ["main", "special", "nc"].includes(t),
            ) as Anime_LocalFileType[]
        } else if (condition.startsWith("!type=")) {
            const typeString = condition.substring(6)
            result.excludeTypes = typeString.split("|").filter(t =>
                ["main", "special", "nc"].includes(t),
            ) as Anime_LocalFileType[]
        }
        // Handle AniDB episode conditions
        else if (condition.startsWith("anidb>=")) {
            const anidbStr = condition.substring(7)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: ">=", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("anidb<=")) {
            const anidbStr = condition.substring(7)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: "<=", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("anidb>")) {
            const anidbStr = condition.substring(6)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: ">", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("anidb<")) {
            const anidbStr = condition.substring(6)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: "<", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("anidb!=")) {
            const anidbStr = condition.substring(7)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: "!=", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("!anidb=")) {
            const anidbStr = condition.substring(7)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: "!=", value: parsed.number, prefix: parsed.prefix })
        } else if (condition.startsWith("anidb=")) {
            const anidbStr = condition.substring(6)
            const parsed = parseAniDBEpisode(anidbStr)
            result.anidbConditions.push({ operator: "=", value: parsed.number, prefix: parsed.prefix })
        }
        // Handle episode number conditions
        else if (condition.startsWith(">=")) {
            const value = parseInt(condition.substring(2))
            if (!isNaN(value)) result.episodeConditions.push({ operator: ">=", value })
        } else if (condition.startsWith("<=")) {
            const value = parseInt(condition.substring(2))
            if (!isNaN(value)) result.episodeConditions.push({ operator: "<=", value })
        } else if (condition.startsWith(">")) {
            const value = parseInt(condition.substring(1))
            if (!isNaN(value)) result.episodeConditions.push({ operator: ">", value })
        } else if (condition.startsWith("<")) {
            const value = parseInt(condition.substring(1))
            if (!isNaN(value)) result.episodeConditions.push({ operator: "<", value })
        } else if (condition.startsWith("!=")) {
            const value = parseInt(condition.substring(2))
            if (!isNaN(value)) result.episodeConditions.push({ operator: "!=", value })
        } else if (condition.startsWith("!")) {
            const value = parseInt(condition.substring(1))
            if (!isNaN(value)) result.episodeConditions.push({ operator: "!=", value })
        } else if (condition.startsWith("=")) {
            const value = parseInt(condition.substring(1))
            if (!isNaN(value)) result.episodeConditions.push({ operator: "=", value })
        } else if (/^\d+$/.test(condition)) {
            // Direct number means equality
            const value = parseInt(condition)
            if (!isNaN(value)) result.episodeConditions.push({ operator: "=", value })
        }
    }

    return result
}

function episodeMatchesSearch(episode: number,
    type: Anime_LocalFileType,
    anidbEpisode: string,
    searchCriteria: ReturnType<typeof parseEpisodeSearch>,
): boolean {
    // Check episode conditions
    for (const condition of searchCriteria.episodeConditions) {
        switch (condition.operator) {
            case ">=":
                if (episode < condition.value) return false
                break
            case "<=":
                if (episode > condition.value) return false
                break
            case ">":
                if (episode <= condition.value) return false
                break
            case "<":
                if (episode >= condition.value) return false
                break
            case "=":
                if (episode !== condition.value) return false
                break
            case "!=":
                if (episode === condition.value) return false
                break
        }
    }

    // Check AniDB episode conditions
    for (const condition of searchCriteria.anidbConditions) {
        const parsedAniDB = parseAniDBEpisode(anidbEpisode)

        // If condition has a prefix, check prefix match first
        if (condition.prefix !== undefined && condition.prefix !== parsedAniDB.prefix) {
            return false
        }

        // Check numeric condition
        switch (condition.operator) {
            case ">=":
                if (parsedAniDB.number < condition.value) return false
                break
            case "<=":
                if (parsedAniDB.number > condition.value) return false
                break
            case ">":
                if (parsedAniDB.number <= condition.value) return false
                break
            case "<":
                if (parsedAniDB.number >= condition.value) return false
                break
            case "=":
                if (parsedAniDB.number !== condition.value) return false
                break
            case "!=":
                if (parsedAniDB.number === condition.value) return false
                break
        }
    }

    // Check type inclusion
    if (searchCriteria.types && !searchCriteria.types.includes(type)) return false

    // Check type exclusion
    if (searchCriteria.excludeTypes && searchCriteria.excludeTypes.includes(type)) return false

    return true
}

function applyEpisodeReplacement(episode: number, replaceText: string, index: number): number {
    let newEpisode = episode

    if (replaceText.startsWith("increment=")) {
        const increment = parseInt(replaceText.substring(10))
        if (!isNaN(increment)) {
            newEpisode = episode + increment
        }
    } else if (replaceText.startsWith("decrement=")) {
        const decrement = parseInt(replaceText.substring(10))
        if (!isNaN(decrement)) {
            newEpisode = episode - decrement
        }
    } else {
        const directValue = parseInt(replaceText)
        if (!isNaN(directValue)) {
            newEpisode = directValue
        }
    }

    return validateEpisodeNumber(newEpisode)
}

function applyMetadataEdits(
    originalMetadata: Anime_LocalFileMetadata | undefined,
    metadataEdits: MetadataEdit[],
    index: number,
): Anime_LocalFileMetadata | undefined {
    if (!originalMetadata || metadataEdits.length === 0) return originalMetadata

    // Store the changes that each rule would make, with later rules taking priority
    let episodeChange: number | undefined = undefined
    let anidbEpisodeChange: string | undefined = undefined
    let typeChange: Anime_LocalFileType | undefined = undefined

    // Evaluate each rule against the original metadata
    for (const edit of metadataEdits) {
        switch (edit.type) {
            case "episode": {
                const searchCriteria = parseEpisodeSearch(edit.searchText)
                if (episodeMatchesSearch(originalMetadata.episode, originalMetadata.type, originalMetadata.aniDBEpisode, searchCriteria)) {
                    episodeChange = applyEpisodeReplacement(originalMetadata.episode, edit.replaceText, index)
                }
                break
            }
            case "anidb": {
                // First check if file matches the AniDB search criteria
                let matchesSearch = true
                if (edit.anidbSearchText && edit.anidbSearchText.trim()) {
                    const anidbSearchCriteria = parseEpisodeSearch(edit.anidbSearchText)
                    matchesSearch = episodeMatchesSearch(originalMetadata.episode,
                        originalMetadata.type,
                        originalMetadata.aniDBEpisode,
                        anidbSearchCriteria)
                }

                if (matchesSearch) {
                    try {
                        if (edit.useRegex) {
                            const flags = edit.caseSensitive ? "g" : "gi"
                            const regex = new RegExp(edit.searchText, flags)
                            if (regex.test(originalMetadata.aniDBEpisode)) {
                                let newAniDB = originalMetadata.aniDBEpisode.replace(regex, edit.replaceText)
                                // Apply enumeration if patterns are found
                                newAniDB = applyEnumerationPattern(newAniDB, index)
                                anidbEpisodeChange = validateAniDBEpisode(newAniDB)
                            }
                        } else {
                            const searchStr = edit.caseSensitive ? edit.searchText : edit.searchText.toLowerCase()
                            const origStr = edit.caseSensitive ? originalMetadata.aniDBEpisode : originalMetadata.aniDBEpisode.toLowerCase()
                            if (origStr.includes(searchStr)) {
                                let newAniDB = originalMetadata.aniDBEpisode.replace(
                                    new RegExp(edit.searchText.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), edit.caseSensitive ? "g" : "gi"),
                                    edit.replaceText,
                                )
                                // Apply enumeration if patterns are found
                                newAniDB = applyEnumerationPattern(newAniDB, index)
                                anidbEpisodeChange = validateAniDBEpisode(newAniDB)
                            }
                        }
                    }
                    catch (error) {
                        // Invalid regex, skip this edit
                        console.warn("Invalid regex in AniDB edit:", edit.searchText, error)
                    }
                }
                break
            }
            case "type": {
                const searchCriteria = parseEpisodeSearch(edit.searchText)
                if (episodeMatchesSearch(originalMetadata.episode, originalMetadata.type, originalMetadata.aniDBEpisode, searchCriteria)) {
                    typeChange = validateFileType(edit.replaceText)
                }
                break
            }
        }
    }

    // Apply the final changes (latest rule wins for each field)
    const newMetadata = { ...originalMetadata }
    if (episodeChange !== undefined) {
        newMetadata.episode = episodeChange
    }
    if (anidbEpisodeChange !== undefined) {
        newMetadata.aniDBEpisode = anidbEpisodeChange
    }
    if (typeChange !== undefined) {
        newMetadata.type = typeChange
    }

    return newMetadata
}

function applyTextFormatting(text: string, formatting: TextFormattingOption): string {
    switch (formatting) {
        case "lowercase":
            return text.toLowerCase()
        case "uppercase":
            return text.toUpperCase()
        case "titlecase":
            return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase()
        case "capitalize":
            return text.replace(/\b\w/g, l => l.toUpperCase())
        default:
            return text
    }
}

function getNewFileName(originalName: string, options: SuperUpdateFormData & { textFormatting: TextFormattingOption }, index: number): string {
    // Separate filename and extension
    const lastDotIndex = originalName.lastIndexOf(".")
    let nameWithoutExt = originalName
    let extension = ""

    if (lastDotIndex > 0 && lastDotIndex < originalName.length - 1) {
        nameWithoutExt = originalName.substring(0, lastDotIndex)
        extension = originalName.substring(lastDotIndex) // includes the dot
    }

    let newName = nameWithoutExt

    if (!options.searchText) {
        // Apply enumeration if enabled
        if (options.enumerateItems) {
            newName = applyEnumerationPattern(newName, index)
        }
        // Apply text formatting
        newName = applyTextFormatting(newName, options.textFormatting)
        return newName + extension
    }

    try {
        if (options.useRegex) {
            const flags = options.caseSensitive ? "g" : "gi"
            const regex = new RegExp(options.searchText, flags)

            if (options.matchAllOccurrences) {
                newName = nameWithoutExt.replace(regex, options.replaceText)
            } else {
                newName = nameWithoutExt.replace(regex, options.replaceText)
            }
        } else {
            const searchStr = options.caseSensitive ? options.searchText : options.searchText.toLowerCase()
            const origStr = options.caseSensitive ? nameWithoutExt : nameWithoutExt.toLowerCase()

            if (options.matchAllOccurrences) {
                const parts = origStr.split(searchStr)
                newName = parts.join(options.replaceText)

                // Preserve original casing for non-matched parts
                if (!options.caseSensitive) {
                    let result = ""
                    let searchIndex = 0
                    for (let i = 0; i < parts.length; i++) {
                        if (i > 0) {
                            result += options.replaceText
                        }
                        if (parts[i]) {
                            result += nameWithoutExt.substr(searchIndex, parts[i].length)
                            searchIndex += parts[i].length + (i < parts.length - 1 ? options.searchText.length : 0)
                        }
                    }
                    newName = result
                }
            } else {
                const index = origStr.indexOf(searchStr)
                if (index !== -1) {
                    newName = nameWithoutExt.substring(0, index) + options.replaceText + nameWithoutExt.substring(index + options.searchText.length)
                }
            }
        }

        // Apply enumeration if enabled
        if (options.enumerateItems) {
            newName = applyEnumerationPattern(newName, index)
        }

        // Apply text formatting
        newName = applyTextFormatting(newName, options.textFormatting)

        // Reattach the extension
        return newName + extension
    }
    catch (error) {
        // If regex is invalid, return original name
        return originalName
    }
}

export function LibraryExplorerSuperUpdateDrawer(props: LibraryExplorerSuperUpdateDrawerProps) {
    const { fileNodes } = props

    const [isOpen, setIsOpen] = useAtom(libraryExplorer_superUpdateDrawerOpenAtom)
    const selectedPaths = useAtomValue(libraryExplorer_selectedPathsAtom)
    const [textFormatting, setTextFormatting] = useState<TextFormattingOption>("none")
    const [metadataEdits, setMetadataEdits] = useState<MetadataEdit[]>([])
    const [formData, setFormData] = useState<SuperUpdateFormData>({
        searchText: "",
        replaceText: "",
        useRegex: true,
        caseSensitive: false,
        matchAllOccurrences: false,
        enumerateItems: true,
        editMetadata: false,
    })
    const { mutate: superUpdate, isPending } = useSuperUpdateLocalFiles()

    const selectedFileNodes = useMemo(() => {
        return fileNodes?.filter(n => selectedPaths.has(n.path) && n.kind === "file") || []
    }, [fileNodes, selectedPaths])

    const previewItems = useMemo((): PreviewItem[] => {
        const options = { ...formData, textFormatting }
        return selectedFileNodes.map((node, index) => {
            const originalName = node.name
            const rawNewName = getNewFileName(originalName, options, index)
            const newName = validateFileName(rawNewName, originalName)

            // Apply metadata edits if enabled
            const originalMetadata = node.localFile?.metadata
            const newMetadata = formData.editMetadata
                ? applyMetadataEdits(originalMetadata, metadataEdits, index)
                : originalMetadata

            const metadataWillChange = formData.editMetadata && originalMetadata && newMetadata && (
                originalMetadata.episode !== newMetadata.episode ||
                originalMetadata.aniDBEpisode !== newMetadata.aniDBEpisode ||
                originalMetadata.type !== newMetadata.type
            )

            return {
                originalPath: node.path,
                originalName,
                newName,
                willChange: newName !== originalName,
                originalMetadata,
                newMetadata,
                metadataWillChange: !!metadataWillChange,
            }
        })
    }, [selectedFileNodes, formData, textFormatting, metadataEdits])

    const changedItems = previewItems.filter(item => item.willChange || item.metadataWillChange)

    const handleFormSubmit = (data: SuperUpdateFormData) => {
        // Validate and filter out invalid changes
        const validChanges = changedItems.filter(item => {
            // Check if filename change is valid
            if (item.willChange) {
                if (!item.newName || item.newName.trim() === "" || item.newName === item.originalName) {
                    return false
                }
            }

            // Metadata changes are always valid at this point due to validation functions
            return item.willChange || item.metadataWillChange
        })

        if (validChanges.length === 0) {
            toast.error("No valid changes to apply. Please check your settings.")
            return
        }

        const filesToUpdate = validChanges.map(item => ({
            path: item.originalPath,
            ...(item.willChange && { newName: item.newName }),
            ...(item.metadataWillChange && { metadata: item.newMetadata }),
        }))

        if (validChanges.length < changedItems.length) {
            const skipped = changedItems.length - validChanges.length
            toast.warning(`Applying ${validChanges.length} changes, skipping ${skipped} invalid changes`)
        }

        console.log("filesToUpdate", filesToUpdate)

        superUpdate({
            files: filesToUpdate,
        }, {
            onSuccess: () => {
                const renamedCount = filesToUpdate.filter(f => f.newName).length
                const metadataCount = filesToUpdate.filter(f => f.metadata).length

                let message = "Successfully updated "
                if (renamedCount > 0 && metadataCount > 0) {
                    message += `${renamedCount} filename(s) and ${metadataCount} metadata`
                } else if (renamedCount > 0) {
                    message += `${renamedCount} filename(s)`
                } else if (metadataCount > 0) {
                    message += `${metadataCount} metadata`
                }

                toast.success(message)
                setIsOpen(false)
                // Reset form
                setFormData({
                    searchText: "",
                    replaceText: "",
                    useRegex: true,
                    caseSensitive: false,
                    matchAllOccurrences: false,
                    enumerateItems: true,
                    editMetadata: false,
                })
                setTextFormatting("none")
                setMetadataEdits([])
            },
            onError: (error) => {
                toast.error("Failed to rename files: " + error.message)
            },
        })
    }

    const addMetadataEdit = () => {
        const newEdit: MetadataEdit = {
            id: Date.now().toString(),
            type: "episode",
            searchText: "",
            anidbSearchText: "",
            replaceText: "",
        }
        setMetadataEdits(prev => [...prev, newEdit])
    }

    const updateMetadataEdit = (id: string, updates: Partial<MetadataEdit>) => {
        setMetadataEdits(prev => prev.map(edit =>
            edit.id === id ? { ...edit, ...updates } : edit,
        ))
    }

    const removeMetadataEdit = (id: string) => {
        setMetadataEdits(prev => prev.filter(edit => edit.id !== id))
    }

    return (
        <Vaul
            open={isOpen}
            onOpenChange={setIsOpen}
        >
            <VaulContent
                className={cn(
                    "bg-gray-950 h-[90%] lg:h-[85%] bg-opacity-95 firefox:bg-opacity-100 lg:mx-[2rem]",
                )}
            >
                <p className="p-4 pb-0">
                    <span className="text-sm text-[--muted]">
                        Update multiple file names and metadata at once.
                    </span>
                </p>
                <div className="p-6 flex-1 overflow-hidden flex flex-col">
                    <div className="mb-6">
                        <Form
                            schema={superUpdateSchema}
                            onSubmit={handleFormSubmit}
                            onChange={(data) => setFormData(prev => ({ ...prev, ...data }))}
                            defaultValues={formData}
                            id="super-update-form"
                        >
                            {(form) => (
                                <>
                                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
                                        <Field.Text
                                            name="searchText"
                                            label="Search for"
                                            placeholder="Enter text to search for..."
                                        />
                                        <Field.Text
                                            name="replaceText"
                                            label={<div className="flex items-center gap-1">
                                                <span>Replace with</span>
                                                <Popover className="w-full max-w-2xl"
                                                         trigger={
                                                             <AiOutlineExclamationCircle className="transition-opacity opacity-45 hover:opacity-90" />}
                                                >
                                                    <div className="p-3 bg-gray-800 rounded-md">
                                                        <p className="text-sm text-gray-300 mb-2">Enumeration patterns:</p>
                                                        <div className="text-xs text-gray-400 space-y-1 font-mono">
                                                            <div>${"{}"} - Simple counter (0, 1, 2...)</div>
                                                            <div>${"{start=5}"} - Start from 5 (5, 6, 7...)</div>
                                                            <div>${"{increment=2}"} - Increment by 2 (0, 2, 4...)</div>
                                                            <div>${"{padding=3}"} - Pad with zeros (000, 001, 002...)</div>
                                                            <div>${"{padding=3;start=10}"} - Combined (010, 011, 012...)</div>
                                                            <div>${"{padding=2;increment=5}"} - Pad + increment (00, 05, 10...)</div>
                                                            <div>${"{increment=2;start=1;padding=3}"} - All combined (001, 003, 005...)</div>
                                                        </div>
                                                    </div>
                                                </Popover>
                                            </div>}
                                            placeholder="Enter replacement text..."

                                        />
                                    </div>

                                    <div className="flex flex-wrap gap-4 mb-4">
                                        <Field.Checkbox
                                            name="useRegex"
                                            label="Use regex"
                                            fieldClass="w-fit"
                                        />
                                        <Field.Checkbox
                                            name="caseSensitive"
                                            label="Case sensitive"
                                            fieldClass="w-fit"
                                        />
                                        <Field.Checkbox
                                            name="matchAllOccurrences"
                                            label="Match all occurrences"
                                            fieldClass="w-fit"
                                        />
                                        <Field.Checkbox
                                            name="enumerateItems"
                                            label="Enumerate items"
                                            fieldClass="w-fit"
                                        />
                                    </div>

                                    <div className="mb-4">
                                        <label className="block text-sm font-medium text-gray-300 mb-2">Text formatting</label>
                                        <div className="flex gap-2">
                                            {[
                                                { value: "none", label: "None" },
                                                { value: "lowercase", label: "aa" },
                                                { value: "uppercase", label: "AA" },
                                                { value: "titlecase", label: "Aa" },
                                                { value: "capitalize", label: "Aa Aa" },
                                            ].map((option) => (
                                                <button
                                                    key={option.value}
                                                    type="button"
                                                    className={cn(
                                                        "px-3 py-1 rounded border text-sm font-mono",
                                                        textFormatting === option.value
                                                            ? "bg-brand-500 border-brand-500 text-white"
                                                            : "bg-gray-800 border-gray-600 text-gray-300 hover:bg-gray-700",
                                                    )}
                                                    onClick={() => setTextFormatting(option.value as TextFormattingOption)}
                                                >
                                                    {option.label}
                                                </button>
                                            ))}
                                        </div>
                                    </div>

                                    <div className="mb-4">
                                        <Field.Checkbox
                                            name="editMetadata"
                                            label="Edit file metadata"
                                            fieldClass="w-fit"
                                        />
                                    </div>

                                    {form.watch("editMetadata") && (
                                        <div className="mb-4 p-4 bg-gray-900 border rounded-md max-h-[230px] overflow-y-auto">
                                            <div className="flex items-center justify-between mb-3">
                                                <h4 className="text-sm font-medium text-gray-300"></h4>
                                                <Button
                                                    leftIcon={<BiPlus />}
                                                    intent="gray-outline"
                                                    size="sm"
                                                    onClick={addMetadataEdit}
                                                >
                                                    Add Rule
                                                </Button>
                                            </div>

                                            {metadataEdits.length === 0 ? (
                                                <p className="text-sm text-gray-500 text-center py-4">
                                                    No metadata edit rules. Click "Add Rule" to create one.
                                                </p>
                                            ) : (
                                                <div className="space-y-4">
                                                    {metadataEdits.map((edit, index) => (
                                                        <MetadataEditRule
                                                            key={edit.id}
                                                            edit={edit}
                                                            index={index}
                                                            onUpdate={(updates) => updateMetadataEdit(edit.id, updates)}
                                                            onRemove={() => removeMetadataEdit(edit.id)}
                                                        />
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    )}

                                </>
                            )}
                        </Form>
                    </div>

                    <Separator className="mb-4" />

                    <div className="flex-1 overflow-hidden">
                        <div className="mb-4 flex justify-between items-center">
                            <div className="flex gap-6">
                                <span className="text-sm text-gray-300">
                                    Original ({selectedFileNodes.length})
                                </span>
                                <span className="text-sm text-gray-300">
                                    Renamed ({changedItems.length})
                                </span>
                            </div>
                            <Button
                                type="submit"
                                form="super-update-form"
                                disabled={changedItems.length === 0 || isPending}
                                loading={isPending}
                                leftIcon={<BiListCheck />}
                                intent="white"
                                size="sm"
                            >
                                Apply ({changedItems.length})
                            </Button>
                        </div>

                        <div className={cn("overflow-y-auto flex-1 bg-gray-950 border rounded-md p-2 h-[calc(100%-55px)]")}>
                            {previewItems.length === 0 ? (
                                <div className="text-center text-gray-500 py-8">
                                    No files selected
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    {previewItems.map((item, index) => (
                                        <div
                                            key={item.originalPath}
                                            className={cn(
                                                "flex items-center gap-3 p-2 rounded",
                                                (item.willChange || item.metadataWillChange)
                                                    ? "bg-green-900/20"
                                                    : "bg-gray-800/20 hover:bg-gray-800/30",
                                            )}
                                        >
                                            <div className="flex-1 min-w-0">
                                                <div className="text-sm text-gray-300 truncate tracking-wide select-text">
                                                    {item.originalName}
                                                </div>
                                                {item.willChange && (
                                                    <div className="text-md text-green-200 truncate tracking-wide">
                                                        {item.newName}
                                                    </div>
                                                )}
                                                {item.metadataWillChange && item.originalMetadata && item.newMetadata && (
                                                    <div className="text-xs text-blue-300 mt-1 space-y-1">
                                                        {item.originalMetadata.episode !== item.newMetadata.episode && (
                                                            <div>Episode: {item.originalMetadata.episode} → {item.newMetadata.episode}</div>
                                                        )}
                                                        {item.originalMetadata.aniDBEpisode !== item.newMetadata.aniDBEpisode && (
                                                            <div>AniDB: "{item.originalMetadata.aniDBEpisode}" →
                                                                 "{item.newMetadata.aniDBEpisode}"</div>
                                                        )}
                                                        {item.originalMetadata.type !== item.newMetadata.type && (
                                                            <div>Type: {item.originalMetadata.type} → {item.newMetadata.type}</div>
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                            {(item.willChange || item.metadataWillChange) && (
                                                <div className="text-xs flex-shrink-0 flex flex-col gap-1">
                                                    {item.willChange && (
                                                        <div
                                                            className={cn(
                                                                item.newName === item.originalName ? "text-yellow-500" : "text-green-500",
                                                            )}
                                                        >
                                                            {item.newName === item.originalName ? "No Change" : "Renamed"}
                                                        </div>
                                                    )}
                                                    {item.metadataWillChange && (
                                                        <div className="text-blue-500">Metadata</div>
                                                    )}
                                                </div>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </VaulContent>
        </Vaul>
    )
}

type MetadataEditRuleProps = {
    edit: MetadataEdit
    index: number
    onUpdate: (updates: Partial<MetadataEdit>) => void
    onRemove: () => void
}

function MetadataEditRule({ edit, index, onUpdate, onRemove }: MetadataEditRuleProps) {
    const getSearchPlaceholder = () => {
        switch (edit.type) {
            case "episode":
                return "e.g.: >=1;<=12;!=5;type=main|special;!type=nc"
            case "anidb":
                return "Enter text or regex pattern"
            case "type":
                return "e.g.: >=1;<12;=5;type=main;!type=special"
            default:
                return ""
        }
    }

    const getAnidbSearchPlaceholder = () => {
        return "e.g.: anidb>=1;anidb=S1;anidb!=C2;type=special"
    }

    const getReplacePlaceholder = () => {
        switch (edit.type) {
            case "episode":
                return "e.g.: increment=1, decrement=1, or direct value like 5"
            case "anidb":
                return "Replacement text (supports enumeration patterns)"
            case "type":
                return "Select new type"
            default:
                return ""
        }
    }

    const getSearchHelp = () => {
        switch (edit.type) {
            case "episode":
                return "Operators: >=, <=, >, <, =, !=, ! | Types: type=main|special|nc, !type=special"
            case "anidb":
                return "Supports regex patterns and case sensitivity options"
            case "type":
                return "Same operators as episode: >=, <=, >, <, =, !=, ! | Types: type=main, !type=special"
            default:
                return ""
        }
    }

    const getAnidbSearchHelp = () => {
        return "AniDB operators: anidb>=, anidb<=, anidb=, anidb!=, !anidb= | Format: numbers (1,12) or prefixed (S1,C2,T1)"
    }

    return (
        <div className="p-3 bg-gray-950 rounded-xl border">
            <div className="flex items-center gap-2 mb-3">
                <span className="text-xs font-mono text-gray-400">#{index + 1}</span>
                <Select
                    options={[
                        { value: "episode", label: "Episode Number" },
                        { value: "anidb", label: "AniDB Episode" },
                        { value: "type", label: "File Type" },
                    ]}
                    value={edit.type}
                    onValueChange={(value) => {
                        onUpdate({ type: value as MetadataEditType })
                        if (value === "anidb") {
                            onUpdate({ searchText: "^(.*)$" })
                            onUpdate({ useRegex: true })
                        }
                    }}
                    size="sm"
                />
                <div className="flex-1" />
                <IconButton
                    icon={<BiTrash />}
                    intent="alert-subtle"
                    size="sm"
                    onClick={onRemove}
                />
            </div>

            <div className="space-y-3 mb-3">


                <div className="flex flex-col lg:flex-row gap-3">
                    {edit.type === "anidb" && (
                        <div className="flex-1">
                            <label className="block text-xs font-medium text-gray-300 mb-1">Filter (Optional)</label>
                            <TextInput
                                placeholder={getAnidbSearchPlaceholder()}
                                value={edit.anidbSearchText || ""}
                                onValueChange={(value: string | undefined) => onUpdate({ anidbSearchText: value || "" })}
                                size="sm"
                            />
                            <p className="text-xs text-gray-500 mt-1">{getAnidbSearchHelp()}</p>
                        </div>
                    )}
                    <div className="flex-1">
                        <label className="block text-xs font-medium text-gray-300 mb-1">
                            {edit.type === "anidb" ? "Find (Text/Regex)" : "Search"}
                        </label>
                        <TextInput
                            placeholder={edit.type === "anidb" ? "Enter text or regex pattern" : getSearchPlaceholder()}
                            value={edit.searchText}
                            onValueChange={(value: string | undefined) => onUpdate({ searchText: value || "" })}
                            size="sm"
                        />
                    </div>
                    <div className="flex-1">
                        <label className="block text-xs font-medium text-gray-300 mb-1">Replace</label>
                        {edit.type === "type" ? (
                            <Select
                                options={[
                                    { value: "main", label: "Main" },
                                    { value: "special", label: "Special" },
                                    { value: "nc", label: "NC" },
                                ]}
                                value={edit.replaceText}
                                onValueChange={(value: string | undefined) => onUpdate({ replaceText: value || "" })}
                                placeholder="Select type"
                                size="sm"
                            />
                        ) : (
                            <TextInput
                                placeholder={getReplacePlaceholder()}
                                value={edit.replaceText}
                                onValueChange={(value: string | undefined) => onUpdate({ replaceText: value || "" })}
                                size="sm"
                            />
                        )}
                    </div>
                </div>
            </div>

            {edit.type === "anidb" && (
                <div className="flex gap-4 mb-2">
                    <Checkbox
                        label="Use regex"
                        value={edit.useRegex || false}
                        onValueChange={(value: boolean | "indeterminate") => onUpdate({ useRegex: !!value })}
                        size="sm"
                        labelClass="text-xs text-gray-300"
                    />
                    <Checkbox
                        label="Case sensitive"
                        value={edit.caseSensitive || false}
                        onValueChange={(value: boolean | "indeterminate") => onUpdate({ caseSensitive: !!value })}
                        size="sm"
                        labelClass="text-xs text-gray-300"
                    />
                </div>
            )}

            <p className="text-xs text-gray-500">{getSearchHelp()}</p>
        </div>
    )
}
