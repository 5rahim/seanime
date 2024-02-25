"use client"
import { MainSidebar } from "@/components/application/main-sidebar"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import dynamic from "next/dynamic"
import React from "react"

const GlobalSearch = dynamic(() => import("@/components/application/global-search").then((mod) => mod.GlobalSearch))

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    return (
        <>
            <AppSidebarProvider>
                <AppLayout withSidebar sidebarSize="slim">
                    <AppLayoutSidebar>
                        <MainSidebar/>
                    </AppLayoutSidebar>
                    <AppLayout>
                        <AppLayoutContent>
                            {children}
                        </AppLayoutContent>
                    </AppLayout>
                </AppLayout>
            </AppSidebarProvider>
            <GlobalSearch/>
        </>
    )
}
