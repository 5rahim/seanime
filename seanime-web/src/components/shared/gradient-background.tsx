import { motion } from "motion/react"
import React, { useEffect, useRef } from "react"

interface GradientBackgroundProps {
    /**
     * Initial size of the radial gradient, defining the starting width.
     * @default 110
     */
    startingGap?: number;

    /**
     * Enables or disables the breathing animation effect.
     * @default false
     */
    Breathing?: boolean;

    /**
     * Array of colors to use in the radial gradient.
     * Each color corresponds to a stop percentage in `gradientStops`.
     * @default ["#0A0A0A", "#2979FF", "#FF80AB", "#FF6D00", "#FFD600", "#00E676", "#3D5AFE"]
     */
    gradientColors?: string[];

    /**
     * Array of percentage stops corresponding to each color in `gradientColors`.
     * The values should range between 0 and 100.
     * @default [35, 50, 60, 70, 80, 90, 100]
     */
    gradientStops?: number[];

    /**
     * Speed of the breathing animation.
     * Lower values result in slower animation.
     * @default 0.02
     */
    animationSpeed?: number;

    /**
     * Maximum range for the breathing animation in percentage points.
     * Determines how much the gradient "breathes" by expanding and contracting.
     * @default 5
     */
    breathingRange?: number;

    /**
     * Additional class names for the gradient container.
     * @default ""
     */
    containerClassName?: string;

    /**
     * Additional top offset for the gradient container form the top to have a more flexible control over the gradient.
     * @default 0
     */
    topOffset?: number;

    duration?: number;
}

export const GradientBackground: React.FC<GradientBackgroundProps> = ({
    startingGap = 125,
    Breathing = true,
    gradientColors = [
        "transparent",
        "#312887",
        "#3D5AFE",
        "#FF80AB",
        "#FF6D00",
        "#FFD600",
        "#00E676",
    ],
    gradientStops = [35, 50, 60, 70, 80, 90, 100],
    animationSpeed = 0.02,
    breathingRange = 5,
    topOffset = 0,
    containerClassName = "",
    duration = 2,
}) => {

    if (gradientColors.length !== gradientStops.length) {
        throw new Error(
            `GradientColors and GradientStops must have the same length.
     Received gradientColors length: ${gradientColors.length},
     gradientStops length: ${gradientStops.length}`,
        )
    }

    const containerRef = useRef<HTMLDivElement | null>(null)

    useEffect(() => {
        let animationFrame: number
        let width = startingGap
        let directionWidth = 1

        const animateGradient = () => {
            if (width >= startingGap + breathingRange) directionWidth = -1
            if (width <= startingGap - breathingRange) directionWidth = 1

            if (!Breathing) directionWidth = 0
            width += directionWidth * animationSpeed

            const gradientStopsString = gradientStops
                .map((stop, index) => `${gradientColors[index]} ${stop}%`)
                .join(", ")

            const gradient = `radial-gradient(${width}% ${width + topOffset}% at 50% 20%, ${gradientStopsString})`

            if (containerRef.current) {
                containerRef.current.style.background = gradient
            }

            animationFrame = requestAnimationFrame(animateGradient)
        }

        animationFrame = requestAnimationFrame(animateGradient)

        return () => cancelAnimationFrame(animationFrame) // Cleanup animation
    }, [startingGap, Breathing, gradientColors, gradientStops, animationSpeed, breathingRange, topOffset])

    return (
        <motion.div
            key="animated-gradient-background"
            initial={{
                opacity: 0,
                scale: 1.5,
            }}
            animate={{
                opacity: 0.2,
                scale: 1,
                transition: {
                    duration: duration,
                    ease: [0.25, 0.1, 0.25, 1], // Cubic bezier easing
                },
            }}
            className={`absolute inset-0 z-[0] overflow-hidden ${containerClassName}`}
        >
            <div
                ref={containerRef}
                className="absolute inset-0 transition-transform"
            />
        </motion.div>
    )
}
