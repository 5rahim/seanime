import { __mediastream_autoNextAtom, __mediastream_autoPlayAtom } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { submenuClass, VdsSubmenuButton } from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { Switch } from "@/components/ui/switch"
import { Menu } from "@vidstack/react"
import { useAtom } from "jotai/react"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { MdPlaylistPlay } from "react-icons/md"

export function MediastreamPlaybackSubmenu() {

    const [autoPlay, setAutoPlay] = useAtom(__mediastream_autoPlayAtom)
    const [autoNext, setAutoNext] = useAtom(__mediastream_autoNextAtom)

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
        </>
    )
}
