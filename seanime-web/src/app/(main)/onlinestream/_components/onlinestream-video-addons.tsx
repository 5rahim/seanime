import { useOnlineStreamEmptyCache } from "@/api/hooks/onlinestream.hooks"
import { useOnlinestreamManagerContext } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import {
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { Button } from "@/components/ui/button"
import { Modal, ModalProps } from "@/components/ui/modal"
import { Popover, PopoverProps } from "@/components/ui/popover"
import { RadioGroup } from "@/components/ui/radio-group"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { useWindowSize } from "@uidotdev/usehooks"
import { Menu, Tooltip } from "@vidstack/react"
import { ChevronLeftIcon, ChevronRightIcon, RadioButtonIcon, RadioButtonSelectedIcon } from "@vidstack/react/icons"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { HiOutlineCog6Tooth } from "react-icons/hi2"
import { LuGlobe, LuSpeech } from "react-icons/lu"
import { MdHighQuality, MdOutlineSubtitles } from "react-icons/md"
import { TbCloudSearch } from "react-icons/tb"

type OnlinestreamServerButtonProps = {
    children?: React.ReactNode
}

export const buttonClass = "ring-media-focus group relative mr-0.5 inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-[--radius-md] outline-none ring-inset hover:bg-white/20 data-[focus]:ring-4 aria-hidden:hidden"

export const tooltipClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm bg-black/90 px-2 py-0.5 text-sm font-medium text-white group-data-[open]/parent:hidden"

export const menuClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[open]:animate-in data-[open]:fade-in data-[open]:slide-in-from-bottom-4 flex h-[var(--menu-height)] max-h-[400px] min-w-[260px] flex-col overflow-y-auto overscroll-y-contain rounded-[--radius-md] border border-white/10 bg-black/95 p-2.5 font-sans text-[15px] font-medium outline-none backdrop-blur-sm transition-[height] duration-300 will-change-[height] data-[resizing]:overflow-hidden"

export const submenuClass =
    "hidden w-full flex-col items-start justify-center outline-none data-[keyboard]:mt-[3px] data-[open]:inline-block"

const radioGroupItemContainerClass = "px-2 py-1.5 rounded-[--radius-md] hover:bg-[--subtle]"

export function OnlinestreamVideoQualitySubmenu() {

    const { customQualities, videoSource, changeQuality } = useOnlinestreamManagerContext()

    return (
        <Menu.Root>
            <VdsSubmenuButton
                label={`Quality `}
                hint={videoSource?.quality || ""}
                disabled={false}
                icon={MdHighQuality}
            />
            <Menu.Content className={submenuClass}>
                <Menu.RadioGroup value={videoSource?.quality || "-"}>
                    {customQualities.map((v) => (
                        <Radio
                            value={v}
                            onSelect={e => {
                                if (e.target.checked) {
                                    changeQuality(v)
                                }
                            }}
                            key={v}
                        >
                            {v}
                        </Radio>
                    ))}
                </Menu.RadioGroup>
            </Menu.Content>
        </Menu.Root>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function OnlinestreamParametersButton({ mediaId }: { mediaId: number }) {

    const { servers, providerExtensionOptions, changeProvider, changeServer } = useOnlinestreamManagerContext()

    const router = useRouter()
    const [provider] = useAtom(__onlinestream_selectedProviderAtom)
    const [selectedServer] = useAtom(__onlinestream_selectedServerAtom)

    const { mutate: emptyCache, isPending } = useOnlineStreamEmptyCache()

    return (
        <>
            <Select
                value={provider || ""}
                options={[
                    ...providerExtensionOptions,
                    {
                        value: "add-provider",
                        label: "Find other providers",
                    },
                ]}
                onValueChange={(v) => {
                    if (v === "add-provider") {
                        router.push(`/extensions`)
                        return
                    }
                    changeProvider(v)
                }}
                size="sm"
                leftAddon={<LuGlobe />}
                fieldClass="w-fit"
                className="rounded-full rounded-l-none w-fit"
                addonClass="rounded-full rounded-r-none"
            />
            {!!servers.length && <Select
                size="sm"
                value={selectedServer}
                options={servers.map((server) => ({ label: server, value: server }))}
                onValueChange={(v) => {
                    changeServer(v)
                }}
                fieldClass="w-fit"
                className="rounded-full w-fit !px-4"
                addonClass="rounded-full rounded-r-none"
            />}
            <IsomorphicPopover
                title="Stream"
                trigger={<Button
                    intent="gray-basic"
                    size="sm"
                    className="rounded-full"
                    leftIcon={<HiOutlineCog6Tooth className="text-xl" />}
                >
                    Cache
                </Button>}
            >
                <p className="text-sm text-[--muted]">
                    Empty the cache if you are experiencing issues with the stream.
                </p>
                <Button
                    size="sm"
                    intent="alert-subtle"
                    onClick={() => emptyCache({ mediaId })}
                    loading={isPending}
                >
                    Empty stream cache
                </Button>
            </IsomorphicPopover>
        </>
    )


}

function IsomorphicPopover(props: PopoverProps & ModalProps) {
    const { title, children, ...rest } = props
    const { width } = useWindowSize()

    if (width && width > 1024) {
        return <Popover
            {...rest}
            className="max-w-xl !w-full overflow-hidden space-y-2"
        >
            {children}
        </Popover>
    }

    return <Modal
        {...rest}
        title={title}
    >
        {children}
    </Modal>
}

export function OnlinestreamProviderButton(props: OnlinestreamServerButtonProps) {

    const {
        children,
        ...rest
    } = props

    const { changeProvider, providerExtensionOptions, servers, changeServer } = useOnlinestreamManagerContext()

    const [provider] = useAtom(__onlinestream_selectedProviderAtom)
    const [selectedServer] = useAtom(__onlinestream_selectedServerAtom)

    if (!servers.length || !selectedServer) return null

    return (
        <Menu.Root className="parent">
            <Tooltip.Root>
                <Tooltip.Trigger asChild>
                    <Menu.Button className={buttonClass}>
                        <TbCloudSearch className="text-3xl" />
                    </Menu.Button>
                </Tooltip.Trigger>
                <Tooltip.Content className={tooltipClass} placement="top">
                    Provider
                </Tooltip.Content>
            </Tooltip.Root>
            <Menu.Content className={menuClass} placement="top">
                <p className="text-white px-2 py-1">
                    Provider
                </p>
                <RadioGroup
                    value={provider || ""}
                    options={providerExtensionOptions}
                    onValueChange={(v) => {
                        changeProvider(v)
                    }}
                    itemContainerClass={radioGroupItemContainerClass}
                />
                <Separator className="my-1" />
                <p className="text-white px-2 py-1">
                    Server
                </p>
                <RadioGroup
                    value={selectedServer}
                    options={servers.map((server) => ({ label: server, value: server }))}
                    onValueChange={(v) => {
                        changeServer(v)
                    }}
                    itemContainerClass={radioGroupItemContainerClass}
                />
            </Menu.Content>
        </Menu.Root>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export interface RadioProps extends Menu.RadioProps {
}

function Radio({ children, ...props }: RadioProps) {
    return (
        <Menu.Radio
            className="ring-media-focus group relative flex w-full cursor-pointer select-none items-center justify-start rounded-sm p-2.5 outline-none data-[hocus]:bg-white/10 data-[focus]:ring-[3px]"
            {...props}
        >
            <RadioButtonIcon className="h-4 w-4 text-white group-data-[checked]:hidden" />
            <RadioButtonSelectedIcon className="text-media-brand hidden h-4 w-4 group-data-[checked]:block" />
            <span className="ml-2">{children}</span>
        </Menu.Radio>
    )
}

export interface VdsSubmenuButtonProps {
    label: string;
    hint: string;
    disabled?: boolean;
    icon: any;
}

export function VdsSubmenuButton({ label, hint, icon: Icon, disabled }: VdsSubmenuButtonProps) {
    return (
        <Menu.Button className="vds-menu-button" disabled={disabled}>
            <ChevronLeftIcon className="vds-menu-button-close-icon" />
            <Icon className="vds-menu-button-icon" />
            <span className="vds-menu-button-label mr-2">{label}</span>
            <span className="vds-menu-button-hint">{hint}</span>
            <ChevronRightIcon className="vds-menu-button-open-icon" />
        </Menu.Button>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function SwitchSubOrDubButton() {
    const [dubbed] = useAtom(__onlinestream_selectedDubbedAtom)
    const { selectedExtension, toggleDubbed } = useOnlinestreamManagerContext()

    if (!selectedExtension || !selectedExtension?.supportsDub) return null

    return (
        <Button
            className=""
            rounded
            intent="gray-basic"
            size="sm"
            leftIcon={!dubbed ? <LuSpeech className="text-xl" /> : <MdOutlineSubtitles className="text-xl" />}
            onClick={() => toggleDubbed()}
        >
            {dubbed ? "Switch to subs" : "Switch to dub"}
        </Button>
    )
}
