import Page from "@/app/(main)/(offline)/offline/entry/anime/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/offline/entry/anime/")({
    component: Page,
})
