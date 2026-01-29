import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/(offline)/offline/manga/page"

export const Route = createFileRoute("/_main/offline/manga/")({
    component: Page,
})
