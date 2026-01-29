import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import React, { useRef, useState } from "react"

type SettingsCardProps = {
    title?: string
    description?: string
    children: React.ReactNode
}

export function SettingsNavCard({ title, children }: SettingsCardProps) {
    const [position, setPosition] = useState({ x: 0, y: 0 })
    const cardRef = useRef<HTMLDivElement>(null)

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        if (!cardRef.current) return
        const rect = cardRef.current.getBoundingClientRect()
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top
        setPosition({ x, y })
    }

    return (
        <div className="pb-4">
            <div
                ref={cardRef}
                onMouseMove={handleMouseMove}
                className="p-0"
                // className=" contents lg:block relative group/settings-nav overflow-hidden"
            >
                {/* <div
                    className="pointer-events-none absolute -inset-px transition-opacity duration-300 opacity-0 group-hover/settings-nav:opacity-100 hidden lg:block"
                    style={{
                        background: `radial-gradient(250px circle at ${position.x}px ${position.y}px, rgb(255 255 255 / 0.025), transparent 40%)`,
                    }}
                 /> */}
                {children}
            </div>
        </div>
    )
}

export function SettingsCard({ title, description, children, className }: SettingsCardProps & { className?: string }) {
    const [position, setPosition] = useState({ x: 0, y: 0 })
    const [isHovered, setIsHovered] = useState(false)
    const cardRef = useRef<HTMLDivElement>(null)

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        if (!cardRef.current) return
        const rect = cardRef.current.getBoundingClientRect()
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top
        setPosition({ x, y })
    }

    return (
        <>
            <Card
                ref={cardRef}
                className={cn("group/settings-card relative lg:bg-gray-950/80 rounded-xl", className)}
                onMouseMove={handleMouseMove}
            >
                {/* <div
                    className="pointer-events-none absolute -inset-px transition-opacity duration-300 opacity-0 group-hover/settings-card:opacity-100"
                    style={{
                        background: `radial-gradient(700px circle at ${position.x}px ${position.y}px, rgb(255 255 255 / 0.025), transparent 40%)`,
                    }}
                 /> */}
                {title && <CardHeader className="p-0 pb-2 flex flex-col lg:flex-row items-center gap-0 mx-3 mt-3 space-y-0">
                    {/* <CardTitle className="font-semibold tracking-wide text-base transition-colors duration-300 group-hover/settings-card:text-white bg-gradient-to-br group-hover/settings-card:from-brand-500/10 group-hover/settings-card:to-purple-500/5 px-4 py-2 bg-[--subtle] w-fit rounded-tl-md rounded-br-md ">
                     {title}
                     </CardTitle> */}
                    <CardTitle
                        className={cn(
                            "font-semibold text-[1rem] tracking-wide transition-colors duration-300 px-4 py-1 border w-fit rounded-xl bg-gray-800/40",
                            "group-hover/settings-card:bg-brand-500/10 group-hover/settings-card:text-white flex-none",
                        )}
                    >
                        {title}
                    </CardTitle>
                    {description && <CardDescription className="px-4 py-2 lg:py-0 w-fit">
                        {description}
                    </CardDescription>}
                </CardHeader>}
                <CardContent
                    className={cn(
                        !title && "pt-4",
                        "space-y-3 flex-wrap",
                    )}
                >
                    {children}
                </CardContent>
            </Card>
        </>
    )
}

export function SettingsPageHeader({ title, description, icon: Icon }: { title: string, description: string, icon: React.ElementType }) {
    return (
        <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-gradient-to-br from-brand-500/10 to-purple-500/10 border border-brand-500/15">
                <Icon className="text-2xl text-brand-600 dark:text-brand-400" />
            </div>
            <div>
                <h3 className="text-xl font-semibold">{title}</h3>
                <p className="text-base text-[--muted]">{description}</p>
            </div>
        </div>
    )
}
