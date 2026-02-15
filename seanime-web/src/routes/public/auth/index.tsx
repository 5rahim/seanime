import Page from "@/app/public/auth/page"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/public/auth/")({
    component: Page,
})
