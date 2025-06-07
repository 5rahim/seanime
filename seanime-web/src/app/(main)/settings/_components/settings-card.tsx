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
                className="lg:p-2 lg:border lg:rounded-[--radius] lg:bg-[--paper] contents lg:block relative group/settings-nav overflow-hidden"
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
                className={cn("group/settings-card relative overflow-hidden bg-[--paper]", className)}
                onMouseMove={handleMouseMove}
            >
                {/* <div
                    className="pointer-events-none absolute -inset-px transition-opacity duration-300 opacity-0 group-hover/settings-card:opacity-100"
                    style={{
                        background: `radial-gradient(700px circle at ${position.x}px ${position.y}px, rgb(255 255 255 / 0.025), transparent 40%)`,
                    }}
                 /> */}
                {title && <CardHeader className="p-0 pb-4">
                    <CardTitle className="font-semibold tracking-wide text-base transition-colors group-hover/settings-card:text-white px-4 py-2 bg-[--subtle] w-fit rounded-br-md">
                        {title}
                    </CardTitle>
                    {description && <CardDescription className="px-4">
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
