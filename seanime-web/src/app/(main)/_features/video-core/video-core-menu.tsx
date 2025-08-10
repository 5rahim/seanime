import { vc_menuOpen } from "@/app/(main)/_features/video-core/video-core"
import { Popover } from "@/components/ui/popover"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { motion } from "motion/react"
import React, { useRef } from "react"
import { LuCheck, LuChevronLeft, LuChevronRight } from "react-icons/lu"

const vc_menuSectionOpen = atom<string | null>(null)

type VideoCoreMenuProps = {
    trigger: React.ReactElement
    children?: React.ReactNode
}

export function VideoCoreMenu(props: VideoCoreMenuProps) {

    const { trigger, children, ...rest } = props
    const [menuOpen, setMenuOpen] = useAtom(vc_menuOpen)

    const [open, setOpen] = React.useState(false)

    const [openSection, setOpenSection] = useAtom(vc_menuSectionOpen)

    const t = useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (!menuOpen) {
            t.current = setTimeout(() => {
                setOpenSection(null)
            }, 300)
        }
        return () => {
            if (t.current) {
                clearTimeout(t.current)
            }
        }
    }, [menuOpen])

    return (
        <Popover
            open={open}
            onOpenChange={v => {
                setOpen(v)
                setMenuOpen(v)
            }}
            trigger={<div>{trigger}</div>}
            sideOffset={4}
            align="center"
            modal={false}
            className="bg-black/85 rounded-xl p-3 overflow-hidden backdrop-blur-sm"
        >
            <div className="h-auto">
                {children}
            </div>
        </Popover>
    )

}

export function VideoCoreMenuTitle(props: { children: React.ReactNode }) {

    const { children, ...rest } = props
    return (
        <div className="text-white/70 font-bold text-sm pb-3 text-center" {...rest}>
            {children}
        </div>
    )

}

export function VideoCoreMenuSectionBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    const [openSection, setOpen] = useAtom(vc_menuSectionOpen)

    return (
        <div className="vc-menu-section-body overflow-hidden">
            {/*<AnimatePresence mode="wait">*/}
            {!openSection && (
                <motion.div
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

export function VideoCoreMenuSubmenuBody(props: { children: React.ReactNode }) {
    const { children, ...rest } = props

    const [openSection, setOpen] = useAtom(vc_menuSectionOpen)

    return (
        <div className="vc-menu-submenu-body overflow-hidden">
            {/*<AnimatePresence mode="wait">*/}
            {openSection && (
                <motion.div
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
                    key={title}
                >
                    <button
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

                    <div className="max-h-[18rem] overflow-y-auto">
                        {children}
                    </div>
                </div>
            )}
        </>
    )
}

type VideoCoreSettingSelectProps<T extends string | number> = {
    options: {
        label: string
        value: T
        moreInfo?: string
        description?: string
    }[]
    value: T
    onValueChange: (value: T) => void
}

export function VideoCoreSettingSelect<T extends string | number>(props: VideoCoreSettingSelectProps<T>) {
    const { options, value, onValueChange } = props
    return (
        <div className="block">
            {options.map(option => (
                <div
                    key={option.value}
                    role="button"
                    className="w-full p-2 flex items-center justify-between rounded-lg group/vc-menu-option hover:bg-white/10 active:bg-white/20 transition-colors"
                    onClick={() => {
                        onValueChange(option.value)
                    }}
                >
                    <span className="w-8 flex justify-start items-center h-full">
                        {value === option.value && <LuCheck className="text-lg" />}
                    </span>
                    <span className="w-full flex flex-1 text-sm font-medium">
                        {option.label}
                    </span>
                    {option.moreInfo && <span className="text-sm font-medium tracking-wide text-[--muted] mr-2">
                        {option.moreInfo}
                    </span>}

                </div>
            ))}
        </div>
    )
}
