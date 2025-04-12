import { RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { Button, ButtonProps } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DatePicker } from "@/components/ui/date-picker"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { RadioGroup } from "@/components/ui/radio-group"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import React, { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"
import * as z from "zod"
import {
    usePluginListenFieldRefSetValueEvent,
    usePluginListenFormResetEvent,
    usePluginListenFormSetValuesEvent,
    usePluginSendEventHandlerTriggeredEvent,
    usePluginSendFieldRefSendValueEvent,
    usePluginSendFormSubmittedEvent,
} from "../generated/plugin-events"
import { usePluginTray } from "../tray/plugin-tray"

type FieldRef<T> = {
    current: T
    __ID: string
}


///////////////////


interface PluginButtonProps {
    label?: string
    style?: React.CSSProperties
    intent?: ButtonProps["intent"]
    onClick?: string
    disabled?: boolean
    loading?: boolean
    size?: "xs" | "sm" | "md" | "lg"
}

export function PluginButton(props: PluginButtonProps) {
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()

    function handleClick() {
        if (props.onClick) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onClick,
                event: {},
            }, trayIcon.extensionId)
        }
    }

    return (
        <Button
            intent={props.intent || "white-subtle"}
            style={props.style}
            onClick={handleClick}
            disabled={props.disabled}
            loading={props.loading}
            size={props.size || "sm"}
        >
            {props.label || "Button"}
        </Button>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fields
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface InputProps {
    placeholder?: string
    label?: string
    id?: string
    style?: React.CSSProperties
    value?: string
    onChange?: string
    fieldRef?: FieldRef<string>
    disabled?: boolean
    size?: "sm" | "md" | "lg"
}

export function PluginInput(props: InputProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const firstRender = React.useRef(true)
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (props.onChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onChange,
                event: {
                    value: debouncedValue,
                },
            }, trayIcon.extensionId)
        }
        if (props.fieldRef) {
            sendFieldRefSendValueEvent({
                fieldRef: props.fieldRef.__ID,
                value: debouncedValue,
            }, trayIcon.extensionId)
        }
    }, [debouncedValue])

    usePluginListenFieldRefSetValueEvent((data) => {
        if (data.fieldRef === props.fieldRef?.__ID) {
            setValue(data.value)
        }
    }, trayIcon.extensionId)

    return (
        <TextInput
            id={props.id}
            label={props.label}
            placeholder={props.placeholder}
            style={props.style}
            value={value}
            onValueChange={(value) => setValue(value)}
            disabled={props.disabled}
            size={props.size || "md"}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface SelectProps {
    options: Array<{
        label: string
        value: string
    }>
    id?: string
    label?: string
    onChange?: string
    fieldRef?: FieldRef<string>
    style?: React.CSSProperties
    value?: string
    disabled?: boolean
    size?: "sm" | "md" | "lg"
}

export function PluginSelect(props: SelectProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const firstRender = React.useRef(true)
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (props.onChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onChange,
                event: { value: debouncedValue },
            }, trayIcon.extensionId)
        }
        if (props.fieldRef) {
            sendFieldRefSendValueEvent({
                fieldRef: props.fieldRef.__ID,
                value: debouncedValue,
            }, trayIcon.extensionId)
        }
    }, [debouncedValue])

    usePluginListenFieldRefSetValueEvent((data) => {
        if (data.fieldRef === props.fieldRef?.__ID) {
            setValue(data.value)
        }
    }, trayIcon.extensionId)

    return (
        <Select
            id={props.id}
            label={props.label}
            style={props.style}
            options={props.options}
            value={value}
            onValueChange={(value) => setValue(value)}
            disabled={props.disabled}
            size={props.size || "md"}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface CheckboxProps {
    label?: string
    id?: string
    style?: React.CSSProperties
    value?: boolean
    onChange?: string
    fieldRef?: FieldRef<boolean>
    disabled?: boolean
    size?: "sm" | "md" | "lg"
}

export function PluginCheckbox(props: CheckboxProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const firstRender = React.useRef(true)
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (props.onChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onChange,
                event: { value: value },
            }, trayIcon.extensionId)
        }
        if (props.fieldRef) {
            sendFieldRefSendValueEvent({
                fieldRef: props.fieldRef.__ID,
                value: value,
            }, trayIcon.extensionId)
        }
    }, [debouncedValue])

    usePluginListenFieldRefSetValueEvent((data) => {
        if (data.fieldRef === props.fieldRef?.__ID) {
            setValue(data.value)
        }
    }, trayIcon.extensionId)

    return (
        <Checkbox
            id={props.id}
            label={props.label}
            style={props.style}
            value={value}
            onValueChange={(value) => typeof value === "boolean" && setValue(value)}
            disabled={props.disabled}
            size={props.size || "md"}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface SwitchProps {
    label?: string
    id?: string
    style?: React.CSSProperties
    value?: boolean
    onChange?: string
    fieldRef?: FieldRef<boolean>
    disabled?: boolean
    size?: "sm" | "md" | "lg"
    side?: "left" | "right"
}

export function PluginSwitch(props: SwitchProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const firstRender = React.useRef(true)
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (props.onChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onChange,
                event: { value: value },
            }, trayIcon.extensionId)
        }
        if (props.fieldRef) {
            sendFieldRefSendValueEvent({
                fieldRef: props.fieldRef.__ID,
                value: value,
            }, trayIcon.extensionId)
        }
    }, [debouncedValue])

    usePluginListenFieldRefSetValueEvent((data) => {
        if (data.fieldRef === props.fieldRef?.__ID) {
            setValue(data.value)
        }
    }, trayIcon.extensionId)

    return (
        <Switch
            side={props.side || "right"}
            id={props.id}
            label={props.label}
            style={props.style}
            value={value}
            onValueChange={(value) => typeof value === "boolean" && setValue(value)}
            disabled={props.disabled}
            size={props.size || "sm"}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface RadioGroupProps {
    options: Array<{
        label: string
        value: string
    }>
    id?: string
    label?: string
    onChange?: string
    fieldRef?: FieldRef<string>
    style?: React.CSSProperties
    value?: string
    disabled?: boolean
    size?: "sm" | "md" | "lg"
}

export function PluginRadioGroup(props: RadioGroupProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const firstRender = React.useRef(true)
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (props.onChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onChange,
                event: { value: value },
            }, trayIcon.extensionId)
        }
        if (props.fieldRef) {
            sendFieldRefSendValueEvent({
                fieldRef: props.fieldRef.__ID,
                value: value,
            }, trayIcon.extensionId)
        }
    }, [debouncedValue])

    usePluginListenFieldRefSetValueEvent((data) => {
        if (data.fieldRef === props.fieldRef?.__ID) {
            setValue(data.value)
        }
    }, trayIcon.extensionId)

    return (
        <RadioGroup
            id={props.id}
            label={props.label}
            options={props.options}
            value={value}
            onValueChange={(value) => setValue(value)}
            disabled={props.disabled}
            size={props.size || "md"}
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

export function PluginForm({ name, fields: _fields }: FormProps) {
    const { sendPluginMessage } = useWebsocketSender()

    const [fields, setFields] = React.useState(_fields)

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

    const { trayIcon } = usePluginTray()

    const { sendFormSubmittedEvent } = usePluginSendFormSubmittedEvent()

    const onSubmit = (data: FormData) => {
        sendFormSubmittedEvent({
            formName: name,
            data: data,
        }, trayIcon.extensionId)
    }

    usePluginListenFormResetEvent((data) => {
        if (data.formName === name) {
            if (!!data.fieldToReset) {
                form.resetField(data.fieldToReset)
            } else {
                form.reset()
                setFields([])
                setTimeout(() => {
                    setFields(_fields)
                }, 250)
            }
        }
    }, trayIcon.extensionId)

    usePluginListenFormSetValuesEvent((data) => {
        if (data.formName === name) {
            for (const [key, value] of Object.entries(data.data)) {
                form.setValue(key, value)
            }
        }
    }, trayIcon.extensionId)


    return (
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {!fields?.length ? <LoadingSpinner /> :
                fields.map((field) => {
                    if (!field.name && field.type !== "submit") return null // Skip fields without names

                    switch (field.type) {
                        case "input":
                            return (
                                <TextInput
                                    key={field.id}
                                    label={field.label}
                                    placeholder={field.placeholder}
                                    {...form.register(field.name)}
                                    // value={form.watch(field.name)}
                                    // onValueChange={(value) => form.setValue(field.name, value)}
                                />
                            )
                        case "number":
                            return (
                                <TextInput
                                    key={field.id}
                                    type="number"
                                    label={field.label}
                                    placeholder={field.placeholder}
                                    {...form.register(field.name)}
                                    // value={form.watch(field.name)}
                                    // onValueChange={(value) => form.setValue(field.name, Number(value))}
                                />
                            )
                        case "select":
                            return (
                                <Controller
                                    key={field.id}
                                    control={form.control}
                                    name={field.name}
                                    defaultValue={field.value ?? ""}
                                    render={({ field: fField }) => (
                                        <Select
                                            key={field.id}
                                            label={field.label}
                                            name={field.name}
                                            options={field.options?.map(opt => ({
                                                label: opt.label,
                                                value: String(opt.value),
                                            })) ?? []}
                                            placeholder={field.placeholder}
                                            value={fField.value}
                                            onValueChange={(value) => fField.onChange(value)}
                                        />
                                    )}
                                />
                            )
                        case "checkbox":
                            return (
                                <Controller
                                    key={field.id}
                                    control={form.control}
                                    name={field.name}
                                    defaultValue={field.value ?? false}
                                    render={({ field: fField }) => (
                                        <Checkbox
                                            key={field.id}
                                            label={field.label}
                                            value={fField.value}
                                            onValueChange={(value) => fField.onChange(value)}
                                        />
                                    )}
                                />
                            )
                        case "switch":
                            return (
                                <Controller
                                    key={field.id}
                                    control={form.control}
                                    name={field.name}
                                    defaultValue={field.value ?? false}
                                    render={({ field: fField }) => (
                                        <Switch
                                            key={field.id}
                                            label={field.label}
                                            value={fField.value}
                                            onValueChange={(value) => fField.onChange(value)}
                                        />
                                    )}
                                />
                            )
                        case "radio":
                            return (
                                <Controller
                                    key={field.id}
                                    control={form.control}
                                    name={field.name}
                                    defaultValue={field.value ?? ""}
                                    render={({ field: fField }) => (
                                        <RadioGroup
                                            key={field.id}
                                            label={field.label}
                                            name={field.name}
                                            options={field.options?.map(opt => ({
                                                label: opt.label,
                                                value: String(opt.value),
                                            })) ?? []}
                                            value={fField.value}
                                            onValueChange={(value) => fField.onChange(value)}
                                        />
                                    )}
                                />
                            )
                        case "date":
                            return (
                                <Controller
                                    key={field.id}
                                    control={form.control}
                                    name={field.name}
                                    defaultValue={field.value ?? ""}
                                    render={({ field: fField }) => (
                                        <DatePicker
                                            key={field.id}
                                            name={field.name}
                                            label={field.label}
                                            value={fField.value}
                                            onValueChange={(date) => fField.onChange(date)}
                                        />
                                    )}
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
