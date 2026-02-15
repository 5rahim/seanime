import Page from "@/app/(main)/manga/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/manga/")({
    component: Page,
})
