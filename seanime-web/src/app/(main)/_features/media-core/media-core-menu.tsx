import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { Popover } from "@/components/ui/popover"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import { motion } from "motion/react"
import React, { useRef } from "react"
import { AiFillInfoCircle } from "react-icons/ai"
import { LuCheck, LuChevronLeft, LuChevronRight } from "react-icons/lu"
import { MediaCoreSelectOption } from "./media-core.types"

export interface MediaCoreMenuProps {
    name: string
    trigger: React.ReactElement
    children?: React.ReactNode
    className?: string
    sideOffset?: number
    isDrawer?: boolean
    openMenu: string | null
    onOpenMenuChange: (menu: string | null) => void
    onOpenSectionChange?: (section: string | null) => void
    onOpenSubSectionChange?: (subSection: string | null) => void
    isFullscreen?: boolean
    containerElement?: HTMLElement | null
}

export function MediaCoreMenu(props: MediaCoreMenuProps) {
    const {
        trigger,
        children,
        className,
        name,
        sideOffset = 4,
        isDrawer,
        openMenu,
        onOpenMenuChange,
        onOpenSectionChange,
        onOpenSubSectionChange,
        isFullscreen,
        containerElement,
    } = props

    const t = useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (openMenu === null) {
            t.current = setTimeout(() => {
                onOpenSectionChange?.(null)
                onOpenSubSectionChange?.(null)
            }, 300)
        }
        return () => {
            if (t.current) {
                clearTimeout(t.current)
            }
        }
    }, [openMenu])

    if (isDrawer) {
        return (
            <Drawer
                data-vc-element="menu-drawer"
                open={openMenu === name}
                onOpenChange={v => {
                    onOpenMenuChange(v ? name : null)
                }}
                trigger={<div>{trigger}</div>}
                allowOutsideInteraction={true}
                contentClass={cn(
                    "bg-black/85 rounded-xl p-3 backdrop-blur-sm w-[20rem] z-[100]",
                    className,
                )}
                portalContainer={isFullscreen ? containerElement || undefined : undefined}
            >
                <div className="h-auto" data-vc-element="menu-drawer-body">
                    {children}
                </div>
            </Drawer>
        )
    }

    return (
        <Popover
            data-vc-element="menu"
            open={openMenu === name}
            onOpenChange={v => {
                onOpenMenuChange(v ? name : null)
            }}
            trigger={<div>{trigger}</div>}
            sideOffset={sideOffset}
            align="center"
            side="top"
            modal={false}
            className={cn(
                "bg-black/85 rounded-xl p-3 backdrop-blur-sm w-[20rem] z-[100]",
                className,
            )}
            portalContainer={isFullscreen ? containerElement || undefined : undefined}
        >
            <div className="h-auto" data-vc-element="menu-body">
                {children}
            </div>
        </Popover>
    )
}

export function MediaCoreMenuTitle(props: { children: React.ReactNode }) {
    return (
        <div
            data-vc-element="menu-title"
            className="text-white/70 font-bold text-sm pb-3 text-center border-b mb-3 flex items-center gap-2 justify-center relative"
        >
            {props.children}
        </div>
    )
}

export function MediaCoreMenuBody(props: { children: React.ReactNode }) {
    return (
        <div data-vc-element="menu-body" className="max-h-[18rem] overflow-y-auto">
            {props.children}
        </div>
    )
}

export function MediaCoreMenuSectionBody(props: { children: React.ReactNode; show: boolean }) {
    if (!props.show) return null
    return (
        <div data-vc-element="menu-section-body">
            <motion.div
                data-vc-element="menu-section-motion-body"
                key="section-body"
                className="h-auto"
                initial={{ opacity: 0, scale: 1.0, x: -10 }}
                animate={{ opacity: 1, scale: 1, x: 0 }}
                exit={{ opacity: 0, scale: 1.0, x: -10 }}
                transition={{ duration: 0.15 }}
            >
                {props.children}
            </motion.div>
        </div>
    )
}

export function MediaCoreMenuSubmenuBody(props: { children: React.ReactNode; show: boolean }) {
    if (!props.show) return null
    return (
        <div data-vc-element="menu-submenu-body">
            <motion.div
                data-vc-element="menu-submenu-motion-body"
                key="section-body"
                className="h-auto"
                initial={{ opacity: 0, scale: 1.0, x: 10 }}
                animate={{ opacity: 1, scale: 1, x: 0 }}
                exit={{ opacity: 0, scale: 1.0, x: 10 }}
                transition={{ duration: 0.15 }}
            >
                {props.children}
            </motion.div>
        </div>
    )
}

export function MediaCoreMenuSubSubmenuBody(props: { children: React.ReactNode; show: boolean }) {
    if (!props.show) return null
    return (
        <div data-vc-element="menu-sub-submenu-body">
            <motion.div
                data-vc-element="menu-sub-submenu-motion-body"
                key="sub-section-body"
                className="h-auto"
                initial={{ opacity: 0, scale: 1.0, x: 10 }}
                animate={{ opacity: 1, scale: 1, x: 0 }}
                exit={{ opacity: 0, scale: 1.0, x: 10 }}
                transition={{ duration: 0.15 }}
            >
                {props.children}
            </motion.div>
        </div>
    )
}

export interface MediaCoreMenuOptionProps {
    title: string
    value?: string
    icon: React.ElementType
    children?: React.ReactNode
    onClick?: () => void
    openSection: string | null
    onOpenSectionChange: (section: string | null) => void
}

export function MediaCoreMenuOption(props: MediaCoreMenuOptionProps) {
    const { children, title, icon: Icon, onClick, value, openSection, onOpenSectionChange } = props

    function handleClick() {
        if (onClick) {
            onClick()
            return
        }
        onOpenSectionChange(title)
    }

    return (
        <>
            {!openSection && (
                <button
                    data-vc-element="menu-option"
                    role="button"
                    className="w-full p-2 h-10 flex items-center justify-between rounded-lg group/vc-menu-option hover:bg-white/10 active:bg-white/20 transition-colors"
                    onClick={handleClick}
                >
                    <span className="w-8 flex justify-start items-center h-full">
                        <Icon className="text-xl" />
                    </span>
                    <span className="w-full flex flex-1 text-sm font-medium">
                        {title}
                    </span>
                    {value && (
                        <span className="text-sm font-medium tracking-wide text-[--muted] mr-2">
                            {value}
                        </span>
                    )}
                    <LuChevronRight className="text-lg" />
                </button>
            )}

            {openSection === title && (
                <div data-vc-element="menu-section" key={title}>
                    <button
                        data-vc-element="menu-section-close"
                        role="button"
                        className="w-full pb-2 h-10 mb-2 flex items-center justify-between rounded-lg transition-colors border-b"
                        onClick={() => onOpenSectionChange(null)}
                    >
                        <span className="w-8 flex justify-start items-center h-full">
                            <LuChevronLeft className="text-lg" />
                        </span>
                        <span className="w-full flex flex-1 text-sm font-medium">
                            {title}
                        </span>
                    </button>

                    <MediaCoreMenuBody>{children}</MediaCoreMenuBody>
                </div>
            )}
        </>
    )
}

export interface MediaCoreMenuSubOptionProps {
    title: string
    value?: string
    icon: React.ElementType
    children?: React.ReactNode
    onClick?: () => void
    parentId: string
    openSection: string | null
    openSubSection: string | null
    onOpenSubSectionChange: (subSection: string | null) => void
}

export function MediaCoreMenuSubOption(props: MediaCoreMenuSubOptionProps) {
    const {
        children,
        title,
        icon: Icon,
        onClick,
        value,
        parentId,
        openSection,
        openSubSection,
        onOpenSubSectionChange,
    } = props

    const itemId = `${parentId}::${title}`

    function handleClick() {
        if (onClick) {
            onClick()
            return
        }
        onOpenSubSectionChange(itemId)
    }

    return (
        <>
            {openSection && !openSubSection && (
                <button
                    data-vc-element="menu-sub-option"
                    role="button"
                    className="w-full p-2 h-10 flex items-center justify-between rounded-lg group/vc-menu-option hover:bg-white/10 active:bg-white/20 transition-colors"
                    onClick={handleClick}
                >
                    <span className="w-8 flex justify-start items-center h-full">
                        <Icon className="text-xl" />
                    </span>
                    <span className="w-full flex flex-1 text-sm font-medium">
                        {title}
                    </span>
                    {value && (
                        <span className="text-sm font-medium tracking-wide text-[--muted] mr-2">
                            {value}
                        </span>
                    )}
                    <LuChevronRight className="text-lg" />
                </button>
            )}

            {openSubSection === itemId && (
                <div data-vc-element="menu-sub-section" key={itemId}>
                    <button
                        data-vc-element="menu-sub-section-close"
                        role="button"
                        className="w-full pb-2 h-10 mb-2 flex items-center justify-between rounded-lg transition-colors border-b"
                        onClick={() => onOpenSubSectionChange(null)}
                    >
                        <span className="w-8 flex justify-start items-center h-full">
                            <LuChevronLeft className="text-lg" />
                        </span>
                        <span className="w-full flex flex-1 text-sm font-medium">
                            {title}
                        </span>
                    </button>

                    <MediaCoreMenuBody>{children}</MediaCoreMenuBody>
                </div>
            )}
        </>
    )
}

export interface MediaCoreSettingSelectProps {
    options: MediaCoreSelectOption[]
    value: any
    onValueChange: (value: any) => void
    isFullscreen?: boolean
    containerElement?: HTMLElement | null
}

export function MediaCoreSettingSelect(props: MediaCoreSettingSelectProps) {
    const { options, value, onValueChange, isFullscreen, containerElement } = props
    return (
        <div className="block" data-vc-element="setting-select">
            {options.map(option => (
                <div
                    data-vc-element="setting-select-option"
                    key={option.value}
                    role="button"
                    className="w-full p-2 flex items-center overflow-hidden justify-between rounded-lg group/vc-menu-option hover:bg-white/10 active:bg-white/20 transition-colors"
                    onClick={() => {
                        onValueChange(option.value)
                    }}
                >
                    <span data-vc-element="setting-select-option-indicator" className="w-8 flex justify-start items-center h-full flex-none">
                        {value === option.value && <LuCheck className="text-lg" />}
                    </span>
                    <div data-vc-element="setting-select-option-body" className="flex-wrap flex flex-1 gap-2 items-center">
                        <span data-vc-element="setting-select-option-label" className="w-fit flex-none text-sm font-medium line-clamp-2">
                            {option.label}
                        </span>
                        <span className="flex-1" data-vc-element="setting-select-option-separator"></span>
                        {(option.moreInfo || option.description) && (
                            <div className="w-fit flex-none ml-2 flex gap-2 items-center" data-vc-element="setting-select-option-description">
                                {option.moreInfo && (
                                    <span className="text-xs font-medium tracking-wide text-[--muted]">
                                        {option.moreInfo}
                                    </span>
                                )}
                                {option.description && (
                                    <Tooltip
                                        trigger={<AiFillInfoCircle className="text-sm" />}
                                        portalContainer={isFullscreen ? containerElement || undefined : undefined}
                                        className="z-[150]"
                                    >
                                        {option.description}
                                    </Tooltip>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            ))}
        </div>
    )
}

export interface MediaCoreSettingTextInputProps {
    value: string
    onValueChange: (value: string) => void
    label?: string
    help?: string
}

export function MediaCoreSettingTextInput(props: MediaCoreSettingTextInputProps) {
    const { value, onValueChange, label, help } = props
    return (
        <div className="block" data-vc-element="setting-text-input">
            <TextInput label={label} value={value} onValueChange={onValueChange} help={help} />
        </div>
    )
}
