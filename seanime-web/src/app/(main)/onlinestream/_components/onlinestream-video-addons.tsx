import { useOnlineStreamEmptyCache } from "@/api/hooks/onlinestream.hooks"
import { __mediaplayer_discreteControlsAtom } from "@/app/(main)/_atoms/builtin-mediaplayer.atoms"
import { useOnlinestreamManagerContext } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import {
    __onlinestream_autoNextAtom,
    __onlinestream_autoPlayAtom,
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { Alert } from "@/components/ui/alert"
import { Button, IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { RadioGroup } from "@/components/ui/radio-group"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { Tooltip as CTooltip } from "@/components/ui/tooltip"
import { Menu, Tooltip } from "@vidstack/react"
import { ChevronLeftIcon, ChevronRightIcon, RadioButtonIcon, RadioButtonSelectedIcon } from "@vidstack/react/icons"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillPlayCircle, AiOutlineCloudServer } from "react-icons/ai"
import { MdHighQuality, MdPlaylistPlay, MdVideoSettings } from "react-icons/md"
import { RxSlider } from "react-icons/rx"

type OnlinestreamServerButtonProps = {
    children?: React.ReactNode
}

export const buttonClass = "ring-media-focus group relative mr-0.5 inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-md outline-none ring-inset hover:bg-white/20 data-[focus]:ring-4 aria-hidden:hidden"

export const tooltipClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm bg-black/90 px-2 py-0.5 text-sm font-medium text-white group-data-[open]/parent:hidden"

export const menuClass =
    "animate-out fade-out slide-out-to-bottom-2 data-[open]:animate-in data-[open]:fade-in data-[open]:slide-in-from-bottom-4 flex h-[var(--menu-height)] max-h-[400px] min-w-[260px] flex-col overflow-y-auto overscroll-y-contain rounded-md border border-white/10 bg-black/95 p-2.5 font-sans text-[15px] font-medium outline-none backdrop-blur-sm transition-[height] duration-300 will-change-[height] data-[resizing]:overflow-hidden"

export const submenuClass =
    "hidden w-full flex-col items-start justify-center outline-none data-[keyboard]:mt-[3px] data-[open]:inline-block"

const radioGroupItemContainerClass = "px-2 py-1.5 rounded-md hover:bg-[--subtle]"

export function OnlinestreamVideoQualitySubmenu() {

    const { customQualities, videoSource, changeQuality } = useOnlinestreamManagerContext()

    return (
        <Menu.Root>
            <VdsSubmenuButton
                label={`Quality`}
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

export function OnlinestreamPlaybackSubmenu() {

    const [autoPlay, setAutoPlay] = useAtom(__onlinestream_autoPlayAtom)
    const [autoNext, setAutoNext] = useAtom(__onlinestream_autoNextAtom)
    const [discreteControls, setDiscreteControls] = useAtom(__mediaplayer_discreteControlsAtom)

    return (
        <>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Auto Play`}
                    hint={autoPlay ? "On" : "Off"}
                    disabled={false}
                    icon={AiFillPlayCircle}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Auto play"
                        fieldClass="py-2 px-2"
                        value={autoPlay}
                        onValueChange={setAutoPlay}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Play Next`}
                    hint={autoNext ? "On" : "Off"}
                    disabled={false}
                    icon={MdPlaylistPlay}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Auto play next"
                        fieldClass="py-2 px-2"
                        value={autoNext}
                        onValueChange={setAutoNext}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Discrete controls`}
                    hint={discreteControls ? "On" : "Off"}
                    disabled={false}
                    icon={RxSlider}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Discrete controls"
                        help="Only show the controls when the mouse is over the bottom part. (Large screens only)"
                        fieldClass="py-2 px-2"
                        value={discreteControls}
                        onValueChange={setDiscreteControls}
                    />
                </Menu.Content>
            </Menu.Root>
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function OnlinestreamParametersButton({ mediaId }: { mediaId: number }) {

    const { servers, providerExtensionOptions, changeProvider, changeServer } = useOnlinestreamManagerContext()

    const [provider] = useAtom(__onlinestream_selectedProviderAtom)
    const [selectedServer] = useAtom(__onlinestream_selectedServerAtom)

    const { mutate: emptyCache, isPending } = useOnlineStreamEmptyCache()

    return (
        <Modal
            title="Stream Parameters"
            trigger={<IconButton intent="gray-basic" icon={<MdVideoSettings />} />}
        >
            <Alert
                intent="info-basic"
                description="Empty the cache if you are experiencing issues with the stream."
            />
            <Select
                label="Provider"
                value={provider || ""}
                options={providerExtensionOptions}
                onValueChange={(v) => {
                    changeProvider(v)
                }}
            />
            {!!servers.length && <Select
                label="Server"
                value={selectedServer}
                options={servers.map((server) => ({ label: server, value: server }))}
                onValueChange={(v) => {
                    changeServer(v)
                }}
            />}

            <Separator />

            <Button
                size="sm"
                intent="alert-subtle"
                onClick={() => emptyCache({ mediaId })}
                loading={isPending}
            >
                Empty stream cache
            </Button>
        </Modal>
    )
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
                        <AiOutlineCloudServer className="text-3xl" />
                    </Menu.Button>
                </Tooltip.Trigger>
                <Tooltip.Content className={tooltipClass} placement="top">
                    Provider
                </Tooltip.Content>
            </Tooltip.Root>
            <Menu.Content className={menuClass} placement="top">
                <RadioGroup
                    value={provider || ""}
                    options={providerExtensionOptions}
                    onValueChange={(v) => {
                        changeProvider(v)
                    }}
                    itemContainerClass={radioGroupItemContainerClass}
                />
                <Separator />
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
        <Menu.Button
            className="ring-media-focus group/parent left-0 z-10 flex w-full cursor-pointer select-none items-center justify-start rounded-sm bg-black/60 p-2.5 outline-none ring-inset data-[open]:sticky data-[open]:-top-2.5 data-[hocus]:bg-white/10 data-[focus]:ring-[3px] aria-disabled:hidden"
            disabled={disabled}
        >
            <ChevronLeftIcon className="group-data-[open]/parent:block -ml-0.5 mr-1.5 hidden h-[18px] w-[18px]" />
            <div className="contents group-data-[open]/parent:hidden">
                <Icon className="text-xl" />
            </div>
            <span className="ml-1.5 group-data-[open]/parent:ml-0">{label}</span>
            <span className="ml-auto text-sm text-white/50">{hint}</span>
            <ChevronRightIcon className="group-data-[open]/parent:hidden ml-0.5 h-[18px] w-[18px] text-sm text-white/50" />
        </Menu.Button>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function SwitchSubOrDubButton() {
    const [dubbed] = useAtom(__onlinestream_selectedDubbedAtom)
    const { selectedExtension, toggleDubbed } = useOnlinestreamManagerContext()

    if (!selectedExtension || !selectedExtension?.supportsDub) return null

    return (
        <CTooltip
            trigger={<Button
                className=""
                rounded
                intent="gray-outline"
                size="sm"
                onClick={() => toggleDubbed()}
            >
                {dubbed ? "Dubbed" : "Subbed"}
            </Button>}
        >
            {dubbed ? "Switch to subs" : "Switch to dub"}
        </CTooltip>
    )
}
