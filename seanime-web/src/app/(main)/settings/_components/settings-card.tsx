import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import React from "react"

type SettingsCardProps = {
    title?: string
    description?: string
    children: React.ReactNode
}

export function SettingsNavCard({ title, children }: SettingsCardProps) {
    return (
        <div className="pb-4">
            <div className="lg:p-2 lg:border lg:rounded-md lg:bg-[--paper] contents lg:block">
                {children}
            </div>
        </div>
    )
}

export function SettingsCard({ title, description, children }: SettingsCardProps) {
    return (
        <>
            <Card className="group/settings-card">
                {title && <CardHeader>
                    <CardTitle className="font-semibold text-xl text-[--muted] transition-colors group-hover/settings-card:text-[--foreground]">
                        {title}
                    </CardTitle>
                    {description && <CardDescription>
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
