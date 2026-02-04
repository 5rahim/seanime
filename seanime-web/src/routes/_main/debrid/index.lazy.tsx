import Page from "@/app/(main)/debrid/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/debrid/")({
    component: Page,
})
