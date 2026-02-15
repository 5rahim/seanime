import Page from "@/app/(main)/schedule/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/schedule/")({
    component: Page,
})
