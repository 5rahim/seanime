import {atomWithStorage} from "jotai/utils"
import {Settings} from "@/lib/server/types"
import {useAtom} from "jotai/react"

export const settingsAtom = atomWithStorage<Settings | undefined>("sea-settings", undefined, undefined,
    {getOnInit: true})

export function useStoredSettings() {

    const [settings, setSettings] = useAtom(settingsAtom)

    return {
        settings,
        setSettings,
    }

}