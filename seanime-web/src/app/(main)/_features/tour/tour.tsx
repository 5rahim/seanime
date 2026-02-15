import { useLocation, useNavigate } from "@tanstack/react-router"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { ReactNode, useCallback, useEffect, useRef } from "react"

export type TourStepPlacement = "top" | "bottom" | "left" | "right"

// When `target` is set, the step highlights an element with a popover next to it
// When `target` is omitted, the step shows a centered modal card
export type TourStep = {
    id: string
    // CSS selector of the element to highlight, omit for modal-style step
    target?: string
    content: ReactNode
    title?: ReactNode
    // image/hero rendered above the title
    image?: ReactNode
    // navigate to this route before showing this step
    route?: string
    // e.g. switch a tab, open a drawer
    prepare?: () => Promise<void> | void
    // popover placement relative to the target, default 'bottom'
    placement?: TourStepPlacement
    // px padding around the spotlight cutout, default 8
    spotlightPadding?: number
    // clicking the highlighted element advances to the next step
    advanceOnTargetClick?: boolean
    // disable back button for one-way flows
    disableBack?: boolean
    // custom label for the next button, e.g. "Got it!"
    nextLabel?: string
    prevLabel?: string
    // hide nav buttons entirely
    hideControls?: boolean
    // custom width for the popover/modal card
    popoverWidth?: number | string
    // sync check before showing the step, return false to trigger fallback
    // e.g. () => someSettingIsEnabled
    condition?: () => boolean
    // what to do when condition returns false, default: "modal"
    conditionFailBehavior?: "skip" | "modal"
    // how long to wait for the target element before fallback (ms), default 3000
    waitForTargetMs?: number
    // what to do when the target element isn't found after timeout
    // "skip" = skip to next step, "modal" = show as modal instead
    // default: "modal"
    targetNotFoundBehavior?: "skip" | "modal"
    // ignore outside clicks, default: false
    ignoreOutsideClick?: boolean
}

// 'idle': waiting for user | 'navigating': changing routes | 'seeking': waiting for DOM element | 'preparing': running prepare()
export type TourStatus = "idle" | "navigating" | "seeking" | "preparing"

export type TourState = {
    active: boolean
    tourId: string | null
    steps: TourStep[]
    currentIndex: number
    status: TourStatus
    onEnd?: () => void
}

const INITIAL_STATE: TourState = {
    active: false,
    tourId: null,
    steps: [],
    currentIndex: 0,
    status: "idle",
    onEnd: undefined,
}

export const tourAtom = atom<TourState>(INITIAL_STATE)

export const currentTourStepAtom = atom<TourStep | null>((get) => {
    const { steps, currentIndex, active } = get(tourAtom)
    if (!active) return null
    return steps[currentIndex] ?? null
})

export function useTour() {
    const [state, setTour] = useAtom(tourAtom)
    const navigate = useNavigate()
    const location = useLocation()

    // mutable ref so closures always see current state
    const stateRef = useRef(state)
    stateRef.current = state

    // wait for a DOM element to exist, resolves null on timeout
    const waitForElement = useCallback(
        (selector: string, timeout = 3000): Promise<Element | null> => {
            return new Promise((resolve) => {
                const existing = document.querySelector(selector)
                if (existing) return resolve(existing)

                const observer = new MutationObserver(() => {
                    const el = document.querySelector(selector)
                    if (el) {
                        resolve(el)
                        observer.disconnect()
                    }
                })

                observer.observe(document.body, {
                    childList: true,
                    subtree: true,
                    attributes: true,
                })

                setTimeout(() => {
                    observer.disconnect()
                    const lastTry = document.querySelector(selector)
                    resolve(lastTry) // null if not found
                }, timeout)
            })
        },
        [],
    )

    const endTour = useCallback(() => {
        const { onEnd } = stateRef.current
        setTour(INITIAL_STATE)
        onEnd?.()
    }, [setTour])

    const goToStep = useCallback(
        async (index: number) => {
            const { steps } = stateRef.current
            const step = steps[index]
            if (!step) return

            try {
                // Condition check (runs before routing/preparing)
                if (step.condition && !step.condition()) {
                    const fallback = step.conditionFailBehavior ?? "modal"
                    if (fallback === "skip") {
                        if (index < steps.length - 1) {
                            goToStep(index + 1)
                        } else {
                            endTour()
                        }
                        return
                    }
                    // fallback === "modal" -> clear target, show as modal
                    step.target = undefined
                }

                // Route
                if (step.route && location.pathname !== step.route) {
                    setTour((prev) => ({ ...prev, status: "navigating" }))
                    await navigate({ to: step.route })
                    await new Promise((r) => setTimeout(r, 150))
                }

                // Prepare
                if (step.prepare) {
                    setTour((prev) => ({ ...prev, status: "preparing" }))
                    await step.prepare()
                    await new Promise((r) => setTimeout(r, 100))
                }

                // Wait for target element
                if (step.target) {
                    setTour((prev) => ({ ...prev, status: "seeking" }))
                    const el = await waitForElement(step.target, step.waitForTargetMs ?? 3000)

                    if (!el) {
                        const fallback = step.targetNotFoundBehavior ?? "modal"
                        if (fallback === "skip") {
                            // skip this step
                            if (index < steps.length - 1) {
                                goToStep(index + 1)
                            } else {
                                endTour()
                            }
                            return
                        }
                        // fallback === "modal" -> clear target, show as modal
                        step.target = undefined
                    }
                }

                // Done
                setTour((prev) => ({
                    ...prev,
                    currentIndex: index,
                    status: "idle",
                }))
            }
            catch (err) {
                console.error("[Tour] Step transition failed:", err)
                // skip broken step instead of killing the tour
                if (index < steps.length - 1) {
                    goToStep(index + 1)
                } else {
                    endTour()
                }
            }
        },
        [location.pathname, navigate, setTour, waitForElement, endTour],
    )

    const start = useCallback(
        (steps: TourStep[], tourId?: string, onEnd?: () => void) => {
            if (steps.length === 0) return

            setTour({
                active: true,
                tourId: tourId ?? null,
                steps,
                currentIndex: 0,
                status: "idle",
                onEnd,
            })

            // kick off first step after state commits
            setTimeout(() => goToStep(0), 0)
        },
        [setTour, goToStep],
    )

    const next = useCallback(() => {
        const { currentIndex, steps } = stateRef.current
        if (currentIndex < steps.length - 1) {
            goToStep(currentIndex + 1)
        } else {
            endTour()
        }
    }, [goToStep, endTour])

    const prev = useCallback(() => {
        const { currentIndex, steps } = stateRef.current
        const step = steps[currentIndex]
        if (step?.disableBack) return
        if (currentIndex > 0) {
            goToStep(currentIndex - 1)
        }
    }, [goToStep])

    const stop = useCallback(() => {
        endTour()
    }, [endTour])

    const goTo = useCallback(
        (stepId: string) => {
            const idx = stateRef.current.steps.findIndex((s) => s.id === stepId)
            if (idx !== -1) goToStep(idx)
        },
        [goToStep],
    )

    // Keyboard nav
    useEffect(() => {
        if (!state.active) return

        const handler = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                e.preventDefault()
                stop()
            }
            if (e.key === "ArrowRight" || e.key === "Enter") {
                e.preventDefault()
                next()
            }
            if (e.key === "ArrowLeft") {
                e.preventDefault()
                prev()
            }
        }
        window.addEventListener("keydown", handler)
        return () => window.removeEventListener("keydown", handler)
    }, [state.active, next, prev, stop])

    return {
        active: state.active,
        tourId: state.tourId,
        currentStep: state.steps[state.currentIndex] ?? null,
        currentIndex: state.currentIndex,
        totalSteps: state.steps.length,
        status: state.status,
        start,
        next,
        prev,
        stop,
        goTo,
    }
}

export const tourHelpers = {
    // click a Radix trigger to open it
    openPopover(triggerSelector: string, delayMs = 200): Promise<void> {
        return new Promise((resolve) => {
            const trigger = document.querySelector(triggerSelector) as HTMLElement | null
            trigger?.click()
            setTimeout(resolve, delayMs)
        })
    },

    closePopover(triggerSelector: string, delayMs = 150): Promise<void> {
        return new Promise((resolve) => {
            const trigger = document.querySelector(triggerSelector) as HTMLElement | null
            // radix popovers close on Escape
            trigger?.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }))
            setTimeout(resolve, delayMs)
        })
    },
    // wait for a selector to appear in the DOM
    waitForSelector(selector: string, timeoutMs = 3000): Promise<Element | null> {
        return new Promise((resolve) => {
            const existing = document.querySelector(selector)
            if (existing) return resolve(existing)

            const observer = new MutationObserver(() => {
                const el = document.querySelector(selector)
                if (el) {
                    observer.disconnect()
                    resolve(el)
                }
            })

            observer.observe(document.body, { childList: true, subtree: true })

            setTimeout(() => {
                observer.disconnect()
                resolve(document.querySelector(selector))
            }, timeoutMs)
        })
    },

    // click an element by selector
    click(selector: string, delayMs = 100): Promise<void> {
        return new Promise((resolve) => {
            const el = document.querySelector(selector) as HTMLElement | null
            el?.click()
            setTimeout(resolve, delayMs)
        })
    },
}
