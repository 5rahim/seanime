import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/(offline)/offline/page"

export const Route = createFileRoute("/_main/offline/")({
    component: Page,
})
