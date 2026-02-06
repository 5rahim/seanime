import Page from "@/app/(main)/_features/offline/entry/manga/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/offline/entry/manga/")({
    component: Page,
})
