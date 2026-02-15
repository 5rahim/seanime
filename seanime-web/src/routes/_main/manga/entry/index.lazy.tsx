import Page from "@/app/(main)/manga/entry/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/manga/entry/")({
    component: Page,
})
