import Page from "@/app/(main)/_features/offline/page"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/offline/")({
    component: Page,
})
