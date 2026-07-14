import type { AnyRouter, ParsedLocation } from "@tanstack/react-router"

type ScrollPosition = {
    left: number
    top: number
}

const positions = new Map<string, ScrollPosition>()

export function setupDenshiScrollRestoration(router: AnyRouter) {
    router.subscribe("onBeforeLoad", ({ fromLocation }) => {
        if (!fromLocation || !isRestorablePath(fromLocation.pathname)) return

        positions.set(getScrollKey(fromLocation), {
            left: window.scrollX,
            top: window.scrollY,
        })
    })

    router.subscribe("onRendered", ({ toLocation }) => {
        if (!isRestorablePath(toLocation.pathname)) return

        const position = positions.get(getScrollKey(toLocation))
        if (!position) return

        window.scrollTo({
            left: position.left,
            top: position.top,
            behavior: "auto",
        })
    })
}

function getScrollKey(location: ParsedLocation) {
    return location.state.__TSR_key || location.href
}

function isRestorablePath(pathname: string) {
    return pathname === "/search" || pathname === "/lists"
}
