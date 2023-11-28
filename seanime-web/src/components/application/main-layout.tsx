"use client"
import { AppLayout, AppSidebarProvider } from "@/components/ui/app-layout"
import React from "react"
import dynamic from "next/dynamic"
import { MainSidebar } from "@/components/application/main-sidebar"

const GlobalSearch = dynamic(() => import("@/components/application/global-search").then((mod) => mod.GlobalSearch))

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    return (
        <>
            <AppSidebarProvider>
                <AppLayout withSidebar sidebarSize="slim">
                    <AppLayout.Sidebar>
                        <MainSidebar/>
                    </AppLayout.Sidebar>
                    <AppLayout>
                        <AppLayout.Content>
                            {children}
                        </AppLayout.Content>
                    </AppLayout>
                </AppLayout>
            </AppSidebarProvider>
            <GlobalSearch/>
        </>
    )
}