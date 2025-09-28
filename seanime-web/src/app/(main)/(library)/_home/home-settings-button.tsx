import { __home_settingsModalOpen } from "@/app/(main)/(library)/_home/home-settings-modal"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { LuSettings2 } from "react-icons/lu"

// by default, the button will be highlighted until the user clicks it for the first time
// this is not applied to the empty home page
export const __home_settingsToolbarButtonDiscovered = atomWithStorage("sea-v3-home-settings-toolbar-button-discovered", false)

type HomeSettingsButtonProps = {
    type: "toolbar" | "empty-home"
}

export function HomeSettingsButton(props: HomeSettingsButtonProps) {
    const { type } = props
    const [discoveredOnce, setDiscoveredOnce] = useAtom(__home_settingsToolbarButtonDiscovered)
    const [isModalOpen, setIsModalOpen] = useAtom(__home_settingsModalOpen)

    // if(type === "toolbar" && !discoveredOnce) return (
    //     <>
    //         <Tooltip
    //             open
    //             className="bg-fuchsia-500/100 flex gap-2 items-center font-semibold"
    //             sideOffset={5}
    //             trigger={<IconButton
    //                 data-library-toolbar-switch-view-button
    //                 intent="white-glass"
    //                 icon={<BsHouseGear className="text-2xl" />}
    //                 className={cn(
    //                     "animate-bounce",
    //                 )}
    //                 onClick={() => {
    //                     setDiscoveredOnce(true)
    //                     setIsModalOpen(true)
    //                 }}
    //             />}
    //         >
    //             <LuSparkles /> New
    //         </Tooltip>
    //     </>
    // )

    if (type === "toolbar") return (
        <>
            <Tooltip
                trigger={<IconButton
                    data-library-toolbar-switch-view-button
                    intent="white-subtle"
                    icon={<LuSettings2 className="text-2xl" />}
                    onClick={() => {
                        setIsModalOpen(true)
                    }}
                />}
            >
                Home Settings
            </Tooltip>
        </>
    )


}
