import { vc_containerElement, vc_isFullscreen } from "@/app/(main)/_features/video-core/video-core"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { Popover } from "@/components/ui/popover"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import { motion } from "motion/react"
import React, { useRef } from "react"
import { AiFillInfoCircle } from "react-icons/ai"
import { LuCheck, LuChevronLeft, LuChevronRight } from "react-icons/lu"

export const vc_menuOpen = atom<string | null>(null)
export const vc_menuSectionOpen = atom<string | null>(null)
export const vc_menuSubSectionOpen = atom<string | null>(null)

type VideoCoreMenuProps = {
    name: string
    trigger: React.ReactElement
    children?: React.ReactNode
    className?: string
    sideOffset?: number
    isDrawer?: boolean
}

export function VideoCoreMenu(props: VideoCoreMenuProps) {

    const { trigger, children, className, name, sideOffset = 4, isDrawer, ...rest } = props

    const [open, setOpen] = useAtom(vc_menuOpen)

    const [openSection, setOpenSection] = useAtom(vc_menuSectionOpen)
    const [openSubSection, setOpenSubSection] = useAtom(vc_menuSubSectionOpen)

    // Get fullscreen state and container element for proper portal mounting
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)

    const t = useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (!open) {
            t.current = setTimeout(() => {
                setOpenSection(null)
                setOpenSubSection(null)
            }, 300)
        }
        return () => {
            if (t.current) {
                clearTimeout(t.current)
            }
        }
    }, [open])

    if (isDrawer) {
        return <Drawer
            data-vc-element="menu-drawer"
            open={open === name}
            onOpenChange={v => {
                setOpen(v ? name : null)
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
    }

    return (
        <Popover
            data-vc-element="menu"
            open={open === name}
            onOpenChange={v => {
                setOpen(v ? name : null)
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

export function VideoCoreMenuTitle(props: { children: React.ReactNode }) {

    const { children, ...rest } = props
    return (
        <div
            data-vc-element="menu-title"
            className="text-white/70 font-bold text-sm pb-3 text-center border-b mb-3 flex items-center gap-2 justify-center relative" {...rest}>
            {children}
        </div>
    )

}

export function VideoCoreMenuSectionBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    const [openSection, setOpen] = useAtom(vc_menuSectionOpen)

    return (
        <div data-vc-element="menu-section-body">
            {/*<AnimatePresence mode="wait">*/}
            {!openSection && (
                <motion.div
                    data-vc-element="menu-section-motion-body"
                    key="section-body"
                    className="h-auto"
                    initial={{ opacity: 0, scale: 1.0, x: -10 }}
                    animate={{ opacity: 1, scale: 1, x: 0 }}
                    exit={{ opacity: 0, scale: 1.0, x: -10 }}
                    transition={{ duration: 0.15 }}
                >
                    {children}
                </motion.div>
            )}
            {/*</AnimatePresence>*/}
        </div>
    )
}

export function VideoCoreMenuBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    return (
        <div data-vc-element="menu-body" className="max-h-[18rem] overflow-y-auto">
            {children}
        </div>
    )
}

export function VideoCoreMenuSubmenuBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    const [openSection, setOpen] = useAtom(vc_menuSectionOpen)
    const [openSubSection] = useAtom(vc_menuSubSectionOpen)

    return (
        <div data-vc-element="menu-submenu-body">
            {/*<AnimatePresence mode="wait">*/}
            {openSection && !openSubSection && (
                <motion.div
                    data-vc-element="menu-submenu-motion-body"
                    key="section-body"
                    className="h-auto"
                    initial={{ opacity: 0, scale: 1.0, x: 10 }}
                    animate={{ opacity: 1, scale: 1, x: 0 }}
                    exit={{ opacity: 0, scale: 1.0, x: 10 }}
                    transition={{ duration: 0.15 }}
                >
                    {children}
                </motion.div>
            )}
            {/*</AnimatePresence>*/}
        </div>
    )
}

export function VideoCoreMenuSubSubmenuBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    const [openSubSection] = useAtom(vc_menuSubSectionOpen)

    return (
        <div data-vc-element="menu-sub-submenu-body">
            {openSubSection && (
                <motion.div
                    data-vc-element="menu-sub-submenu-motion-body"
                    key="sub-section-body"
                    className="h-auto"
                    initial={{ opacity: 0, scale: 1.0, x: 10 }}
                    animate={{ opacity: 1, scale: 1, x: 0 }}
                    exit={{ opacity: 0, scale: 1.0, x: 10 }}
                    transition={{ duration: 0.15 }}
                >
                    {children}
                </motion.div>
            )}
        </div>
    )
}

export function VideoCoreMenuOption(props: {
    title: string,
    value?: string,
    icon: React.ElementType,
    children?: React.ReactNode,
    onClick?: () => void
}) {
    const { children, title, icon: Icon, onClick, value, ...rest } = props

    const [openSection, setOpen] = useAtom(vc_menuSectionOpen)

    function handleClick() {
        if (onClick) {
            onClick()
            return
        }

        // open the section
        setOpen(title)
    }

    return (
        <>
            {!openSection && <button
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
                {value && <span className="text-sm font-medium tracking-wide text-[--muted] mr-2">
                    {value}
                </span>}
                <LuChevronRight className="text-lg" />
            </button>}

            {openSection === title && (
                <div
                    data-vc-element="menu-section"
                    key={title}
                >
                    <button
                        data-vc-element="menu-section-close"
                        role="button"
                        className="w-full pb-2 h-10 mb-2 flex items-center justify-between rounded-lg transition-colors border-b"
                        onClick={() => setOpen(null)}
                    >
                        <span className="w-8 flex justify-start items-center h-full">
                            <LuChevronLeft className="text-lg" />
                        </span>
                        <span className="w-full flex flex-1 text-sm font-medium">
                            {title}
                        </span>
                    </button>

                    <VideoCoreMenuBody>
                        {children}
                    </VideoCoreMenuBody>
                </div>
            )}
        </>
    )
}

export function VideoCoreMenuSubOption(props: {
    title: string,
    value?: string,
    icon: React.ElementType,
    children?: React.ReactNode,
    onClick?: () => void,
    parentId: string
}) {
    const { children, title, icon: Icon, onClick, value, parentId, ...rest } = props

    const [openSection] = useAtom(vc_menuSectionOpen)
    const [openSubSection, setOpenSubSection] = useAtom(vc_menuSubSectionOpen)

    const itemId = `${parentId}::${title}`

    function handleClick() {
        if (onClick) {
            onClick()
            return
        }

        // open the sub-section
        setOpenSubSection(itemId)
    }

    return (
        <>
            {openSection && !openSubSection && <button
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
                {value && <span className="text-sm font-medium tracking-wide text-[--muted] mr-2">
                    {value}
                </span>}
                <LuChevronRight className="text-lg" />
            </button>}

            {openSubSection === itemId && (
                <div
                    data-vc-element="menu-sub-section"
                    key={itemId}
                >
                    <button
                        data-vc-element="menu-sub-section-close"
                        role="button"
                        className="w-full pb-2 h-10 mb-2 flex items-center justify-between rounded-lg transition-colors border-b"
                        onClick={() => setOpenSubSection(null)}
                    >
                        <span className="w-8 flex justify-start items-center h-full">
                            <LuChevronLeft className="text-lg" />
                        </span>
                        <span className="w-full flex flex-1 text-sm font-medium">
                            {title}
                        </span>
                    </button>

                    <VideoCoreMenuBody>
                        {children}
                    </VideoCoreMenuBody>
                </div>
            )}
        </>
    )
}

type VideoCoreSettingSelectProps = {
    options: {
        label: string
        value: any
        moreInfo?: string
        description?: string
    }[]
    value: any
    onValueChange: (value: any) => void
    isFullscreen?: boolean
    containerElement?: HTMLElement | null | undefined
}

export function VideoCoreSettingSelect(props: VideoCoreSettingSelectProps) {
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
                        {(option.moreInfo || option.description) &&
                            <div className="w-fit flex-none ml-2 flex gap-2 items-center" data-vc-element="setting-select-option-description">
                            {option.moreInfo && <span className="text-xs font-medium tracking-wide text-[--muted]">
                                {option.moreInfo}
                            </span>}
                            {option.description && <Tooltip
                                trigger={<AiFillInfoCircle className="text-sm" />}
                                portalContainer={isFullscreen ? containerElement || undefined : undefined}
                                className="z-[150]"
                            >
                                {option.description}
                            </Tooltip>}
                        </div>}
                    </div>

                </div>
            ))}
        </div>
    )
}

type VideoCoreSettingTextInputProps = {
    value: string
    onValueChange: (value: string) => void
    label?: string
    help?: string
}

export function VideoCoreSettingTextInput(props: VideoCoreSettingTextInputProps) {
    const { value, onValueChange, label, help } = props
    return (
        <div className="block" data-vc-element="setting-text-input">
            <TextInput
                label={label}
                value={value}
                onValueChange={onValueChange}
                help={help}
            />
        </div>
    )
}
