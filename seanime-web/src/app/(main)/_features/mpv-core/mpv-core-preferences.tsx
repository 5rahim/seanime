import { useSaveMediaPlayerSettings } from "@/api/hooks/settings.hooks"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { upath } from "@/lib/helpers/upath"
import { atom, useAtom, useAtomValue } from "jotai"
import React from "react"
import { useServerStatus } from "../../_hooks/use-server-status"
import {
    mc_defaultKeybindings,
    mc_initialSettings,
    mc_keybindingsAtom,
    mc_settings,
    mpvCore_stateAtom,
    type MpvCoreKeybindings,
} from "./mpv-core.atoms"

export const mpvCorePreferencesModalAtom = atom(false)

const tabsRootClass = cn("w-full contents space-y-4")
const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)
const tabsListClass = cn("w-full flex flex-row lg:flex-row flex-wrap h-fit !mt-4")
const tabContentClass = cn("space-y-4 animate-in fade-in-0 duration-300")

const KeybindingValueInput = ({
    actionKey,
    value,
    onValueChange,
}: {
    actionKey: keyof MpvCoreKeybindings
    value: number
    onValueChange: (value: number) => void
}) => (
    <NumberInput
        value={value}
        onValueChange={onValueChange}
        size="sm"
        fieldClass="w-16"
        hideControls
        min={0}
        step={actionKey.includes("Speed") ? 0.25 : 1}
    />
)

const KeybindingRow = ({
    action,
    actionKey,
    hasValue = false,
    valueLabel = "",
    editedKeybindings,
    setEditedKeybindings,
    recordingKey,
    handleKeyRecord,
    formatKeyDisplay,
}: {
    action: string
    actionKey: keyof MpvCoreKeybindings
    hasValue?: boolean
    valueLabel?: string
    editedKeybindings: MpvCoreKeybindings
    setEditedKeybindings: React.Dispatch<React.SetStateAction<MpvCoreKeybindings>>
    recordingKey: string | null
    handleKeyRecord: (actionKey: keyof MpvCoreKeybindings) => void
    formatKeyDisplay: (keyCode: string) => string
}) => (
    <div className="flex items-center justify-between py-2 border rounded-lg px-3 bg-[--paper]">
        <div className="flex-1">
            <div className="font-medium text-sm">{action}</div>
            {hasValue && (
                <div className="flex items-center gap-2 mt-1">
                    <span className="text-xs text-muted-foreground">{valueLabel}:</span>
                    <KeybindingValueInput
                        actionKey={actionKey}
                        value={"value" in editedKeybindings[actionKey]
                            ? Number((editedKeybindings[actionKey] as { value: number }).value)
                            : 0}
                        onValueChange={value => {
                            setEditedKeybindings(previous => ({
                                ...previous,
                                [actionKey]: { ...previous[actionKey], value: value || 0 },
                            }))
                        }}
                    />
                </div>
            )}
        </div>
        <div className="flex items-center gap-2">
            <Button
                intent={recordingKey === actionKey ? "white-subtle" : "gray-subtle"}
                size="sm"
                onClick={() => handleKeyRecord(actionKey)}
                className={cn(
                    "h-8 px-3 text-lg font-mono",
                    recordingKey === actionKey && "!text-xs text-white",
                )}
            >
                {recordingKey === actionKey ? "Press key..." : formatKeyDisplay(editedKeybindings[actionKey].key)}
            </Button>
        </div>
    </div>
)

export function MpvCorePreferencesModal(props: {
    fullscreen: boolean
    containerElement: HTMLElement | null
    onTerminate?: (reason: string) => void
}) {
    const { fullscreen, containerElement, onTerminate } = props
    const [open, setOpen] = useAtom(mpvCorePreferencesModalAtom)
    const [keybindings, setKeybindings] = useAtom(mc_keybindingsAtom)
    const [settings, setSettings] = useAtom(mc_settings)
    const mpvCoreState = useAtomValue(mpvCore_stateAtom)
    const [editedKeybindings, setEditedKeybindings] = React.useState<MpvCoreKeybindings>(keybindings)
    const [editedSubLanguage, setEditedSubLanguage] = React.useState(settings.preferredSubtitleLanguage)
    const [editedAudioLanguage, setEditedAudioLanguage] = React.useState(settings.preferredAudioLanguage)
    const [editedSubsBlacklist, setEditedSubsBlacklist] = React.useState(settings.preferredSubtitleBlacklist)
    const [editedSubtitleDelay, setEditedSubtitleDelay] = React.useState(settings.subtitleDelay)
    const [recordingKey, setRecordingKey] = React.useState<string | null>(null)
    const [tab, setTab] = React.useState("keybinds")

    const serverStatus = useServerStatus()
    const { mutate: saveMediaPlayerSettings } = useSaveMediaPlayerSettings()
    const mediaPlayerSettings = serverStatus?.settings?.mediaPlayer
    const [editedScreenshotDir, setEditedScreenshotDir] = React.useState(mediaPlayerSettings?.screenshotDir ?? "")

    const isAbsolute = React.useMemo(() => {
        if (!editedScreenshotDir) return true
        return upath.isAbsolute(editedScreenshotDir)
    }, [editedScreenshotDir])

    React.useEffect(() => {
        if (!open) return
        setEditedKeybindings(keybindings)
        setEditedSubLanguage(settings.preferredSubtitleLanguage)
        setEditedAudioLanguage(settings.preferredAudioLanguage)
        setEditedSubsBlacklist(settings.preferredSubtitleBlacklist)
        setEditedSubtitleDelay(settings.subtitleDelay)
        setEditedScreenshotDir(mediaPlayerSettings?.screenshotDir ?? "")
    }, [open, keybindings, settings, mediaPlayerSettings])

    const handleKeyRecord = (actionKey: keyof MpvCoreKeybindings) => {
        setRecordingKey(actionKey)
        const handleKeyDown = (event: KeyboardEvent) => {
            event.preventDefault()
            event.stopPropagation()
            event.stopImmediatePropagation()
            setEditedKeybindings(previous => ({
                ...previous,
                [actionKey]: { ...previous[actionKey], key: event.code },
            }))
            setRecordingKey(null)
            document.removeEventListener("keydown", handleKeyDown, true)
        }
        document.addEventListener("keydown", handleKeyDown, true)
    }

    const saveSettings = () => {
        setKeybindings(editedKeybindings)
        setSettings({
            ...settings,
            preferredSubtitleLanguage: editedSubLanguage,
            preferredAudioLanguage: editedAudioLanguage,
            preferredSubtitleBlacklist: editedSubsBlacklist,
            subtitleDelay: editedSubtitleDelay,
        })

        const currentMediaPlayer = serverStatus?.settings?.mediaPlayer
        if (currentMediaPlayer) {
            saveMediaPlayerSettings({
                mediaPlayer: {
                    ...currentMediaPlayer,
                    screenshotDir: editedScreenshotDir,
                },
            })
        }
        setOpen(false)
    }

    const handleSave = () => {
        saveSettings()
    }

    const handleReset = () => {
        setEditedKeybindings(mc_defaultKeybindings)
        setEditedSubLanguage(mc_initialSettings.preferredSubtitleLanguage)
        setEditedAudioLanguage(mc_initialSettings.preferredAudioLanguage)
        setEditedSubsBlacklist(mc_initialSettings.preferredSubtitleBlacklist)
        setEditedSubtitleDelay(mc_initialSettings.subtitleDelay)
        setEditedScreenshotDir(mediaPlayerSettings?.screenshotDir ?? "")
    }

    const formatKeyDisplay = (keyCode: string) => {
        const keyMap: Record<string, string> = {
            KeyA: "A", KeyB: "B", KeyC: "C", KeyD: "D", KeyE: "E", KeyF: "F",
            KeyG: "G", KeyH: "H", KeyI: "I", KeyJ: "J", KeyK: "K", KeyL: "L",
            KeyM: "M", KeyN: "N", KeyO: "O", KeyP: "P", KeyQ: "Q", KeyR: "R",
            KeyS: "S", KeyT: "T", KeyU: "U", KeyV: "V", KeyW: "W", KeyX: "X",
            KeyY: "Y", KeyZ: "Z",
            ArrowUp: "↑", ArrowDown: "↓", ArrowLeft: "←", ArrowRight: "→",
            BracketLeft: "[", BracketRight: "]",
            Space: "⎵",
        }
        return keyMap[keyCode] || keyCode
    }

    const rowProps = {
        editedKeybindings,
        setEditedKeybindings,
        recordingKey,
        handleKeyRecord,
        formatKeyDisplay,
    }

    return (
        <>
            <Modal
                title="Preferences"
                open={open}
                onOpenChange={setOpen}
                contentClass="max-w-5xl focus:outline-none focus-visible:outline-none outline-none bg-[--background] backdrop-blur-sm z-[101]"
                overlayClass="z-[150] bg-black/50"
                portalContainer={props.fullscreen ? props.containerElement || undefined : undefined}
            >
                <Tabs
                    value={tab}
                    onValueChange={setTab}
                    className={tabsRootClass}
                    triggerClass={tabsTriggerClass}
                    listClass={tabsListClass}
                    variant="pill"
                >
                    <TabsList className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-xl">
                        <TabsTrigger value="keybinds">Keyboard Shortcuts</TabsTrigger>
                        <TabsTrigger value="subtitles">Subtitles & Audio</TabsTrigger>
                        <TabsTrigger value="general">General</TabsTrigger>
                    </TabsList>

                    <TabsContent value="general" className={tabContentClass}>
                        <div className="space-y-4">
                            <DirectorySelector
                                value={editedScreenshotDir}
                                onSelect={setEditedScreenshotDir}
                                label="Screenshot Directory"
                                help="Configure the directory where screenshots will be saved"
                                error={!isAbsolute ? "Must be an absolute path" : ""}
                            />

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
                                        disabled={!isAbsolute}
                                    >
                                        Save
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </TabsContent>

                    <TabsContent value="keybinds" className={tabContentClass}>
                        <div className="space-y-3 hidden lg:block">
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                                <div>
                                    <div className="space-y-3">
                                        <KeybindingRow action="Seek Forward (Fine)" actionKey="seekForwardFine" hasValue valueLabel="Seconds" {...rowProps} />
                                        <KeybindingRow action="Seek Backward (Fine)" actionKey="seekBackwardFine" hasValue valueLabel="Seconds" {...rowProps} />
                                        <KeybindingRow action="Seek Forward" actionKey="seekForward" hasValue valueLabel="Seconds" {...rowProps} />
                                        <KeybindingRow action="Seek Backward" actionKey="seekBackward" hasValue valueLabel="Seconds" {...rowProps} />
                                        <KeybindingRow action="Increase Speed" actionKey="increaseSpeed" hasValue valueLabel="increment" {...rowProps} />
                                        <KeybindingRow action="Decrease Speed" actionKey="decreaseSpeed" hasValue valueLabel="increment" {...rowProps} />
                                    </div>
                                </div>
                                <div>
                                    <div className="space-y-3">
                                        <KeybindingRow action="Next Chapter" actionKey="nextChapter" {...rowProps} />
                                        <KeybindingRow action="Previous Chapter" actionKey="previousChapter" {...rowProps} />
                                        <KeybindingRow action="Next Episode" actionKey="nextEpisode" {...rowProps} />
                                        <KeybindingRow action="Previous Episode" actionKey="previousEpisode" {...rowProps} />
                                        <KeybindingRow action="Cycle Subtitles" actionKey="cycleSubtitles" {...rowProps} />
                                        <KeybindingRow action="Fullscreen" actionKey="fullscreen" {...rowProps} />
                                        <KeybindingRow action="Picture in Picture" actionKey="pictureInPicture" {...rowProps} />
                                        <KeybindingRow action="Take Screenshot" actionKey="takeScreenshot" {...rowProps} />
                                        <KeybindingRow action="Display Characters" actionKey="openInSight" {...rowProps} />
                                    </div>
                                </div>
                                <div>
                                    <div className="space-y-3">
                                        <KeybindingRow action="Volume Up" actionKey="volumeUp" hasValue valueLabel="Percent" {...rowProps} />
                                        <KeybindingRow action="Volume Down" actionKey="volumeDown" hasValue valueLabel="Percent" {...rowProps} />
                                        <KeybindingRow action="Mute" actionKey="mute" {...rowProps} />
                                        <KeybindingRow action="Cycle Audio" actionKey="cycleAudio" {...rowProps} />
                                        <KeybindingRow action="Stats for Nerds" actionKey="statsForNerds" {...rowProps} />
                                    </div>
                                </div>
                            </div>
                            <div className="flex items-center justify-between pt-6">
                                <Button intent="gray-outline" onClick={handleReset}>Reset all</Button>
                                <div className="flex gap-2">
                                    <Button intent="gray-outline" onClick={() => setOpen(false)}>Cancel</Button>
                                    <Button intent="primary" onClick={handleSave}>Save</Button>
                                </div>
                            </div>
                        </div>
                    </TabsContent>

                    <TabsContent value="subtitles" className={tabContentClass}>
                        <div className="space-y-3">
                            <h3 className="text-lg font-semibold text-white">Defaults</h3>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-muted-foreground">Preferred Subtitle Language</label>
                                    <TextInput
                                        value={editedSubLanguage}
                                        onValueChange={setEditedSubLanguage}
                                        placeholder="eng,jpn,spa"
                                        onKeyDown={event => event.stopPropagation()}
                                        onInput={event => event.stopPropagation()}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-muted-foreground">Preferred Audio Language</label>
                                    <TextInput
                                        value={editedAudioLanguage}
                                        onValueChange={setEditedAudioLanguage}
                                        placeholder="jpn,eng,kor"
                                        onKeyDown={event => event.stopPropagation()}
                                        onInput={event => event.stopPropagation()}
                                    />
                                </div>
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-muted-foreground">Ignored Subtitle Names</label>
                                <TextInput
                                    value={editedSubsBlacklist}
                                    onValueChange={setEditedSubsBlacklist}
                                    placeholder="e.g. signs & songs,signs/songs"
                                    onKeyDown={event => event.stopPropagation()}
                                    onInput={event => event.stopPropagation()}
                                    help="Subtitle tracks that will not be selected by default if they match the preferred languages. Separate multiple names with commas."
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-muted-foreground">Subtitle Delay</label>
                                <NumberInput
                                    value={editedSubtitleDelay}
                                    onValueChange={value => setEditedSubtitleDelay(value || 0)}
                                    min={-30}
                                    max={30}
                                    step={0.1}
                                />
                            </div>
                        </div>
                        <div className="flex items-center justify-between pt-6">
                            <Button intent="gray-outline" onClick={handleReset}>Reset all</Button>
                            <div className="flex gap-2">
                                <Button intent="gray-outline" onClick={() => setOpen(false)}>Cancel</Button>
                                <Button intent="primary" onClick={handleSave}>Save</Button>
                            </div>
                        </div>
                    </TabsContent> </Tabs>
            </Modal>
        </>
    )
}
