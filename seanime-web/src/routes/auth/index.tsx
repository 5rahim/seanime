import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/public/auth/page"

export const Route = createFileRoute("/auth/")({
    component: Page,
})
