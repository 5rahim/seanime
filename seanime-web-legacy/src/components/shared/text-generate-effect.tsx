import { cn } from "@/components/ui/core/styling"
import { motion, stagger, useAnimate } from "motion/react"
import React, { useEffect } from "react"

export const TextGenerateEffect = ({
    words,
    className,
    style,
    ...rest
}: {
    words: string;
    className?: string;
    style?: any
} & React.HTMLAttributes<HTMLDivElement>) => {
    const [scope, animate] = useAnimate()
    let wordsArray = words.split(" ")

    useEffect(() => {
        animate(
            "span",
            {
                opacity: 1,
            },
            {
                duration: 2,
                delay: stagger(0.2),
            },
        )
    }, [words])

    const renderWords = () => {
        return (
            <motion.div ref={scope}>
                {wordsArray.map((word, idx) => {
                    return (
                        <motion.span
                            key={word + idx}
                            className="opacity-0"
                        >
                            {word}{" "}
                        </motion.span>
                    )
                })}
            </motion.div>
        )
    }

    return (
        <div className={cn("font-bold", className)} style={style} {...rest}>
            {renderWords()}
        </div>
    )
}
