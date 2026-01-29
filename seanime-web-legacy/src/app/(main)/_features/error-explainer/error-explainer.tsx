import { cn } from "@/components/ui/core/styling"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

const __errorExplainer_overlayOpenAtom = atom(false)
const __errorExplainer_errorAtom = atom<string | null>(null)

export function ErrorExplainer() {
    const [open, setOpen] = useAtom(__errorExplainer_overlayOpenAtom)

    return (
        <>
            {open && <div
                className={cn(
                    "error-explainer-ui",
                    "fixed z-[100] bottom-8 w-fit left-20 h-fit flex",
                    "transition-all duration-300 select-none",
                    // !isRecording && "hover:translate-y-[-2px]",
                    // isRecording && "justify-end",
                )}
            >
                <div
                    className={cn(
                        "p-4 bg-gray-900 border text-white rounded-xl",
                        "transition-colors duration-300",
                        // isRecording && "p-0 border-transparent bg-transparent",
                    )}
                >
                </div>
            </div>}
        </>
    )
}

export function useErrorExplainer() {
    const [error, setError] = useAtom(__errorExplainer_errorAtom)

    const explaination = React.useMemo(() => {
        if (!error) return null

        if (error.includes("could not open and play video")) {

        }

        return ""
    }, [error])

    return { error, setError }
}
