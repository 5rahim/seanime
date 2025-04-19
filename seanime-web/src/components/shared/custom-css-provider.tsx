import { useAtom } from "jotai"
import { atomWithStorage } from "jotai/utils"
import React, { useEffect, useState } from "react"
import { useWindowSize } from "react-use"

const customCSSAtom = atomWithStorage("sea-custom-css", {
    customCSS: "",
    mobileCustomCSS: "",
}, undefined, { getOnInit: true })

export function CustomCSSProvider({ children }: { children: React.ReactNode }) {
    const [customCSS, setCustomCSS] = useAtom(customCSSAtom)
    const [mounted, setMounted] = useState(false)
    const { width } = useWindowSize()

    const isMobile = width < 1024

    const usedCSS = React.useMemo(() => isMobile ? customCSS.mobileCustomCSS : customCSS.customCSS, [isMobile, customCSS])

    useEffect(() => {
        setMounted(true)
    }, [])

    return (
        <>
            {children}
            {mounted && usedCSS && (
                <style id="sea-custom-css" dangerouslySetInnerHTML={{ __html: usedCSS }} />
            )}
        </>
    )
}

export function useCustomCSS() {
    const [customCSS, setCustomCSS] = useAtom(customCSSAtom)

    return { customCSS, setCustomCSS }
}
