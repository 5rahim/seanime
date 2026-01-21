import { Button, CloseButton, IconButton } from "@/components/ui/button"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { useFieldArray } from "react-hook-form"
import { BiPlus } from "react-icons/bi"

type TextArrayFieldProps<T extends string | number> = {
    name: string
    control: any
    type?: "text" | "number"
    label?: string
    placeholder?: string
    separatorText?: string
    suggestions?: string[]
    suggestionLabels?: string[]
}

export function TextArrayField<T extends string | number>(props: TextArrayFieldProps<T>) {
    const { fields, append, remove } = useFieldArray({
        control: props.control,
        name: props.name,
    })

    return (
        <div className="space-y-2">
            {props.label && <div className="flex items-center">
                <div className="text-base font-semibold">{props.label}</div>
            </div>}
            {!!props.suggestions?.length && (
                <div className="flex flex-wrap gap-2 mb-2">
                    {props.suggestions.map((suggestion, index) => (
                        <Button
                            key={suggestion}
                            intent="gray-subtle"
                            size="sm"
                            className="rounded-full"
                            onClick={() => append(props.type === "number" ? parseInt(suggestion) : suggestion)}
                        >
                            {props.suggestionLabels?.[index] || suggestion}
                        </Button>
                    ))}
                </div>
            )}
            {fields.map((field, index) => (
                <React.Fragment key={field.id}>
                    <div className="flex gap-2 items-center">
                        {props.type === "text" && (
                            <TextInput
                                {...props.control.register(`${props.name}.${index}`)}
                                placeholder={props.placeholder}
                            />
                        )}
                        {props.type === "number" && (
                            <TextInput
                                type="number"
                                {...props.control.register(`${props.name}.${index}`, {
                                    valueAsNumber: true,
                                    min: 1,
                                    validate: (value: number) => !isNaN(value),
                                })}
                            />
                        )}
                        <CloseButton
                            size="sm"
                            intent="alert-subtle"
                            onClick={() => remove(index)}
                        />
                    </div>
                    {(!!props.separatorText && index < fields.length - 1) && (
                        <p className="text-center text-[--muted]">{props.separatorText}</p>
                    )}
                </React.Fragment>
            ))}
            <IconButton
                intent="success-glass"
                className="rounded-full"
                onClick={() => append(props.type === "number" ? 1 : "")}
                icon={<BiPlus />}
            />
        </div>
    )
}

type ReleaseGroupsFieldProps = {
    name: string
    control: any
}

export function ReleaseGroupsField(props: ReleaseGroupsFieldProps) {
    const suggestions = [
        "SubsPlease",
        "Erai-raws",
        "VARYG",
        "EMBER",
        "Judas",
        "ASW",
        "Tsundere-Raws",
    ]

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Release Groups</div>
            <p className="text-sm">
                List of release groups to look for. If empty, any release group will be accepted.
            </p>

            <TextArrayField
                name={props.name}
                control={props.control}
                type="text"
                placeholder="e.g. SubsPlease"
                separatorText="OR"
                suggestions={suggestions}
            />
        </div>
    )
}

type ResolutionsFieldProps = {
    name: string
    control: any
}

export function ResolutionsField(props: ResolutionsFieldProps) {
    const suggestions = ["1080p", "720p", "2160p", "480p"]

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Resolutions</div>
            <p className="text-sm">
                List of resolutions to look for. If empty, the highest resolution will be accepted.
            </p>

            <TextArrayField
                name={props.name}
                control={props.control}
                type="text"
                placeholder="e.g. 1080p"
                separatorText="OR"
                suggestions={suggestions}
            />
        </div>
    )
}

type AdditionalTermsFieldProps = {
    name: string
    control: any
    defaultOpen?: boolean
}

export function AdditionalTermsField(props: AdditionalTermsFieldProps) {
    const suggestions = [
        { label: "HEVC/H.265", value: "H265,H.265,H 265,x265,HEVC" },
        { label: "10-bit", value: "10bit,10-bit,10 bit" },
        { label: "Dual Audio", value: "Dual Audio,DUAL-AUDIO,Dual-Audio" },
        { label: "Multi Audio", value: "Multi Audio,MULTI-AUDIO,Multi-Audio" },
        { label: "FLAC", value: "FLAC" },
        { label: "Hi10P", value: "Hi10P,Hi10,Hi10p" },
        { label: "AVC/H.264", value: "H264,H.264,H 264,x264,AVC" },
        { label: "AAC", value: "AAC,AAC2.0" },
        { label: "E-AC3", value: "E-AC3,EAC3,EAC-3,E-AC-3" },
        { label: "Multi-Sub", value: "Multi-Sub,Multi Sub,Multisub" },
        { label: "WEB-DL", value: "WEB-DL,WEBDL,WEB DL" },
        { label: "BluRay", value: "BluRay,Blu-Ray,BLURAY,BDRip" },
        { label: "Opus", value: "Opus" },
    ]

    return (
        <div className="border rounded-[--radius] p-4 relative !mt-8 space-y-3">
            <div className="absolute -top-2.5 tracking-wide font-semibold uppercase text-sm left-4 bg-gray-950 px-2">Additional
                                                                                                                     terms
            </div>
            <div>
                <p className="text-sm -top-2 relative"><span className="font-semibold">
                    All options must be included for the torrent to be accepted.</span> Within each option, you can
                                                                                        include variations separated by
                                                                                        commas.</p>
            </div>

            <TextArrayField
                name={props.name}
                control={props.control}
                type="text"
                placeholder="e.g. H265,H.265,H 265,x265"
                separatorText="AND"
                suggestions={suggestions.map(s => s.value)}
                suggestionLabels={suggestions.map(s => s.label)}
            />
        </div>
    )
}
