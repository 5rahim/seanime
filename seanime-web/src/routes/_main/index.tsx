import { HomeScreen } from "@/app/(main)/(library)/_home/home-screen"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display.tsx"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/")({
    component: HomeScreen,
    pendingComponent: MediaEntryPageLoadingDisplay,
    pendingMs: 250,
})
