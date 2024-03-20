import {
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
    onlinestream_providers,
} from "@/app/(main)/onlinestream/_lib/episodes"
import { useOnlinestreamManagerContext } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import { RadioGroup } from "@/components/ui/radio-group"
import { Menu, Tooltip } from "@vidstack/react"
import { useAtom } from "jotai/react"
import React from "react"
import { AiOutlineCloudServer } from "react-icons/ai"
import { MdVideoSettings } from "react-icons/md"

type OnlinestreamServerButtonProps = {
    children?: React.ReactNode
}

export const buttonClass = "ring-media-focus group relative mr-0.5 inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-md outline-none ring-inset hover:bg-white/20 data-[focus]:ring-4 aria-hidden:hidden"

export const tooltipClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm bg-black/90 px-2 py-0.5 text-sm font-medium text-white parent-data-[open]:hidden"

export const menuClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[open]:animate-in data-[open]:fade-in data-[open]:slide-in-from-bottom-4 flex h-[var(--menu-height)] max-h-[auto] min-w-[260px] flex-col overflow-y-auto overscroll-y-contain rounded-md border border-white/10 bg-black/95 p-2.5 font-sans text-[15px] font-medium outline-none backdrop-blur-sm transition-[height] duration-300 will-change-[height] data-[resizing]:overflow-hidden"

export const submenuClass =
    "hidden w-full flex-col items-start justify-center outline-none data-[keyboard]:mt-[3px] data-[open]:inline-block"

const radioGroupItemContainerClass = "px-2 py-1.5 rounded-md hover:bg-[--subtle]"

export function OnlinestreamServerButton(props: OnlinestreamServerButtonProps) {

    const {
        children,
        ...rest
    } = props

    const { servers } = useOnlinestreamManagerContext()

    const [selectedServer, setServer] = useAtom(__onlinestream_selectedServerAtom)

    if (!servers.length || !selectedServer) return null

    return (
        <Menu.Root className="parent">
            <Tooltip.Root>
                <Tooltip.Trigger asChild>
                    <Menu.Button className={buttonClass}>
                        <AiOutlineCloudServer className="text-3xl" />
                    </Menu.Button>
                </Tooltip.Trigger>
                <Tooltip.Content className={tooltipClass} placement="top">
                    Server
                </Tooltip.Content>
            </Tooltip.Root>
            <Menu.Content className={menuClass} placement="top">
                <RadioGroup
                    value={selectedServer}
                    options={servers.map((server) => ({ label: server, value: server }))}
                    onValueChange={(v) => {
                        setServer(v)
                    }}
                    itemContainerClass={radioGroupItemContainerClass}
                />
            </Menu.Content>
        </Menu.Root>
    )
}

export function OnlinestreamProviderButton(props: OnlinestreamServerButtonProps) {

    const {
        children,
        ...rest
    } = props

    const { servers } = useOnlinestreamManagerContext()

    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

    if (!servers.length || !provider) return null

    return (
        <Menu.Root className="parent">
            <Tooltip.Root>
                <Tooltip.Trigger asChild>
                    <Menu.Button className={buttonClass}>
                        <MdVideoSettings className="text-3xl" />
                    </Menu.Button>
                </Tooltip.Trigger>
                <Tooltip.Content className={tooltipClass} placement="top">
                    Provider
                </Tooltip.Content>
            </Tooltip.Root>
            <Menu.Content className={menuClass} placement="top">
                <RadioGroup
                    value={provider}
                    options={onlinestream_providers}
                    onValueChange={(v) => {
                        setProvider(v)
                    }}
                    itemContainerClass={radioGroupItemContainerClass}
                />
            </Menu.Content>
        </Menu.Root>
    )
}
