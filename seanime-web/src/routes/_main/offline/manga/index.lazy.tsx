import Page from "@/app/(main)/_features/offline/manga/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/offline/manga/")({
    component: Page,
})
