import type { AnyRouter, ParsedLocation } from "@tanstack/react-router"
import { afterEach, describe, expect, it, vi } from "vitest"
import { setupDenshiScrollRestoration } from "./denshi-scroll-restoration"

type Handler = (event: {
    fromLocation?: ParsedLocation
    toLocation?: ParsedLocation
}) => void

function setup() {
    const handlers = new Map<string, Handler>()
    const router = {
        subscribe: (event: string, handler: Handler) => {
            handlers.set(event, handler)
            return () => handlers.delete(event)
        },
    } as unknown as AnyRouter

    setupDenshiScrollRestoration(router)

    return handlers
}

function getLocation(pathname: string, key: string): ParsedLocation {
    return {
        href: pathname,
        pathname,
        state: { __TSR_key: key },
    } as ParsedLocation
}

describe("setupDenshiScrollRestoration", () => {
    afterEach(() => {
        vi.unstubAllGlobals()
    })

    it.each(["/search", "/lists"])("restores the same %s history entry", pathname => {
        const scrollTo = vi.fn()
        vi.stubGlobal("window", { scrollX: 12, scrollY: 640, scrollTo })
        const handlers = setup()
        const location = getLocation(pathname, `${pathname}-entry`)

        handlers.get("onBeforeLoad")?.({ fromLocation: location })
        handlers.get("onRendered")?.({ toLocation: location })

        expect(scrollTo).toHaveBeenCalledWith({
            left: 12,
            top: 640,
            behavior: "auto",
        })
    })

    it("does not restore a new history entry for the same route", () => {
        const scrollTo = vi.fn()
        vi.stubGlobal("window", { scrollX: 0, scrollY: 480, scrollTo })
        const handlers = setup()

        handlers.get("onBeforeLoad")?.({ fromLocation: getLocation("/lists", "old-entry") })
        handlers.get("onRendered")?.({ toLocation: getLocation("/lists", "new-entry") })

        expect(scrollTo).not.toHaveBeenCalled()
    })

    it("ignores other routes", () => {
        const scrollTo = vi.fn()
        vi.stubGlobal("window", { scrollX: 0, scrollY: 320, scrollTo })
        const handlers = setup()
        const location = getLocation("/discover", "discover-entry")

        handlers.get("onBeforeLoad")?.({ fromLocation: location })
        handlers.get("onRendered")?.({ toLocation: location })

        expect(scrollTo).not.toHaveBeenCalled()
    })
})
