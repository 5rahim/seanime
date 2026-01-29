import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/schedule/page"

export const Route = createFileRoute("/_main/schedule/")({
    component: Page,
})
