import { LibraryExplorer } from "@/app/(main)/_features/library-explorer/library-explorer"
import { libraryExplorer_drawerOpenAtom } from "@/app/(main)/_features/library-explorer/library-explorer.atoms"
import { cn } from "@/components/ui/core/styling"
import { Vaul, VaulContent } from "@/components/vaul"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useAtom } from "jotai/react"
import React from "react"

export function LibraryExplorerDrawer(props: {}) {

    const ts = useThemeSettings()
    const [open, setOpen] = useAtom(libraryExplorer_drawerOpenAtom)

    return (
        <Vaul
            open={open}
            onOpenChange={v => setOpen(v)}
        >

            <VaulContent
                className={cn(
                    "bg-gray-950 h-[90%] lg:h-[80%] bg-opacity-95 firefox:bg-opacity-100 lg:mx-[2rem]",
                )}
            >
                <LibraryExplorer />
                <div className="block lg:hidden">
                    <p className="text-center text-white text-lg font-semibold py-4">
                        Library explorer can only be rendered on larger screens.
                    </p>
                </div>
            </VaulContent>
        </Vaul>
    )

}
