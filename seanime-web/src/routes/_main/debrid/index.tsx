import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/debrid/page"

export const Route = createFileRoute("/_main/debrid/")({
    component: Page,
})
