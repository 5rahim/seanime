import { RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button, ButtonProps } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { DatePicker } from "@/components/ui/date-picker"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { RadioGroup } from "@/components/ui/radio-group"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { Textarea } from "@/components/ui/textarea"
import { Tooltip } from "@/components/ui/tooltip"
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


/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface TooltipProps {
    item?: any
    text: string
}

export function PluginTooltip({ item, text }: TooltipProps) {
    if (!item) return
    return (
        <Tooltip
            trigger={<div className="w-fit"><RenderPluginComponents data={item} /></div>}
        >
            {text}
        </Tooltip>
    )
}

///////////////////


interface PluginCSSProps {
    css: string
}

export function PluginCSS(props: PluginCSSProps) {
    const scopeId = React.useId().replace(/:/g, "_")

    // Scope CSS to only apply to siblings and their children
    const scopedCSS = scopeCSSToSiblings(props.css, scopeId)

    return (
        <>
            <style>{scopedCSS}</style>
            <span data-css-scope={scopeId} style={{ display: "none" }} />
        </>
    )
}

// Helper function to scope CSS rules to siblings
function scopeCSSToSiblings(css: string, scopeId: string): string {
    const scope = `[data-css-scope="${scopeId}"]`

    // Parse and scope CSS rules
    // This regex matches CSS selectors before opening braces
    return css.replace(
        /([^{}]+)\{/g,
        (match, selector) => {
            // Clean up the selector
            const cleanSelector = selector.trim()

            // Skip @rules like @media, @keyframes, etc.
            if (cleanSelector.startsWith("@")) {
                return match
            }

            // Split multiple selectors separated by commas
            const selectors = cleanSelector.split(",").map((s: string) => s.trim())

            // Scope each selector to apply only to siblings and their descendants
            const scopedSelectors = selectors.map((sel: string) => {
                // Use general sibling combinator (~) to target all siblings after the scope marker
                return `${scope} ~ * ${sel}, ${scope} ~ ${sel}`
            }).join(", ")

            return `${scopedSelectors} {`
        },
    )
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
    className?: string
}

export function PluginButton(props: PluginButtonProps) {
    const { onClick, label, style, className, size, loading, disabled, intent, ...rest } = props
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()

    function handleClick() {
        if (onClick) {
            sendEventHandlerTriggeredEvent({
                handlerName: onClick,
                event: {},
            }, trayIcon.extensionId)
        }
    }

    return (
        <Button
            intent={intent || "white-subtle"}
            style={style}
            onClick={handleClick}
            disabled={disabled}
            loading={loading}
            size={size || "sm"}
            className={className}
        >
            {label || "Button"}
        </Button>
    )
}

///////////////////

interface PluginAnchorProps {
    text?: string
    href?: string
    target?: string
    onClick?: string
    style?: React.CSSProperties
    className?: string
}

export function PluginAnchor(props: PluginAnchorProps) {
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()

    function handleClick(e: React.MouseEvent) {
        if (props.onClick) {
            e.preventDefault()
            sendEventHandlerTriggeredEvent({
                handlerName: props.onClick,
                event: {
                    href: props.href,
                    text: props.text,
                },
            }, trayIcon.extensionId)
        }
    }

    return (
        <a
            href={props.href}
            target={props.target || "_blank"}
            rel="noopener noreferrer"
            style={props.style}
            onClick={handleClick}
            className={cn("underline cursor-pointer", props.className)}
        >
            {props.text || "Link"}
        </a>
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
    onSelect?: string
    fieldRef?: FieldRef<string>
    disabled?: boolean
    size?: "sm" | "md" | "lg"
    className?: string
    textarea?: boolean
}

export function PluginInput(props: InputProps) {
    const { trayIcon } = usePluginTray()
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { sendFieldRefSendValueEvent } = usePluginSendFieldRefSendValueEvent()
    const [value, setValue] = React.useState(props.value || props.fieldRef?.current)
    const debouncedValue = useDebounce(value, 200)

    const inputRef = React.useRef<HTMLInputElement>(null)
    const textareaRef = React.useRef<HTMLTextAreaElement>(null)

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

    const [selectedText, setSelectedText] = React.useState<{ value: string, cursorStart: number, cursorEnd: number } | null>(null)
    const debouncedSelectedText = useDebounce(selectedText, 400)

    function handleTextSelected(e: any) {
        if (props.onSelect) {
            const cursorStart = props.textarea ? textareaRef.current?.selectionStart : inputRef.current?.selectionStart
            const cursorEnd = props.textarea ? textareaRef.current?.selectionEnd : inputRef.current?.selectionEnd
            const selectedText = props.textarea ? textareaRef.current?.value.slice(cursorStart ?? 0, cursorEnd ?? 0) : inputRef.current?.value.slice(
                cursorStart ?? 0,
                cursorEnd ?? 0)

            setSelectedText({ value: selectedText ?? "", cursorStart: cursorStart ?? 0, cursorEnd: cursorEnd ?? 0 })
        }
    }

    useEffect(() => {
        if (props.onSelect && debouncedSelectedText) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onSelect,
                event: {
                    value: debouncedSelectedText.value,
                    cursorStart: debouncedSelectedText.cursorStart,
                    cursorEnd: debouncedSelectedText.cursorEnd,
                },
            }, trayIcon.extensionId)
        }
    }, [debouncedSelectedText?.value, debouncedSelectedText?.cursorStart, debouncedSelectedText?.cursorEnd])

    if (props.textarea) {
        return (
            <Textarea
                id={props.id}
                label={props.label}
                placeholder={props.placeholder}
                style={props.style}
                value={value}
                onValueChange={(value) => setValue(value)}
                onSelect={handleTextSelected}
                disabled={props.disabled}
                fieldClass={props.className}
                ref={textareaRef}
            />
        )
    }

    return (
        <TextInput
            id={props.id}
            label={props.label}
            placeholder={props.placeholder}
            style={props.style}
            value={value}
            onValueChange={(value) => setValue(value)}
            onSelect={handleTextSelected}
            disabled={props.disabled}
            size={props.size || "md"}
            fieldClass={props.className}
            ref={inputRef}
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
    className?: string
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
            fieldClass={props.className}
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
    className?: string
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
            fieldClass={props.className}
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
    className?: string
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
            onValueChange={(value) => setValue(value)}
            disabled={props.disabled}
            size={props.size || "sm"}
            fieldClass={props.className}
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
    className?: string
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
            fieldClass={props.className}
        />
    )
}


/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface FlexProps {
    items?: any[]
    direction?: "row" | "column"
    gap?: number
    style?: React.CSSProperties
    className?: string
}

export function PluginFlex({ items = [], direction = "row", gap = 2, style, className }: FlexProps) {
    return (
        <div
            className={cn("flex", className)}
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
    className?: string
}

export function PluginStack({ items = [], style, gap = 2, className }: StackProps) {
    return (
        <div
            className={cn("flex", className)}
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
    onClick?: string
    className?: string
}

export function PluginDiv({ items = [], style, onClick, className }: DivProps) {
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()

    function handleClick() {
        if (onClick) {
            sendEventHandlerTriggeredEvent({
                handlerName: onClick,
                event: {},
            }, trayIcon.extensionId)
        }
    }
    return (
        <div
            className={cn("relative", className)}
            style={style}
            onClick={handleClick}
        >
            {items && items.length > 0 && <RenderPluginComponents data={items} />}
        </div>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface TextProps {
    text: string
    style?: React.CSSProperties
    className?: string
}

export function PluginText({ text, style, className }: TextProps) {
    return <p className={cn("w-full break-all", className)} style={style}>{text}</p>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Modal
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginModalProps {
    trigger?: any
    title?: string
    description?: string
    items?: any[]
    footer?: any[]
    open?: boolean
    onOpenChange?: string
    className?: string
}

export function PluginModal(props: PluginModalProps) {
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()
    const [isOpen, setIsOpen] = React.useState(props.open ?? false)

    function handleOpenChange(open: boolean) {
        setIsOpen(open)
        if (props.onOpenChange) {
            sendEventHandlerTriggeredEvent({
                handlerName: props.onOpenChange,
                event: { open },
            }, trayIcon.extensionId)
        }
    }

    return (
        <Modal
            trigger={props.trigger ? <div className="w-fit"><RenderPluginComponents data={props.trigger} /></div> : undefined}
            title={props.title}
            description={props.description}
            open={isOpen}
            onOpenChange={handleOpenChange}
            footer={props.footer ? <RenderPluginComponents data={props.footer} /> : undefined}
            contentClass={props.className}
        >
            {props.items && <RenderPluginComponents data={props.items} />}
        </Modal>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Dropdown Menu
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginDropdownMenuProps {
    trigger?: any
    items?: any[]
    className?: string
}

export function PluginDropdownMenu(props: PluginDropdownMenuProps) {
    return (
        <DropdownMenu
            trigger={props.trigger ? <div className="w-fit"><RenderPluginComponents data={props.trigger} /></div> : <div />}
            className={props.className}
        >
            {props.items && <RenderPluginComponents data={props.items} />}
        </DropdownMenu>
    )
}

interface PluginDropdownMenuItemProps {
    item?: any
    onClick?: string
    disabled?: boolean
    className?: string
}

export function PluginDropdownMenuItem(props: PluginDropdownMenuItemProps) {
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
        <DropdownMenuItem
            onClick={handleClick}
            disabled={props.disabled}
            className={props.className}
        >
            {props.item && <RenderPluginComponents data={props.item} />}
        </DropdownMenuItem>
    )
}

interface PluginDropdownMenuSeparatorProps {
    className?: string
}

export function PluginDropdownMenuSeparator(props: PluginDropdownMenuSeparatorProps) {
    return <DropdownMenuSeparator className={props.className} />
}

interface PluginDropdownMenuLabelProps {
    label?: string
    className?: string
}

export function PluginDropdownMenuLabel(props: PluginDropdownMenuLabelProps) {
    return <DropdownMenuLabel className={props.className}>{props.label}</DropdownMenuLabel>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Popover
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginPopoverProps {
    trigger?: any
    items?: any[]
    className?: string
}

export function PluginPopover(props: PluginPopoverProps) {
    return (
        <Popover
            trigger={props.trigger ? <RenderPluginComponents data={props.trigger} /> : <div />}
            className={props.className}
        >
            {props.items && <RenderPluginComponents data={props.items} />}
        </Popover>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// A (Anchor with items)
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginAProps {
    href?: string
    items?: any[]
    target?: string
    onClick?: string
    style?: React.CSSProperties
    className?: string
}

export function PluginA(props: PluginAProps) {
    const { sendEventHandlerTriggeredEvent } = usePluginSendEventHandlerTriggeredEvent()
    const { trayIcon } = usePluginTray()

    function handleClick(e: React.MouseEvent) {
        if (props.onClick) {
            e.preventDefault()
            sendEventHandlerTriggeredEvent({
                handlerName: props.onClick,
                event: { href: props.href },
            }, trayIcon.extensionId)
        }
    }

    return (
        <a
            href={props.href}
            target={props.target || "_blank"}
            rel="noopener noreferrer"
            style={props.style}
            onClick={handleClick}
            className={cn("underline cursor-pointer", props.className)}
        >
            {props.items && <RenderPluginComponents data={props.items} />}
        </a>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// P (Paragraph with items)
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginPProps {
    items?: any[]
    style?: React.CSSProperties
    className?: string
}

export function PluginP(props: PluginPProps) {
    return (
        <p className={cn("w-full", props.className)} style={props.style}>
            {props.items && <RenderPluginComponents data={props.items} />}
        </p>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Alert
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginAlertProps {
    title?: string
    description?: string
    intent?: "info" | "success" | "warning" | "alert"
    className?: string
}

export function PluginAlert(props: PluginAlertProps) {
    return (
        <Alert
            title={props.title}
            description={props.description}
            intent={props.intent || "info"}
            className={props.className}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Tabs
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginTabsProps {
    defaultValue?: string
    items?: any[]
    className?: string
}

export function PluginTabs(props: PluginTabsProps) {
    return (
        <Tabs defaultValue={props.defaultValue} className={props.className}>
            {props.items && <RenderPluginComponents data={props.items} />}
        </Tabs>
    )
}

interface PluginTabsListProps {
    items?: any[]
    className?: string
}

export function PluginTabsList(props: PluginTabsListProps) {
    return (
        <TabsList className={props.className}>
            {props.items && <RenderPluginComponents data={props.items} />}
        </TabsList>
    )
}

interface PluginTabsTriggerProps {
    value?: string
    item?: any
    className?: string
}

export function PluginTabsTrigger(props: PluginTabsTriggerProps) {
    return (
        <TabsTrigger value={props.value || ""} className={props.className}>
            {props.item && <RenderPluginComponents data={props.item} />}
        </TabsTrigger>
    )
}

interface PluginTabsContentProps {
    value?: string
    items?: any[]
    className?: string
}

export function PluginTabsContent(props: PluginTabsContentProps) {
    return (
        <TabsContent value={props.value || ""} className={props.className}>
            {props.items && <RenderPluginComponents data={props.items} />}
        </TabsContent>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Badge
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginBadgeProps {
    text?: string
    intent?: "gray" | "primary" | "success" | "warning" | "alert" | "info" | "blue"
    size?: "sm" | "md" | "lg" | "xl"
    className?: string
}

export function PluginBadge(props: PluginBadgeProps) {
    return (
        <Badge
            intent={props.intent || "gray"}
            size={props.size || "md"}
            className={props.className}
        >
            {props.text}
        </Badge>
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Span
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

interface PluginSpanProps {
    text: string
    items?: any[]
    style?: React.CSSProperties
    className?: string
}

export function PluginSpan(props: PluginSpanProps) {
    return (
        <span className={props.className} style={props.style}>
            {!!props.text && props.text}
            {props.items && <RenderPluginComponents data={props.items} />}
        </span>
    )
}

///////////////////

interface PluginImgProps {
    src?: string
    alt?: string
    width?: string
    height?: string
    style?: React.CSSProperties
    className?: string
}

export function PluginImg(props: PluginImgProps) {
    return (
        <img
            src={props.src}
            alt={props.alt || ""}
            width={props.width}
            height={props.height}
            style={props.style}
            className={props.className}
        />
    )
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Form
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
