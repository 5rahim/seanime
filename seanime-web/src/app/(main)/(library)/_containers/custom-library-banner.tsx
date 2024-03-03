"use client"
import { cn } from "@/components/ui/core/styling"
import { getAssetUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "framer-motion"
import Image from "next/image"
import React, { useEffect, useState } from "react"
import { useWindowScroll } from "react-use"


export function CustomLibraryBanner() {

    const ts = useThemeSettings()
    const image = React.useMemo(() => getAssetUrl(ts.libraryScreenCustomBanner), [ts.libraryScreenCustomBanner])
    const [dimmed, setDimmed] = useState(false)

    const { y } = useWindowScroll()

    useEffect(() => {
        if (y > 100)
            setDimmed(true)
        else
            setDimmed(false)
    }, [(y > 100)])

    return (
        <>
            <div className="py-10"></div>
            <div className="__header h-[20rem] z-[-1] top-0 w-full fixed group/library-header">
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.2 }}
                    className={cn(
                        "h-[30rem] z-[0] w-full flex-none absolute top-0 overflow-hidden",
                    )}
                >
                    <div
                        className="w-full absolute z-[2] top-0 h-[10rem] opacity-80 bg-gradient-to-b from-[--background] to-transparent via"
                    />
                    <div
                        className={cn(
                            "z-[1] absolute inset-0 w-full h-full bg-cover bg-no-repeat transition-opacity duration-1000 opacity-100",
                            dimmed && "opacity-10",
                        )}
                        style={{
                            backgroundImage: `url(${image})`,
                            backgroundPosition: ts.libraryScreenBannerPosition || "50% 50%",
                            backgroundRepeat: "no-repeat",
                            backgroundSize: "cover",
                        }}
                    />

                    {/*{(!!image) && <Image*/}
                    {/*    src={image}*/}
                    {/*    alt="banner image"*/}
                    {/*    fill*/}
                    {/*    quality={100}*/}
                    {/*    priority*/}
                    {/*    sizes="100vw"*/}
                    {/*    className={cn(*/}
                    {/*        "object-cover object-center z-[1] opacity-100 transition-all duration-1000",*/}
                    {/*        dimmed && "opacity-10",*/}
                    {/*    )}*/}
                    {/*/>}*/}
                    <div
                        className="w-full z-[2] absolute bottom-0 h-[25rem] bg-gradient-to-t from-[--background] via-opacity-50 via-10% to-transparent"
                    />
                    <div className="h-full absolute z-[2] w-full xl-right-48">
                        <Image
                            src={"/mask-2.png"}
                            alt="mask"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            className={cn(
                                "object-cover object-left z-[2] transition-opacity duration-1000 opacity-30",
                            )}
                        />
                    </div>
                    <div className="h-full absolute z-[2] w-full xl:-right-48">
                        <Image
                            src={"/mask.png"}
                            alt="mask"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            className={cn(
                                "object-cover object-right z-[2] transition-opacity duration-1000 opacity-20",
                            )}
                        />
                    </div>
                </motion.div>
            </div>
        </>
    )

}
