import { cn } from "@/components/ui/core/styling"
import { useAtomValue } from "jotai/react"
import { AnimatePresence, motion } from "motion/react"
import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react"
import { tourAtom, TourStep, TourStepPlacement, useTour } from "./tour"

type Rect = { top: number; left: number; width: number; height: number }

const EMPTY_RECT: Rect = { top: 0, left: 0, width: 0, height: 0 }

// returns a viewport-relative rect with padding applied
function getElementRect(el: Element, padding: number): Rect {
    const r = el.getBoundingClientRect()
    return {
        top: r.top - padding,
        left: r.left - padding,
        width: r.width + padding * 2,
        height: r.height + padding * 2,
    }
}

type PopoverPosition = { top: number; left: number }

function computePopoverPosition(
    targetRect: Rect,
    popoverWidth: number,
    popoverHeight: number,
    placement: TourStepPlacement,
    gap: number = 12,
): PopoverPosition {
    const vw = window.innerWidth
    const vh = window.innerHeight

    let top = 0
    let left = 0

    switch (placement) {
        case "bottom":
            top = targetRect.top + targetRect.height + gap
            left = targetRect.left + targetRect.width / 2 - popoverWidth / 2
            break
        case "top":
            top = targetRect.top - popoverHeight - gap
            left = targetRect.left + targetRect.width / 2 - popoverWidth / 2
            break
        case "left":
            top = targetRect.top + targetRect.height / 2 - popoverHeight / 2
            left = targetRect.left - popoverWidth - gap
            break
        case "right":
            top = targetRect.top + targetRect.height / 2 - popoverHeight / 2
            left = targetRect.left + targetRect.width + gap
            break
    }

    // clamp to viewport
    const margin = 16
    if (left < margin) left = margin
    if (left + popoverWidth > vw - margin) left = vw - margin - popoverWidth
    if (top < margin) top = margin
    if (top + popoverHeight > vh - margin) top = vh - margin - popoverHeight

    return { top, left }
}

// picks the placement with the most available space if preferred doesn't fit
function resolvePlacement(
    targetRect: Rect,
    popoverWidth: number,
    popoverHeight: number,
    preferred: TourStepPlacement,
    gap: number = 12,
): TourStepPlacement {
    const vw = window.innerWidth
    const vh = window.innerHeight

    const spaceTop = targetRect.top - gap
    const spaceBottom = vh - (targetRect.top + targetRect.height + gap)
    const spaceLeft = targetRect.left - gap
    const spaceRight = vw - (targetRect.left + targetRect.width + gap)

    const fits = (p: TourStepPlacement) => {
        switch (p) {
            case "bottom":
                return spaceBottom >= popoverHeight
            case "top":
                return spaceTop >= popoverHeight
            case "left":
                return spaceLeft >= popoverWidth
            case "right":
                return spaceRight >= popoverWidth
        }
    }

    if (fits(preferred)) return preferred

    // fallback: pick the side with most room
    const placements: TourStepPlacement[] = ["bottom", "top", "right", "left"]
    const spaces: Record<TourStepPlacement, number> = {
        bottom: spaceBottom,
        top: spaceTop,
        left: spaceLeft,
        right: spaceRight,
    }

    return placements.sort((a, b) => spaces[b] - spaces[a])[0]
}


// full-screen SVG with a transparent cutout over the target element
function SpotlightOverlay({ rect, onClick }: { rect: Rect; onClick?: (e: React.MouseEvent) => void }) {
    const vw = window.innerWidth
    const vh = window.innerHeight

    const borderRadius = 8
    const r = borderRadius

    // outer rect (CW) + inner rounded-rect cutout (CCW) for even-odd fill
    const outerPath = `M0,0 H${vw} V${vh} H0 Z`
    const innerPath = rect.width > 0 && rect.height > 0
        ? [
            `M${rect.left + r},${rect.top}`,
            `H${rect.left + rect.width - r}`,
            `Q${rect.left + rect.width},${rect.top} ${rect.left + rect.width},${rect.top + r}`,
            `V${rect.top + rect.height - r}`,
            `Q${rect.left + rect.width},${rect.top + rect.height} ${rect.left + rect.width - r},${rect.top + rect.height}`,
            `H${rect.left + r}`,
            `Q${rect.left},${rect.top + rect.height} ${rect.left},${rect.top + rect.height - r}`,
            `V${rect.top + r}`,
            `Q${rect.left},${rect.top} ${rect.left + r},${rect.top} Z`,
        ].join(" ")
        : ""

    return (
        <svg
            className="fixed inset-0 z-[9998] pointer-events-auto"
            width={vw}
            height={vh}
            style={{ width: vw, height: vh }}
            onClick={onClick}
        >
            <defs>
                <filter id="tour-spotlight-glow" x="-20%" y="-20%" width="140%" height="140%">
                    <feGaussianBlur stdDeviation="3" result="blur" />
                    <feMerge>
                        <feMergeNode in="blur" />
                        <feMergeNode in="SourceGraphic" />
                    </feMerge>
                </filter>
            </defs>
            <motion.path
                d={outerPath + " " + innerPath}
                fillRule="evenodd"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.3 }}
                fill="rgba(0, 0, 0, 0.7)"
            />
            {/* subtle glow border around the spotlight */}
            {rect.width > 0 && (
                <motion.rect
                    x={rect.left}
                    y={rect.top}
                    width={rect.width}
                    height={rect.height}
                    rx={borderRadius}
                    ry={borderRadius}
                    fill="none"
                    stroke="rgba(166, 135, 244, 0.4)"
                    strokeWidth={2}
                    filter="url(#tour-spotlight-glow)"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.4, delay: 0.1 }}
                />
            )}
        </svg>
    )
}

function ModalOverlay({ onClick }: { onClick?: (e: React.MouseEvent) => void }) {
    return (
        <motion.div
            className="fixed inset-0 z-[9998] bg-black/70 pointer-events-auto"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            onClick={onClick}
        />
    )
}

interface TourCardProps {
    step: TourStep
    currentIndex: number
    totalSteps: number
    style: React.CSSProperties
    onNext: () => void
    onPrev: () => void
    onStop: () => void
    isModal?: boolean
}

const TourCard = React.forwardRef<HTMLDivElement, TourCardProps>(
    ({ step, currentIndex, totalSteps, style, onNext, onPrev, onStop, isModal }, ref) => {
        const isFirstStep = currentIndex === 0
        const isLastStep = currentIndex === totalSteps - 1
        const showBack = !isFirstStep && !step.disableBack

        const content = (
            <motion.div
                ref={ref}
                className={cn(
                    "z-[9999] pointer-events-auto",
                    "rounded-xl border bg-[--paper] shadow-2xl",
                    "flex flex-col overflow-hidden",
                )}
                style={{
                    ...style,
                    ...(isModal ? {} : { position: "fixed" as const }),
                }}
                initial={{ opacity: 0, scale: 0.95, y: 8 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.95, y: 8 }}
                transition={{ duration: 0.25, ease: [0.23, 1, 0.32, 1] }}
            >
                {step.image && (
                    <div className="w-full overflow-hidden rounded-t-xl">
                        {step.image}
                    </div>
                )}

                <div className="p-4 flex flex-col gap-3">
                    {step.title && (
                        <div className="font-semibold text-base text-gray-100 leading-snug">
                            {step.title}
                        </div>
                    )}

                    <div className="text-sm text-[--muted] leading-relaxed">
                        {step.content}
                    </div>

                    {!step.hideControls && (
                        <div className="flex items-center justify-between pt-1 gap-3">
                            {totalSteps > 1 && (
                                <div className="flex items-center gap-1.5 flex-shrink-0">
                                    {Array.from({ length: totalSteps }).map((_, i) => (
                                        <span
                                            key={i}
                                            className={cn(
                                                "inline-block rounded-full transition-all duration-300",
                                                i === currentIndex
                                                    ? "w-5 h-1.5 bg-brand-400"
                                                    : i < currentIndex
                                                        ? "w-1.5 h-1.5 bg-brand-400/50"
                                                        : "w-1.5 h-1.5 bg-gray-700",
                                            )}
                                        />
                                    ))}
                                </div>
                            )}

                            <div className="flex items-center gap-2 ml-auto">
                                {!isLastStep && (
                                    <button
                                        onClick={onStop}
                                        className={cn(
                                            "text-xs text-gray-500 hover:text-gray-300 transition-colors",
                                            "px-2 py-1 rounded-md hover:bg-gray-800",
                                        )}
                                    >
                                        Skip
                                    </button>
                                )}

                                {showBack && (
                                    <button
                                        onClick={onPrev}
                                        className={cn(
                                            "text-sm font-medium text-gray-400 hover:text-gray-200 transition-colors",
                                            "px-3 py-1.5 rounded-lg hover:bg-gray-800",
                                        )}
                                    >
                                        {step.prevLabel ?? "Back"}
                                    </button>
                                )}

                                <button
                                    onClick={onNext}
                                    className={cn(
                                        "text-sm font-semibold rounded-lg px-4 py-1.5 transition-all",
                                        "bg-brand-500 hover:bg-brand-600 text-white",
                                        "shadow-sm shadow-brand-500/20",
                                    )}
                                >
                                    {step.nextLabel ?? (isLastStep ? "Done" : "Next")}
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </motion.div>
        )

        // modal mode: wrap in a fixed centering container so framer-motion's
        // transform doesn't conflict with CSS centering
        if (isModal) {
            return (
                <div className="fixed inset-0 z-[9999] flex items-center justify-center pointer-events-none">
                    {content}
                </div>
            )
        }

        return content
    },
)

TourCard.displayName = "TourCard"

function TourLoadingIndicator() {
    return (
        <motion.div
            className="fixed inset-0 z-[10000] flex items-center justify-center pointer-events-none"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
        >
            <div className="flex items-center gap-3 rounded-xl bg-gray-900 border border-[rgba(255,255,255,0.1)] px-5 py-3 shadow-xl">
                <svg
                    className="animate-spin h-4 w-4 text-brand-400"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                >
                    <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                    />
                    <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                </svg>
                <span className="text-sm text-gray-300">Loading…</span>
            </div>
        </motion.div>
    )
}

export function TourOverlay() {
    const tourState = useAtomValue(tourAtom)
    const { active, currentStep, currentIndex, totalSteps, status, next, prev, stop } = useTour()

    const cardRef = useRef<HTMLDivElement>(null)

    const [targetRect, setTargetRect] = useState<Rect>(EMPTY_RECT)
    const [cardSize, setCardSize] = useState({ width: 0, height: 0 })

    const isSpotlightStep = !!currentStep?.target
    const isModalStep = !currentStep?.target && active && status === "idle"
    const isLoading = active && status !== "idle"

    // measure the card so we can position it
    useLayoutEffect(() => {
        if (!cardRef.current) return
        const { width, height } = cardRef.current.getBoundingClientRect()
        setCardSize((prev) => {
            if (prev.width === width && prev.height === height) return prev
            return { width, height }
        })
    }, [currentStep?.id, status])

    // track the target element's rect (scroll + resize)
    useEffect(() => {
        if (!active || !currentStep?.target || status !== "idle") {
            setTargetRect(EMPTY_RECT)
            return
        }

        const el = document.querySelector(currentStep.target)
        if (!el) {
            setTargetRect(EMPTY_RECT)
            return
        }

        const padding = currentStep.spotlightPadding ?? 8

        const update = () => {
            setTargetRect(getElementRect(el, padding))
        }

        update()

        // scroll the element into view
        el.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" })

        // re-measure on scroll/resize via RAF throttle
        let rafId = 0
        const onLayout = () => {
            cancelAnimationFrame(rafId)
            rafId = requestAnimationFrame(update)
        }

        window.addEventListener("scroll", onLayout, true)
        window.addEventListener("resize", onLayout)

        const ro = new ResizeObserver(onLayout)
        ro.observe(el)

        return () => {
            cancelAnimationFrame(rafId)
            window.removeEventListener("scroll", onLayout, true)
            window.removeEventListener("resize", onLayout)
            ro.disconnect()
        }
    }, [active, currentStep?.target, currentStep?.spotlightPadding, status, currentIndex])

    // prevent background scroll while tour is active
    useEffect(() => {
        if (!active) return
        const original = document.body.style.overflow
        document.body.style.overflow = "hidden"
        return () => {
            document.body.style.overflow = original
        }
    }, [active])

    // advanceOnTargetClick
    useEffect(() => {
        if (!active || !currentStep?.target || !currentStep.advanceOnTargetClick || status !== "idle") return

        const el = document.querySelector(currentStep.target)
        if (!el) return

        const handler = () => next()
        el.addEventListener("click", handler)
        return () => el.removeEventListener("click", handler)
    }, [active, currentStep?.target, currentStep?.advanceOnTargetClick, status, next])

    // popover positioning
    const popoverWidth = typeof currentStep?.popoverWidth === "number"
        ? currentStep.popoverWidth
        : typeof currentStep?.popoverWidth === "string"
            ? parseInt(currentStep.popoverWidth, 10) || 340
            : isModalStep
                ? 440
                : 340

    const resolvedPlacement = useMemo(() => {
        if (!isSpotlightStep || targetRect === EMPTY_RECT) return "bottom" as TourStepPlacement
        return resolvePlacement(
            targetRect,
            popoverWidth,
            cardSize.height || 200,
            currentStep?.placement ?? "bottom",
        )
    }, [isSpotlightStep, targetRect, popoverWidth, cardSize.height, currentStep?.placement])

    const popoverPos = useMemo(() => {
        if (!isSpotlightStep || targetRect === EMPTY_RECT) return { top: 0, left: 0 }
        return computePopoverPosition(
            targetRect,
            popoverWidth,
            cardSize.height || 200,
            resolvedPlacement,
        )
    }, [isSpotlightStep, targetRect, popoverWidth, cardSize.height, resolvedPlacement])

    const handleOverlayClick = useCallback((e: React.MouseEvent) => {
        if (currentStep?.ignoreOutsideClick) return
        // don't close if clicking inside the spotlight hole
        if (isSpotlightStep && targetRect.width > 0) {
            const x = e.clientX
            const y = e.clientY
            if (
                x >= targetRect.left &&
                x <= targetRect.left + targetRect.width &&
                y >= targetRect.top &&
                y <= targetRect.top + targetRect.height
            ) {
                return
            }
        }
        stop()
    }, [isSpotlightStep, targetRect, stop, currentStep?.ignoreOutsideClick])

    if (!active) return null

    return (
        <AnimatePresence mode="wait">
            {active && (
                <div key="tour-container" className="tour-overlay-root">
                    <AnimatePresence>
                        {isLoading && <TourLoadingIndicator key="tour-loading" />}
                    </AnimatePresence>

                    <AnimatePresence mode="wait">
                        {status === "idle" && (
                            isSpotlightStep ? (
                                <SpotlightOverlay
                                    key={`spotlight-${currentIndex}`}
                                    rect={targetRect}
                                    onClick={handleOverlayClick}
                                />
                            ) : (
                                <ModalOverlay
                                    key={`modal-overlay-${currentIndex}`}
                                    onClick={handleOverlayClick}
                                />
                            )
                        )}
                    </AnimatePresence>

                    <AnimatePresence mode="wait">
                        {status === "idle" && currentStep && (
                            <TourCard
                                key={`card-${currentStep.id}`}
                                ref={cardRef}
                                step={currentStep}
                                currentIndex={currentIndex}
                                totalSteps={totalSteps}
                                onNext={next}
                                onPrev={prev}
                                onStop={stop}
                                isModal={isModalStep}
                                style={
                                    isModalStep
                                        ? { width: popoverWidth }
                                        : {
                                            top: popoverPos.top,
                                            left: popoverPos.left,
                                            width: popoverWidth,
                                        }
                                }
                            />
                        )}
                    </AnimatePresence>
                </div>
            )}
        </AnimatePresence>
    )
}
