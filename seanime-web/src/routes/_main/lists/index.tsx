import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/lists/page"

export const Route = createFileRoute("/_main/lists/")({
    component: Page,
})
