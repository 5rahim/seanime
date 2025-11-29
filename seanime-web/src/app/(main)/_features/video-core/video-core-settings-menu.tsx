import {
    vc_containerElement,
    vc_dispatchAction,
    vc_isFullscreen,
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
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React, { useState } from "react"
import { HiFastForward } from "react-icons/hi"
import { ImFileText } from "react-icons/im"
import { IoCaretForwardCircleOutline } from "react-icons/io5"
import { LuChevronUp, LuHeading, LuPaintbrush, LuPalette, LuSettings2, LuSparkles, LuTvMinimalPlay } from "react-icons/lu"
import { MdOutlineSubtitles, MdSpeed } from "react-icons/md"
import { RiShadowLine } from "react-icons/ri"
import { TbArrowForwardUp } from "react-icons/tb"
import { VscTextSize } from "react-icons/vsc"

export function VideoCoreSettingsMenu() {
    const action = useSetAtom(vc_dispatchAction)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const playbackRate = useAtomValue(vc_playbackRate)
    const setPlaybackRate = useSetAtom(vc_storedPlaybackRateAtom)
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const subtitleManager = useAtomValue(vc_subtitleManager)

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

    const [settings, setSettings] = useAtom(vc_settings)

    const [editedSubCustomization, setEditedSubCustomization] = useState<VideoCoreSettings["subtitleCustomization"]>(
        settings.subtitleCustomization || vc_initialSettings.subtitleCustomization,
    )

    const [subFontName, setSubFontName] = useState<string>(editedSubCustomization?.fontName || "")

    React.useEffect(() => {
        if (openMenuSection === "Subtitle Styles") {
            setEditedSubCustomization(settings.subtitleCustomization || vc_initialSettings.subtitleCustomization)
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
                        ["default", LuChevronUp],
                    ]}
                    state="default"
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
                    <VideoCoreMenuOption
                        title="Subtitle Styles"
                        icon={MdOutlineSubtitles}
                        value={editedSubCustomization?.enabled ? `On${!!editedSubCustomization?.fontName ? ", Font" : ""}` : "Off"}
                    />
                    <VideoCoreMenuOption title="Player Appearance" icon={LuTvMinimalPlay} />
                    <VideoCoreMenuOption title="Preferences" icon={LuSettings2} onClick={() => setKeybindingsModelOpen(true)} />
                </VideoCoreMenuSectionBody>
                <VideoCoreMenuSubmenuBody>
                    <VideoCoreMenuOption title="Subtitle Styles" icon={MdOutlineSubtitles}>
                        <p className="text-sm text-[--muted] mb-2">Subtitle customization will not override ASS/SAA tracks that contain multiple
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
                                value={editedSubCustomization.fontName?.slice(0,
                                    11) + (!!editedSubCustomization.fontName?.length && editedSubCustomization.fontName?.length > 10
                                    ? "..."
                                    : "")}
                            />
                            <VideoCoreMenuSubOption title="Font Size" icon={VscTextSize} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Outline" icon={ImFileText} parentId="Subtitle Styles" />
                            <VideoCoreMenuSubOption title="Shadow" icon={RiShadowLine} parentId="Subtitle Styles" />
                        </>}
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
                                Place the font files in a folder named <strong>assets</strong> in the Seanime data directory. The file name must match
                                the font name exactly.
                            </p>
                            <div className="space-y-2">
                                <VideoCoreSettingTextInput
                                    label="File Name"
                                    value={subFontName ?? ""}
                                    onValueChange={(v: string) => setSubFontName(v)}
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
                            options={[
                                { label: "Small", value: 54 },
                                { label: "Medium", value: 62 },
                                { label: "Large", value: 72 },
                                { label: "Extra Large", value: 82 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("fontSize", v)}
                            value={editedSubCustomization.fontSize ?? 62}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Text Color" icon={LuPalette} parentId="Subtitle Styles">
                        <VideoCoreSettingSelect
                            options={[
                                { label: "White", value: "#FFFFFF" },
                                { label: "Black", value: "#000000" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("primaryColor", v)}
                            value={editedSubCustomization.primaryColor ?? "#FFFFFF"}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Outline" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Outline Width</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "None", value: 0 },
                                { label: "Small", value: 2 },
                                { label: "Medium", value: 3 },
                                { label: "Large", value: 4 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("outline", v)}
                            value={editedSubCustomization.outline ?? 3}
                        />
                        <p className="text-[--muted] text-sm my-2">Outline Color</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "Black", value: "#000000" },
                                { label: "White", value: "#FFFFFF" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("outlineColor", v)}
                            value={editedSubCustomization.outlineColor ?? "#000000"}
                        />
                    </VideoCoreMenuSubOption>
                    <VideoCoreMenuSubOption title="Shadow" icon={LuPalette} parentId="Subtitle Styles">
                        <p className="text-[--muted] text-sm mb-2">Shadow Depth</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "None", value: 0 },
                                { label: "Small", value: 1 },
                                { label: "Medium", value: 2 },
                                { label: "Large", value: 3 },
                            ]}
                            onValueChange={(v: number) => handleSubtitleCustomizationChange("shadow", v)}
                            value={editedSubCustomization.shadow ?? 0}
                        />
                        <p className="text-[--muted] text-sm my-2">Shadow Color</p>
                        <VideoCoreSettingSelect
                            options={[
                                { label: "Black", value: "#000000" },
                                { label: "White", value: "#FFFFFF" },
                                { label: "Yellow", value: "#FFD700" },
                                { label: "Cyan", value: "#00FFFF" },
                                { label: "Pink", value: "#FF69B4" },
                                { label: "Purple", value: "#9370DB" },
                                { label: "Lime", value: "#00FF00" },
                            ]}
                            onValueChange={(v: string) => handleSubtitleCustomizationChange("backColor", v)}
                            value={editedSubCustomization.backColor ?? "#000000"}
                        />
                    </VideoCoreMenuSubOption>
                </VideoCoreMenuSubSubmenuBody>
            </VideoCoreMenu>
        </>
    )
}
