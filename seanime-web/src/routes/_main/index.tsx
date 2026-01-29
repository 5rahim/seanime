import { HomeScreen } from "@/app/(main)/(library)/_home/home-screen"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/")({
    component: HomeScreen,
})
