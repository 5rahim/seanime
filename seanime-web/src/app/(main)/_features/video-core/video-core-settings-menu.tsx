import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import {
    vc_containerElement,
    vc_dispatchAction,
    vc_isFullscreen,
    vc_isMobile,
    vc_mediaCaptionsManager,
    vc_miniPlayer,
    vc_playbackRate,
    vc_subtitleManager,
} from "@/app/(main)/_features/video-core/video-core"
import { anime4kOptions, getAnime4KOptionByValue, vc_anime4kOption } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import { Anime4KOption } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { VideoCoreControlButtonIcon } from "@/app/(main)/_features/video-core/video-core-control-bar"
import {
    vc_menuOpen,
    vc_menuSectionOpen,
    vc_menuSubSectionOpen,
    VideoCoreMenu,
    VideoCoreMenuOption,
    VideoCoreMenuSectionBody,
    VideoCoreMenuSubmenuBody,
    VideoCoreMenuSubOption,
    VideoCoreMenuSubSubmenuBody,
    VideoCoreMenuTitle,
    VideoCoreSettingSelect,
    VideoCoreSettingTextInput,
} from "@/app/(main)/_features/video-core/video-core-menu"
import { videoCorePreferencesModalAtom } from "@/app/(main)/_features/video-core/video-core-preferences"
import {
    vc_autoNextAtom,
    vc_autoPlayVideoAtom,
    vc_autoSkipOPEDAtom,
    vc_beautifyImageAtom,
    vc_highlightOPEDChaptersAtom,
    vc_initialSettings,
    vc_settings,
    vc_showChapterMarkersAtom,
    vc_storedPlaybackRateAtom,
    VideoCoreSettings,
} from "@/app/(main)/_features/video-core/video-core.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { upath } from "@/lib/helpers/upath"
import { useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React, { useState } from "react"
import { HiFastForward } from "react-icons/hi"
import { ImFileText } from "react-icons/im"
import { IoCaretForwardCircleOutline } from "react-icons/io5"
import { LuChevronUp, LuHeading, LuPaintbrush, LuPalette, LuSettings, LuSettings2, LuSparkles, LuTvMinimalPlay } from "react-icons/lu"
import { MdOutlineAccessTime, MdOutlineSubtitles, MdSpeed } from "react-icons/md"
import { RiShadowLine } from "react-icons/ri"
import { TbArrowForwardUp } from "react-icons/tb"
import { VscTextSize } from "react-icons/vsc"

const SUBTITLE_STYLES_FONT_SIZE_OPTIONS = [
    { label: "Small", value: 54 },
    { label: "Medium", value: 62 },
    { label: "Large", value: 72 },
    { label: "Extra Large", value: 82 },
]

const SUBTITLE_STYLES_COLOR_OPTIONS = [
    { label: "White", value: "#FFFFFF" },
    { label: "Black", value: "#000000" },
    { label: "Gray", value: "#808080" },
    { label: "Yellow", value: "#FFD700" },
    { label: "Cyan", value: "#00FFFF" },
    { label: "Pink", value: "#FF69B4" },
    { label: "Purple", value: "#9370DB" },
    { label: "Lime", value: "#00FF00" },
]

const SUBTITLE_STYLES_OUTLINE_WIDTH_OPTIONS = [
    { label: "None", value: 0 },
    { label: "Small", value: 2 },
    { label: "Medium", value: 3 },
    { label: "Large", value: 4 },
]

const SUBTITLE_STYLES_SHADOW_DEPTH_OPTIONS = [
    { label: "None", value: 0 },
    { label: "Small", value: 1 },
    { label: "Medium", value: 2 },
    { label: "Large", value: 3 },
]

export const SUBTITLE_STYLES_BACK_COLOR_OPACITY_OPTIONS = [
    { label: "100%", value: 0 },
    { label: "80%", value: 64 },
    { label: "70%", value: 77 },
    { label: "50%", value: 150 },
    { label: "25%", value: 200 },
    { label: "0%", value: 255 },
]

export const vc_subtitleStylesDefaults: VideoCoreSettings["subtitleCustomization"] = {
    enabled: false,
    fontName: "",
    fontSize: SUBTITLE_STYLES_FONT_SIZE_OPTIONS[1].value,
    primaryColor: SUBTITLE_STYLES_COLOR_OPTIONS[0].value,
    outlineColor: SUBTITLE_STYLES_COLOR_OPTIONS[1].value,
    backColor: SUBTITLE_STYLES_COLOR_OPTIONS[1].value,
    backColorOpacity: SUBTITLE_STYLES_BACK_COLOR_OPACITY_OPTIONS[0].value,
    outline: SUBTITLE_STYLES_OUTLINE_WIDTH_OPTIONS[2].value,
    shadow: SUBTITLE_STYLES_SHADOW_DEPTH_OPTIONS[0].value,
}

export function vc_getSubtitleStyle<T extends keyof VideoCoreSettings["subtitleCustomization"]>(settings: VideoCoreSettings["subtitleCustomization"] | undefined,
    key: T,
): NonNullable<VideoCoreSettings["subtitleCustomization"][T]> {
    return settings?.[key] ?? vc_subtitleStylesDefaults[key] as any
}

export function vc_getSubtitleStyleLabel<T extends keyof VideoCoreSettings["subtitleCustomization"]>(settings: VideoCoreSettings["subtitleCustomization"] | undefined,
    key: T,
): string {
    switch (key) {
        case "fontSize":
            return SUBTITLE_STYLES_FONT_SIZE_OPTIONS.find(o => o.value === vc_getSubtitleStyle(settings, key))?.label ?? ""
        case "outline":
            return SUBTITLE_STYLES_OUTLINE_WIDTH_OPTIONS.find(o => o.value === vc_getSubtitleStyle(settings, key))?.label ?? ""
        case "shadow":
            return SUBTITLE_STYLES_SHADOW_DEPTH_OPTIONS.find(o => o.value === vc_getSubtitleStyle(settings, key))?.label ?? ""
        case "primaryColor":
        case "outlineColor":
        case "backColor":
            return SUBTITLE_STYLES_COLOR_OPTIONS.find(o => o.value === vc_getSubtitleStyle(settings, key))?.label ?? ""
        // case "backColorOpacity":
        //     return `${((vc_getSubtitleStyle(settings, "backColorOpacity")-255) / 255 * 100).toFixed(0)}%`
    }
    return ""
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const CAPTION_STYLES_FONT_SIZE_OPTIONS = [
    { label: "Small", value: 4 },
    { label: "Medium", value: 5 },
    { label: "Large", value: 5.7 },
    { label: "Extra Large", value: 6.1 },
]

export const CAPTION_STYLES_TEXT_SHADOW_OPTIONS = [
    { label: "None", value: 0 },
    { label: "Small", value: 2 },
    { label: "Medium", value: 4 },
    { label: "Large", value: 6 },
]

export const CAPTION_STYLES_BACKGROUND_OPACITY_OPTIONS = [
    { label: "0%", value: 0 },
    { label: "25%", value: 0.25 },
    { label: "50%", value: 0.5 },
    { label: "70%", value: 0.7 },
    { label: "80%", value: 0.8 },
    { label: "100%", value: 1 },
]

export const CAPTION_STYLES_COLOR_OPTIONS = SUBTITLE_STYLES_COLOR_OPTIONS

export const vc_captionsStylesDefaults: VideoCoreSettings["captionCustomization"] = {
    fontSize: CAPTION_STYLES_FONT_SIZE_OPTIONS[1].value,
    textColor: CAPTION_STYLES_COLOR_OPTIONS[0].value,
    backgroundColor: SUBTITLE_STYLES_COLOR_OPTIONS[1].value,
    textShadow: CAPTION_STYLES_TEXT_SHADOW_OPTIONS[2].value,
    textShadowColor: CAPTION_STYLES_COLOR_OPTIONS[1].value,
    backgroundOpacity: CAPTION_STYLES_BACKGROUND_OPACITY_OPTIONS[3].value,
}

export function vc_getCaptionStyle<T extends keyof VideoCoreSettings["captionCustomization"]>(settings: VideoCoreSettings["captionCustomization"] | undefined,
    key: T,
): NonNullable<VideoCoreSettings["captionCustomization"][T]> {
    return settings?.[key] ?? vc_captionsStylesDefaults[key] as any
}

export function vc_getCaptionStyleLabel<T extends keyof VideoCoreSettings["captionCustomization"]>(settings: VideoCoreSettings["captionCustomization"] | undefined,
    key: T,
): string {
    switch (key) {
        case "fontSize":
            return CAPTION_STYLES_FONT_SIZE_OPTIONS.find(o => o.value === vc_getCaptionStyle(settings, key))?.label ?? ""
        case "textShadow":
            return CAPTION_STYLES_TEXT_SHADOW_OPTIONS.find(o => o.value === vc_getCaptionStyle(settings, key))?.label ?? ""
        case "backgroundColor":
        case "textShadowColor":
        case "textColor":
            return CAPTION_STYLES_COLOR_OPTIONS.find(o => o.value === vc_getCaptionStyle(settings, key))?.label ?? ""
        case "backgroundOpacity":
            return `${(vc_getCaptionStyle(settings, "backgroundOpacity") * 100).toFixed(0)}%`
    }
    return ""
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function VideoCoreSettingsMenu() {
    const serverStatus = useServerStatus()
    const isMobile = useAtomValue(vc_isMobile)
    const action = useSetAtom(vc_dispatchAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const playbackRate = useAtomValue(vc_playbackRate)
    const setPlaybackRate = useSetAtom(vc_storedPlaybackRateAtom)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)

    const [anime4kOption, setAnime4kOption] = useAtom(vc_anime4kOption)
    const currentAnime4kOption = getAnime4KOptionByValue(anime4kOption)

    const [, setKeybindingsModelOpen] = useAtom(videoCorePreferencesModalAtom)

    const [showChapterMarkers, setShowChapterMarkers] = useAtom(vc_showChapterMarkersAtom)
    const [highlightOPEDChapters, setHighlightOPEDChapters] = useAtom(vc_highlightOPEDChaptersAtom)
    const [beautifyImage, setBeautifyImage] = useAtom(vc_beautifyImageAtom)
    const [autoNext, setAutoNext] = useAtom(vc_autoNextAtom)
    const [autoPlay, setAutoPlay] = useAtom(vc_autoPlayVideoAtom)
    const [autoSkipOPED, setAutoSkipOPED] = useAtom(vc_autoSkipOPEDAtom)

    const [menuOpen, setMenuOpen] = useAtom(vc_menuOpen)
    const [openMenuSection, setOpenMenuSection] = useAtom(vc_menuSectionOpen)
    const [openMenuSubSection, setOpenMenuSubSection] = useAtom(vc_menuSubSectionOpen)

    const { mutate: openInExplorer, isPending: isOpeningInExplorer } = useOpenInExplorer()

    const [settings, setSettings] = useAtom(vc_settings)

    const [editedSubCustomization, setEditedSubCustomization] = useState<VideoCoreSettings["subtitleCustomization"]>(
        settings.subtitleCustomization || vc_initialSettings.subtitleCustomization,
    )

    const [editedCaptionCustomization, setEditedCaptionCustomization] = useState<VideoCoreSettings["captionCustomization"]>(
        settings.captionCustomization || vc_initialSettings.captionCustomization,
    )

    const [editedSubtitleDelay, setEditedSubtitleDelay] = useState(settings.subtitleDelay ?? 0)

    const [subFontName, setSubFontName] = useState<string>(editedSubCustomization?.fontName || "")

    React.useEffect(() => {
        if (openMenuSection === "Subtitle Styles") {
            setEditedSubCustomization(settings.subtitleCustomization || vc_initialSettings.subtitleCustomization)
        }
        if (openMenuSection === "Caption Styles") {
            setEditedCaptionCustomization(settings.captionCustomization || vc_initialSettings.captionCustomization)
        }
        if (openMenuSection === "Subtitle Delay") {
            setEditedSubtitleDelay(settings.subtitleDelay)
        }
    }, [openMenuSection, settings])

    const handleSaveSettings = (customization?: VideoCoreSettings["subtitleCustomization"]) => {
        const newSettings = {
            ...settings,
            subtitleCustomization: customization || editedSubCustomization,
        }
        setSettings(newSettings)
        subtitleManager?.updateSettings(newSettings)

        // // Go back to submenu after saving from sub-submenu
        // setOpenMenuSubSection(null)
    }

    const handleSaveCaptionSettings = (customization?: VideoCoreSettings["captionCustomization"]) => {
        const newSettings = {
            ...settings,
            captionCustomization: customization || editedCaptionCustomization,
        }
        setSettings(newSettings)
        mediaCaptionsManager?.updateSettings(newSettings)
    }

    const handleSubtitleCustomizationChange = <K extends keyof VideoCoreSettings["subtitleCustomization"]>(
        key: K,
        value: VideoCoreSettings["subtitleCustomization"][K],
    ): void => {
        const newCustomization = {
            ...editedSubCustomization,
            [key]: value,
        }
        setEditedSubCustomization(newCustomization)
        React.startTransition(() => {
            handleSaveSettings(newCustomization)
        })
    }

    const handleCaptionCustomizationChange = <K extends keyof VideoCoreSettings["captionCustomization"]>(
        key: K,
        value: VideoCoreSettings["captionCustomization"][K],
    ): void => {
        const newCustomization = {
            ...editedCaptionCustomization,
            [key]: value,
        }
        setEditedCaptionCustomization(newCustomization)
        React.startTransition(() => {
            handleSaveCaptionSettings(newCustomization)
        })
    }

    const handleSubtitleDelayChange = (delay: number): void => {
        setEditedSubtitleDelay(delay)
        const newSettings = {
            ...settings,
            subtitleDelay: delay,
        }
        setSettings(newSettings)
        subtitleManager?.updateSettings(newSettings)
        mediaCaptionsManager?.updateSettings(newSettings)
    }

    if (isMiniPlayer) return null

    return (
        <>
            {playbackRate !== 1 && (
                <p
                    className="text-sm text-[--muted] cursor-pointer" onClick={() => {
                    setMenuOpen("settings")
                    React.startTransition(() => {
                        setOpenMenuSection("Playback Speed")
                    })
                }}
                >
                    {`${(playbackRate).toFixed(2)}x`}
                </p>
            )}
            <VideoCoreMenu
                name="settings"
                trigger={<VideoCoreControlButtonIcon
                    icons={[
                        ["default", isMobile ? LuSettings : LuChevronUp],
                    ]}
                    state="default"
                    className={isMobile ? "text-xl" : ""}
                    onClick={() => {
                    }}
                />}
            >
                <VideoCoreMenuSectionBody>
                    <VideoCoreMenuTitle>Settings</VideoCoreMenuTitle>
                    <VideoCoreMenuOption title="Playback Speed" icon={MdSpeed} value={`${(playbackRate).toFixed(2)}x`} />
                    <VideoCoreMenuOption title="Auto Play" icon={IoCaretForwardCircleOutline} value={autoPlay ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Auto Next" icon={HiFastForward} value={autoNext ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Skip OP/ED" icon={TbArrowForwardUp} value={autoSkipOPED ? "On" : "Off"} />
                    <VideoCoreMenuOption title="Anime4K" icon={LuSparkles} value={currentAnime4kOption?.label || "Off"} />
                    {(subtitleManager || mediaCaptionsManager) && <VideoCoreMenuOption
                        title="Subtitle Delay"
                        icon={MdOutlineAccessTime}
                        value={`${settings.subtitleDelay.toFixed(1)}s`}
                    />}
                    {subtitleManager && <VideoCoreMenuOption
                        title="Subtitle Styles"
                        icon={MdOutlineSubtitles}
                        value={editedSubCustomization?.enabled ? `On${!!editedSubCustomization?.fontName ? ", Font" : ""}` : "Off"}
                    />}
                    {mediaCaptionsManager && <VideoCoreMenuOption
                        title="Caption Styles"
                        icon={MdOutlineSubtitles}
                    />}
                    <VideoCoreMenuOption title="Player Appearance" icon={LuTvMinimalPlay} />
                    <VideoCoreMenuOption title="Preferences" icon={LuSettings2} onClick={() => setKeybindingsModelOpen(true)} />
                </VideoCoreMenuSectionBody>
                <VideoCoreMenuSubmenuBody>
                    <VideoCoreMenuOption title="Subtitle Styles" icon={MdOutlineSubtitles}>
                        <p className="text-sm text-[--muted] mb-2">Subtitle customization will not override ASS/SSA tracks that contain multiple
                                                                   styles.</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("enabled", v === 1)}
                            value={editedSubCustomization.enabled ? 1 : 0}
                        />
                        {editedSubCustomization.enabled && <>
                            <p className="text-[--muted] text-sm my-2">Options</p>
                            <VideoCoreMenuSubOption
                                title="Font"
                                icon={LuHeading}
                                parentId="Subtitle Styles"
                                value={!editedSubCustomization.fontName ? "Default" : editedSubCustomization.fontName?.slice(0,
                                    11) + (!!editedSubCustomization.fontName?.length && editedSubCustomization.fontName?.length > 10
                                    ? "..."
                                    : "")}
                            />
                            <VideoCoreMenuSubOption
                                title="Font Size"
                                icon={VscTextSize}
                                parentId="Subtitle Styles"
                                value={vc_getSubtitleStyleLabel(settings.subtitleCustomization, "fontSize")}
                            />
                            <VideoCoreMenuSubOption
                                title="Text Color"
                                icon={LuPalette}
                                parentId="Subtitle Styles"
                                value={vc_getSubtitleStyleLabel(settings.subtitleCustomization, "primaryColor")}
                            />
                            <VideoCoreMenuSubOption
                                title="Outline"
                                icon={ImFileText}
                                parentId="Subtitle Styles"
                                value={`${vc_getSubtitleStyleLabel(settings.subtitleCustomization,
                                    "outline")}, ${vc_getSubtitleStyleLabel(settings.subtitleCustomization, "outlineColor")}`}
                            />
                            <VideoCoreMenuSubOption
                                title="Shadow"
                                icon={RiShadowLine}
                                parentId="Subtitle Styles"
                                value={`${vc_getSubtitleStyleLabel(settings.subtitleCustomization,
                                    "shadow")}, ${vc_getSubtitleStyleLabel(settings.subtitleCustomization, "backColor")}`}
                            />
                        </>}
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Caption Styles" icon={MdOutlineSubtitles}>
                        <p className="text-sm text-[--muted] mb-2">This only applies to non-ASS subtitles.</p>
                        {/*<VideoCoreSettingSelect*/}
                        {/*    options={[*/}
                        {/*        { label: "On", value: 1 },*/}
                        {/*        { label: "Off", value: 0 },*/}
                        {/*    ]}*/}
                        {/*    onValueChange={(v: number) => handleCaptionCustomizationChange("enabled", v === 1)}*/}
                        {/*    value={editedCaptionCustomization.enabled ? 1 : 0}*/}
                        {/*/>*/}
                        {/*{editedCaptionCustomization.enabled && <>*/}
                        <p className="text-[--muted] text-sm my-2">Options</p>
                        <VideoCoreMenuSubOption
                            title="Font Size"
                            icon={VscTextSize}
                            parentId="Caption Styles"
                            value={vc_getCaptionStyleLabel(settings.captionCustomization, "fontSize")}
                        />
                        {/*<VideoCoreMenuSubOption title="Font Family" icon={LuHeading} parentId="Caption Styles" />*/}
                        <VideoCoreMenuSubOption
                            title="Text Color"
                            icon={LuPalette}
                            parentId="Caption Styles"
                            value={vc_getCaptionStyleLabel(settings.captionCustomization, "textColor")}
                        />
                        <VideoCoreMenuSubOption
                            title="Background"
                            icon={LuPaintbrush}
                            parentId="Caption Styles"
                            value={`${vc_getCaptionStyleLabel(settings.captionCustomization,
                                "backgroundOpacity")}, ${vc_getCaptionStyleLabel(settings.captionCustomization, "backgroundColor")}`}
                        />
                        {/*<VideoCoreMenuSubOption title="Outline" icon={ImFileText} parentId="Caption Styles" />*/}
                        <VideoCoreMenuSubOption
                            title="Shadow"
                            icon={RiShadowLine}
                            parentId="Caption Styles"
                            value={`${vc_getCaptionStyleLabel(settings.captionCustomization,
                                "textShadow")}, ${vc_getCaptionStyleLabel(settings.captionCustomization, "textShadowColor")}`}
                        />
                        {/*</>}*/}
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Subtitle Delay" icon={MdOutlineAccessTime}>
                        <p className="text-sm text-[--muted] mb-2">Positive values delay subtitles, negative values advance them.</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "-2.0s", value: -2.0 },
                                { label: "-1.0s", value: -1.0 },
                                { label: "-0.9s", value: -0.9 },
                                { label: "-0.8s", value: -0.8 },
                                { label: "-0.7s", value: -0.7 },
                                { label: "-0.6s", value: -0.6 },
                                { label: "-0.5s", value: -0.5 },
                                { label: "-0.4s", value: -0.4 },
                                { label: "-0.3s", value: -0.3 },
                                { label: "-0.2s", value: -0.2 },
                                { label: "-0.1s", value: -0.1 },
                                { label: "0s", value: 0 },
                                { label: "0.1s", value: 0.1 },
                                { label: "0.2s", value: 0.2 },
                                { label: "0.3s", value: 0.3 },
                                { label: "0.4s", value: 0.4 },
                                { label: "0.5s", value: 0.5 },
                                { label: "0.6s", value: 0.6 },
                                { label: "0.7s", value: 0.7 },
                                { label: "0.8s", value: 0.8 },
                                { label: "0.9s", value: 0.9 },
                                { label: "1.0s", value: 1.0 },
                                { label: "2.0s", value: 2.0 },
                            ]}
                            onValueChange={(v: number) => {
                                handleSubtitleDelayChange(v)
                            }}
                            value={editedSubtitleDelay}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Playback Speed" icon={MdSpeed}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "0.5x", value: 0.5 },
                                { label: "0.9x", value: 0.9 },
                                { label: "1x", value: 1 },
                                { label: "1.1x", value: 1.1 },
                                { label: "1.5x", value: 1.5 },
                                { label: "2x", value: 2 },
                            ]}
                            onValueChange={(v: number) => {
                                setPlaybackRate(v)
                            }}
                            value={playbackRate}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Auto Play" icon={IoCaretForwardCircleOutline}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoPlay(!!v)
                            }}
                            value={autoPlay ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Auto Next" icon={HiFastForward}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoNext(!!v)
                            }}
                            value={autoNext ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Skip OP/ED" icon={TbArrowForwardUp}>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "On", value: 1 },
                                { label: "Off", value: 0 },
                            ]}
                            onValueChange={(v: number) => {
                                setAutoSkipOPED(!!v)
                            }}
                            value={autoSkipOPED ? 1 : 0}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Anime4K" icon={LuSparkles}>
                        <p className="text-[--muted] text-sm mb-2">
                            Real-time sharpening. GPU-intensive.
                        </p>
                        <VideoCoreSettingSelect
                            isFullscreen={isFullscreen}
                            containerElement={containerElement}
                            options={anime4kOptions.map(option => ({
                                label: `${option.label}`,
                                value: option.value,
                                moreInfo: option.performance === "heavy" ? "Heavy" : undefined,
                                description: option.description,
                            }))}
                            onValueChange={(value: Anime4KOption) => {
                                setAnime4kOption(value)
                            }}
                            value={anime4kOption}
                        />
                    </VideoCoreMenuOption>
                    <VideoCoreMenuOption title="Player Appearance" icon={LuPaintbrush}>
                        <Switch
                            label="Show Chapter Markers"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={showChapterMarkers}
                            onValueChange={setShowChapterMarkers}
                        />
                        <Switch
                            label="Highlight OP/ED Chapters"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={highlightOPEDChapters}
                            onValueChange={setHighlightOPEDChapters}
                        />
                        <Switch
                            label="Increase Saturation"
                            side="right"
                            fieldClass="hover:bg-transparent hover:border-transparent px-0 ml-0 w-full"
                            size="sm"
                            value={beautifyImage}
                            onValueChange={setBeautifyImage}
                        />
                    </VideoCoreMenuOption>
                </VideoCoreMenuSubmenuBody>
                <VideoCoreMenuSubSubmenuBody>
                    <VideoCoreMenuSubOption title="Font" icon={VscTextSize} parentId="Subtitle Styles">
                        <div className="">
                            <p className="text-sm mb-2">Custom Font</p>
                            <p className="text-sm text-[--muted] mb-2">
                                Place the font file in the <span
                                className="text-indigo-300 cursor-pointer underline underline-offset-2"
                                onClick={() => {
                                    openInExplorer({ path: upath.normalize(`${serverStatus?.dataDir}/assets`) })
                                }}
                            >Seanime assets directory</span>. The file name must match
                                the font name exactly.
                            </p>
                            <div className="space-y-2">
                                <VideoCoreSettingTextInput
                                    label="File Name"
                                    value={subFontName ?? ""}
                                    onValueChange={(v: string) => setSubFontName(v)}
                                    help="Example: Noto Sans JP.woff2"
                                />
                                <div className="flex w-full">
                                    <Button
                                        size="sm" intent="gray-glass" onClick={() => {
                                        handleSubtitleCustomizationChange("fontName", subFontName)
                                    }}
                                    >
                                        Save
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Font Size" icon={LuHeading} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Font Size</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_FONT_SIZE_OPTIONS}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("fontSize", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "fontSize")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Subtitle Styles">
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("primaryColor", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "primaryColor")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Outline" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Outline Width</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_OUTLINE_WIDTH_OPTIONS}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("outline", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "outline")}
                        />
                        <p className="text-[--muted] text-sm my-2">Outline Color</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("outlineColor", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "outlineColor")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Shadow" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Shadow Depth</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_SHADOW_DEPTH_OPTIONS}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("shadow", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "shadow")}
                        />
                        <p className="text-[--muted] text-sm my-2">Shadow Opacity</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_BACK_COLOR_OPACITY_OPTIONS}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("backColorOpacity", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "backColorOpacity")}
                        />
                        <p className="text-[--muted] text-sm my-2">Shadow Color</p>
                        <VideoCoreSettingSelect
                            options={SUBTITLE_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("backColor", v)}
                            value={vc_getSubtitleStyle(editedSubCustomization, "backColor")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Font Size" icon={VscTextSize} parentId="Caption Styles">
                        {/*<p className="text-[--muted] text-sm mb-2">Font size as percentage of video height</p>*/}
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_FONT_SIZE_OPTIONS}
                            onValueChange={(v: number) => handleCaptionCustomizationChange("fontSize", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "fontSize")}
                        />
                    </VideoCoreMenuSubOption>
                    {/*<VideoCoreMenuSubOption title="Font Family" icon={LuHeading} parentId="Caption Styles">*/}
                    {/*    /!*<p className="text-[--muted] text-sm mb-2">Font family for captions</p>*!/*/}
                    {/*    <VideoCoreSettingSelect*/}
                    {/*        options={[*/}
                    {/*            { label: "Inter", value: "Inter, Arial, sans-serif" },*/}
                    {/*            { label: "Arial", value: "Arial, sans-serif" },*/}
                    {/*            { label: "Courier", value: "Courier New, monospace" },*/}
                    {/*            { label: "Georgia", value: "Georgia, serif" },*/}
                    {/*            { label: "Times", value: "Times New Roman, serif" },*/}
                    {/*        ]}*/}
                    {/*        onValueChange={(v: string) => handleCaptionCustomizationChange("fontFamily", v)}*/}
                    {/*        value={editedCaptionCustomization.fontFamily ?? "Inter, Arial, sans-serif"}*/}
                    {/*    />*/}
                    {/*</VideoCoreMenuSubOption>*/}
                    <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Caption Styles">
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleCaptionCustomizationChange("textColor", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "textColor")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Background" icon={LuPaintbrush} parentId="Caption Styles">
                        <p className="text-[--muted] text-sm my-2">Background Opacity</p>
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_BACKGROUND_OPACITY_OPTIONS}
                            onValueChange={(v: number) => handleCaptionCustomizationChange("backgroundOpacity", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "backgroundOpacity")}
                        />
                        <p className="text-[--muted] text-sm mb-2">Background Color</p>
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleCaptionCustomizationChange("backgroundColor", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "backgroundColor")}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Shadow" icon={RiShadowLine} parentId="Caption Styles">
                        <p className="text-[--muted] text-sm mb-2">Text shadow</p>
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_TEXT_SHADOW_OPTIONS}
                            onValueChange={(v: number) => handleCaptionCustomizationChange("textShadow", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "textShadow")}
                        />
                        <p className="text-[--muted] text-sm my-2">Shadow Color</p>
                        <VideoCoreSettingSelect
                            options={CAPTION_STYLES_COLOR_OPTIONS}
                            onValueChange={(v: string) => handleCaptionCustomizationChange("textShadowColor", v)}
                            value={vc_getCaptionStyle(editedCaptionCustomization, "textShadowColor")}
                        />
                    </VideoCoreMenuSubOption>
                </VideoCoreMenuSubSubmenuBody>
            </VideoCoreMenu>
        </>
    )
}
