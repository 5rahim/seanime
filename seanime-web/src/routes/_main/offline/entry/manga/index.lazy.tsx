import Page from "@/app/(main)/(offline)/offline/entry/manga/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/offline/entry/manga/")({
    component: Page,
})
