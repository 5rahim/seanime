"use client"
import { cn } from "@/components/ui/core/styling"
import { getAssetUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "motion/react"
import React from "react"

type CustomBackgroundImageProps = React.ComponentPropsWithoutRef<"div"> & {}

export function CustomBackgroundImage(props: CustomBackgroundImageProps) {

    const {
        className,
        ...rest
    } = props

    const ts = useThemeSettings()

    return (
        <>
            {!!ts.libraryScreenCustomBackgroundImage && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.1 }}
                    className="fixed w-full h-full inset-0"
                >

                    {ts.libraryScreenCustomBackgroundBlur !== "" && <div
                        className="fixed w-full h-full inset-0 z-[0]"
                        style={{ backdropFilter: `blur(${ts.libraryScreenCustomBackgroundBlur})` }}
                    >
                    </div>}

                    <div
                        className={cn(
                            "fixed w-full h-full inset-0 z-[-1] bg-no-repeat bg-cover bg-center transition-opacity duration-1000",
                            className,
                        )}
                        style={{
                            backgroundImage: `url(${getAssetUrl(ts.libraryScreenCustomBackgroundImage)})`,
                            opacity: ts.libraryScreenCustomBackgroundOpacity / 100,
                        }}
                        {...rest}
                    />
                </motion.div>
            )}
        </>
    )
}
