import Page from "@/app/(main)/(offline)/offline/manga/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/offline/manga/")({
    component: Page,
})
