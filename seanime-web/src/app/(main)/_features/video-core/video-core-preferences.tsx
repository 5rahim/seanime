import { useSaveMediaPlayerSettings } from "@/api/hooks/settings.hooks"
import {
    vc_audioManager,
    vc_containerElement,
    vc_dispatchAction,
    vc_isFullscreen,
    vc_isMuted,
    vc_mediaCaptionsManager,
    vc_pip,
    vc_subtitleManager,
    vc_volume,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { vc_fullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import { useVideoCoreOverlayFeedback } from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { vc_pipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import {
    vc_defaultKeybindings,
    vc_initialSettings,
    vc_keybindingsAtom,
    vc_settings,
    vc_storedMutedAtom,
    vc_storedVolumeAtom,
    vc_useLibassRendererAtom,
    VideoCoreKeybindings,
} from "@/app/(main)/_features/video-core/video-core.atoms"
import { AlphaBadge } from "@/components/shared/beta-badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { logger } from "@/lib/helpers/debug"
import { atom, useAtom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { UseFormReturn } from "react-hook-form"
import { toast } from "sonner"
import { useServerStatus } from "../../_hooks/use-server-status"
import { useVideoCoreScreenshot } from "./video-core-screenshot"

export const videoCorePreferencesModalAtom = atom(false)

const tabsRootClass = cn("w-full contents space-y-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)

const tabsListClass = cn(
    "w-full flex flex-row lg:flex-row flex-wrap h-fit !mt-4",
)

const tabContentClass = cn(
    "space-y-4 animate-in fade-in-0 duration-300",
)

const translationSettingsSchema = defineSchema(({ z, presets }) => z.object({
    vcTranslate: z.boolean().default(false),
    vcTranslateProvider: z.string().default("google"),
    vcTranslateTargetLanguage: z.string().default("en"),
    vcTranslateApiKey: z.string().default(""),
}))

const KeybindingValueInput = ({
    actionKey,
    value,
    onValueChange,
}: {
    actionKey: keyof VideoCoreKeybindings
    value: number
    onValueChange: (value: number) => void
}) => {
    return (
        <NumberInput
            value={value}
            onValueChange={onValueChange}
            size="sm"
            fieldClass="w-16"
            hideControls
            min={0}
            step={actionKey.includes("Speed") ? 0.25 : 1}
            // onKeyDown={(e) => e.stopPropagation()}
            // onInput={(e) => e.stopPropagation()}
        />
    )
}

const KeybindingRow = ({
    action,
    description,
    actionKey,
    hasValue = false,
    valueLabel = "",
    editedKeybindings,
    setEditedKeybindings,
    recordingKey,
    handleKeyRecord,
    formatKeyDisplay = (actionKey: keyof VideoCoreKeybindings) => actionKey,
}: {
    action: string
    description: string
    actionKey: keyof VideoCoreKeybindings
    hasValue?: boolean
    valueLabel?: string
    editedKeybindings: VideoCoreKeybindings
    setEditedKeybindings: React.Dispatch<React.SetStateAction<VideoCoreKeybindings>>
    recordingKey: string | null
    handleKeyRecord: (actionKey: keyof VideoCoreKeybindings) => void
    formatKeyDisplay?: (actionKey: keyof VideoCoreKeybindings) => keyof VideoCoreKeybindings | string
}) => (
    <div className="flex items-center justify-between py-2 border rounded-lg px-3 bg-[--paper]">
        <div className="flex-1">
            <div className="font-medium text-sm">{action}</div>
            {hasValue && (
                <div className="flex items-center gap-2 mt-1">
                    <span className="text-xs text-muted-foreground">{valueLabel}:</span>
                    <KeybindingValueInput
                        actionKey={actionKey}
                        value={("value" in editedKeybindings[actionKey]) ? (editedKeybindings[actionKey] as any).value : 0}
                        onValueChange={(value) => {
                            setEditedKeybindings(prev => ({
                                ...prev,
                                [actionKey]: { ...prev[actionKey], value: value || 0 },
                            }))
                        }}
                    />
                </div>
            )}
        </div>
        <div className="flex items-center gap-2">
            <Button
                intent={recordingKey === actionKey ? "white-subtle" : "gray-glass"}
                size="sm"
                onClick={() => handleKeyRecord(actionKey)}
                className={cn(
                    "h-8 px-3 text-lg font-mono",
                    recordingKey === actionKey && "!text-xs text-white",
                )}
            >
                {recordingKey === actionKey ? "Press key..." : formatKeyDisplay(editedKeybindings?.[actionKey]?.key as any ?? "" as any)}
            </Button>
        </div>
    </div>
)

export function VideoCorePreferencesModal() {
    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const [open, setOpen] = useAtom(videoCorePreferencesModalAtom)
    const [keybindings, setKeybindings] = useAtom(vc_keybindingsAtom)
    const [editedKeybindings, setEditedKeybindings] = useState<VideoCoreKeybindings>(keybindings)
    const [useLibassRenderer, setUseLibassRenderer] = useAtom(vc_useLibassRendererAtom)
    const [editedUseLibassRenderer, setEditedUseLibassRenderer] = useState(useLibassRenderer)

    const [recordingKey, setRecordingKey] = useState<string | null>(null)

    const [tab, setTab] = useState("keybinds")
    const { mutate: saveMediaPlayerSettings } = useSaveMediaPlayerSettings()
    const serverStatus = useServerStatus()
    const translationFormRef = useRef<UseFormReturn<any>>(null)

    const [settings, setSettings] = useAtom(vc_settings)
    const [editedSubLanguage, setEditedSubLanguage] = useState(settings.preferredSubtitleLanguage)
    const [editedAudioLanguage, setEditedAudioLanguage] = useState(settings.preferredAudioLanguage)
    const [editedSubsBlacklist, setEditedSubsBlacklist] = useState(settings.preferredSubtitleBlacklist)
    const [editedSubtitleDelay, setEditedSubtitleDelay] = useState(settings.subtitleDelay ?? 0)
    // const [editedSubCustomization, setEditedSubCustomization] = useState<VideoCoreSettings["subtitleCustomization"]>(
    //     settings.subtitleCustomization || vc_initialSettings.subtitleCustomization
    // )
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)

    // Reset edited keybindings and language preferences when modal opens
    useEffect(() => {
        if (open) {
            setEditedKeybindings(keybindings)
            setEditedSubLanguage(settings.preferredSubtitleLanguage)
            setEditedAudioLanguage(settings.preferredAudioLanguage)
            setEditedSubsBlacklist(settings.preferredSubtitleBlacklist)
            setEditedSubtitleDelay(settings.subtitleDelay ?? 0)
            setEditedUseLibassRenderer(useLibassRenderer)
            // setEditedSubCustomization(settings.subtitleCustomization || vc_initialSettings.subtitleCustomization)
        }
    }, [open, keybindings, settings, useLibassRenderer])

    const handleKeyRecord = (actionKey: keyof VideoCoreKeybindings) => {
        setRecordingKey(actionKey)

        const handleKeyDown = (e: KeyboardEvent) => {
            e.preventDefault()
            e.stopPropagation()
            e.stopImmediatePropagation()

            setEditedKeybindings(prev => ({
                ...prev,
                [actionKey]: {
                    ...prev[actionKey],
                    key: e.code,
                },
            }))

            setRecordingKey(null)
            document.removeEventListener("keydown", handleKeyDown, true)
        }

        document.addEventListener("keydown", handleKeyDown, true)
    }

    const handleSave = () => {
        setKeybindings(editedKeybindings)
        const newSettings = {
            ...settings,
            preferredSubtitleLanguage: editedSubLanguage,
            preferredAudioLanguage: editedAudioLanguage,
            preferredSubtitleBlacklist: editedSubsBlacklist,
            subtitleDelay: editedSubtitleDelay,
            // subtitleCustomization: editedSubCustomization,
        }
        setSettings(newSettings)
        setUseLibassRenderer(editedUseLibassRenderer)
        // Update subtitle manager with new settings
        subtitleManager?.updateSettings(newSettings)
        mediaCaptionsManager?.updateSettings(newSettings)
        setOpen(false)
    }

    const handleReset = () => {
        setEditedKeybindings(vc_defaultKeybindings)
        setEditedSubLanguage(vc_initialSettings.preferredSubtitleLanguage)
        setEditedAudioLanguage(vc_initialSettings.preferredAudioLanguage)
        setEditedSubsBlacklist(vc_initialSettings.preferredSubtitleBlacklist)
        setEditedSubtitleDelay(vc_initialSettings.subtitleDelay)
        setEditedUseLibassRenderer(true)
        // setEditedSubCustomization(vc_initialSettings.subtitleCustomization)
    }

    const formatKeyDisplay = (keyCode: string) => {
        const keyMap: Record<string, string> = {
            "KeyA": "A", "KeyB": "B", "KeyC": "C", "KeyD": "D", "KeyE": "E", "KeyF": "F",
            "KeyG": "G", "KeyH": "H", "KeyI": "I", "KeyJ": "J", "KeyK": "K", "KeyL": "L",
            "KeyM": "M", "KeyN": "N", "KeyO": "O", "KeyP": "P", "KeyQ": "Q", "KeyR": "R",
            "KeyS": "S", "KeyT": "T", "KeyU": "U", "KeyV": "V", "KeyW": "W", "KeyX": "X",
            "KeyY": "Y", "KeyZ": "Z",
            "ArrowUp": "↑", "ArrowDown": "↓", "ArrowLeft": "←", "ArrowRight": "→",
            "BracketLeft": "[", "BracketRight": "]",
            "Space": "⎵",
        }
        return keyMap[keyCode] || keyCode
    }

    return (
        <Modal
            title="Preferences"
            open={open}
            onOpenChange={setOpen}
            contentClass="max-w-5xl focus:outline-none focus-visible:outline-none outline-none bg-[--background] backdrop-blur-sm z-[101]"
            overlayClass="z-[150] bg-black/50"
            portalContainer={isFullscreen ? containerElement || undefined : undefined}
        >

            <Tabs
                value={tab}
                onValueChange={setTab}
                className={tabsRootClass}
                triggerClass={tabsTriggerClass}
                listClass={tabsListClass}
            >
                <TabsList className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-xl">
                    <TabsTrigger value="keybinds">Keyboard Shortcuts</TabsTrigger>
                    <TabsTrigger value="subtitles">Subtitles</TabsTrigger>
                    <TabsTrigger value="translation">Translation <AlphaBadge /></TabsTrigger>
                    {/*<TabsTrigger value="browser-client">Rendering</TabsTrigger>*/}
                </TabsList>

                <TabsContent value="keybinds" className={tabContentClass}>
                    <div className="space-y-3 hidden lg:block">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                            <div>
                                {/* <h3 className="text-lg font-semibold mb-4 text-white">Playback</h3> */}
                                <div className="space-y-3">
                                    <KeybindingRow
                                        action="Seek Forward (Fine)"
                                        description="Seek forward (fine)"
                                        actionKey="seekForwardFine"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Seconds"
                                    />
                                    <KeybindingRow
                                        action="Seek Backward (Fine)"
                                        description="Seek backward (fine)"
                                        actionKey="seekBackwardFine"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Seconds"
                                    />
                                    <KeybindingRow
                                        action="Seek Forward"
                                        description="Seek forward"
                                        actionKey="seekForward"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Seconds"
                                    />
                                    <KeybindingRow
                                        action="Seek Backward"
                                        description="Seek backward"
                                        actionKey="seekBackward"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Seconds"
                                    />
                                    <KeybindingRow
                                        action="Increase Speed"
                                        description="Increase playback speed"
                                        actionKey="increaseSpeed"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="increment"
                                    />
                                    <KeybindingRow
                                        action="Decrease Speed"
                                        description="Decrease playback speed"
                                        actionKey="decreaseSpeed"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="increment"
                                    />
                                </div>
                            </div>

                            <div>
                                {/* <h3 className="text-lg font-semibold mb-4 text-white">Navigation</h3> */}
                                <div className="space-y-3">
                                    <KeybindingRow
                                        action="Next Chapter"
                                        description="Skip to next chapter"
                                        actionKey="nextChapter"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Previous Chapter"
                                        description="Skip to previous chapter"
                                        actionKey="previousChapter"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Next Episode"
                                        description="Play next episode"
                                        actionKey="nextEpisode"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Previous Episode"
                                        description="Play previous episode"
                                        actionKey="previousEpisode"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Cycle Subtitles"
                                        description="Cycle through subtitle tracks"
                                        actionKey="cycleSubtitles"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Fullscreen"
                                        description="Toggle fullscreen"
                                        actionKey="fullscreen"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Picture in Picture"
                                        description="Toggle picture in picture"
                                        actionKey="pictureInPicture"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Take Screenshot"
                                        description="Take screenshot"
                                        actionKey="takeScreenshot"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                </div>
                            </div>

                            <div>
                                {/* <h3 className="text-lg font-semibold mb-4 text-white">Audio</h3> */}
                                <div className="space-y-3">
                                    <KeybindingRow
                                        action="Volume Up"
                                        description="Increase volume"
                                        actionKey="volumeUp"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Percent"
                                    />
                                    <KeybindingRow
                                        action="Volume Down"
                                        description="Decrease volume"
                                        actionKey="volumeDown"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                        hasValue={true}
                                        valueLabel="Percent"
                                    />
                                    <KeybindingRow
                                        action="Mute"
                                        description="Toggle mute"
                                        actionKey="mute"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                    <KeybindingRow
                                        action="Cycle Audio"
                                        description="Cycle through audio tracks"
                                        actionKey="cycleAudio"
                                        editedKeybindings={editedKeybindings}
                                        setEditedKeybindings={setEditedKeybindings}
                                        recordingKey={recordingKey}
                                        handleKeyRecord={handleKeyRecord}
                                        formatKeyDisplay={formatKeyDisplay}
                                    />
                                </div>
                            </div>
                        </div>

                        <div className="flex items-center justify-between pt-6">
                            <Button
                                intent="gray-outline"
                                onClick={handleReset}
                            >
                                Reset all
                            </Button>
                            <div className="flex gap-2">
                                <Button
                                    intent="gray-outline"
                                    onClick={() => setOpen(false)}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    intent="primary"
                                    onClick={handleSave}
                                >
                                    Save
                                </Button>
                            </div>
                        </div>
                    </div>
                </TabsContent>
                <TabsContent value="subtitles" className={tabContentClass}>
                    <div className="space-y-3">
                        <h3 className="text-lg font-semibold text-white">Defaults</h3>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-muted-foreground">
                                    Preferred Subtitle Language
                                </label>
                                <TextInput
                                    value={editedSubLanguage}
                                    onValueChange={setEditedSubLanguage}
                                    placeholder="eng,jpn,spa"
                                    onKeyDown={(e) => e.stopPropagation()}
                                    onInput={(e) => e.stopPropagation()}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-muted-foreground">
                                    Preferred Audio Language
                                </label>
                                <TextInput
                                    value={editedAudioLanguage}
                                    onValueChange={setEditedAudioLanguage}
                                    placeholder="jpn,eng,kor"
                                    onKeyDown={(e) => e.stopPropagation()}
                                    onInput={(e) => e.stopPropagation()}
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-muted-foreground">
                                Ignored Subtitle Names
                            </label>
                            <TextInput
                                value={editedSubsBlacklist}
                                onValueChange={setEditedSubsBlacklist}
                                placeholder="e.g., sign & songs"
                                onKeyDown={(e) => e.stopPropagation()}
                                onInput={(e) => e.stopPropagation()}
                                help="Subtitle tracks that will not be selected by default if they match the preferred lanauges. Separate multiple names with commas."
                            />
                        </div>
                    </div>

                    <div className="space-y-3">
                        <h3 className="text-lg font-semibold text-white">Rendering</h3>
                        <div className="space-y-2">
                            <Switch
                                side="right"
                                label="Convert Soft Subs to ASS"
                                value={editedUseLibassRenderer}
                                onValueChange={setEditedUseLibassRenderer}
                                help="The player will convert other subtitle formats (SRT, VTT, ...) to ASS. In case your language is not supported, you can add a new font or disable this feature. Reloading the player is required after changing this setting."
                            />

                            {/*<div className="space-y-2">*/}
                            {/*    <label className="text-sm font-medium text-muted-foreground">*/}
                            {/*        Subtitle Delay (seconds)*/}
                            {/*    </label>*/}
                            {/*    <NumberInput*/}
                            {/*        value={editedSubtitleDelay}*/}
                            {/*        onValueChange={setEditedSubtitleDelay}*/}
                            {/*        fieldClass="w-32"*/}
                            {/*        step={0.1}*/}
                            {/*        hideControls={true}*/}
                            {/*        onKeyDown={(e) => e.stopPropagation()}*/}
                            {/*        onInput={(e) => e.stopPropagation()}*/}
                            {/*    />*/}
                            {/*</div>*/}
                        </div>
                    </div>

                    <div className="flex items-center justify-between pt-6">
                        <Button
                            intent="gray-outline"
                            onClick={handleReset}
                        >
                            Reset all
                        </Button>
                        <div className="flex gap-2">
                            <Button
                                intent="gray-outline"
                                onClick={() => setOpen(false)}
                            >
                                Cancel
                            </Button>
                            <Button
                                intent="primary"
                                onClick={handleSave}
                            >
                                Save
                            </Button>
                        </div>
                    </div>
                </TabsContent>
                <TabsContent value="translation" className={tabContentClass}>
                    <Form
                        schema={translationSettingsSchema}
                        onSubmit={data => {
                            const currentMediaPlayer = serverStatus?.settings?.mediaPlayer!

                            saveMediaPlayerSettings({
                                mediaPlayer: {
                                    ...currentMediaPlayer,
                                    vcTranslate: data.vcTranslate,
                                    vcTranslateTargetLanguage: data.vcTranslateTargetLanguage.toLowerCase(),
                                    vcTranslateProvider: data.vcTranslateProvider,
                                    vcTranslateApiKey: data.vcTranslateApiKey,
                                },
                            }, {
                                onSuccess: () => {
                                    toast.success("Translation settings saved")
                                    translationFormRef.current?.reset(translationFormRef.current.getValues())

                                    subtitleManager?.updateShouldTranslate(data.vcTranslate ? data.vcTranslateTargetLanguage : null)
                                    mediaCaptionsManager?.updateShouldTranslate(data.vcTranslate ? data.vcTranslateTargetLanguage : null)
                                },
                            })
                        }}
                        defaultValues={{
                            vcTranslate: serverStatus?.settings?.mediaPlayer?.vcTranslate ?? false,
                            vcTranslateProvider: serverStatus?.settings?.mediaPlayer?.vcTranslateProvider || "deepl",
                            vcTranslateTargetLanguage: serverStatus?.settings?.mediaPlayer?.vcTranslateTargetLanguage?.toLowerCase() || "en",
                            vcTranslateApiKey: serverStatus?.settings?.mediaPlayer?.vcTranslateApiKey || "",
                        }}
                        stackClass="space-y-4 relative"
                        mRef={translationFormRef}
                    >
                        {(f) => (
                            <div className="space-y-4">
                                <div className="space-y-4">
                                    <Field.Switch
                                        name="vcTranslate"
                                        side="right"
                                        label="Enable Translation"
                                        help="Automatically translate subtitle tracks to your selected language"
                                    />
                                    <div className="space-y-2">
                                        <Field.Select
                                            label="Provider"
                                            name="vcTranslateProvider"
                                            options={[
                                                { value: "deepl", label: "DeepL" },
                                                { value: "openai", label: "OpenAI" },
                                            ]}
                                            contentClass="z-[999]"
                                        />
                                    </div>

                                    {f.watch("vcTranslateProvider") === "deepl" && (
                                        <p>
                                            Note: DeepL does not support all target languages.
                                        </p>
                                    )}

                                    <div className="space-y-2">
                                        <Field.Select
                                            label="Target Language"
                                            name="vcTranslateTargetLanguage"
                                            options={[
                                                { value: "en", label: "English" },
                                                { value: "es", label: "Spanish" },
                                                { value: "fr", label: "French" },
                                                { value: "de", label: "German" },
                                                { value: "it", label: "Italian" },
                                                { value: "pt", label: "Portuguese" },
                                                { value: "ru", label: "Russian" },
                                                { value: "ja", label: "Japanese" },
                                                { value: "ko", label: "Korean" },
                                                { value: "zh", label: "Chinese" },
                                                { value: "ar", label: "Arabic" },
                                                { value: "hi", label: "Hindi" },
                                                { value: "tr", label: "Turkish" },
                                                { value: "pl", label: "Polish" },
                                                { value: "nl", label: "Dutch" },
                                                { value: "sv", label: "Swedish" },
                                                { value: "no", label: "Norwegian" },
                                                { value: "da", label: "Danish" },
                                                { value: "fi", label: "Finnish" },
                                                { value: "el", label: "Greek" },
                                                { value: "cs", label: "Czech" },
                                                { value: "hu", label: "Hungarian" },
                                                { value: "ro", label: "Romanian" },
                                                { value: "th", label: "Thai" },
                                                { value: "vi", label: "Vietnamese" },
                                                { value: "id", label: "Indonesian" },
                                                { value: "ms", label: "Malay" },
                                                { value: "uk", label: "Ukrainian" },
                                                { value: "bg", label: "Bulgarian" },
                                                { value: "hr", label: "Croatian" },
                                                { value: "sr", label: "Serbian" },
                                                { value: "sk", label: "Slovak" },
                                                { value: "sl", label: "Slovenian" },
                                                { value: "et", label: "Estonian" },
                                                { value: "lv", label: "Latvian" },
                                                { value: "lt", label: "Lithuanian" },
                                                { value: "he", label: "Hebrew" },
                                                { value: "fa", label: "Persian" },
                                                { value: "bn", label: "Bengali" },
                                                { value: "ur", label: "Urdu" },
                                                { value: "ta", label: "Tamil" },
                                                { value: "te", label: "Telugu" },
                                                { value: "mr", label: "Marathi" },
                                                { value: "kn", label: "Kannada" },
                                                { value: "ml", label: "Malayalam" },
                                                { value: "pa", label: "Punjabi" },
                                                { value: "sw", label: "Swahili" },
                                                { value: "af", label: "Afrikaans" },
                                            ]}
                                            contentClass="z-[999]"
                                            help="Select the language you want subtitles to be translated to"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Field.Text
                                            label="API Key"
                                            name="vcTranslateApiKey"
                                            placeholder="Enter your API key"
                                            onKeyDown={(e) => e.stopPropagation()}
                                            onInput={(e) => e.stopPropagation()}
                                        />
                                    </div>
                                </div>

                                <p className="text-[--muted]">
                                    Reloading the player is required only when switching languages or API Key.
                                </p>

                                <div className="flex items-center justify-end pt-6">
                                    <div className="flex gap-2">
                                        <Button
                                            type="button"
                                            intent="gray-outline"
                                            onClick={() => setOpen(false)}
                                        >
                                            Cancel
                                        </Button>
                                        <Button
                                            type="submit"
                                            intent="primary"
                                        >
                                            Save
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </Form>
                </TabsContent>
            </Tabs>

        </Modal>
    )
}

export function VideoCoreKeybindingController(props: {
    active: boolean
    videoRef: React.RefObject<HTMLVideoElement>,
    chapterCues: VideoCoreChapterCue[],
    introEndTime: number | undefined,
    introStartTime: number | undefined
    endingEndTime: number | undefined,
    endingStartTime: number | undefined
}) {
    const {
        active,
        videoRef,
        chapterCues,
        introEndTime,
        introStartTime,
        endingEndTime,
        endingStartTime,
    } = props

    const [keybindings] = useAtom(vc_keybindingsAtom)
    const isKeybindingsModalOpen = useAtomValue(videoCorePreferencesModalAtom)
    const fullscreen = useAtomValue(vc_isFullscreen)
    const pip = useAtomValue(vc_pip)
    const volume = useAtomValue(vc_volume)
    const setVolume = useSetAtom(vc_storedVolumeAtom)
    const muted = useAtomValue(vc_isMuted)
    const setMuted = useSetAtom(vc_storedMutedAtom)
    const { showOverlayFeedback } = useVideoCoreOverlayFeedback()

    const action = useSetAtom(vc_dispatchAction)

    const subtitleManager = useAtomValue(vc_subtitleManager)
    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)
    const audioManager = useAtomValue(vc_audioManager)
    const fullscreenManager = useAtomValue(vc_fullscreenManager)
    const pipManager = useAtomValue(vc_pipManager)

    // Rate limiting for seeking operations
    const lastSeekTime = useRef(0)
    const SEEK_THROTTLE_MS = 100 // Minimum time between seek operations

    function seek(seconds: number) {
        const isPaused = videoRef.current?.paused
        if (!isPaused) {
            videoRef.current?.pause()
        }
        action({ type: "seek", payload: { time: seconds, flashTime: true } })
        if (!isPaused) {
            videoRef.current?.play()?.catch()
        }
    }

    function seekTo(to: number) {
        const isPaused = videoRef.current?.paused
        if (!isPaused) {
            videoRef.current?.pause()
        }
        action({ type: "seekTo", payload: { time: to, flashTime: true } })
        if (!isPaused) {
            videoRef.current?.play()?.catch()
        }
    }

    const { takeScreenshot } = useVideoCoreScreenshot()

    //
    // Keyboard shortcuts
    //

    const handleKeyboardShortcuts = useCallback(async (e: KeyboardEvent) => {
        // Don't handle shortcuts if in an input/textarea or while keybindings modal is open
        if (isKeybindingsModalOpen || e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
            return
        }

        // Ignore combinations with modifier keys
        if (e.ctrlKey || e.shiftKey || e.altKey || e.metaKey) {
            return
        }

        if (!videoRef.current || !active) {
            return
        }

        const video = videoRef.current


        if (e.code === "Space" || e.code === "Enter") {
            e.preventDefault()
            if (video.paused) {
                await video.play()
                showOverlayFeedback({ message: "PLAY", type: "icon" })
            } else {
                video.pause()
                showOverlayFeedback({ message: "PAUSE", type: "icon" })
            }
            return
        }

        // Home, go to beginning
        if (e.code === "Home") {
            e.preventDefault()
            seekTo(0)
            showOverlayFeedback({ message: "Beginning" })
            return
        }

        // End, go to end
        if (e.code === "End") {
            e.preventDefault()
            seekTo(video.duration)
            showOverlayFeedback({ message: "End" })
            return
        }

        // Escape - Exit fullscreen
        if (e.code === "Escape" && fullscreen) {
            e.preventDefault()
            fullscreenManager?.exitFullscreen()
            return
        }

        // Number keys 0-9, seek to percentage (0%, 10%, 20%, ..., 90%)
        if (e.code.startsWith("Digit") && e.code.length === 6) {
            e.preventDefault()
            const digit = parseInt(e.code.slice(-1))
            const percentage = digit * 10
            const seekTime = Math.max(0, Math.min(video.duration, (video.duration * percentage) / 100))
            seekTo(seekTime)
            // showOverlayFeedback({ message: `${percentage}%` })
            return
        }

        // frame-by-frame seeking, assuming 24fps
        if (e.code === "Comma") {
            e.preventDefault()
            seek(-1 / 24)
            showOverlayFeedback({ message: "Previous Frame" })
            return
        }

        if (e.code === "Period") {
            e.preventDefault()
            seek(1 / 24)
            showOverlayFeedback({ message: "Next Frame" })
            return
        }

        // Helper function to check if seeking is rate limited
        const canSeek = () => {
            const now = Date.now()
            if (now - lastSeekTime.current < SEEK_THROTTLE_MS) {
                return false
            }
            lastSeekTime.current = now
            return true
        }

        // Check which shortcut was pressed
        if (e.code === keybindings.seekForward.key) {
            e.preventDefault()
            if (!canSeek()) return

            if (props.introEndTime && props.introStartTime && video.currentTime < props.introEndTime && video.currentTime >= props.introStartTime) {
                seekTo(props.introEndTime)
                showOverlayFeedback({ message: "Skipped Opening" })
                return
            }
            if (props.endingEndTime && props.endingStartTime && video.currentTime < props.endingEndTime && video.currentTime >= props.endingStartTime) {
                seekTo(props.endingEndTime)
                showOverlayFeedback({ message: "Skipped Ending" })
                return
            }
            seek(keybindings.seekForward.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekBackward.key) {
            e.preventDefault()
            if (!canSeek()) return
            seek(-keybindings.seekBackward.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekForwardFine.key) {
            e.preventDefault()
            // if (!canSeek()) return
            video.dispatchEvent(new Event("seeking"))
            seek(keybindings.seekForwardFine.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekBackwardFine.key) {
            e.preventDefault()
            // if (!canSeek()) return
            video.dispatchEvent(new Event("seeking"))
            seek(-keybindings.seekBackwardFine.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.nextChapter.key) {
            e.preventDefault()
            handleNextChapter()
        } else if (e.code === keybindings.previousChapter.key) {
            e.preventDefault()
            handlePreviousChapter()
        } else if (e.code === keybindings.volumeUp.key) {
            e.preventDefault()
            const newVolume = Math.min(1, volume + keybindings.volumeUp.value / 100)
            setVolume(newVolume)
        } else if (e.code === keybindings.volumeDown.key) {
            e.preventDefault()
            const newVolume = Math.max(0, volume - keybindings.volumeDown.value / 100)
            setVolume(newVolume)
        } else if (e.code === keybindings.mute.key) {
            e.preventDefault()
            setMuted(!muted)
        } else if (e.code === keybindings.cycleSubtitles.key) {
            e.preventDefault()
            handleCycleSubtitles()
        } else if (e.code === keybindings.cycleAudio.key) {
            e.preventDefault()
            handleCycleAudio()
        } else if (e.code === keybindings.nextEpisode.key) {
            e.preventDefault()
            handleNextEpisode()
        } else if (e.code === keybindings.previousEpisode.key) {
            e.preventDefault()
            handlePreviousEpisode()
        } else if (e.code === keybindings.fullscreen.key) {
            e.preventDefault()
            handleToggleFullscreen()
        } else if (e.code === keybindings.pictureInPicture.key) {
            e.preventDefault()
            handleTogglePictureInPicture()
        } else if (e.code === keybindings.increaseSpeed.key) {
            e.preventDefault()
            const newRate = Math.min(8, video.playbackRate + keybindings.increaseSpeed.value)
            video.playbackRate = newRate
            showOverlayFeedback({ message: `Speed: ${newRate.toFixed(2)}x` })
        } else if (e.code === keybindings.decreaseSpeed.key) {
            e.preventDefault()
            const newRate = Math.max(0.20, video.playbackRate - keybindings.decreaseSpeed.value)
            video.playbackRate = newRate
            showOverlayFeedback({ message: `Speed: ${newRate.toFixed(2)}x` })
        } else if (e.code === keybindings.takeScreenshot.key) {
            e.preventDefault()
            takeScreenshot()
        }
    }, [keybindings, volume, muted, seek, active, fullscreen, pip, showOverlayFeedback, introEndTime, introStartTime, isKeybindingsModalOpen])

    // Keyboard shortcut handlers
    const handleNextChapter = useCallback(() => {
        if (!videoRef.current || !chapterCues) return

        const currentTime = videoRef.current.currentTime

        // Sort chapters by start time to ensure proper order
        const sortedChapters = [...chapterCues].sort((a, b) => a.startTime - b.startTime)

        // Find the next chapter (with a small buffer to avoid edge cases)
        const nextChapter = sortedChapters.find(chapter => chapter.startTime > currentTime + 1)
        if (nextChapter) {
            seekTo(nextChapter.startTime)
            // Try to get chapter name from video track cues
            const chapterName = nextChapter.text
            showOverlayFeedback({ message: chapterName ? `Chapter: ${chapterName}` : `Chapter ${sortedChapters.indexOf(nextChapter) + 1}` })
        } else {
            // If no next chapter, go to the end
            const lastChapter = sortedChapters[sortedChapters.length - 1]
            if (lastChapter && lastChapter.endTime) {
                seekTo(lastChapter.endTime)
                showOverlayFeedback({ message: "End of chapters" })
            }
        }
    }, [chapterCues, seekTo, showOverlayFeedback])

    const handlePreviousChapter = useCallback(() => {
        if (!videoRef.current || !chapterCues) return

        const currentTime = videoRef.current.currentTime

        // Sort chapters by start time to ensure proper order
        const sortedChapters = [...chapterCues].sort((a, b) => a.startTime - b.startTime)

        // Find the current chapter first
        const currentChapterIndex = sortedChapters.findIndex((chapter, index) => {
            const nextChapter = sortedChapters[index + 1]
            return chapter.startTime <= currentTime && (!nextChapter || currentTime < nextChapter.startTime)
        })

        if (currentChapterIndex > 0) {
            // Go to previous chapter
            const previousChapter = sortedChapters[currentChapterIndex - 1]
            seekTo(previousChapter.startTime)
            const chapterName = previousChapter.text
            showOverlayFeedback({ message: chapterName ? `Chapter: ${chapterName}` : `Chapter ${currentChapterIndex}` })
        } else if (currentChapterIndex === 0) {
            // Already in first chapter, go to the beginning
            seekTo(0)
            const firstChapter = sortedChapters[0]
            const chapterName = firstChapter.text
            showOverlayFeedback({ message: chapterName ? `Chapter: ${chapterName}` : "Chapter 1" })
        } else {
            // If we can't determine current chapter, just go to the beginning
            seekTo(0)
            showOverlayFeedback({ message: "Beginning" })
        }
    }, [chapterCues, seekTo, showOverlayFeedback])


    const handleCycleSubtitles = useCallback(() => {
        if (!videoRef.current) return
        // TODO: make it work when both types are combined
        let found = false
        if (subtitleManager) {
            // Cycle to next track or disable if we're at the end
            const nextTrackNumber = subtitleManager.getNextTrackNumber(subtitleManager.getSelectedTrackNumberOrNull())

            // Enable next track if available
            if (nextTrackNumber > -1) {
                subtitleManager?.selectTrack(nextTrackNumber)
                const trackName = subtitleManager.getTrack(nextTrackNumber)?.label || `Track ${nextTrackNumber}`
                showOverlayFeedback({ message: `Subtitles: ${trackName}` })
                found = true
            }
        }
        if (mediaCaptionsManager) {
            const currentTrackIdx = mediaCaptionsManager.getSelectedTrackIndexOrNull() ?? -1
            const nextTrackIdx = currentTrackIdx + 1
            const nextTrack = mediaCaptionsManager.getTrack(nextTrackIdx)

            // Enable next track if available
            if (nextTrack) {
                mediaCaptionsManager?.selectTrack(nextTrackIdx)
                const trackName = mediaCaptionsManager.getTrack(nextTrackIdx)?.label || `Track ${nextTrackIdx}`
                showOverlayFeedback({ message: `Subtitles: ${trackName}` })
                found = true
            }
        }

        if (!found) {
            showOverlayFeedback({ message: "Subtitles: Off" })
            subtitleManager?.setNoTrack()
            mediaCaptionsManager?.setNoTrack()
        }
    }, [subtitleManager, mediaCaptionsManager])

    const handleCycleAudio = useCallback(() => {
        if (!videoRef.current || !audioManager) return

        // HLS stream
        if (audioManager.isHlsStream()) {
            const currentTrackNumber = audioManager.getSelectedTrackNumberOrNull()
            if (currentTrackNumber === null) {
                showOverlayFeedback({ message: "No additional audio tracks" })
                return
            }
            const audioTracks = audioManager.getHlsAudioTracks()

            const nextTrackNumber = (currentTrackNumber + 1) % (audioTracks.length)

            const nextTrack = audioTracks.find(n => n.id === nextTrackNumber)
            if (nextTrack) {
                const trackName = nextTrack.name || nextTrack.language || `Track ${nextTrack.id + 1}`
                showOverlayFeedback({ message: `Audio: ${trackName}` })
                audioManager.selectTrack(nextTrackNumber)
            }

            return
        }

        const audioTracks = videoRef.current.audioTracks
        if (!audioTracks || audioTracks.length <= 1) {
            showOverlayFeedback({ message: "No additional audio tracks" })
            return
        }

        // Find currently enabled track
        let currentTrackIndex = -1
        for (let i = 0; i < audioTracks.length; i++) {
            if (audioTracks[i].enabled) {
                currentTrackIndex = i
                break
            }
        }

        // Cycle to next track
        const nextIndex = (currentTrackIndex + 1) % audioTracks.length

        // Disable all tracks first
        for (let i = 0; i < audioTracks.length; i++) {
            audioTracks[i].enabled = false
        }

        // Enable next track
        audioTracks[nextIndex].enabled = true
        audioManager?.selectTrack(nextIndex)

        const trackName = audioTracks[nextIndex].label || audioTracks[nextIndex].language || `Track ${nextIndex + 1}`
        showOverlayFeedback({ message: `Audio: ${trackName}` })
    }, [audioManager])

    const log = logger("VideoCoreKeybindings")

    const handleNextEpisode = useCallback(() => {
        // Placeholder for next episode functionality
        log.info("Next episode shortcut pressed - not implemented yet")
    }, [])

    const handlePreviousEpisode = useCallback(() => {
        // Placeholder for previous episode functionality
        log.info("Previous episode shortcut pressed - not implemented yet")
    }, [])

    const handleToggleFullscreen = useCallback(() => {
        fullscreenManager?.toggleFullscreen()

        React.startTransition(() => {
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        })
    }, [fullscreenManager])

    const handleTogglePictureInPicture = useCallback(() => {
        pipManager?.enterPip()

        React.startTransition(() => {
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        })
    }, [pip, pipManager])

    // Add keyboard event listeners
    useEffect(() => {
        if (!active) return

        document.addEventListener("keydown", handleKeyboardShortcuts)

        return () => {
            document.removeEventListener("keydown", handleKeyboardShortcuts)
        }
    }, [handleKeyboardShortcuts, active])

    // Handle fullscreen state changes to ensure video gets focused
    useEffect(() => {
        if (!active) return

        const handleFullscreenChange = () => {
            // Small delay to ensure fullscreen transition is complete
            setTimeout(() => {
                if (document.fullscreenElement && videoRef.current) {
                    videoRef.current.focus()
                }
            }, 100)
        }

        document.addEventListener("fullscreenchange", handleFullscreenChange)

        return () => {
            document.removeEventListener("fullscreenchange", handleFullscreenChange)
        }
    }, [active])

    return null
}
