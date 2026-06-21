import {
    MediaCoreMenu,
    MediaCoreMenuTitle,
    MediaCoreMenuBody,
    MediaCoreMenuSectionBody,
    MediaCoreMenuSubmenuBody,
    MediaCoreMenuSubSubmenuBody,
    MediaCoreMenuOption,
    MediaCoreMenuSubOption,
    MediaCoreSettingSelect,
    MediaCoreSettingTextInput,
} from "@/app/(main)/_features/media-core/media-core-menu"
import { useAtom, useAtomValue } from "jotai"
import { vc_menuOpen, vc_menuSectionOpen, vc_menuSubSectionOpen, vc_isFullscreen, vc_containerElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import React from "react"

export const VideoCoreMenuTitle = MediaCoreMenuTitle
export const VideoCoreMenuBody = MediaCoreMenuBody

export function VideoCoreMenuSectionBody(props: { children: React.ReactNode }) {
    const openSection = useAtomValue(vc_menuSectionOpen)
    return <MediaCoreMenuSectionBody show={!openSection}>{props.children}</MediaCoreMenuSectionBody>
}

export function VideoCoreMenuSubmenuBody(props: { children: React.ReactNode }) {
    const openSection = useAtomValue(vc_menuSectionOpen)
    const openSubSection = useAtomValue(vc_menuSubSectionOpen)
    return <MediaCoreMenuSubmenuBody show={!!openSection && !openSubSection}>{props.children}</MediaCoreMenuSubmenuBody>
}

export function VideoCoreMenuSubSubmenuBody(props: { children: React.ReactNode }) {
    const openSubSection = useAtomValue(vc_menuSubSectionOpen)
    return <MediaCoreMenuSubSubmenuBody show={!!openSubSection}>{props.children}</MediaCoreMenuSubSubmenuBody>
}

export function VideoCoreMenu(props: any) {
    const [openMenu, setOpenMenu] = useAtom(vc_menuOpen)
    const [, setOpenSection] = useAtom(vc_menuSectionOpen)
    const [, setOpenSubSection] = useAtom(vc_menuSubSectionOpen)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)

    return (
        <MediaCoreMenu
            {...props}
            openMenu={openMenu}
            onOpenMenuChange={setOpenMenu}
            onOpenSectionChange={setOpenSection}
            onOpenSubSectionChange={setOpenSubSection}
            isFullscreen={isFullscreen}
            containerElement={containerElement}
        />
    )
}

export function VideoCoreMenuOption(props: any) {
    const [openSection, setOpenSection] = useAtom(vc_menuSectionOpen)
    return (
        <MediaCoreMenuOption
            {...props}
            openSection={openSection}
            onOpenSectionChange={setOpenSection}
        />
    )
}

export function VideoCoreMenuSubOption(props: any) {
    const openSection = useAtomValue(vc_menuSectionOpen)
    const [openSubSection, setOpenSubSection] = useAtom(vc_menuSubSectionOpen)
    return (
        <MediaCoreMenuSubOption
            {...props}
            openSection={openSection}
            openSubSection={openSubSection}
            onOpenSubSectionChange={setOpenSubSection}
        />
    )
}

export const VideoCoreSettingSelect = MediaCoreSettingSelect
export const VideoCoreSettingTextInput = MediaCoreSettingTextInput
