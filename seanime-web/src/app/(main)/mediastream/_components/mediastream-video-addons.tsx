import { __mediaplayer_discreteControlsAtom } from "@/app/(main)/_atoms/builtin-mediaplayer.atoms"
import { __mediastream_autoNextAtom, __mediastream_autoPlayAtom } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { submenuClass, VdsSubmenuButton } from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { Switch } from "@/components/ui/switch"
import { Menu } from "@vidstack/react"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { MdPlaylistPlay } from "react-icons/md"
import { RxSlider } from "react-icons/rx"

export function MediastreamPlaybackSubmenu() {

    const [autoPlay, setAutoPlay] = useAtom(__mediastream_autoPlayAtom)
    const [autoNext, setAutoNext] = useAtom(__mediastream_autoNextAtom)
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
                    label={`Auto Play Next Episode`}
                    hint={autoNext ? "On" : "Off"}
                    disabled={false}
                    icon={MdPlaylistPlay}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Auto play next episode"
                        fieldClass="py-2 px-2"
                        value={autoNext}
                        onValueChange={setAutoNext}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Discrete Controls`}
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
