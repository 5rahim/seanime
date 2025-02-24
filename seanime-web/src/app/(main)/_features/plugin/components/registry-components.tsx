import { RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DatePicker } from "@/components/ui/date-picker"
import { RadioGroup } from "@/components/ui/radio-group"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import type React from "react"
import { useForm } from "react-hook-form"
import * as z from "zod"
import { usePluginListenFormResetEvent, usePluginSendFormSubmittedEvent } from "../generated/plugin-events"
import { useTrayPlugin } from "../tray/tray-plugin"

interface ButtonProps {
    label?: string
    style?: React.CSSProperties
}

export function PluginButton(props: ButtonProps) {
    return (
        <Button
            intent="white-subtle"
            style={props.style}
        >
            {props.label || "Button"}
        </Button>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface InputProps {
    placeholder?: string
    label?: string
    id?: string
    style?: React.CSSProperties
}

export function PluginInput(props: InputProps) {
    return (
        <TextInput
            id={props.id}
            label={props.label}
            placeholder={props.placeholder}
            style={props.style}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface FlexProps {
    items?: any[]
    direction?: "row" | "column"
    gap?: number
    style?: React.CSSProperties
}

export function PluginFlex({ items = [], direction = "row", gap = 2, style }: FlexProps) {
    return (
        <div
            className="flex"
            style={{
                ...(style || {}),
                gap: `${gap * 0.25}rem`,
                flexDirection: direction,
            }}
        >
            {items && items.length > 0 && <RenderPluginComponents data={items} />}
        </div>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface StackProps {
    items?: any[]
    style?: React.CSSProperties,
    gap?: number
}

export function PluginStack({ items = [], style, gap = 2 }: StackProps) {
    return (
        <div
            className="flex"
            style={{
                ...(style || {}),
                gap: `${gap * 0.25}rem`,
                flexDirection: "column",
            }}
        >
            {items && items.length > 0 && <RenderPluginComponents data={items} />}
        </div>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface DivProps {
    items?: any[]
    style?: React.CSSProperties
}

export function PluginDiv({ items = [], style }: DivProps) {
    return (
        <div
            className="relative"
            style={style}
        >
            {items && items.length > 0 && <RenderPluginComponents data={items} />}
        </div>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface TextProps {
    text: string
    style?: React.CSSProperties
}

export function PluginText({ text, style }: TextProps) {
    return <p className="w-full break-all" style={style}>{text}</p>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface FormProps {
    name: string
    fields: Array<{
        id: string
        type: string
        name: string
        label: string
        placeholder?: string
        value?: any
        options?: Array<{
            label: string
            value: any
        }>
    }>
}

export function PluginForm({ name, fields }: FormProps) {
    const { sendPluginMessage } = useWebsocketSender()

    // Create a dynamic schema based on the fields
    const schema = z.object(
        fields.reduce((acc, field) => {
            if (!field.name) return acc // Skip fields without names

            switch (field.type) {
                case "input":
                    acc[field.name] = z.string().optional()
                    break
                case "number":
                    acc[field.name] = z.number().optional()
                    break
                case "select":
                    acc[field.name] = z.string().optional()
                    break
                case "checkbox":
                    acc[field.name] = z.boolean().optional()
                    break
                case "radio":
                    acc[field.name] = z.string().optional()
                    break
                case "date":
                    acc[field.name] = z.date().optional()
                    break
            }
            return acc
        }, {} as { [key: string]: any }),
    )

    type FormData = z.infer<typeof schema>

    const form = useForm<FormData>({
        // resolver: zodResolver(schema),
        defaultValues: fields.reduce((acc, field) => {
            if (!field.name) return acc // Skip fields without names
            acc[field.name] = field.value ?? ""
            return acc
        }, {} as { [key: string]: any }),
    })

    const { extensionID } = useTrayPlugin()

    const { sendFormSubmittedEvent } = usePluginSendFormSubmittedEvent()

    const onSubmit = (data: FormData) => {
        // console.log("submitted", data)
        sendFormSubmittedEvent({
            formName: name,
            data: data,
        }, extensionID)
    }

    usePluginListenFormResetEvent((data) => {
        if (data.formName === name) {
            if (data.fieldToReset) {
                form.resetField(data.fieldToReset)
            } else {
                form.reset()
            }
        }
    }, extensionID)

    return (
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {fields.map((field) => {
                if (!field.name && field.type !== "submit") return null // Skip fields without names

                const { register } = form
                switch (field.type) {
                    case "input":
                        return (
                            <TextInput
                                key={field.id}
                                label={field.label}
                                placeholder={field.placeholder}
                                {...register(field.name)}
                            />
                        )
                    case "number":
                        return (
                            <TextInput
                                key={field.id}
                                type="number"
                                label={field.label}
                                placeholder={field.placeholder}
                                {...register(field.name, { valueAsNumber: true })}
                            />
                        )
                    case "select":
                        return (
                            <Select
                                key={field.id}
                                label={field.label}
                                options={field.options?.map(opt => ({
                                    label: opt.label,
                                    value: String(opt.value),
                                })) ?? []}
                                placeholder={field.placeholder}
                                value={form.watch(field.name)}
                                onValueChange={(value) => form.setValue(field.name, value)}
                            />
                        )
                    case "checkbox":
                        return (
                            <Checkbox
                                key={field.id}
                                label={field.label}
                                value={form.watch(field.name)}
                                onValueChange={(value) => form.setValue(field.name, value)}
                            />
                        )
                    case "radio":
                        return (
                            <RadioGroup
                                key={field.id}
                                label={field.label}
                                options={field.options?.map(opt => ({
                                    label: opt.label,
                                    value: String(opt.value),
                                })) ?? []}
                                value={form.watch(field.name)}
                                onValueChange={(value) => form.setValue(field.name, value)}
                            />
                        )
                    case "date":
                        return (
                            <DatePicker
                                key={field.id}
                                label={field.label}
                                value={form.watch(field.name)}
                                onValueChange={(date) => form.setValue(field.name, date)}
                            />
                        )
                    case "submit":
                        return (
                            <Button key={field.id} type="submit">
                                {field.label}
                            </Button>
                        )
                    default:
                        return null
                }
            })}
        </form>
    )
}
